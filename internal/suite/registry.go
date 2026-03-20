package suite

import (
	"fmt"
	"sort"
)

// Container describes a concrete container in a suite.
type Container struct {
	Name  string
	Image string
	Role  string
}

// PortMapping describes a published suite port.
type PortMapping struct {
	Port    int
	Service string
}

// VolumeMount describes a workspace-backed mount.
type VolumeMount struct {
	HostPath      string
	ContainerPath string
}

// Suite defines a named collection of containers, ports, and mounts.
type Suite struct {
	Name            string
	Containers      []Container
	Ports           []PortMapping
	Volumes         []VolumeMount
	ComposeTemplate string
	GPURequired     bool
}

var registry = map[string]Suite{
	"boosting": {
		Name: "boosting",
		Containers: []Container{
			{Name: "gradient-boost-core", Image: "python:3.12-slim", Role: "Core ML"},
			{Name: "gradient-boost-lab", Image: "quay.io/jupyter/base-notebook:python-3.11.6", Role: "JupyterLab"},
			{Name: "gradient-boost-track", Image: "ghcr.io/mlflow/mlflow:2.14", Role: "MLflow"},
		},
		Ports: []PortMapping{
			{Port: 8888, Service: "JupyterLab"},
			{Port: 5000, Service: "MLflow"},
		},
		Volumes: []VolumeMount{
			{HostPath: "data", ContainerPath: "/data"},
			{HostPath: "notebooks", ContainerPath: "/notebooks"},
			{HostPath: "models", ContainerPath: "/models"},
			{HostPath: "outputs", ContainerPath: "/outputs"},
			{HostPath: "mlruns", ContainerPath: "/mlruns"},
		},
		ComposeTemplate: "boosting",
		GPURequired:     false,
	},
	"neural": {
		Name: "neural",
		Containers: []Container{
			{Name: "gradient-neural-torch", Image: "pytorch/pytorch:2.6.0-cuda12.4-cudnn9-runtime", Role: "Training"},
			{Name: "gradient-neural-infer", Image: "nvidia/cuda:12.4-runtime-ubuntu24.04", Role: "Inference"},
			{Name: "gradient-neural-lab", Image: "quay.io/jupyter/base-notebook:python-3.11.6", Role: "JupyterLab"},
		},
		Ports: []PortMapping{
			{Port: 8000, Service: "vLLM API"},
			{Port: 8080, Service: "llama.cpp / Airflow"},
			{Port: 8888, Service: "JupyterLab"},
		},
		Volumes: []VolumeMount{
			{HostPath: "data", ContainerPath: "/data"},
			{HostPath: "notebooks", ContainerPath: "/notebooks"},
			{HostPath: "models", ContainerPath: "/models"},
			{HostPath: "outputs", ContainerPath: "/outputs"},
		},
		ComposeTemplate: "neural",
		GPURequired:     true,
	},
	"flow": {
		Name: "flow",
		Containers: []Container{
			{Name: "gradient-flow-mlflow", Image: "ghcr.io/mlflow/mlflow:2.14", Role: "Tracking"},
			{Name: "gradient-flow-airflow", Image: "apache/airflow:2.9.0", Role: "Orchestration"},
			{Name: "gradient-flow-prometheus", Image: "prom/prometheus:v2.51.0", Role: "Metrics"},
			{Name: "gradient-flow-grafana", Image: "grafana/grafana:10.4.0", Role: "Dashboards"},
			{Name: "gradient-flow-store", Image: "minio/minio:RELEASE.2024-04-06T05-26-02Z", Role: "Artifacts"},
			{Name: "gradient-flow-serve", Image: "bentoml/bentoml:1.2.0", Role: "Serving"},
		},
		Ports: []PortMapping{
			{Port: 5000, Service: "MLflow"},
			{Port: 8080, Service: "Airflow"},
			{Port: 9090, Service: "Prometheus"},
			{Port: 3000, Service: "Grafana"},
			{Port: 9001, Service: "MinIO console"},
			{Port: 3100, Service: "BentoML"},
		},
		Volumes: []VolumeMount{
			{HostPath: "mlruns", ContainerPath: "/mlruns"},
			{HostPath: "dags", ContainerPath: "/dags"},
			{HostPath: "outputs", ContainerPath: "/outputs"},
			{HostPath: "models", ContainerPath: "/models"},
		},
		ComposeTemplate: "flow",
		GPURequired:     false,
	},
	"forge": {
		Name:       "forge",
		Containers: []Container{},
		Ports: []PortMapping{
			{Port: 8888, Service: "JupyterLab"},
			{Port: 8000, Service: "vLLM API"},
			{Port: 8080, Service: "llama.cpp / Airflow"},
			{Port: 5000, Service: "MLflow"},
			{Port: 3000, Service: "Grafana"},
			{Port: 9090, Service: "Prometheus"},
			{Port: 9001, Service: "MinIO console"},
			{Port: 3100, Service: "BentoML"},
		},
		Volumes: []VolumeMount{
			{HostPath: "data", ContainerPath: "/data"},
			{HostPath: "notebooks", ContainerPath: "/notebooks"},
			{HostPath: "models", ContainerPath: "/models"},
			{HostPath: "outputs", ContainerPath: "/outputs"},
			{HostPath: "mlruns", ContainerPath: "/mlruns"},
			{HostPath: "dags", ContainerPath: "/dags"},
		},
		ComposeTemplate: "forge",
		GPURequired:     false,
	},
}

// All returns all known suites in stable order.
func All() []Suite {
	names := Names()
	suites := make([]Suite, 0, len(names))
	for _, name := range names {
		suites = append(suites, registry[name])
	}
	return suites
}

// Names returns suite names in stable order.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Get returns a suite by name.
func Get(name string) (Suite, error) {
	s, ok := registry[name]
	if !ok {
		return Suite{}, fmt.Errorf("unknown suite %q", name)
	}
	return s, nil
}
