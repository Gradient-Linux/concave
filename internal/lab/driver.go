package lab

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ArchiveRef identifies a cold-tier archive produced by Driver.Archive.
//
// PeerID is the mesh-addressable identifier of the node that produced the
// archive (machine-id or hostname). Carrying it in the manifest lets a user
// restore an env archived on peer A onto peer B via
// /api/v1/lab/envs/restore.
type ArchiveRef struct {
	EnvID      string    `json:"env_id"`
	Path       string    `json:"path"`
	PeerID     string    `json:"peer_id,omitempty"`
	SizeBytes  int64     `json:"size_bytes"`
	CreatedAt  time.Time `json:"created_at"`
	OriginSpec EnvSpec   `json:"origin_spec"`
}

// Driver abstracts the backend (docker | slurm | proxmox) that actually runs
// an ephemeral environment.
type Driver interface {
	// Name returns the canonical driver name, matching what users pass in
	// EnvSpec.Driver.
	Name() string

	// OwnsTTL reports whether the driver itself is responsible for enforcing
	// TTL (e.g. Slurm's `--time` flag). Drivers that return false rely on the
	// reaper to expire environments.
	OwnsTTL() bool

	// Launch starts a new env from spec and returns the populated Env record.
	Launch(ctx context.Context, spec EnvSpec) (Env, error)

	// Inspect returns the current state of an env.
	Inspect(ctx context.Context, env Env) (Env, error)

	// ExtendTTL pushes an env's expiry to the supplied time. Drivers that
	// own TTL must relay the extension to the underlying scheduler.
	ExtendTTL(ctx context.Context, env Env, until time.Time) (Env, error)

	// Archive copies the env's hot-tier state to the cold tier and returns a
	// reference that can be used to restore it later.
	Archive(ctx context.Context, env Env) (ArchiveRef, error)

	// Destroy tears down the running env. Storage is not touched; callers are
	// expected to call Archive first if they want the state preserved.
	Destroy(ctx context.Context, env Env) error
}

// ErrDriverNotRegistered is returned by the registry when a lookup fails.
var ErrDriverNotRegistered = errors.New("lab driver not registered")

// Registry is a process-wide map of driver name → Driver.
type Registry struct {
	mu      sync.RWMutex
	drivers map[string]Driver
	active  string
}

// NewRegistry creates an empty driver registry.
func NewRegistry() *Registry {
	return &Registry{drivers: make(map[string]Driver)}
}

// Register installs a driver. The first registered driver becomes the active
// one unless SetActive is called afterwards.
func (r *Registry) Register(d Driver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drivers[d.Name()] = d
	if r.active == "" {
		r.active = d.Name()
	}
}

// SetActive selects the default driver used when an EnvSpec does not specify
// one. Returns ErrDriverNotRegistered if the driver is unknown.
func (r *Registry) SetActive(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.drivers[name]; !ok {
		return fmt.Errorf("%w: %q", ErrDriverNotRegistered, name)
	}
	r.active = name
	return nil
}

// Active returns the currently selected default driver name.
func (r *Registry) Active() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.active
}

// Get resolves a driver by name. An empty name returns the active driver.
func (r *Registry) Get(name string) (Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if name == "" {
		name = r.active
	}
	if name == "" {
		return nil, ErrDriverNotRegistered
	}
	d, ok := r.drivers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrDriverNotRegistered, name)
	}
	return d, nil
}

// Names returns the sorted list of registered driver names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.drivers))
	for name := range r.drivers {
		names = append(names, name)
	}
	return names
}
