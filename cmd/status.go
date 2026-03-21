package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show container status for installed suites",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	state, err := loadState()
	if err != nil {
		return err
	}

	ui.Line("Gradient Linux — Suite Status")
	ui.Line("─────────────────────────────────────────────────────────")
	ui.Line(fmt.Sprintf("%-11s %-26s %-10s %s", "Suite", "Container", "Status", "Port"))
	ui.Line("─────────────────────────────────────────────────────────")

	if len(state.Installed) == 0 {
		ui.Line("not installed")
		ui.Line("─────────────────────────────────────────────────────────")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	for _, name := range orderInstalledSuites(state.Installed, false) {
		s, err := currentSuiteDefinition(name)
		if err != nil {
			return err
		}
		for idx, container := range s.Containers {
			status, err := dockerContainerStatus(ctx, container.Name)
			if err != nil {
				status = "error"
			}
			suiteColumn := ""
			if idx == 0 {
				suiteColumn = name
			}
			ui.Line(fmt.Sprintf("%-11s %-26s %-10s %s", suiteColumn, container.Name, status, containerPortSummary(s, container)))
		}
	}

	ui.Line("─────────────────────────────────────────────────────────")
	if gpuLine, ok := currentGPULine(); ok {
		ui.Line(gpuLine)
	}
	ui.Line(currentWorkspaceLine())
	return nil
}

func containerPortSummary(s suite.Suite, container suite.Container) string {
	ports := make([]string, 0, 2)
	for _, mapping := range s.Ports {
		switch {
		case strings.Contains(container.Name, "lab") && mapping.Service == "JupyterLab":
			ports = append(ports, fmt.Sprintf("%d", mapping.Port))
		case strings.Contains(container.Name, "track") && mapping.Service == "MLflow":
			ports = append(ports, fmt.Sprintf("%d", mapping.Port))
		case strings.Contains(container.Name, "mlflow") && mapping.Service == "MLflow":
			ports = append(ports, fmt.Sprintf("%d", mapping.Port))
		case strings.Contains(container.Name, "infer") && (mapping.Service == "vLLM API" || mapping.Service == "llama.cpp"):
			ports = append(ports, fmt.Sprintf("%d", mapping.Port))
		case strings.Contains(container.Name, "airflow") && mapping.Service == "Airflow":
			ports = append(ports, fmt.Sprintf("%d", mapping.Port))
		case strings.Contains(container.Name, "prometheus") && mapping.Service == "Prometheus":
			ports = append(ports, fmt.Sprintf("%d", mapping.Port))
		case strings.Contains(container.Name, "grafana") && mapping.Service == "Grafana":
			ports = append(ports, fmt.Sprintf("%d", mapping.Port))
		case strings.Contains(container.Name, "store") && mapping.Service == "MinIO console":
			ports = append(ports, fmt.Sprintf("%d", mapping.Port))
		case strings.Contains(container.Name, "serve") && mapping.Service == "BentoML endpoint":
			ports = append(ports, fmt.Sprintf("%d", mapping.Port))
		}
	}
	if len(ports) == 0 {
		return "—"
	}
	return strings.Join(ports, ",")
}

func currentGPULine() (string, bool) {
	state, err := gpuDetectState()
	if err != nil {
		return "", false
	}
	if state == gpu.GPUStateAMD {
		return "GPU         AMD detected", true
	}
	if state != gpu.GPUStateNVIDIA {
		return "", false
	}

	out, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader").CombinedOutput()
	if err != nil {
		return "GPU         NVIDIA detected", true
	}

	fields := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(fields) < 2 {
		return "GPU         NVIDIA detected", true
	}
	return fmt.Sprintf("GPU         %s %s", strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])), true
}

func currentWorkspaceLine() string {
	out, err := exec.Command("df", "-h", workspaceRoot()).CombinedOutput()
	if err != nil {
		return "Workspace   " + workspaceRoot()
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return "Workspace   " + workspaceRoot()
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return "Workspace   " + workspaceRoot()
	}
	return fmt.Sprintf("Workspace   %s %s free", workspaceRoot(), fields[3])
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
