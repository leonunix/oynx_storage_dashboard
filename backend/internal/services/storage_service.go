package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/config"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/system"
)

var ErrMutationsDisabled = fmt.Errorf("destructive storage operations are disabled (set ONYX_DASHBOARD_ALLOW_DM_MUTATIONS=true)")

type StorageService struct {
	cfg            config.OperationsConfig
	runner         *system.Runner
	storageTimeout time.Duration
}

func NewStorageService(cfg config.OperationsConfig, runner *system.Runner, storageTimeout time.Duration) *StorageService {
	return &StorageService{cfg: cfg, runner: runner, storageTimeout: storageTimeout}
}

func (s *StorageService) requireMutations() error {
	if !s.cfg.AllowDestructiveDM {
		return ErrMutationsDisabled
	}
	return nil
}

// ── Layout (read-only topology) ───────────────────────────────────

func (s *StorageService) Layout(ctx context.Context) (domain.StorageLayout, error) {
	layout := domain.StorageLayout{
		AllowMutations: s.cfg.AllowDestructiveDM,
	}

	// Block devices
	if output, err := s.runner.Run(ctx, "lsblk", "--json", "-o", "NAME,TYPE,SIZE,STATE,PKNAME"); err == nil {
		var payload struct {
			BlockDevices []struct {
				Name   string `json:"name"`
				Type   string `json:"type"`
				Size   string `json:"size"`
				State  string `json:"state"`
				Parent string `json:"pkname"`
			} `json:"blockdevices"`
		}
		if json.Unmarshal([]byte(output), &payload) == nil {
			for _, device := range payload.BlockDevices {
				layout.BlockDevices = append(layout.BlockDevices, domain.DeviceSummary{
					Name:   device.Name,
					Type:   device.Type,
					Size:   device.Size,
					State:  device.State,
					Parent: device.Parent,
				})
			}
		} else {
			layout.Warnings = append(layout.Warnings, "failed to parse lsblk output")
		}
	} else {
		layout.Warnings = append(layout.Warnings, err.Error())
	}

	// dm targets
	if output, err := s.runner.Run(ctx, "dmsetup", "ls", "--tree"); err == nil {
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			layout.DMTargets = append(layout.DMTargets, domain.DMTarget{
				Name:   line,
				State:  "observed",
				Target: "dm",
			})
		}
	} else {
		layout.Warnings = append(layout.Warnings, err.Error())
	}

	// LVM logical volumes
	if output, err := s.runner.Run(ctx, "lvs", "--reportformat", "json", "-o", "lv_name,vg_name,lv_attr,lv_size"); err == nil {
		var payload struct {
			Report []struct {
				LV []struct {
					Name   string `json:"lv_name"`
					VGName string `json:"vg_name"`
					Attr   string `json:"lv_attr"`
					Size   string `json:"lv_size"`
				} `json:"lv"`
			} `json:"report"`
		}
		if json.Unmarshal([]byte(output), &payload) == nil {
			for _, report := range payload.Report {
				for _, lv := range report.LV {
					layout.LogicalVolumes = append(layout.LogicalVolumes, domain.LvmVolume{
						Name:   lv.Name,
						VGName: lv.VGName,
						Attr:   lv.Attr,
						Size:   lv.Size,
					})
				}
			}
		} else {
			layout.Warnings = append(layout.Warnings, "failed to parse lvs output")
		}
	} else {
		layout.Warnings = append(layout.Warnings, err.Error())
	}

	// RAID arrays
	layout.RaidArrays = s.listRaidArrays(ctx)

	// Physical volumes
	layout.PhysicalVolumes = s.listPhysicalVolumes(ctx, &layout.Warnings)

	// Volume groups
	layout.VolumeGroups = s.listVolumeGroups(ctx, &layout.Warnings)

	return layout, nil
}

func (s *StorageService) listRaidArrays(ctx context.Context) []domain.RaidArray {
	output, err := s.runner.Run(ctx, "mdadm", "--detail", "--scan")
	if err != nil {
		return nil
	}

	var arrays []domain.RaidArray
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "ARRAY") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := fields[1]
		arr, err := s.raidDetail(ctx, name)
		if err != nil {
			arrays = append(arrays, domain.RaidArray{Name: name, State: "unknown"})
			continue
		}
		arrays = append(arrays, arr)
	}
	return arrays
}

var raidDetailKV = regexp.MustCompile(`^\s*(.+?)\s*:\s*(.+?)\s*$`)

func (s *StorageService) raidDetail(ctx context.Context, devPath string) (domain.RaidArray, error) {
	output, err := s.runner.Run(ctx, "mdadm", "--detail", devPath)
	if err != nil {
		return domain.RaidArray{}, err
	}

	arr := domain.RaidArray{Name: devPath}
	var devices []string

	for _, line := range strings.Split(output, "\n") {
		m := raidDetailKV.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		key, val := strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
		switch key {
		case "Raid Level":
			arr.Level = val
		case "Array Size":
			arr.Size = val
		case "State":
			arr.State = strings.ToLower(val)
		case "Active Devices":
			arr.ActiveDevs, _ = strconv.Atoi(val)
		case "Total Devices":
			arr.TotalDevs, _ = strconv.Atoi(val)
		case "UUID":
			arr.UUID = val
		}
		// Device lines: "N M N /dev/sdX"
		if strings.Contains(line, "/dev/") && !strings.HasPrefix(strings.TrimSpace(line), "Raid") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "/dev/") && p != devPath {
					devices = append(devices, p)
				}
			}
		}
	}
	arr.Devices = devices
	return arr, nil
}

func (s *StorageService) listPhysicalVolumes(ctx context.Context, warnings *[]string) []domain.PhysicalVolume {
	output, err := s.runner.Run(ctx, "pvs", "--reportformat", "json", "-o", "pv_name,vg_name,pv_size,pv_free,pv_attr,pv_uuid")
	if err != nil {
		*warnings = append(*warnings, err.Error())
		return nil
	}

	var payload struct {
		Report []struct {
			PV []struct {
				Name   string `json:"pv_name"`
				VGName string `json:"vg_name"`
				Size   string `json:"pv_size"`
				Free   string `json:"pv_free"`
				Attr   string `json:"pv_attr"`
				UUID   string `json:"pv_uuid"`
			} `json:"pv"`
		} `json:"report"`
	}
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		*warnings = append(*warnings, "failed to parse pvs output")
		return nil
	}

	var pvs []domain.PhysicalVolume
	for _, report := range payload.Report {
		for _, pv := range report.PV {
			pvs = append(pvs, domain.PhysicalVolume{
				Name:   strings.TrimSpace(pv.Name),
				VGName: strings.TrimSpace(pv.VGName),
				Size:   strings.TrimSpace(pv.Size),
				Free:   strings.TrimSpace(pv.Free),
				Attr:   strings.TrimSpace(pv.Attr),
				UUID:   strings.TrimSpace(pv.UUID),
			})
		}
	}
	return pvs
}

func (s *StorageService) listVolumeGroups(ctx context.Context, warnings *[]string) []domain.VolumeGroup {
	output, err := s.runner.Run(ctx, "vgs", "--reportformat", "json", "-o", "vg_name,vg_size,vg_free,pv_count,lv_count,vg_attr,vg_uuid")
	if err != nil {
		*warnings = append(*warnings, err.Error())
		return nil
	}

	var payload struct {
		Report []struct {
			VG []struct {
				Name    string `json:"vg_name"`
				Size    string `json:"vg_size"`
				Free    string `json:"vg_free"`
				PVCount string `json:"pv_count"`
				LVCount string `json:"lv_count"`
				Attr    string `json:"vg_attr"`
				UUID    string `json:"vg_uuid"`
			} `json:"vg"`
		} `json:"report"`
	}
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		*warnings = append(*warnings, "failed to parse vgs output")
		return nil
	}

	var vgs []domain.VolumeGroup
	for _, report := range payload.Report {
		for _, vg := range report.VG {
			pvCount, _ := strconv.Atoi(strings.TrimSpace(vg.PVCount))
			lvCount, _ := strconv.Atoi(strings.TrimSpace(vg.LVCount))
			vgs = append(vgs, domain.VolumeGroup{
				Name:    strings.TrimSpace(vg.Name),
				Size:    strings.TrimSpace(vg.Size),
				Free:    strings.TrimSpace(vg.Free),
				PVCount: pvCount,
				LVCount: lvCount,
				Attr:    strings.TrimSpace(vg.Attr),
				UUID:    strings.TrimSpace(vg.UUID),
			})
		}
	}
	return vgs
}

// ── Provision planning ────────────────────────────────────────────

func (s *StorageService) PlanProvision(_ context.Context, req domain.ProvisionRequest) domain.ProvisionPlan {
	plan := domain.ProvisionPlan{
		Name: req.Name,
		SafetyChecks: []string{
			"验证所选设备未挂载且未被现有 VG/LV 占用",
			"确认 mdadm/dm-raid 参数与 data_disks、strip_size 一致",
			"确认 Onyx buffer/data device 映射使用不同 LV，避免元数据与数据混用",
			"在执行前要求二次确认并记录审计日志",
		},
		Commands: []string{
			fmt.Sprintf("pvcreate %s", strings.Join(req.Devices, " ")),
			fmt.Sprintf("vgcreate %s %s", req.VGName, strings.Join(req.Devices, " ")),
			fmt.Sprintf("lvcreate -n %s -L <size> %s", req.MetaLVName, req.VGName),
			fmt.Sprintf("lvcreate -n %s -l 100%%FREE %s", req.DataLVName, req.VGName),
			fmt.Sprintf("onyx-storage -c config/default.toml create-volume -n %s -s <bytes> --compression lz4", req.Name),
		},
		Warnings: []string{
			"当前 dashboard 默认只生成计划，不直接执行破坏性 dm/LVM 命令",
			"建议把真正的执行动作放到受控 agent 或 root-only sidecar 中",
		},
		ExecutionReady: s.cfg.AllowDestructiveDM,
	}

	if !s.cfg.AllowDestructiveDM {
		plan.Warnings = append(plan.Warnings, "ONYX_DASHBOARD_ALLOW_DM_MUTATIONS=false，当前环境禁止直接落盘执行")
	}

	return plan
}

// ── RAID management ───────────────────────────────────────────────

func (s *StorageService) RaidDetail(ctx context.Context, name string) (domain.RaidArray, error) {
	devPath := name
	if !strings.HasPrefix(devPath, "/dev/") {
		devPath = "/dev/" + devPath
	}
	return s.raidDetail(ctx, devPath)
}

var validRaidLevels = map[string]int{
	"raid0":  2,
	"raid1":  2,
	"raid5":  3,
	"raid6":  4,
	"raid10": 4,
}

func (s *StorageService) RaidCreate(ctx context.Context, req domain.RaidCreateRequest) error {
	if err := s.requireMutations(); err != nil {
		return err
	}

	minDevs, ok := validRaidLevels[req.Level]
	if !ok {
		return fmt.Errorf("unsupported RAID level: %s (supported: raid0, raid1, raid5, raid6, raid10)", req.Level)
	}
	if len(req.Devices) < minDevs {
		return fmt.Errorf("%s requires at least %d devices, got %d", req.Level, minDevs, len(req.Devices))
	}
	if req.Name == "" {
		return fmt.Errorf("RAID device name is required")
	}

	args := []string{
		"--create", req.Name,
		"--level=" + req.Level,
		"--raid-devices=" + strconv.Itoa(len(req.Devices)),
	}
	if req.ChunkKB > 0 {
		args = append(args, "--chunk="+strconv.Itoa(req.ChunkKB))
	}
	if req.Force {
		args = append(args, "--run", "--force")
	}
	args = append(args, req.Devices...)

	_, err := s.runner.RunWithTimeout(ctx, s.storageTimeout, "mdadm", args...)
	return err
}

func (s *StorageService) RaidStop(ctx context.Context, req domain.RaidStopRequest) error {
	if err := s.requireMutations(); err != nil {
		return err
	}
	devPath := req.Name
	if !strings.HasPrefix(devPath, "/dev/") {
		devPath = "/dev/" + devPath
	}
	_, err := s.runner.Run(ctx, "mdadm", "--stop", devPath)
	return err
}

// ── LVM PV management ────────────────────────────────────────────

func (s *StorageService) PVCreate(ctx context.Context, req domain.PVCreateRequest) error {
	if err := s.requireMutations(); err != nil {
		return err
	}
	if req.Device == "" {
		return fmt.Errorf("device path is required")
	}
	args := []string{req.Device}
	if req.Force {
		args = append([]string{"-f"}, args...)
	}
	_, err := s.runner.Run(ctx, "pvcreate", args...)
	return err
}

func (s *StorageService) PVRemove(ctx context.Context, req domain.PVRemoveRequest) error {
	if err := s.requireMutations(); err != nil {
		return err
	}
	if req.Device == "" {
		return fmt.Errorf("device path is required")
	}
	args := []string{req.Device}
	if req.Force {
		args = append([]string{"-f"}, args...)
	}
	_, err := s.runner.Run(ctx, "pvremove", args...)
	return err
}

// ── LVM VG management ────────────────────────────────────────────

func (s *StorageService) VGCreate(ctx context.Context, req domain.VGCreateRequest) error {
	if err := s.requireMutations(); err != nil {
		return err
	}
	if req.Name == "" {
		return fmt.Errorf("volume group name is required")
	}
	if len(req.Devices) == 0 {
		return fmt.Errorf("at least one device is required")
	}
	args := append([]string{req.Name}, req.Devices...)
	_, err := s.runner.Run(ctx, "vgcreate", args...)
	return err
}

func (s *StorageService) VGRemove(ctx context.Context, name string, force bool) error {
	if err := s.requireMutations(); err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("volume group name is required")
	}
	args := []string{name}
	if force {
		args = append([]string{"-f"}, args...)
	}
	_, err := s.runner.Run(ctx, "vgremove", args...)
	return err
}

// ── LVM LV management ────────────────────────────────────────────

func (s *StorageService) LVCreate(ctx context.Context, req domain.LVCreateRequest) error {
	if err := s.requireMutations(); err != nil {
		return err
	}
	if req.Name == "" || req.VGName == "" || req.Size == "" {
		return fmt.Errorf("name, vgName, and size are all required")
	}

	sizeFlag := "-L"
	if strings.Contains(req.Size, "%") {
		sizeFlag = "-l"
	}

	_, err := s.runner.Run(ctx, "lvcreate", "-n", req.Name, sizeFlag, req.Size, req.VGName)
	return err
}

func (s *StorageService) LVRemove(ctx context.Context, req domain.LVRemoveRequest) error {
	if err := s.requireMutations(); err != nil {
		return err
	}
	if req.Name == "" || req.VGName == "" {
		return fmt.Errorf("name and vgName are required")
	}
	lvPath := req.VGName + "/" + req.Name
	_, err := s.runner.Run(ctx, "lvremove", "-f", lvPath)
	return err
}

func (s *StorageService) LVResize(ctx context.Context, req domain.LVResizeRequest) error {
	if err := s.requireMutations(); err != nil {
		return err
	}
	if req.Name == "" || req.VGName == "" || req.Size == "" {
		return fmt.Errorf("name, vgName, and size are required")
	}

	lvPath := req.VGName + "/" + req.Name
	cmd := "lvresize"
	if strings.HasPrefix(req.Size, "+") {
		cmd = "lvextend"
	}
	_, err := s.runner.RunWithTimeout(ctx, s.storageTimeout, cmd, "-L", req.Size, lvPath)
	return err
}

// ── Provision execution ───────────────────────────────────────────

func (s *StorageService) ExecuteProvision(ctx context.Context, req domain.ProvisionExecuteRequest) (domain.ProvisionExecuteResult, error) {
	if err := s.requireMutations(); err != nil {
		return domain.ProvisionExecuteResult{}, err
	}
	if len(req.Commands) == 0 {
		return domain.ProvisionExecuteResult{Success: true}, nil
	}

	var results []domain.CommandResult
	for _, cmdLine := range req.Commands {
		parts := strings.Fields(cmdLine)
		if len(parts) == 0 {
			continue
		}
		name := parts[0]
		args := parts[1:]

		stdout, err := s.runner.RunWithTimeout(ctx, s.storageTimeout, name, args...)
		cr := domain.CommandResult{
			Command: cmdLine,
			Stdout:  stdout,
		}
		if err != nil {
			cr.Error = err.Error()
			if exitErr, ok := err.(*exec.ExitError); ok {
				cr.ExitCode = exitErr.ExitCode()
			} else {
				cr.ExitCode = -1
			}
			results = append(results, cr)
			return domain.ProvisionExecuteResult{Success: false, Results: results}, nil
		}
		results = append(results, cr)
	}

	return domain.ProvisionExecuteResult{Success: true, Results: results}, nil
}
