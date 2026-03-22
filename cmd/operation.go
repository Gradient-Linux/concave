package cmd

import (
	"context"
	"os"
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
	return wrapDockerError(waitHealthy(ctx, s, 60*time.Second, func(results []suite.HealthResult) {
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
	}))
}
