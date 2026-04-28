package lab

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Manager is the high-level façade combining a Registry, a Store, and a
// StorageConfig. HTTP handlers and CLI commands interact with the Manager
// rather than the underlying building blocks directly.
type Manager struct {
	registry *Registry
	store    *Store
	storage  StorageConfig
	audit    AuditWriter
	nowFn    func() time.Time

	mu sync.Mutex
}

// NewManager constructs a Manager. Use LoadStorage() to populate the storage
// tier defaults before calling this.
func NewManager(registry *Registry, store *Store, storage StorageConfig) *Manager {
	return &Manager{
		registry: registry,
		store:    store,
		storage:  storage,
		audit:    DiscardAuditWriter{},
		nowFn:    time.Now,
	}
}

// SetAuditWriter installs an audit sink. Defaults to DiscardAuditWriter.
func (m *Manager) SetAuditWriter(w AuditWriter) {
	if w == nil {
		w = DiscardAuditWriter{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.audit = w
}

func (m *Manager) auditWriter() AuditWriter {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.audit
}

func (m *Manager) record(event AuditEvent) {
	if event.Time.IsZero() {
		event.Time = m.nowFn()
	}
	if event.PeerID == "" {
		event.PeerID = localPeerID()
	}
	_ = m.auditWriter().Write(event)
}

// SetNowFn is a test seam.
func (m *Manager) SetNowFn(fn func() time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nowFn = fn
}

// Storage returns a copy of the current storage configuration.
func (m *Manager) Storage() StorageConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.storage
}

// UpdateStorage persists a new storage configuration.
func (m *Manager) UpdateStorage(next StorageConfig) error {
	if err := SaveStorage(next); err != nil {
		return err
	}
	if err := EnsureTierDirs(next); err != nil {
		return err
	}
	m.mu.Lock()
	m.storage = next
	m.mu.Unlock()
	return nil
}

// Registry exposes the underlying registry for driver-aware CLI commands.
func (m *Manager) Registry() *Registry { return m.registry }

// Launch provisions a new env via the selected driver and persists the record.
func (m *Manager) Launch(ctx context.Context, spec EnvSpec) (Env, error) {
	m.mu.Lock()
	storage := m.storage
	m.mu.Unlock()

	if spec.HotTierPath == "" {
		spec.HotTierPath = storage.HotTier
	}
	if spec.ColdTierPath == "" {
		spec.ColdTierPath = storage.ColdTier
	}
	if spec.Driver == "" {
		spec.Driver = m.registry.Active()
	}
	driver, err := m.registry.Get(spec.Driver)
	if err != nil {
		return Env{}, err
	}
	env, err := driver.Launch(ctx, spec)
	if err != nil {
		m.record(AuditEvent{Actor: spec.Owner, Action: "lab.launch", Driver: spec.Driver, Error: err.Error()})
		return Env{}, err
	}
	created, err := m.store.Create(env)
	if err != nil {
		m.record(AuditEvent{Actor: spec.Owner, Action: "lab.launch", Driver: spec.Driver, Error: err.Error()})
		return Env{}, err
	}
	m.record(AuditEvent{
		Actor:  created.Owner,
		Action: "lab.launch",
		EnvID:  created.ID,
		Driver: created.Driver,
		Details: map[string]any{
			"image": created.Image,
			"ttl":   spec.TTL.String(),
		},
	})
	return created, nil
}

// Get returns a record by ID with a driver-level inspect refresh.
func (m *Manager) Get(ctx context.Context, id string) (Env, error) {
	env, ok := m.store.Get(id)
	if !ok {
		return Env{}, fmt.Errorf("lab env %q not found", id)
	}
	driver, err := m.registry.Get(env.Driver)
	if err != nil {
		return env, nil
	}
	refreshed, err := driver.Inspect(ctx, env)
	if err != nil {
		return env, nil
	}
	if refreshed.Status != env.Status || refreshed.LastError != env.LastError {
		_ = m.store.Update(refreshed)
	}
	return refreshed, nil
}

// List returns all env records without a driver refresh.
func (m *Manager) List() []Env {
	return m.store.List()
}

// ExtendTTL pushes the expiry of an existing env out.
func (m *Manager) ExtendTTL(ctx context.Context, id string, extension time.Duration) (Env, error) {
	if extension <= 0 {
		return Env{}, errors.New("extension must be positive")
	}
	env, ok := m.store.Get(id)
	if !ok {
		return Env{}, fmt.Errorf("lab env %q not found", id)
	}
	if !env.Active() {
		return Env{}, errors.New("cannot extend a non-active env")
	}
	driver, err := m.registry.Get(env.Driver)
	if err != nil {
		return Env{}, err
	}
	now := m.nowFn()
	base := env.ExpiresAt
	if base.Before(now) {
		base = now
	}
	refreshed, err := driver.ExtendTTL(ctx, env, base.Add(extension))
	if err != nil {
		return Env{}, err
	}
	if err := m.store.Update(refreshed); err != nil {
		return Env{}, err
	}
	m.record(AuditEvent{
		Actor:   env.Owner,
		Action:  "lab.extend",
		EnvID:   env.ID,
		Driver:  env.Driver,
		Details: map[string]any{"extend": extension.String(), "new_expiry": refreshed.ExpiresAt},
	})
	return refreshed, nil
}

// ArchiveAndDestroy archives the env's hot tier to the cold tier, then tears
// the env down. Used by both the reaper and explicit admin actions.
func (m *Manager) ArchiveAndDestroy(ctx context.Context, id string) (Env, error) {
	env, ok := m.store.Get(id)
	if !ok {
		return Env{}, fmt.Errorf("lab env %q not found", id)
	}
	driver, err := m.registry.Get(env.Driver)
	if err != nil {
		return env, err
	}
	env.Status = StatusArchiving
	_ = m.store.Update(env)

	ref, err := driver.Archive(ctx, env)
	if err != nil {
		env.Status = StatusFailed
		env.LastError = "archive: " + err.Error()
		_ = m.store.Update(env)
		return env, err
	}
	if err := driver.Destroy(ctx, env); err != nil {
		env.LastError = "destroy: " + err.Error()
		_ = m.store.Update(env)
		return env, err
	}
	now := m.nowFn()
	env.Status = StatusArchived
	env.ArchiveRef = ref.Path
	env.ArchivedAt = &now
	if err := m.store.Update(env); err != nil {
		return env, err
	}
	m.record(AuditEvent{
		Actor:  env.Owner,
		Action: "lab.archive",
		EnvID:  env.ID,
		Driver: env.Driver,
		PeerID: ref.PeerID,
		Details: map[string]any{
			"archive_ref": ref.Path,
			"size_bytes":  ref.SizeBytes,
		},
	})
	return env, nil
}

// Destroy removes an env without archiving it. Admin-only.
func (m *Manager) Destroy(ctx context.Context, id string) error {
	env, ok := m.store.Get(id)
	if !ok {
		return nil
	}
	driver, err := m.registry.Get(env.Driver)
	if err != nil {
		return err
	}
	if err := driver.Destroy(ctx, env); err != nil {
		return err
	}
	return m.store.Delete(id)
}

// ReapExpired walks the store and archives every expired, active env whose
// driver does not own TTL. Intended to be called on a ticker from
// `concave serve`.
func (m *Manager) ReapExpired(ctx context.Context) []error {
	now := m.nowFn()
	var errs []error
	for _, env := range m.store.List() {
		if !env.Active() || !env.Expired(now) {
			continue
		}
		driver, err := m.registry.Get(env.Driver)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if driver.OwnsTTL() {
			continue
		}
		if _, err := m.ArchiveAndDestroy(ctx, env.ID); err != nil {
			errs = append(errs, fmt.Errorf("reap %s: %w", env.ID, err))
		}
	}
	return errs
}

// RunReaper blocks until ctx is done, calling ReapExpired on each tick.
func (m *Manager) RunReaper(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = m.ReapExpired(ctx)
		}
	}
}
