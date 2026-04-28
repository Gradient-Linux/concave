// Package lab provides backend-agnostic ephemeral lab environment management.
//
// The package exposes a Driver interface with three planned implementations:
//
//   - docker: reference implementation shipped in this package. Launches
//     containerised JupyterLab environments on the local Docker daemon with a
//     TTL reaper that archives the hot-tier workspace to the cold tier before
//     tearing the container down.
//   - slurm: (future) submits sbatch jobs and relies on Slurm for TTL,
//     fair-share, GPU/GRES accounting, and epilog-driven archival.
//   - proxmox: (future) clones a Proxmox VE template via the REST API and
//     snapshots to the cold tier on teardown.
//
// The web and TUI surfaces consume a single HTTP API
// (`/api/v1/lab/envs`) that is driver-agnostic: the driver is chosen at launch
// time and recorded on the Env record. Storage tiers (hot/cold) are orthogonal
// to driver choice and are configured by the sysadmin via
// `concave lab storage`.
package lab

import "time"

// Status is the lifecycle state of an ephemeral lab environment.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusExpiring  Status = "expiring"
	StatusArchiving Status = "archiving"
	StatusArchived  Status = "archived"
	StatusFailed    Status = "failed"
)

// EnvSpec is the user-facing request to launch an ephemeral environment.
type EnvSpec struct {
	Owner        string        `json:"owner"`
	Image        string        `json:"image"`
	DisplayName  string        `json:"display_name,omitempty"`
	GPUs         int           `json:"gpus,omitempty"`
	CPURequest   string        `json:"cpu_request,omitempty"`
	MemRequest   string        `json:"mem_request,omitempty"`
	TTL          time.Duration `json:"ttl"`
	HotTierPath  string        `json:"hot_tier_path,omitempty"`
	ColdTierPath string        `json:"cold_tier_path,omitempty"`
	Driver       string        `json:"driver,omitempty"`
}

// Env is the persisted record for a running (or archived) environment.
type Env struct {
	ID           string     `json:"id"`
	Driver       string     `json:"driver"`
	Owner        string     `json:"owner"`
	Image        string     `json:"image"`
	DisplayName  string     `json:"display_name,omitempty"`
	Status       Status     `json:"status"`
	ContainerID  string     `json:"container_id,omitempty"`
	JupyterURL   string     `json:"jupyter_url,omitempty"`
	Token        string     `json:"token,omitempty"`
	GPUs         int        `json:"gpus"`
	CPURequest   string     `json:"cpu_request,omitempty"`
	MemRequest   string     `json:"mem_request,omitempty"`
	HotTierPath  string     `json:"hot_tier_path"`
	ColdTierPath string     `json:"cold_tier_path"`
	ArchiveRef   string     `json:"archive_ref,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	ArchivedAt   *time.Time `json:"archived_at,omitempty"`
	LastError    string     `json:"last_error,omitempty"`
}

// Remaining reports the TTL budget left on the env, clamped to zero.
func (e Env) Remaining(now time.Time) time.Duration {
	if e.ExpiresAt.IsZero() {
		return 0
	}
	remaining := e.ExpiresAt.Sub(now)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Expired reports whether the env's TTL has passed.
func (e Env) Expired(now time.Time) bool {
	return !e.ExpiresAt.IsZero() && !now.Before(e.ExpiresAt)
}

// Active reports whether the env is still in a runnable/archivable state.
func (e Env) Active() bool {
	switch e.Status {
	case StatusArchived, StatusFailed:
		return false
	}
	return true
}
