package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/system"
	"github.com/Gradient-Linux/concave/internal/ui"
)

func wrapUserError(err error) error {
	return system.WithExitCode(err, system.ExitUserError)
}

func wrapDockerError(err error) error {
	return system.WithExitCode(err, system.ExitDocker)
}

func runLockedOperation(subcommand string, timeout time.Duration, cleanup func(), fn func(context.Context) error) error {
	unlock, err := systemLock(subcommand)
	if err != nil {
		return wrapUserError(err)
	}
	defer ui.EndProgress()

	runCleanup := func() {
		unlock()
		if cleanup != nil {
			cleanup()
		}
	}

	ctx, stop := systemSignalHandler(context.Background(), runCleanup)
	defer stop()
	defer unlock()

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return fn(ctx)
}

func composeCleanup(name string) func() {
	return func() {
		_ = os.Remove(dockerComposePath(name) + ".tmp")
	}
}

func waitForHealthy(ctx context.Context, s suite.Suite) error {
	ui.Info("Health", "Waiting for "+s.Name+" to be healthy...")
	deadlineCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	check := func() ([]suite.HealthResult, bool, error) {
		results := make([]suite.HealthResult, 0, len(s.Containers))
		allRunning := true
		composePath := dockerComposePath(s.Name)
		for _, container := range s.Containers {
			status, err := dockerComposeServiceStatus(deadlineCtx, composePath, container.Name)
			if err != nil {
				status = "error"
			}
			if status != "running" {
				allRunning = false
			}
			results = append(results, suite.HealthResult{
				Container: container.Name,
				Status:    status,
			})
		}
		return results, allRunning, nil
	}

	printProgress := func(results []suite.HealthResult) {
		for _, result := range results {
			switch result.Status {
			case "running":
				ui.Pass(result.Container, "running")
			case "not found":
				ui.Warn(result.Container, "starting...")
			default:
				ui.Warn(result.Container, result.Status)
			}
		}
	}

	results, ok, err := check()
	if err != nil {
		return wrapDockerError(err)
	}
	printProgress(results)
	if ok {
		return nil
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-deadlineCtx.Done():
			return wrapDockerError(fmt.Errorf("%s", healthTimeoutError(s, 60*time.Second, results)))
		case <-ticker.C:
			results, ok, err = check()
			if err != nil {
				return wrapDockerError(err)
			}
			printProgress(results)
			if ok {
				return nil
			}
		}
	}
}

func healthTimeoutError(s suite.Suite, timeout time.Duration, results []suite.HealthResult) string {
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
