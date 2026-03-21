package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Gradient-Linux/concave/internal/workspace"
)

// ImageVersion stores the active and previous tag for a container image.
type ImageVersion struct {
	Current  string `json:"current"`
	Previous string `json:"previous"`
}

// Versions maps suite names to container image versions.
type Versions map[string]map[string]ImageVersion

// LoadVersions reads ~/gradient/config/versions.json or returns an empty structure when missing.
func LoadVersions() (Versions, error) {
	if err := workspace.EnsureLayout(); err != nil {
		return nil, err
	}

	path := workspace.ConfigPath("versions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Versions{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var versions Versions
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}
	if versions == nil {
		versions = Versions{}
	}
	return versions, nil
}

// SaveVersions writes ~/gradient/config/versions.json atomically.
func SaveVersions(versions Versions) error {
	return writeJSONAtomically(workspace.ConfigPath("versions.json"), versions)
}

// SetImageVersion stores a container's current and previous image tags.
func SetImageVersion(versions Versions, suiteName, containerName, current, previous string) {
	if _, ok := versions[suiteName]; !ok {
		versions[suiteName] = map[string]ImageVersion{}
	}
	versions[suiteName][containerName] = ImageVersion{
		Current:  current,
		Previous: previous,
	}
}

// GetImageVersion fetches the stored version information for a container.
func GetImageVersion(versions Versions, suiteName, containerName string) (ImageVersion, bool) {
	containers, ok := versions[suiteName]
	if !ok {
		return ImageVersion{}, false
	}
	version, ok := containers[containerName]
	return version, ok
}

// RemoveSuiteVersions removes all stored image tags for a suite.
func RemoveSuiteVersions(versions Versions, suiteName string) {
	delete(versions, suiteName)
}

// SwapPrevious swaps current and previous image tags for every container in a suite.
func SwapPrevious(versions Versions, suiteName string) error {
	containers, ok := versions[suiteName]
	if !ok {
		return fmt.Errorf("suite %s has no recorded versions", suiteName)
	}
	for name, version := range containers {
		if version.Previous == "" {
			return fmt.Errorf("suite %s container %s has no previous image", suiteName, name)
		}
		version.Current, version.Previous = version.Previous, version.Current
		containers[name] = version
	}
	return nil
}
