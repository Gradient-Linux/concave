package compute

import "time"

// ResourceQuota defines the resource allocation for a user or group.
type ResourceQuota struct {
	CPUCores    float64
	MemoryGB    float64
	GPUFraction float64
	GPUMemoryGB float64
	IOWeightPct int
}

// GPUAllocStrategy is how GPU capacity is divided across users.
type GPUAllocStrategy string

const (
	GPUAllocMIG         GPUAllocStrategy = "mig"
	GPUAllocTimeSlicing GPUAllocStrategy = "time-slicing"
	GPUAllocNone        GPUAllocStrategy = "none"
)

// Group defines a team resource allocation.
type Group struct {
	Name        string
	Preset      string
	Users       []string
	Quota       ResourceQuota
	GPUStrategy GPUAllocStrategy
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Presets maps preset names to resource quotas.
var Presets = map[string]ResourceQuota{
	"research-team":  {CPUCores: 4, MemoryGB: 32, GPUFraction: 0, IOWeightPct: 100},
	"inference-node": {CPUCores: 2, MemoryGB: 16, GPUFraction: 1, IOWeightPct: 100},
	"training-node":  {CPUCores: 8, MemoryGB: 64, GPUFraction: 1, IOWeightPct: 100},
	"student-lab":    {CPUCores: 2, MemoryGB: 8, GPUFraction: 0, GPUMemoryGB: 0, IOWeightPct: 100},
}
