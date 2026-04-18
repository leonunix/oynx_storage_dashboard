package services

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/nakabonne/tstorage"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/config"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
)

var availableTelemetryWindows = []string{"1h", "6h", "24h", "7d", "14d"}

type metricsSampler interface {
	SampleTelemetry(ctx context.Context) (*telemetrySample, error)
}

type telemetrySample struct {
	CapturedAt           time.Time
	EngineMode           string
	VolumeCount          int
	LiveHandleCount      int
	ZoneCount            int
	BufferFillPercent    int
	BufferPendingEntries uint64
	BufferPayloadBytes   uint64
	BufferPayloadLimit   uint64
	AllocatorFreeBlocks  uint64
	AllocatorTotalBlocks uint64
	CompressionRatio     float64
	DedupHitRatePct      float64
	DataReductionRatio   float64
	Metrics              domain.MetricsJSON
}

func newTelemetrySample(capturedAt time.Time, overview domain.Overview, metrics domain.MetricsJSON) *telemetrySample {
	compressionRatio := 1.0
	if metrics.CompressOutputBytes > 0 {
		compressionRatio = float64(metrics.CompressInputBytes) / float64(metrics.CompressOutputBytes)
	}

	dedupHitRatePct := 0.0
	dedupTotal := metrics.DedupHits + metrics.DedupMisses
	if dedupTotal > 0 {
		dedupHitRatePct = (float64(metrics.DedupHits) / float64(dedupTotal)) * 100
	}

	dataReductionRatio := 1.0
	dedupSavedBytes := metrics.DedupHits * 4096
	logicalTotal := metrics.CompressInputBytes + dedupSavedBytes
	if metrics.CompressOutputBytes > 0 && logicalTotal > 0 {
		dataReductionRatio = float64(logicalTotal) / float64(metrics.CompressOutputBytes)
	}

	return &telemetrySample{
		CapturedAt:           capturedAt,
		EngineMode:           overview.EngineMode,
		VolumeCount:          overview.VolumeCount,
		LiveHandleCount:      overview.LiveHandleCount,
		ZoneCount:            overview.ZoneCount,
		BufferFillPercent:    overview.BufferFillPercent,
		BufferPendingEntries: overview.BufferPendingEntries,
		BufferPayloadBytes:   overview.BufferPayloadBytes,
		BufferPayloadLimit:   overview.BufferPayloadLimit,
		AllocatorFreeBlocks:  overview.AllocatorFreeBlocks,
		AllocatorTotalBlocks: overview.AllocatorTotalBlocks,
		CompressionRatio:     compressionRatio,
		DedupHitRatePct:      dedupHitRatePct,
		DataReductionRatio:   dataReductionRatio,
		Metrics:              metrics,
	}
}

func (s *telemetrySample) snapshot() *domain.TelemetrySnapshot {
	if s == nil {
		return nil
	}

	bufferPayloadUtil := 0.0
	if s.BufferPayloadLimit > 0 {
		bufferPayloadUtil = (float64(s.BufferPayloadBytes) / float64(s.BufferPayloadLimit)) * 100
	}

	allocatorUsage := 0.0
	if s.AllocatorTotalBlocks > 0 && s.AllocatorFreeBlocks <= s.AllocatorTotalBlocks {
		usedBlocks := s.AllocatorTotalBlocks - s.AllocatorFreeBlocks
		allocatorUsage = (float64(usedBlocks) / float64(s.AllocatorTotalBlocks)) * 100
	}

	return &domain.TelemetrySnapshot{
		CapturedAt:               s.CapturedAt,
		EngineMode:               s.EngineMode,
		VolumeCount:              s.VolumeCount,
		LiveHandleCount:          s.LiveHandleCount,
		ZoneCount:                s.ZoneCount,
		BufferFillPercent:        s.BufferFillPercent,
		BufferPendingEntries:     s.BufferPendingEntries,
		BufferPayloadBytes:       s.BufferPayloadBytes,
		BufferPayloadLimit:       s.BufferPayloadLimit,
		BufferPayloadUtilPercent: bufferPayloadUtil,
		AllocatorFreeBlocks:      s.AllocatorFreeBlocks,
		AllocatorTotalBlocks:     s.AllocatorTotalBlocks,
		AllocatorUsagePercent:    allocatorUsage,
		CompressionRatio:         s.CompressionRatio,
		DedupHitRatePct:          s.DedupHitRatePct,
		DataReductionRatio:       s.DataReductionRatio,
	}
}

func (s *telemetrySample) rows() []tstorage.Row {
	if s == nil {
		return nil
	}

	ts := s.CapturedAt.Unix()
	add := func(metric string, value float64) tstorage.Row {
		return tstorage.Row{
			Metric: metric,
			DataPoint: tstorage.DataPoint{
				Timestamp: ts,
				Value:     value,
			},
		}
	}

	return []tstorage.Row{
		add("volume_read_ops", float64(s.Metrics.VolumeReadOps)),
		add("volume_read_bytes", float64(s.Metrics.VolumeReadBytes)),
		add("volume_write_ops", float64(s.Metrics.VolumeWriteOps)),
		add("volume_write_bytes", float64(s.Metrics.VolumeWriteBytes)),
		add("buffer_read_ops", float64(s.Metrics.BufferReadOps)),
		add("buffer_read_bytes", float64(s.Metrics.BufferReadBytes)),
		add("buffer_write_ops", float64(s.Metrics.BufferWriteOps)),
		add("buffer_write_bytes", float64(s.Metrics.BufferWriteBytes)),
		add("lv3_read_ops", float64(s.Metrics.Lv3ReadOps)),
		add("lv3_read_compressed_bytes", float64(s.Metrics.Lv3ReadCompressedBytes)),
		add("lv3_read_decompressed_bytes", float64(s.Metrics.Lv3ReadDecompressedBytes)),
		add("lv3_write_ops", float64(s.Metrics.Lv3WriteOps)),
		add("lv3_write_compressed_bytes", float64(s.Metrics.Lv3WriteCompressedBytes)),
		add("volume_read_total_ns", float64(s.Metrics.VolumeReadTotalNs)),
		add("volume_write_total_ns", float64(s.Metrics.VolumeWriteTotalNs)),
		add("dedup_hits", float64(s.Metrics.DedupHits)),
		add("dedup_misses", float64(s.Metrics.DedupMisses)),
		add("gc_blocks_rewritten", float64(s.Metrics.GcBlocksRewritten)),
		add("flush_errors", float64(s.Metrics.FlushErrors)),
		add("buffer_backpressure_events", float64(s.Metrics.BufferBackpressureEvents)),
		add("buffer_fill_pct", float64(s.BufferFillPercent)),
		add("buffer_pending_entries", float64(s.BufferPendingEntries)),
		add("buffer_payload_bytes", float64(s.BufferPayloadBytes)),
		add("buffer_payload_limit", float64(s.BufferPayloadLimit)),
		add("compression_ratio", s.CompressionRatio),
		add("dedup_hit_rate_pct", s.DedupHitRatePct),
		add("data_reduction_ratio", s.DataReductionRatio),
		add("allocator_free_blocks", float64(s.AllocatorFreeBlocks)),
		add("allocator_total_blocks", float64(s.AllocatorTotalBlocks)),
		add("volume_count", float64(s.VolumeCount)),
		add("live_handle_count", float64(s.LiveHandleCount)),
		add("zone_count", float64(s.ZoneCount)),
	}
}

type MetricsHistoryService struct {
	storage        tstorage.Storage
	sampler        metricsSampler
	sampleInterval time.Duration
	retention      time.Duration

	doneCh    chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	wg        sync.WaitGroup

	stateMu  sync.RWMutex
	sampleMu sync.Mutex

	latest   *telemetrySample
	previous *telemetrySample
}

func NewMetricsHistoryService(cfg config.MetricsConfig, sampler metricsSampler) (*MetricsHistoryService, error) {
	if sampler == nil {
		return nil, errors.New("metrics sampler is required")
	}
	if cfg.DataPath == "" {
		cfg.DataPath = "var/metrics-tsdb"
	}
	if cfg.SampleInterval <= 0 {
		cfg.SampleInterval = time.Minute
	}
	if cfg.Retention <= 0 {
		cfg.Retention = 14 * 24 * time.Hour
	}
	if cfg.PartitionDuration <= 0 {
		cfg.PartitionDuration = 24 * time.Hour
	}

	storage, err := tstorage.NewStorage(
		tstorage.WithDataPath(cfg.DataPath),
		tstorage.WithTimestampPrecision(tstorage.Seconds),
		tstorage.WithRetention(cfg.Retention),
		tstorage.WithPartitionDuration(cfg.PartitionDuration),
	)
	if err != nil {
		return nil, err
	}

	return &MetricsHistoryService{
		storage:        storage,
		sampler:        sampler,
		sampleInterval: cfg.SampleInterval,
		retention:      cfg.Retention,
		doneCh:         make(chan struct{}),
	}, nil
}

func (s *MetricsHistoryService) Start() {
	s.startOnce.Do(func() {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			if err := s.sampleAndStore(context.Background(), 0); err != nil {
				log.Printf("dashboard metrics sampler: initial sample failed: %v", err)
			}

			timer := time.NewTimer(delayUntilNextBoundary(time.Now().UTC(), s.sampleInterval))
			defer timer.Stop()

			for {
				select {
				case <-s.doneCh:
					return
				case <-timer.C:
					if err := s.sampleAndStore(context.Background(), 0); err != nil {
						log.Printf("dashboard metrics sampler: background sample failed: %v", err)
					}
					timer.Reset(s.sampleInterval)
				}
			}
		}()
	})
}

func (s *MetricsHistoryService) Stop() error {
	var err error
	s.stopOnce.Do(func() {
		close(s.doneCh)
		s.wg.Wait()
		err = s.storage.Close()
	})
	return err
}

func (s *MetricsHistoryService) Telemetry(ctx context.Context, window time.Duration) (domain.TelemetryResponse, error) {
	if window <= 0 {
		window = 24 * time.Hour
	}

	if err := s.EnsureFresh(ctx, s.sampleInterval+15*time.Second); err != nil {
		return domain.TelemetryResponse{}, err
	}

	latest, previous := s.currentSamples()
	response := domain.TelemetryResponse{
		GeneratedAt:      time.Now().UTC(),
		WindowSeconds:    int64(window.Seconds()),
		StepSeconds:      int64(s.sampleInterval.Seconds()),
		RetentionDays:    int(s.retention.Hours() / 24),
		AvailableWindows: availableTelemetryWindows,
		Latest:           latest.snapshot(),
		Rates:            buildTelemetryRates(previous, latest),
		Series:           map[string][]domain.TelemetryPoint{},
	}
	if latest == nil {
		return response, nil
	}

	start := latest.CapturedAt.Add(-window).Unix()
	end := latest.CapturedAt.Unix() + 1

	var err error
	if response.Series["client_write_bps"], err = s.selectRateSeries("volume_write_bytes", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["buffer_write_bps"], err = s.selectRateSeries("buffer_write_bytes", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["lv3_write_bps"], err = s.selectRateSeries("lv3_write_compressed_bytes", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["client_read_bps"], err = s.selectRateSeries("volume_read_bytes", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["buffer_read_bps"], err = s.selectRateSeries("buffer_read_bytes", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["lv3_read_bps"], err = s.selectRateSeries("lv3_read_compressed_bytes", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["lv3_read_decompressed_bps"], err = s.selectRateSeries("lv3_read_decompressed_bytes", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["client_read_latency_ns"], err = s.selectAvgLatencySeries("volume_read_total_ns", "volume_read_ops", start, end); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["client_write_latency_ns"], err = s.selectAvgLatencySeries("volume_write_total_ns", "volume_write_ops", start, end); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["client_write_iops"], err = s.selectRateSeries("volume_write_ops", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["buffer_write_iops"], err = s.selectRateSeries("buffer_write_ops", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["lv3_write_iops"], err = s.selectRateSeries("lv3_write_ops", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["client_read_iops"], err = s.selectRateSeries("volume_read_ops", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["buffer_read_iops"], err = s.selectRateSeries("buffer_read_ops", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["lv3_read_iops"], err = s.selectRateSeries("lv3_read_ops", start, end, 1); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["buffer_fill_pct"], err = s.selectGaugeSeries("buffer_fill_pct", start, end); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["buffer_pending_entries"], err = s.selectGaugeSeries("buffer_pending_entries", start, end); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["buffer_payload_bytes"], err = s.selectGaugeSeries("buffer_payload_bytes", start, end); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["compression_ratio"], err = s.selectGaugeSeries("compression_ratio", start, end); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["dedup_hit_rate_pct"], err = s.selectGaugeSeries("dedup_hit_rate_pct", start, end); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["data_reduction_ratio"], err = s.selectGaugeSeries("data_reduction_ratio", start, end); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["dedup_hits_per_min"], err = s.selectRateSeries("dedup_hits", start, end, 60); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["dedup_misses_per_min"], err = s.selectRateSeries("dedup_misses", start, end, 60); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["gc_rewrites_per_min"], err = s.selectRateSeries("gc_blocks_rewritten", start, end, 60); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["flush_errors_per_min"], err = s.selectRateSeries("flush_errors", start, end, 60); err != nil {
		return domain.TelemetryResponse{}, err
	}
	if response.Series["backpressure_events_per_min"], err = s.selectRateSeries("buffer_backpressure_events", start, end, 60); err != nil {
		return domain.TelemetryResponse{}, err
	}

	return response, nil
}

func (s *MetricsHistoryService) EnsureFresh(ctx context.Context, maxAge time.Duration) error {
	return s.sampleAndStore(ctx, maxAge)
}

func (s *MetricsHistoryService) sampleAndStore(ctx context.Context, maxAge time.Duration) error {
	s.sampleMu.Lock()
	defer s.sampleMu.Unlock()

	if maxAge > 0 {
		latest, _ := s.currentSamples()
		if latest != nil && time.Since(latest.CapturedAt) <= maxAge {
			return nil
		}
	}

	sample, err := s.sampler.SampleTelemetry(ctx)
	if err != nil {
		return err
	}
	if sample.CapturedAt.IsZero() {
		sample.CapturedAt = time.Now().UTC()
	}
	if err := s.storage.InsertRows(sample.rows()); err != nil {
		return err
	}

	s.stateMu.Lock()
	s.previous = s.latest
	s.latest = sample
	s.stateMu.Unlock()

	return nil
}

func (s *MetricsHistoryService) currentSamples() (*telemetrySample, *telemetrySample) {
	s.stateMu.RLock()
	latest := s.latest
	previous := s.previous
	s.stateMu.RUnlock()
	return latest, previous
}

func (s *MetricsHistoryService) selectGaugeSeries(metric string, start, end int64) ([]domain.TelemetryPoint, error) {
	points, err := s.storage.Select(metric, nil, start, end)
	if errors.Is(err, tstorage.ErrNoDataPoints) {
		return []domain.TelemetryPoint{}, nil
	}
	if err != nil {
		return nil, err
	}

	result := make([]domain.TelemetryPoint, 0, len(points))
	for _, point := range points {
		result = append(result, domain.TelemetryPoint{
			Timestamp: point.Timestamp,
			Value:     point.Value,
		})
	}
	return result, nil
}

func (s *MetricsHistoryService) selectRateSeries(metric string, start, end int64, scale float64) ([]domain.TelemetryPoint, error) {
	points, err := s.storage.Select(metric, nil, start, end)
	if errors.Is(err, tstorage.ErrNoDataPoints) {
		return []domain.TelemetryPoint{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(points) < 2 {
		return []domain.TelemetryPoint{}, nil
	}

	result := make([]domain.TelemetryPoint, 0, len(points)-1)
	for i := 1; i < len(points); i++ {
		prev := points[i-1]
		curr := points[i]
		deltaTime := float64(curr.Timestamp - prev.Timestamp)
		if deltaTime <= 0 {
			continue
		}
		deltaValue := curr.Value - prev.Value
		if deltaValue < 0 {
			deltaValue = 0
		}

		result = append(result, domain.TelemetryPoint{
			Timestamp: curr.Timestamp,
			Value:     (deltaValue / deltaTime) * scale,
		})
	}
	return result, nil
}

// selectAvgLatencySeries returns a per-sample average latency (ns/op) derived
// from two cumulative counters (total ns and op count). For each adjacent
// sample pair, it computes Δns / Δops. Points with no op delta are emitted as
// zero to keep the x-axis aligned with other rate series.
func (s *MetricsHistoryService) selectAvgLatencySeries(totalNsMetric, opsMetric string, start, end int64) ([]domain.TelemetryPoint, error) {
	nsPoints, err := s.storage.Select(totalNsMetric, nil, start, end)
	if errors.Is(err, tstorage.ErrNoDataPoints) {
		nsPoints = nil
	} else if err != nil {
		return nil, err
	}
	opsPoints, err := s.storage.Select(opsMetric, nil, start, end)
	if errors.Is(err, tstorage.ErrNoDataPoints) {
		opsPoints = nil
	} else if err != nil {
		return nil, err
	}
	if len(nsPoints) < 2 || len(opsPoints) < 2 {
		return []domain.TelemetryPoint{}, nil
	}

	// Align by timestamp — both series are sampled from the same snapshot on
	// the same schedule, so the common case is 1:1 matching. We still do an
	// explicit lookup so a missed sample on either side doesn't skew results.
	opsByTs := make(map[int64]float64, len(opsPoints))
	for _, p := range opsPoints {
		opsByTs[p.Timestamp] = p.Value
	}

	result := make([]domain.TelemetryPoint, 0, len(nsPoints)-1)
	for i := 1; i < len(nsPoints); i++ {
		prev := nsPoints[i-1]
		curr := nsPoints[i]
		prevOps, okPrev := opsByTs[prev.Timestamp]
		currOps, okCurr := opsByTs[curr.Timestamp]
		if !okPrev || !okCurr {
			continue
		}
		deltaNs := curr.Value - prev.Value
		deltaOps := currOps - prevOps
		if deltaNs < 0 || deltaOps <= 0 {
			continue
		}
		result = append(result, domain.TelemetryPoint{
			Timestamp: curr.Timestamp,
			Value:     deltaNs / deltaOps,
		})
	}
	return result, nil
}

func buildTelemetryRates(previous, latest *telemetrySample) domain.TelemetryRates {
	if previous == nil || latest == nil {
		return domain.TelemetryRates{}
	}

	windowSeconds := latest.CapturedAt.Sub(previous.CapturedAt).Seconds()
	if windowSeconds <= 0 {
		return domain.TelemetryRates{}
	}

	rate := func(prev, curr uint64) float64 {
		if curr < prev {
			return 0
		}
		return float64(curr-prev) / windowSeconds
	}

	avgLatency := func(prevNs, currNs, prevOps, currOps uint64) float64 {
		if currNs < prevNs || currOps <= prevOps {
			return 0
		}
		deltaOps := currOps - prevOps
		if deltaOps == 0 {
			return 0
		}
		return float64(currNs-prevNs) / float64(deltaOps)
	}

	return domain.TelemetryRates{
		WindowSeconds:          windowSeconds,
		ClientReadBps:          rate(previous.Metrics.VolumeReadBytes, latest.Metrics.VolumeReadBytes),
		ClientWriteBps:         rate(previous.Metrics.VolumeWriteBytes, latest.Metrics.VolumeWriteBytes),
		ClientReadIops:         rate(previous.Metrics.VolumeReadOps, latest.Metrics.VolumeReadOps),
		ClientWriteIops:        rate(previous.Metrics.VolumeWriteOps, latest.Metrics.VolumeWriteOps),
		ClientReadLatencyNs:    avgLatency(previous.Metrics.VolumeReadTotalNs, latest.Metrics.VolumeReadTotalNs, previous.Metrics.VolumeReadOps, latest.Metrics.VolumeReadOps),
		ClientWriteLatencyNs:   avgLatency(previous.Metrics.VolumeWriteTotalNs, latest.Metrics.VolumeWriteTotalNs, previous.Metrics.VolumeWriteOps, latest.Metrics.VolumeWriteOps),
		BufferReadBps:          rate(previous.Metrics.BufferReadBytes, latest.Metrics.BufferReadBytes),
		BufferWriteBps:         rate(previous.Metrics.BufferWriteBytes, latest.Metrics.BufferWriteBytes),
		BufferReadIops:         rate(previous.Metrics.BufferReadOps, latest.Metrics.BufferReadOps),
		BufferWriteIops:        rate(previous.Metrics.BufferWriteOps, latest.Metrics.BufferWriteOps),
		Lv3ReadBps:             rate(previous.Metrics.Lv3ReadCompressedBytes, latest.Metrics.Lv3ReadCompressedBytes),
		Lv3ReadDecompressedBps: rate(previous.Metrics.Lv3ReadDecompressedBytes, latest.Metrics.Lv3ReadDecompressedBytes),
		Lv3WriteBps:            rate(previous.Metrics.Lv3WriteCompressedBytes, latest.Metrics.Lv3WriteCompressedBytes),
		Lv3ReadIops:            rate(previous.Metrics.Lv3ReadOps, latest.Metrics.Lv3ReadOps),
		Lv3WriteIops:           rate(previous.Metrics.Lv3WriteOps, latest.Metrics.Lv3WriteOps),
	}
}

func delayUntilNextBoundary(now time.Time, step time.Duration) time.Duration {
	if step <= 0 {
		return time.Minute
	}
	next := now.Truncate(step).Add(step)
	delay := time.Until(next)
	if delay <= 0 {
		return step
	}
	return delay
}
