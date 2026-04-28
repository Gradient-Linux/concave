package lab

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Gradient-Linux/concave/internal/workspace"
)

// StorageConfig describes the two storage tiers every ephemeral env relies on.
//
// HotTier is where the live environment's scratch workspace lives — typically
// the primary workspace disk (NVMe/SSD). It is short-lived: created at env
// launch, destroyed when the env is destroyed.
//
// ColdTier is the sysadmin-configurable destination for archival tarballs. It
// defaults to the slowest writable non-root block device detected at setup
// time (typically an HDD) and can be reassigned with `concave lab storage
// set-cold <path>`.
type StorageConfig struct {
	HotTier  string `json:"hot_tier"`
	ColdTier string `json:"cold_tier"`
}

// Default returns the default StorageConfig used when no lab.json is found.
func defaultStorage() StorageConfig {
	return StorageConfig{
		HotTier:  filepath.Join(workspace.Root(), "envs"),
		ColdTier: filepath.Join(workspace.Root(), "envs-archive"),
	}
}

var (
	storageMu sync.Mutex
	// storagePath is overridable for tests.
	storagePath = func() string {
		return workspace.ConfigPath("lab.json")
	}
)

// LoadStorage reads the persisted storage config, falling back to defaults
// when the file is missing.
func LoadStorage() (StorageConfig, error) {
	storageMu.Lock()
	defer storageMu.Unlock()

	cfg := defaultStorage()
	data, err := os.ReadFile(storagePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read lab storage: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultStorage(), fmt.Errorf("parse lab storage: %w", err)
	}
	if cfg.HotTier == "" {
		cfg.HotTier = defaultStorage().HotTier
	}
	if cfg.ColdTier == "" {
		cfg.ColdTier = defaultStorage().ColdTier
	}
	return cfg, nil
}

// SaveStorage persists the storage config atomically.
func SaveStorage(cfg StorageConfig) error {
	if err := validateTierPath("hot_tier", cfg.HotTier); err != nil {
		return err
	}
	if err := validateTierPath("cold_tier", cfg.ColdTier); err != nil {
		return err
	}
	if cfg.HotTier == cfg.ColdTier {
		return errors.New("hot_tier and cold_tier must be different paths")
	}

	storageMu.Lock()
	defer storageMu.Unlock()

	path := storagePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("ensure lab config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode lab storage: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write lab storage: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename lab storage: %w", err)
	}
	return nil
}

// EnsureTierDirs ensures both storage tier directories exist.
func EnsureTierDirs(cfg StorageConfig) error {
	for _, dir := range []string{cfg.HotTier, cfg.ColdTier} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("ensure tier dir %s: %w", dir, err)
		}
	}
	return nil
}

func validateTierPath(label, path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("%s is required", label)
	}
	if !filepath.IsAbs(path) {
		return fmt.Errorf("%s must be an absolute path", label)
	}
	return nil
}
