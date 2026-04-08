package services

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/config"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/system"
)

type OnyxService struct {
	cfg    config.OnyxConfig
	runner *system.Runner
}

func NewOnyxService(cfg config.OnyxConfig, runner *system.Runner) *OnyxService {
	return &OnyxService{cfg: cfg, runner: runner}
}

func (s *OnyxService) Overview(ctx context.Context) (domain.Overview, error) {
	raw, err := s.statusRaw(ctx)
	if err != nil {
		return domain.Overview{}, err
	}

	overview := domain.Overview{
		EngineMode:    "unknown",
		EngineRunning: true,
		Metrics:       map[string]any{},
		RawStatus:     raw,
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "mode":
			overview.EngineMode = value
		case "volumes":
			overview.VolumeCount = atoi(value)
		case "live_handles":
			overview.LiveHandleCount = atoi(value)
		case "zones":
			overview.ZoneCount = atoi(value)
		case "buffer_fill_pct":
			overview.BufferFillPercent = atoi(value)
		case "buffer_pending_entries":
			overview.BufferPendingEntries = uint64(atoi(value))
		case "allocator_free_blocks":
			free, total := parseAllocator(value)
			overview.AllocatorFreeBlocks = free
			overview.AllocatorTotalBlocks = total
		default:
			if strings.Contains(value, "=") {
				overview.Metrics[key] = parseMetricsLine(value)
			} else {
				overview.Metrics[key] = value
			}
		}
	}

	return overview, nil
}

func (s *OnyxService) ListVolumes(ctx context.Context) ([]domain.Volume, error) {
	lines, err := s.sendSocketCommand("list-volumes")
	if err == nil {
		volumes := make([]domain.Volume, 0, len(lines))
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}
			volumes = append(volumes, domain.Volume{
				Name:        fields[0],
				SizeBytes:   parseUint(fields[1]),
				ZoneCount:   uint32(parseUint(fields[2])),
				Compression: fields[3],
				Status:      "online",
			})
		}
		return volumes, nil
	}

	output, err := s.runner.Run(ctx, s.cfg.BinaryPath, "-c", s.cfg.ConfigPath, "list-volumes")
	if err != nil {
		return nil, err
	}

	volumes := make([]domain.Volume, 0)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "No volumes" {
			continue
		}
		name, size, zones, compression, ok := parseLegacyVolumeLine(line)
		if !ok {
			continue
		}
		volumes = append(volumes, domain.Volume{
			Name:        name,
			SizeBytes:   size,
			ZoneCount:   zones,
			Compression: compression,
			Status:      "configured",
		})
	}
	return volumes, nil
}

func (s *OnyxService) CreateVolume(ctx context.Context, req domain.CreateVolumeRequest) error {
	command := fmt.Sprintf("create-volume %s %d %s", req.Name, req.SizeBytes, req.Compression)
	if _, err := s.sendSocketCommand(command); err == nil {
		return nil
	}

	_, err := s.runner.Run(
		ctx,
		s.cfg.BinaryPath,
		"-c",
		s.cfg.ConfigPath,
		"create-volume",
		"-n",
		req.Name,
		"-s",
		strconv.FormatUint(req.SizeBytes, 10),
		"--compression",
		req.Compression,
	)
	return err
}

func (s *OnyxService) DeleteVolume(ctx context.Context, name string) error {
	command := fmt.Sprintf("delete-volume %s", name)
	if _, err := s.sendSocketCommand(command); err == nil {
		return nil
	}

	_, err := s.runner.Run(
		ctx,
		s.cfg.BinaryPath,
		"-c",
		s.cfg.ConfigPath,
		"delete-volume",
		"-n",
		name,
	)
	return err
}

func (s *OnyxService) statusRaw(ctx context.Context) (string, error) {
	output, err := s.runner.Run(ctx, s.cfg.BinaryPath, "-c", s.cfg.ConfigPath, "status")
	if err != nil {
		return "", err
	}
	return output, nil
}

func (s *OnyxService) sendSocketCommand(command string) ([]string, error) {
	if _, err := os.Stat(s.cfg.SocketPath); err != nil {
		return nil, err
	}

	conn, err := net.Dial("unix", s.cfg.SocketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if _, err := fmt.Fprintf(conn, "%s\n", command); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(conn)
	lines := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "ok" || strings.HasPrefix(line, "ok ") {
			continue
		}
		if strings.HasPrefix(line, "error:") {
			return nil, fmt.Errorf(strings.TrimSpace(strings.TrimPrefix(line, "error:")))
		}
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}

func atoi(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func parseUint(value string) uint64 {
	parsed, _ := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return parsed
}

func parseAllocator(value string) (uint64, uint64) {
	parts := strings.Split(strings.TrimSpace(value), "/")
	if len(parts) != 2 {
		return 0, 0
	}
	return parseUint(parts[0]), parseUint(parts[1])
}

func parseMetricsLine(value string) map[string]uint64 {
	result := make(map[string]uint64)
	for _, token := range strings.Fields(value) {
		pair := strings.SplitN(token, "=", 2)
		if len(pair) != 2 {
			continue
		}
		result[pair[0]] = parseUint(pair[1])
	}
	return result
}

func parseLegacyVolumeLine(line string) (string, uint64, uint32, string, bool) {
	trimmed := strings.TrimSpace(line)
	parts := strings.Split(trimmed, ":")
	if len(parts) != 2 {
		return "", 0, 0, "", false
	}
	name := strings.TrimSpace(parts[0])
	right := parts[1]
	segments := strings.Split(right, ",")
	if len(segments) < 3 {
		return "", 0, 0, "", false
	}

	sizeText := strings.TrimSpace(strings.TrimSuffix(segments[0], "bytes"))
	zonesText := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(segments[1]), "zones="))
	compression := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(segments[2]), "compression="))

	return name, parseUint(sizeText), uint32(parseUint(zonesText)), compression, true
}
