package docker

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Gradient-Linux/concave/internal/ui"
)

func TestWithRetrySucceedsAfterTransientFailures(t *testing.T) {
	oldJitter := retryJitter
	t.Cleanup(func() { retryJitter = oldJitter })
	retryJitter = func() float64 { return 0.5 }

	var attempts int
	var buf bytes.Buffer
	ui.SetOutput(&buf)
	defer ui.ResetOutput()

	err := WithRetry(context.Background(), RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
		MaxDelay:    time.Millisecond,
	}, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("network error")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithRetry() error = %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
	if got := buf.String(); !strings.Contains(got, "attempt 2/3") || !strings.Contains(got, "attempt 3/3") {
		t.Fatalf("unexpected warning output: %q", got)
	}
}

func TestWithRetryStopsOnContextCancellation(t *testing.T) {
	oldJitter := retryJitter
	t.Cleanup(func() { retryJitter = oldJitter })
	retryJitter = func() float64 { return 0.5 }

	ctx, cancel := context.WithCancel(context.Background())
	var attempts int
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := WithRetry(ctx, RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Second,
		MaxDelay:    time.Second,
	}, func() error {
		attempts++
		return errors.New("temporary error")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WithRetry() error = %v, want context canceled", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
	if time.Since(start) >= 500*time.Millisecond {
		t.Fatalf("retry did not return promptly after cancellation")
	}
}

func TestWithRetryDoesNotRetryManifestUnknown(t *testing.T) {
	var attempts int
	err := WithRetry(context.Background(), DefaultRetryConfig(), func() error {
		attempts++
		return errors.New("manifest unknown: image not found")
	})
	if err == nil {
		t.Fatal("expected retry failure")
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}
