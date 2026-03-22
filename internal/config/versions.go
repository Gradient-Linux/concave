package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Gradient-Linux/concave/internal/logx"
	"github.com/Gradient-Linux/concave/internal/workspace"
)

// ImageVersion stores the active and previous tag for a container image.
type ImageVersion struct {
	Current  string `json:"current"`
	Previous string `json:"previous"`
}

// VersionManifest tracks image versions by suite and container.
type VersionManifest map[string]map[string]ImageVersion

type installRecord interface {
	RecordName() string
	RecordImages() map[string]string
}

// LoadManifest reads ~/gradient/config/versions.json or returns an empty manifest when missing.
func LoadManifest() (VersionManifest, error) {
	if err := workspace.EnsureLayout(); err != nil {
		return nil, err
	}

	path := workspace.ConfigPath("versions.json")
	logx.Debug("manifest load", "path", path)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return VersionManifest{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var manifest VersionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}
	if manifest == nil {
		manifest = VersionManifest{}
	}
	return manifest, nil
}

// SaveManifest writes ~/gradient/config/versions.json atomically.
func SaveManifest(manifest VersionManifest) error {
	logx.Debug("manifest save", "path", workspace.ConfigPath("versions.json"))
	return writeJSONAtomically(workspace.ConfigPath("versions.json"), manifest)
}

// RecordUpdate moves current to previous and stores the requested image as current.
func RecordUpdate(manifest VersionManifest, suiteName, containerName, newImage string) VersionManifest {
	manifest = ensureManifest(manifest)
	logx.Debug("manifest record update", "suite", suiteName, "container", containerName, "image", newImage)
	current := ImageVersion{}
	if containers, ok := manifest[suiteName]; ok {
		current = containers[containerName]
	}
	if _, ok := manifest[suiteName]; !ok {
		manifest[suiteName] = map[string]ImageVersion{}
	}
	manifest[suiteName][containerName] = ImageVersion{
		Current:  newImage,
		Previous: current.Current,
	}
	return manifest
}

// RecordInstall initialises any missing manifest entries for a suite.
func RecordInstall(manifest VersionManifest, s installRecord) VersionManifest {
	manifest = ensureManifest(manifest)
	name := s.RecordName()
	logx.Debug("manifest record install", "suite", name)
	if _, ok := manifest[name]; !ok {
		manifest[name] = map[string]ImageVersion{}
	}
	for containerName, image := range s.RecordImages() {
		if _, exists := manifest[name][containerName]; exists {
			continue
		}
		manifest[name][containerName] = ImageVersion{
			Current:  image,
			Previous: "",
		}
	}
	return manifest
}

// SwapForRollback swaps current and previous image tags for every container in a suite.
func SwapForRollback(manifest VersionManifest, suiteName string) (VersionManifest, error) {
	manifest = ensureManifest(manifest)
	logx.Debug("manifest swap rollback", "suite", suiteName)
	containers, ok := manifest[suiteName]
	if !ok {
		return manifest, fmt.Errorf("nothing to roll back for suite %s", suiteName)
	}
	for name, version := range containers {
		if version.Previous == "" {
			return manifest, fmt.Errorf("no previous version for container %s — run concave update first", name)
		}
		version.Current, version.Previous = version.Previous, version.Current
		containers[name] = version
	}
	return manifest, nil
}

func ensureManifest(manifest VersionManifest) VersionManifest {
	if manifest == nil {
		return VersionManifest{}
	}
	return manifest
}
