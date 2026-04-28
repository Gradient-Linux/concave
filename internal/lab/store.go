package lab

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/Gradient-Linux/concave/internal/workspace"
)

// Store persists env records on disk under ~/gradient/config/lab-envs.json.
//
// The on-disk representation is deliberately append-update-delete style JSON:
// concurrent writes are serialised through the Store mutex and snapshots are
// rewritten atomically via `*.tmp` + rename. This is good enough for a
// single-node control plane and avoids taking on sqlite as a dependency.
type Store struct {
	mu   sync.Mutex
	path string
	envs map[string]Env
}

// defaultEnvStorePath returns the default on-disk path. Overridable in tests.
var envStorePath = func() string {
	return workspace.ConfigPath("lab-envs.json")
}

// NewStore creates a Store rooted at the default path and loads any existing
// records.
func NewStore() (*Store, error) {
	return NewStoreAt(envStorePath())
}

// NewStoreAt is a test seam for targeting an arbitrary path.
func NewStoreAt(path string) (*Store, error) {
	s := &Store{path: path, envs: make(map[string]Env)}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// Create assigns a new ID and persists the record. CreatedAt / ExpiresAt on
// the supplied env are preserved.
func (s *Store) Create(env Env) (Env, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if env.ID == "" {
		id, err := newEnvID()
		if err != nil {
			return Env{}, err
		}
		env.ID = id
	}
	if _, exists := s.envs[env.ID]; exists {
		return Env{}, fmt.Errorf("lab env %q already exists", env.ID)
	}
	s.envs[env.ID] = env
	if err := s.persistLocked(); err != nil {
		delete(s.envs, env.ID)
		return Env{}, err
	}
	return env, nil
}

// Update replaces an existing record in place.
func (s *Store) Update(env Env) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.envs[env.ID]; !exists {
		return fmt.Errorf("lab env %q not found", env.ID)
	}
	prev := s.envs[env.ID]
	s.envs[env.ID] = env
	if err := s.persistLocked(); err != nil {
		s.envs[env.ID] = prev
		return err
	}
	return nil
}

// Delete removes a record and persists the snapshot.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.envs[id]; !exists {
		return nil
	}
	prev := s.envs[id]
	delete(s.envs, id)
	if err := s.persistLocked(); err != nil {
		s.envs[id] = prev
		return err
	}
	return nil
}

// Get returns a record by ID.
func (s *Store) Get(id string) (Env, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	env, ok := s.envs[id]
	return env, ok
}

// List returns a deterministic snapshot of all records.
func (s *Store) List() []Env {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Env, 0, len(s.envs))
	for _, env := range s.envs {
		out = append(out, env)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read lab env store: %w", err)
	}
	var envs []Env
	if err := json.Unmarshal(data, &envs); err != nil {
		return fmt.Errorf("parse lab env store: %w", err)
	}
	for _, env := range envs {
		s.envs[env.ID] = env
	}
	return nil
}

func (s *Store) persistLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("ensure lab env dir: %w", err)
	}
	envs := make([]Env, 0, len(s.envs))
	for _, env := range s.envs {
		envs = append(envs, env)
	}
	sort.Slice(envs, func(i, j int) bool {
		return envs[i].CreatedAt.Before(envs[j].CreatedAt)
	})
	data, err := json.MarshalIndent(envs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode lab env store: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write lab env store: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("rename lab env store: %w", err)
	}
	return nil
}

func newEnvID() (string, error) {
	var buf [6]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate env id: %w", err)
	}
	return "env-" + hex.EncodeToString(buf[:]), nil
}
