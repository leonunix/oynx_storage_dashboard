package domain

import "time"

type Permission string

const (
	PermissionOverviewRead Permission = "overview:read"
	PermissionMetricsRead  Permission = "metrics:read"
	PermissionVolumesRead  Permission = "volumes:read"
	PermissionVolumesWrite Permission = "volumes:write"
	PermissionStorageRead  Permission = "storage:read"
	PermissionStorageWrite Permission = "storage:write"
	PermissionAuditRead    Permission = "audit:read"
	PermissionUsersManage  Permission = "users:manage"
)

type User struct {
	Username    string       `json:"username"`
	DisplayName string       `json:"displayName"`
	Role        string       `json:"role"`
	Permissions []Permission `json:"permissions"`
	Disabled    bool         `json:"disabled"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type SetupStatus struct {
	Initialized       bool   `json:"initialized"`
	SuggestedUsername string `json:"suggestedUsername,omitempty"`
	SuggestedRole     string `json:"suggestedRole,omitempty"`
}

type SetupInitializeRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
}

type UserCreateRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	Password    string `json:"password"`
}

type UserUpdateRequest struct {
	DisplayName *string `json:"displayName"`
	Role        *string `json:"role"`
	Disabled    *bool   `json:"disabled"`
}

type UserPasswordResetRequest struct {
	Password string `json:"password"`
}

type RoleDefinition struct {
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}

type AuditEvent struct {
	ID          string    `json:"id"`
	At          time.Time `json:"at"`
	Actor       string    `json:"actor"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	Result      string    `json:"result"`
	Description string    `json:"description"`
}

type Volume struct {
	Name        string           `json:"name"`
	SizeBytes   uint64           `json:"sizeBytes"`
	ZoneCount   uint32           `json:"zoneCount"`
	Compression string           `json:"compression"`
	Status      string           `json:"status"`
	CreatedAt   uint64           `json:"createdAt,omitempty"`
	Metrics     *VolumeIOMetrics `json:"metrics,omitempty"`
}

type VolumeIOMetrics struct {
	ReadOps     uint64 `json:"readOps"`
	ReadBytes   uint64 `json:"readBytes"`
	WriteOps    uint64 `json:"writeOps"`
	WriteBytes  uint64 `json:"writeBytes"`
	ReadErrors  uint64 `json:"readErrors"`
	WriteErrors uint64 `json:"writeErrors"`
}

type CreateVolumeRequest struct {
	Name        string `json:"name"`
	SizeBytes   uint64 `json:"sizeBytes"`
	Compression string `json:"compression"`
}

type DeviceSummary struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Size   string `json:"size"`
	State  string `json:"state"`
	Parent string `json:"parent,omitempty"`
}

type DMTarget struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Target string `json:"target"`
}

type LvmVolume struct {
	Name   string `json:"name"`
	VGName string `json:"vgName"`
	Attr   string `json:"attr"`
	Size   string `json:"size"`
}

type StorageLayout struct {
	BlockDevices    []DeviceSummary  `json:"blockDevices"`
	DMTargets       []DMTarget       `json:"dmTargets"`
	LogicalVolumes  []LvmVolume      `json:"logicalVolumes"`
	RaidArrays      []RaidArray      `json:"raidArrays"`
	PhysicalVolumes []PhysicalVolume `json:"physicalVolumes"`
	VolumeGroups    []VolumeGroup    `json:"volumeGroups"`
	AllowMutations  bool             `json:"allowMutations"`
	Warnings        []string         `json:"warnings"`
}

// ── RAID types ─────────────────────────────────────────────────────

type RaidArray struct {
	Name       string   `json:"name"`
	Level      string   `json:"level"`
	State      string   `json:"state"`
	Size       string   `json:"size"`
	Devices    []string `json:"devices"`
	ActiveDevs int      `json:"activeDevs"`
	TotalDevs  int      `json:"totalDevs"`
	UUID       string   `json:"uuid,omitempty"`
}

type RaidCreateRequest struct {
	Name    string   `json:"name"`
	Level   string   `json:"level"`
	Devices []string `json:"devices"`
	ChunkKB int      `json:"chunkKb,omitempty"`
	Force   bool     `json:"force,omitempty"`
}

type RaidStopRequest struct {
	Name string `json:"name"`
}

// ── LVM PV types ──────────────────────────────────────────────────

type PhysicalVolume struct {
	Name   string `json:"name"`
	VGName string `json:"vgName"`
	Size   string `json:"size"`
	Free   string `json:"free"`
	Attr   string `json:"attr"`
	UUID   string `json:"uuid,omitempty"`
}

type PVCreateRequest struct {
	Device string `json:"device"`
	Force  bool   `json:"force,omitempty"`
}

type PVRemoveRequest struct {
	Device string `json:"device"`
	Force  bool   `json:"force,omitempty"`
}

// ── LVM VG types ──────────────────────────────────────────────────

type VolumeGroup struct {
	Name    string `json:"name"`
	Size    string `json:"size"`
	Free    string `json:"free"`
	PVCount int    `json:"pvCount"`
	LVCount int    `json:"lvCount"`
	Attr    string `json:"attr"`
	UUID    string `json:"uuid,omitempty"`
}

type VGCreateRequest struct {
	Name    string   `json:"name"`
	Devices []string `json:"devices"`
}

type VGRemoveRequest struct {
	Name  string `json:"name"`
	Force bool   `json:"force,omitempty"`
}

// ── LVM LV types ──────────────────────────────────────────────────

type LVCreateRequest struct {
	Name   string `json:"name"`
	VGName string `json:"vgName"`
	Size   string `json:"size"`
}

type LVRemoveRequest struct {
	Name   string `json:"name"`
	VGName string `json:"vgName"`
	Force  bool   `json:"force,omitempty"`
}

type LVResizeRequest struct {
	Name   string `json:"name"`
	VGName string `json:"vgName"`
	Size   string `json:"size"`
}

// ── Provision execution ───────────────────────────────────────────

type ProvisionExecuteRequest struct {
	Commands []string `json:"commands"`
}

type ProvisionExecuteResult struct {
	Success bool            `json:"success"`
	Results []CommandResult `json:"results"`
}

type CommandResult struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Error    string `json:"error,omitempty"`
}

// ── Provision planning ────────────────────────────────────────────

type ProvisionRequest struct {
	Name        string   `json:"name"`
	Devices     []string `json:"devices"`
	RaidType    string   `json:"raidType"`
	StripSizeKB int      `json:"stripSizeKb"`
	VGName      string   `json:"vgName"`
	DataLVName  string   `json:"dataLvName"`
	MetaLVName  string   `json:"metaLvName"`
}

type ProvisionPlan struct {
	Name           string   `json:"name"`
	SafetyChecks   []string `json:"safetyChecks"`
	Commands       []string `json:"commands"`
	Warnings       []string `json:"warnings"`
	ExecutionReady bool     `json:"executionReady"`
}

// ── Engine TOML configuration (mirrors Rust OnyxConfig) ─────────────

type EngineConfig struct {
	Meta    MetaSection    `json:"meta" toml:"meta"`
	Storage StorageSection `json:"storage" toml:"storage"`
	Buffer  BufferSection  `json:"buffer" toml:"buffer"`
	Ublk    UblkSection    `json:"ublk" toml:"ublk"`
	Flush   FlushSection   `json:"flush" toml:"flush"`
	Engine  EngineSection  `json:"engine" toml:"engine"`
	Gc      GcSection      `json:"gc" toml:"gc"`
	Dedup   DedupSection   `json:"dedup" toml:"dedup"`
	Service ServiceSection `json:"service" toml:"service"`
}

type MetaSection struct {
	RocksdbPath  *string `json:"rocksdb_path,omitempty" toml:"rocksdb_path,omitempty"`
	BlockCacheMB *int    `json:"block_cache_mb,omitempty" toml:"block_cache_mb,omitempty"`
	WalDir       *string `json:"wal_dir,omitempty" toml:"wal_dir,omitempty"`
}

type StorageSection struct {
	DataDevice         *string `json:"data_device,omitempty" toml:"data_device,omitempty"`
	BlockSize          *int    `json:"block_size,omitempty" toml:"block_size,omitempty"`
	UseHugepages       *bool   `json:"use_hugepages,omitempty" toml:"use_hugepages,omitempty"`
	DefaultCompression *string `json:"default_compression,omitempty" toml:"default_compression,omitempty"`
}

type BufferSection struct {
	Device            *string `json:"device,omitempty" toml:"device,omitempty"`
	CapacityMB        *int    `json:"capacity_mb,omitempty" toml:"capacity_mb,omitempty"`
	FlushWatermarkPct *int    `json:"flush_watermark_pct,omitempty" toml:"flush_watermark_pct,omitempty"`
	GroupCommitWaitUs *int    `json:"group_commit_wait_us,omitempty" toml:"group_commit_wait_us,omitempty"`
	Shards            *int    `json:"shards,omitempty" toml:"shards,omitempty"`
}

type UblkSection struct {
	NrQueues   *int `json:"nr_queues,omitempty" toml:"nr_queues,omitempty"`
	QueueDepth *int `json:"queue_depth,omitempty" toml:"queue_depth,omitempty"`
	IoBufBytes *int `json:"io_buf_bytes,omitempty" toml:"io_buf_bytes,omitempty"`
}

type FlushSection struct {
	CompressWorkers     *int `json:"compress_workers,omitempty" toml:"compress_workers,omitempty"`
	CoalesceMaxRawBytes *int `json:"coalesce_max_raw_bytes,omitempty" toml:"coalesce_max_raw_bytes,omitempty"`
	CoalesceMaxLbas     *int `json:"coalesce_max_lbas,omitempty" toml:"coalesce_max_lbas,omitempty"`
}

type EngineSection struct {
	ZoneCount      *int `json:"zone_count,omitempty" toml:"zone_count,omitempty"`
	ZoneSizeBlocks *int `json:"zone_size_blocks,omitempty" toml:"zone_size_blocks,omitempty"`
}

type GcSection struct {
	Enabled              *bool    `json:"enabled,omitempty" toml:"enabled,omitempty"`
	ScanIntervalMs       *int     `json:"scan_interval_ms,omitempty" toml:"scan_interval_ms,omitempty"`
	DeadRatioThreshold   *float64 `json:"dead_ratio_threshold,omitempty" toml:"dead_ratio_threshold,omitempty"`
	BufferUsageMaxPct    *int     `json:"buffer_usage_max_pct,omitempty" toml:"buffer_usage_max_pct,omitempty"`
	BufferUsageResumePct *int     `json:"buffer_usage_resume_pct,omitempty" toml:"buffer_usage_resume_pct,omitempty"`
	MaxRewritePerCycle   *int     `json:"max_rewrite_per_cycle,omitempty" toml:"max_rewrite_per_cycle,omitempty"`
}

type DedupSection struct {
	Enabled                *bool `json:"enabled,omitempty" toml:"enabled,omitempty"`
	Workers                *int  `json:"workers,omitempty" toml:"workers,omitempty"`
	BufferSkipThresholdPct *int  `json:"buffer_skip_threshold_pct,omitempty" toml:"buffer_skip_threshold_pct,omitempty"`
	RescanIntervalMs       *int  `json:"rescan_interval_ms,omitempty" toml:"rescan_interval_ms,omitempty"`
	MaxRescanPerCycle      *int  `json:"max_rescan_per_cycle,omitempty" toml:"max_rescan_per_cycle,omitempty"`
}

type ServiceSection struct {
	SocketPath *string `json:"socket_path,omitempty" toml:"socket_path,omitempty"`
}

type ConfigReloadResponse struct {
	Mode    string `json:"mode"`
	Message string `json:"message"`
}

type Overview struct {
	EngineMode           string            `json:"engineMode"`
	EngineRunning        bool              `json:"engineRunning"`
	VolumeCount          int               `json:"volumeCount"`
	LiveHandleCount      int               `json:"liveHandleCount"`
	ZoneCount            int               `json:"zoneCount"`
	BufferFillPercent    int               `json:"bufferFillPercent"`
	BufferPendingEntries uint64            `json:"bufferPendingEntries"`
	BufferPayloadBytes   uint64            `json:"bufferPayloadBytes"`
	BufferPayloadLimit   uint64            `json:"bufferPayloadLimit"`
	AllocatorFreeBlocks  uint64            `json:"allocatorFreeBlocks"`
	AllocatorTotalBlocks uint64            `json:"allocatorTotalBlocks"`
	UblkDevices          []uint32          `json:"ublkDevices"`
	BufferShards         []BufferShardJSON `json:"bufferShards"`
	CompressionRatio     float64           `json:"compressionRatio"`
	DedupHitRate         float64           `json:"dedupHitRate"`
	DataReductionRatio   float64           `json:"dataReductionRatio"`
	Metrics              map[string]any    `json:"metrics"`
	RawStatus            string            `json:"rawStatus"`
}

// StatusJSON matches the Rust status-json IPC response.
type StatusJSON struct {
	Mode        string            `json:"mode"`
	UblkDevices []uint32          `json:"ublk_devices"`
	Status      *EngineStatusJSON `json:"status"`
}

// EngineStatusJSON matches Rust EngineStatusSnapshot (serialized).
type EngineStatusJSON struct {
	Mode                 string            `json:"mode"`
	VolumeCount          int               `json:"volume_count"`
	LiveHandleCount      int               `json:"live_handle_count"`
	ZoneCount            *int              `json:"zone_count"`
	BufferPendingEntries *uint64           `json:"buffer_pending_entries"`
	BufferFillPct        *int              `json:"buffer_fill_pct"`
	BufferPayloadBytes   *uint64           `json:"buffer_payload_memory_bytes"`
	BufferPayloadLimit   *uint64           `json:"buffer_payload_memory_limit_bytes"`
	BufferShards         []BufferShardJSON `json:"buffer_shards"`
	AllocatorFreeBlocks  *uint64           `json:"allocator_free_blocks"`
	AllocatorTotalBlocks *uint64           `json:"allocator_total_blocks"`
	Metrics              MetricsJSON       `json:"metrics"`
}

type BufferShardJSON struct {
	ShardIdx       int     `json:"shard_idx"`
	UsedBytes      uint64  `json:"used_bytes"`
	CapacityBytes  uint64  `json:"capacity_bytes"`
	FillPct        int     `json:"fill_pct"`
	PendingEntries uint64  `json:"pending_entries"`
	HeadOffset     uint64  `json:"head_offset"`
	TailOffset     uint64  `json:"tail_offset"`
	LogOrderLen    int     `json:"log_order_len"`
	FlushedSeqsLen int     `json:"flushed_seqs_len"`
	HeadSeq        *uint64 `json:"head_seq"`
	HeadRemaining  *uint32 `json:"head_remaining_lbas"`
	HeadAgeMs      *uint64 `json:"head_age_ms"`
	HeadResidencyMs *uint64 `json:"head_residency_ms"`
}

// MetricsJSON matches Rust EngineMetricsSnapshot (serialized).
type MetricsJSON struct {
	UptimeSecs                          uint64 `json:"uptime_secs"`
	VolumeCreateOps                     uint64 `json:"volume_create_ops"`
	VolumeDeleteOps                     uint64 `json:"volume_delete_ops"`
	VolumeOpenOps                       uint64 `json:"volume_open_ops"`
	VolumeReadOps                       uint64 `json:"volume_read_ops"`
	VolumeReadBytes                     uint64 `json:"volume_read_bytes"`
	VolumeWriteOps                      uint64 `json:"volume_write_ops"`
	VolumeWriteBytes                    uint64 `json:"volume_write_bytes"`
	BufferAppends                       uint64 `json:"buffer_appends"`
	BufferAppendBytes                   uint64 `json:"buffer_append_bytes"`
	BufferWriteOps                      uint64 `json:"buffer_write_ops"`
	BufferWriteBytes                    uint64 `json:"buffer_write_bytes"`
	BufferBackpressureEvents            uint64 `json:"buffer_backpressure_events"`
	BufferBackpressureWaitNs            uint64 `json:"buffer_backpressure_wait_ns"`
	BufferHydrationSkippedDueToMemLimit uint64 `json:"buffer_hydration_skipped_due_to_mem_limit"`
	BufferHydrationHeadBypassCount      uint64 `json:"buffer_hydration_head_bypass_count"`
	BufferLookupHits                    uint64 `json:"buffer_lookup_hits"`
	BufferLookupMisses                  uint64 `json:"buffer_lookup_misses"`
	BufferReadOps                       uint64 `json:"buffer_read_ops"`
	BufferReadBytes                     uint64 `json:"buffer_read_bytes"`
	ReadBufferHits                      uint64 `json:"read_buffer_hits"`
	ReadLv3Hits                         uint64 `json:"read_lv3_hits"`
	Lv3ReadOps                          uint64 `json:"lv3_read_ops"`
	Lv3ReadBytes                        uint64 `json:"lv3_read_bytes"`
	Lv3WriteOps                         uint64 `json:"lv3_write_ops"`
	Lv3WriteBytes                       uint64 `json:"lv3_write_bytes"`
	ReadUnmapped                        uint64 `json:"read_unmapped"`
	ReadCrcErrors                       uint64 `json:"read_crc_errors"`
	CoalesceRuns                        uint64 `json:"coalesce_runs"`
	CoalescedUnits                      uint64 `json:"coalesced_units"`
	CoalescedBytes                      uint64 `json:"coalesced_bytes"`
	CompressUnits                       uint64 `json:"compress_units"`
	CompressInputBytes                  uint64 `json:"compress_input_bytes"`
	CompressOutputBytes                 uint64 `json:"compress_output_bytes"`
	DedupHits                           uint64 `json:"dedup_hits"`
	DedupMisses                         uint64 `json:"dedup_misses"`
	DedupSkippedUnits                   uint64 `json:"dedup_skipped_units"`
	DedupHitFailures                    uint64 `json:"dedup_hit_failures"`
	DedupLookupOps                      uint64 `json:"dedup_lookup_ops"`
	DedupLookupNs                       uint64 `json:"dedup_lookup_ns"`
	DedupLiveCheckOps                   uint64 `json:"dedup_live_check_ops"`
	DedupLiveCheckNs                    uint64 `json:"dedup_live_check_ns"`
	DedupStaleIndexEntries              uint64 `json:"dedup_stale_index_entries"`
	DedupStaleDeleteNs                  uint64 `json:"dedup_stale_delete_ns"`
	DedupHitCommitOps                   uint64 `json:"dedup_hit_commit_ops"`
	DedupHitCommitNs                    uint64 `json:"dedup_hit_commit_ns"`
	FlushUnitsWritten                   uint64 `json:"flush_units_written"`
	FlushUnitBytes                      uint64 `json:"flush_unit_bytes"`
	FlushPackedSlotsWritten             uint64 `json:"flush_packed_slots_written"`
	FlushPackedBytes                    uint64 `json:"flush_packed_bytes"`
	FlushErrors                         uint64 `json:"flush_errors"`
	FlushWriterPrecheckLivePbaOps       uint64 `json:"flush_writer_precheck_live_pba_ops"`
	FlushWriterPrecheckLivePbaNs        uint64 `json:"flush_writer_precheck_live_pba_ns"`
	FlushWriterPrecheckLivePbaFailures  uint64 `json:"flush_writer_precheck_live_pba_failures"`
	GcCycles                            uint64 `json:"gc_cycles"`
	GcCandidatesFound                   uint64 `json:"gc_candidates_found"`
	GcBlocksRewritten                   uint64 `json:"gc_blocks_rewritten"`
	GcErrors                            uint64 `json:"gc_errors"`
	DedupRescanCycles                   uint64 `json:"dedup_rescan_cycles"`
	DedupRescanHits                     uint64 `json:"dedup_rescan_hits"`
	DedupRescanErrors                   uint64 `json:"dedup_rescan_errors"`
}

// VolumeJSON matches the enriched volumes-json IPC response.
type VolumeJSON struct {
	ID          string           `json:"id"`
	SizeBytes   uint64           `json:"size_bytes"`
	BlockSize   uint32           `json:"block_size"`
	Compression any              `json:"compression"`
	CreatedAt   uint64           `json:"created_at"`
	ZoneCount   uint32           `json:"zone_count"`
	Metrics     *VolumeIOMetrics `json:"metrics"`
}
