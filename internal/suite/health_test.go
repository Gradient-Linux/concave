package suite

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitHealthyReturnsWhenAllContainersRunning(t *testing.T) {
	oldStatus := healthStatusFn
	t.Cleanup(func() { healthStatusFn = oldStatus })

	var ready atomic.Bool
	healthStatusFn = func(ctx context.Context, name string) (string, error) {
		if !ready.Load() && name == "gradient-boost-lab" {
			return "not found", nil
		}
		return "running", nil
	}

	var snapshots [][]HealthResult
	s := Suite{
		Name: "boosting",
		Containers: []Container{
			{Name: "gradient-boost-core"},
			{Name: "gradient-boost-lab"},
		},
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		ready.Store(true)
	}()

	err := WaitHealthy(context.Background(), s, 3*time.Second, func(results []HealthResult) {
		copied := append([]HealthResult(nil), results...)
		snapshots = append(snapshots, copied)
	})
	if err != nil {
		t.Fatalf("WaitHealthy() error = %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatal("expected progress snapshots")
	}
}

func TestWaitHealthyTimeoutIncludesRecoveryHint(t *testing.T) {
	oldStatus := healthStatusFn
	t.Cleanup(func() { healthStatusFn = oldStatus })

	healthStatusFn = func(ctx context.Context, name string) (string, error) {
		return "stopped", nil
	}

	s := Suite{
		Name: "boosting",
		Containers: []Container{
			{Name: "gradient-boost-lab"},
		},
	}

	err := WaitHealthy(context.Background(), s, 10*time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "Check logs: concave logs boosting --service gradient-boost-lab") {
		t.Fatalf("missing recovery hint: %v", err)
	}
}

func TestWaitHealthyReturnsOnContextCancellation(t *testing.T) {
	oldStatus := healthStatusFn
	t.Cleanup(func() { healthStatusFn = oldStatus })

	healthStatusFn = func(ctx context.Context, name string) (string, error) {
		return "stopped", nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := Suite{
		Name: "boosting",
		Containers: []Container{
			{Name: "gradient-boost-lab"},
		},
	}

	err := WaitHealthy(ctx, s, time.Second, nil)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "not running after") {
		t.Fatalf("unexpected error: %v", err)
	}
}
