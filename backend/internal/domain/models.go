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
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
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
	Name        string `json:"name"`
	SizeBytes   uint64 `json:"sizeBytes"`
	ZoneCount   uint32 `json:"zoneCount"`
	Compression string `json:"compression"`
	Status      string `json:"status"`
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
	BlockDevices   []DeviceSummary `json:"blockDevices"`
	DMTargets      []DMTarget      `json:"dmTargets"`
	LogicalVolumes []LvmVolume     `json:"logicalVolumes"`
	Warnings       []string        `json:"warnings"`
}

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

type Overview struct {
	EngineMode           string         `json:"engineMode"`
	EngineRunning        bool           `json:"engineRunning"`
	VolumeCount          int            `json:"volumeCount"`
	LiveHandleCount      int            `json:"liveHandleCount"`
	ZoneCount            int            `json:"zoneCount"`
	BufferFillPercent    int            `json:"bufferFillPercent"`
	BufferPendingEntries uint64         `json:"bufferPendingEntries"`
	AllocatorFreeBlocks  uint64         `json:"allocatorFreeBlocks"`
	AllocatorTotalBlocks uint64         `json:"allocatorTotalBlocks"`
	Metrics              map[string]any `json:"metrics"`
	RawStatus            string         `json:"rawStatus"`
}
