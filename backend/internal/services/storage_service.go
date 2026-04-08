package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/config"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/system"
)

type StorageService struct {
	cfg    config.OperationsConfig
	runner *system.Runner
}

func NewStorageService(cfg config.OperationsConfig, runner *system.Runner) *StorageService {
	return &StorageService{cfg: cfg, runner: runner}
}

func (s *StorageService) Layout(ctx context.Context) (domain.StorageLayout, error) {
	layout := domain.StorageLayout{}

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

	return layout, nil
}

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
