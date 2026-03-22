package suite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/docker"
)

// HealthResult holds the runtime status for a single suite container.
type HealthResult struct {
	Container string
	Status    string
	ExitCode  int
}

var healthStatusFn = docker.ContainerStatus

// WaitHealthy polls suite containers until they are all running or the timeout expires.
func WaitHealthy(
	ctx context.Context,
	s Suite,
	timeout time.Duration,
	progressFn func(results []HealthResult),
) error {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	deadlineCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	check := func() ([]HealthResult, bool, error) {
		results := make([]HealthResult, 0, len(s.Containers))
		allRunning := true
		for _, container := range s.Containers {
			status, err := healthStatusFn(deadlineCtx, container.Name)
			if err != nil {
				status = "unhealthy"
			}
			if status != "running" {
				allRunning = false
			}
			results = append(results, HealthResult{
				Container: container.Name,
				Status:    status,
			})
		}
		if progressFn != nil {
			progressFn(results)
		}
		return results, allRunning, nil
	}

	results, ok, err := check()
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-deadlineCtx.Done():
			return fmt.Errorf("%s", healthTimeoutError(s, timeout, results))
		case <-ticker.C:
			results, ok, err = check()
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}

func healthTimeoutError(s Suite, timeout time.Duration, results []HealthResult) string {
	lines := make([]string, 0, len(results))
	for _, result := range results {
		if result.Status == "running" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s not running after %s\n   Check logs: concave logs %s --service %s", result.Container, timeout.Round(time.Second), s.Name, result.Container))
	}
	if len(lines) == 0 {
		return fmt.Sprintf("suite %s did not become healthy after %s", s.Name, timeout.Round(time.Second))
	}
	return strings.Join(lines, "\n")
}
