package suite

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/Gradient-Linux/concave/internal/ui"
)

var forgeComponents = []Container{
	{Name: "gradient-boost-core", Image: "python:3.12-slim", Role: "Boosting core"},
	{Name: "gradient-boost-lab", Image: "quay.io/jupyter/base-notebook:python-3.11.6", Role: "Boosting lab"},
	{Name: "gradient-boost-track", Image: "ghcr.io/mlflow/mlflow:2.14", Role: "Boosting MLflow"},
	{Name: "gradient-neural-torch", Image: "pytorch/pytorch:2.6.0-cuda12.4-cudnn9-runtime", Role: "Neural torch"},
	{Name: "gradient-neural-infer", Image: "nvidia/cuda:12.4-runtime-ubuntu24.04", Role: "Neural infer"},
	{Name: "gradient-neural-lab", Image: "quay.io/jupyter/base-notebook:python-3.11.6", Role: "Neural lab"},
	{Name: "gradient-flow-mlflow", Image: "ghcr.io/mlflow/mlflow:2.14", Role: "Flow MLflow"},
	{Name: "gradient-flow-airflow", Image: "apache/airflow:2.9.0", Role: "Flow Airflow"},
	{Name: "gradient-flow-prometheus", Image: "prom/prometheus:v2.51.0", Role: "Flow Prometheus"},
	{Name: "gradient-flow-grafana", Image: "grafana/grafana:10.4.0", Role: "Flow Grafana"},
	{Name: "gradient-flow-store", Image: "minio/minio:RELEASE.2024-04-06T05-26-02Z", Role: "Flow MinIO"},
	{Name: "gradient-flow-serve", Image: "bentoml/bentoml:1.2.0", Role: "Flow BentoML"},
}

// SelectForgeComponents presents a component checklist and returns selected service names.
func SelectForgeComponents() []string {
	items := make([]string, 0, len(forgeComponents))
	for _, component := range forgeComponents {
		items = append(items, component.Name)
	}
	return ui.Checklist(items)
}

// BuildForgeCompose filters the forge template down to selected services and removes disabled profiles.
func BuildForgeCompose(selected []string) ([]byte, error) {
	if len(selected) == 0 {
		return nil, fmt.Errorf("forge requires at least one selected component")
	}

	data, err := readForgeTemplate()
	if err != nil {
		return nil, err
	}

	scanner := strings.Split(string(data), "\n")
	var lines []string
	var block []string
	var service string
	inServices := false
	inNetworks := false

	flushBlock := func() {
		if len(block) == 0 {
			return
		}
		if service == "" || slices.Contains(selected, service) {
			for _, line := range block {
				if strings.Contains(line, "profiles: [\"disabled\"]") {
					continue
				}
				lines = append(lines, line)
			}
		}
		block = nil
		service = ""
	}

	for _, line := range scanner {
		switch {
		case strings.HasPrefix(line, "services:"):
			inServices = true
			lines = append(lines, line)
		case strings.HasPrefix(line, "networks:"):
			flushBlock()
			inNetworks = true
			inServices = false
			lines = append(lines, line)
		case inServices && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") && strings.HasSuffix(strings.TrimSpace(line), ":"):
			flushBlock()
			service = strings.TrimSuffix(strings.TrimSpace(line), ":")
			block = append(block, line)
		case inNetworks:
			lines = append(lines, line)
		case inServices:
			block = append(block, line)
		default:
			lines = append(lines, line)
		}
	}
	flushBlock()

	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

func readForgeTemplate() ([]byte, error) {
	candidates := []string{filepath.Join("templates", "forge.compose.yml")}

	if _, sourceFile, _, ok := runtime.Caller(0); ok {
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "..", ".."))
		candidates = append(candidates, filepath.Join(repoRoot, "templates", "forge.compose.yml"))
	}

	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), "templates", "forge.compose.yml"))
	}

	var failures []string
	for _, candidate := range uniqueStrings(candidates) {
		data, err := os.ReadFile(candidate)
		if err == nil {
			return data, nil
		}
		failures = append(failures, fmt.Sprintf("%s: %v", candidate, err))
	}

	return nil, fmt.Errorf("read forge.compose.yml: %s", strings.Join(failures, "; "))
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
