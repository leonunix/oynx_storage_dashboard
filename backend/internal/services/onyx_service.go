package services

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

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
	// Try JSON IPC first
	if ov, err := s.overviewJSON(); err == nil {
		return ov, nil
	}

	// Fallback to text parsing
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

func (s *OnyxService) overviewJSON() (domain.Overview, error) {
	lines, err := s.sendSocketCommand("status-json")
	if err != nil {
		return domain.Overview{}, err
	}
	raw := strings.Join(lines, "\n")
	var status domain.StatusJSON
	if err := json.Unmarshal([]byte(raw), &status); err != nil {
		return domain.Overview{}, fmt.Errorf("parse status-json: %w", err)
	}

	ov := domain.Overview{
		EngineMode:    status.Mode,
		EngineRunning: true,
		UblkDevices:   status.UblkDevices,
		Metrics:       map[string]any{},
	}
	if s := status.Status; s != nil {
		ov.VolumeCount = s.VolumeCount
		ov.LiveHandleCount = s.LiveHandleCount
		if s.ZoneCount != nil {
			ov.ZoneCount = *s.ZoneCount
		}
		if s.BufferFillPct != nil {
			ov.BufferFillPercent = *s.BufferFillPct
		}
		if s.BufferPendingEntries != nil {
			ov.BufferPendingEntries = *s.BufferPendingEntries
		}
		if s.BufferPayloadBytes != nil {
			ov.BufferPayloadBytes = *s.BufferPayloadBytes
		}
		if s.BufferPayloadLimit != nil {
			ov.BufferPayloadLimit = *s.BufferPayloadLimit
		}
		if s.AllocatorFreeBlocks != nil {
			ov.AllocatorFreeBlocks = *s.AllocatorFreeBlocks
		}
		if s.AllocatorTotalBlocks != nil {
			ov.AllocatorTotalBlocks = *s.AllocatorTotalBlocks
		}
		ov.BufferShards = s.BufferShards
		ov.Metrics["uptime_secs"] = s.Metrics.UptimeSecs
		ov.Metrics["volume_read_ops"] = s.Metrics.VolumeReadOps
		ov.Metrics["volume_write_ops"] = s.Metrics.VolumeWriteOps
		ov.Metrics["volume_read_bytes"] = s.Metrics.VolumeReadBytes
		ov.Metrics["volume_write_bytes"] = s.Metrics.VolumeWriteBytes
		ov.Metrics["buffer_write_ops"] = s.Metrics.BufferWriteOps
		ov.Metrics["buffer_write_bytes"] = s.Metrics.BufferWriteBytes
		ov.Metrics["buffer_read_ops"] = s.Metrics.BufferReadOps
		ov.Metrics["buffer_read_bytes"] = s.Metrics.BufferReadBytes
		ov.Metrics["lv3_read_ops"] = s.Metrics.Lv3ReadOps
		ov.Metrics["lv3_read_bytes"] = s.Metrics.Lv3ReadBytes
		ov.Metrics["lv3_write_ops"] = s.Metrics.Lv3WriteOps
		ov.Metrics["lv3_write_bytes"] = s.Metrics.Lv3WriteBytes
		ov.Metrics["compress_input_bytes"] = s.Metrics.CompressInputBytes
		ov.Metrics["compress_output_bytes"] = s.Metrics.CompressOutputBytes
		ov.Metrics["buffer_payload_bytes"] = ov.BufferPayloadBytes
		ov.Metrics["buffer_payload_limit"] = ov.BufferPayloadLimit
		ov.Metrics["buffer_backpressure_events"] = s.Metrics.BufferBackpressureEvents
		ov.Metrics["buffer_hydration_skips"] = s.Metrics.BufferHydrationSkippedDueToMemLimit
		ov.Metrics["dedup_hits"] = s.Metrics.DedupHits
		ov.Metrics["dedup_misses"] = s.Metrics.DedupMisses
		ov.Metrics["writer_precheck_live_pba_ops"] = s.Metrics.FlushWriterPrecheckLivePbaOps
		ov.Metrics["writer_precheck_live_pba_failures"] = s.Metrics.FlushWriterPrecheckLivePbaFailures
		ov.Metrics["gc_cycles"] = s.Metrics.GcCycles
		ov.Metrics["gc_blocks_rewritten"] = s.Metrics.GcBlocksRewritten
		ov.Metrics["flush_errors"] = s.Metrics.FlushErrors
		ov.Metrics["gc_errors"] = s.Metrics.GcErrors
		ov.Metrics["read_crc_errors"] = s.Metrics.ReadCrcErrors

		// Compute ratios
		m := s.Metrics
		if m.CompressOutputBytes > 0 {
			ov.CompressionRatio = float64(m.CompressInputBytes) / float64(m.CompressOutputBytes)
		} else {
			ov.CompressionRatio = 1.0
		}
		dedupTotal := m.DedupHits + m.DedupMisses
		if dedupTotal > 0 {
			ov.DedupHitRate = float64(m.DedupHits) / float64(dedupTotal)
		}
		// Data reduction: logical bytes (compress input + dedup saved) / physical bytes (compress output)
		dedupSavedBytes := m.DedupHits * 4096
		logicalTotal := m.CompressInputBytes + dedupSavedBytes
		if m.CompressOutputBytes > 0 && logicalTotal > 0 {
			ov.DataReductionRatio = float64(logicalTotal) / float64(m.CompressOutputBytes)
		} else {
			ov.DataReductionRatio = 1.0
		}
	}
	return ov, nil
}

// MetricsJSON fetches structured metrics via JSON IPC.
func (s *OnyxService) MetricsJSON() (*domain.MetricsJSON, error) {
	lines, err := s.sendSocketCommand("metrics-json")
	if err != nil {
		return nil, err
	}
	raw := strings.Join(lines, "\n")
	var m domain.MetricsJSON
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("parse metrics-json: %w", err)
	}
	return &m, nil
}

func (s *OnyxService) SampleTelemetry(ctx context.Context) (*telemetrySample, error) {
	metrics, err := s.MetricsJSON()
	if err != nil {
		return nil, err
	}

	overview, err := s.Overview(ctx)
	if err != nil {
		return nil, err
	}

	return newTelemetrySample(time.Now().UTC(), overview, *metrics), nil
}

func (s *OnyxService) ListVolumes(ctx context.Context) ([]domain.Volume, error) {
	// Try JSON IPC first
	if vols, err := s.listVolumesJSON(); err == nil {
		return vols, nil
	}

	// Fallback: text socket IPC
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

	// Fallback: CLI
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

func (s *OnyxService) listVolumesJSON() ([]domain.Volume, error) {
	lines, err := s.sendSocketCommand("volumes-json")
	if err != nil {
		return nil, err
	}
	raw := strings.Join(lines, "\n")
	var vols []domain.VolumeJSON
	if err := json.Unmarshal([]byte(raw), &vols); err != nil {
		return nil, fmt.Errorf("parse volumes-json: %w", err)
	}
	result := make([]domain.Volume, 0, len(vols))
	for _, v := range vols {
		result = append(result, domain.Volume{
			Name:        v.ID,
			SizeBytes:   v.SizeBytes,
			ZoneCount:   v.ZoneCount,
			Compression: fmt.Sprintf("%v", v.Compression),
			Status:      "online",
			CreatedAt:   v.CreatedAt,
			Metrics:     v.Metrics,
		})
	}
	return result, nil
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
		if line == "" {
			continue
		}
		if line == "ok" || strings.HasPrefix(line, "ok ") {
			break
		}
		if strings.HasPrefix(line, "error:") {
			return nil, errors.New(strings.TrimSpace(strings.TrimPrefix(line, "error:")))
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
