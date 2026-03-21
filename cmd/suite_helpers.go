package cmd

import (
	"fmt"
	"sort"

	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/suite"
)

var suiteOrder = []string{"boosting", "neural", "flow", "forge"}

func installedSuiteTargets(args []string, reverse bool) ([]string, error) {
	if len(args) == 1 {
		if _, err := getSuite(args[0]); err != nil {
			return nil, err
		}
		installed, err := isInstalled(args[0])
		if err != nil {
			return nil, err
		}
		if !installed {
			return nil, fmt.Errorf("suite %s is not installed", args[0])
		}
		return []string{args[0]}, nil
	}

	state, err := loadState()
	if err != nil {
		return nil, err
	}
	return orderInstalledSuites(state.Installed, reverse), nil
}

func orderInstalledSuites(installed []string, reverse bool) []string {
	set := make(map[string]struct{}, len(installed))
	for _, name := range installed {
		set[name] = struct{}{}
	}

	ordered := make([]string, 0, len(installed))
	for _, name := range suiteOrder {
		if _, ok := set[name]; ok {
			ordered = append(ordered, name)
		}
	}
	if !reverse {
		return ordered
	}
	for left, right := 0, len(ordered)-1; left < right; left, right = left+1, right-1 {
		ordered[left], ordered[right] = ordered[right], ordered[left]
	}
	return ordered
}

func currentSuiteDefinition(name string) (suite.Suite, error) {
	s, err := getSuite(name)
	if err != nil {
		return suite.Suite{}, err
	}
	if name != "forge" {
		return s, nil
	}

	manifest, err := loadManifest()
	if err != nil {
		return suite.Suite{}, err
	}
	forgeEntries, ok := manifest[name]
	if !ok || len(forgeEntries) == 0 {
		return suite.Suite{}, fmt.Errorf("forge has no recorded component selection")
	}

	names := make([]string, 0, len(forgeEntries))
	overrides := make(map[string]string, len(forgeEntries))
	for containerName, version := range forgeEntries {
		names = append(names, containerName)
		overrides[containerName] = version.Current
	}
	sort.Strings(names)

	selection, err := forgeSelectionFromNames(names, overrides)
	if err != nil {
		return suite.Suite{}, err
	}
	s.Containers = selection.Containers
	s.Ports = selection.Ports
	s.Volumes = selection.Volumes
	return s, nil
}

func writeComposeForCurrentState(name string) (string, error) {
	if name != "forge" {
		return dockerWriteCompose(name)
	}

	s, err := currentSuiteDefinition(name)
	if err != nil {
		return "", err
	}
	data, err := buildForgeCompose(suite.ForgeSelection{
		Containers: s.Containers,
		Ports:      s.Ports,
		Volumes:    s.Volumes,
	})
	if err != nil {
		return "", err
	}
	return dockerWriteRawCompose(name, data)
}

func currentImageForFirstContainer(name string, manifest config.VersionManifest) string {
	s, err := getSuite(name)
	if err != nil || len(s.Containers) == 0 {
		return ""
	}
	if containers, ok := manifest[name]; ok {
		if version, ok := containers[s.Containers[0].Name]; ok && version.Current != "" {
			return version.Current
		}
	}
	return s.Containers[0].Image
}

func containerNames(s suite.Suite) []string {
	names := make([]string, 0, len(s.Containers))
	for _, container := range s.Containers {
		names = append(names, container.Name)
	}
	return names
}
