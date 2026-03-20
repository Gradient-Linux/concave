package suite

import (
	"fmt"
	"sort"
	"strings"
)

// InstallPlan summarizes the container images and runtime entry points for a suite.
type InstallPlan struct {
	Suite            Suite
	Images           []string
	PrimaryContainer string
	JupyterContainer string
}

// BuildInstallPlan builds a lightweight install plan from the registry definition.
func BuildInstallPlan(name string) (InstallPlan, error) {
	s, err := Get(name)
	if err != nil {
		return InstallPlan{}, err
	}

	images := ImageList(s)
	primary := PrimaryContainer(s)
	jupyter, _ := JupyterContainer(s)

	return InstallPlan{
		Suite:            s,
		Images:           images,
		PrimaryContainer: primary,
		JupyterContainer: jupyter,
	}, nil
}

// ImageList returns the suite's container images in registry order.
func ImageList(s Suite) []string {
	images := make([]string, 0, len(s.Containers))
	for _, container := range s.Containers {
		images = append(images, container.Image)
	}
	return images
}

// ContainerNames returns the suite's container names in registry order.
func ContainerNames(s Suite) []string {
	names := make([]string, 0, len(s.Containers))
	for _, container := range s.Containers {
		names = append(names, container.Name)
	}
	return names
}

// PrimaryContainer returns the first container in a suite.
func PrimaryContainer(s Suite) string {
	if len(s.Containers) == 0 {
		return ""
	}
	return s.Containers[0].Name
}

// JupyterContainer returns the suite container that exposes JupyterLab, if any.
func JupyterContainer(s Suite) (string, bool) {
	for _, container := range s.Containers {
		if strings.Contains(strings.ToLower(container.Role), "jupyter") {
			return container.Name, true
		}
	}
	return "", false
}

// SuitePorts returns a stable string representation of suite ports.
func SuitePorts(s Suite) string {
	ports := make([]int, 0, len(s.Ports))
	for _, port := range s.Ports {
		ports = append(ports, port.Port)
	}
	sort.Ints(ports)

	parts := make([]string, 0, len(ports))
	for _, port := range ports {
		parts = append(parts, fmt.Sprintf("%d", port))
	}
	return strings.Join(parts, ", ")
}
