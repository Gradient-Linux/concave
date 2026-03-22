package docker

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/ui"
)

// RetryConfig controls retry behaviour.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	ShouldRetry func(err error) bool
}

var retryJitter = func() float64 { return rand.Float64() }

// DefaultRetryConfig returns sensible defaults for pull operations.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   2 * time.Second,
		MaxDelay:    30 * time.Second,
		ShouldRetry: defaultShouldRetry,
	}
}

// WithRetry retries a transient operation with exponential backoff and jitter.
func WithRetry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 2 * time.Second
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 30 * time.Second
	}
	if cfg.ShouldRetry == nil {
		cfg.ShouldRetry = defaultShouldRetry
	}

	var lastErr error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt == cfg.MaxAttempts || !cfg.ShouldRetry(err) {
				return err
			}
			delay := jitterDelay(backoffDelay(cfg.BaseDelay, cfg.MaxDelay, attempt-1))
			ui.Warn("Pull", fmt.Sprintf("attempt %d/%d — %s, retrying in %s", attempt+1, cfg.MaxAttempts, compactError(err), delay.Round(time.Second)))
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			continue
		}
		return nil
	}
	return lastErr
}

func defaultShouldRetry(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "manifest unknown"),
		strings.Contains(text, "not found"),
		strings.Contains(text, "no compose template found"),
		strings.Contains(text, "validation"):
		return false
	default:
		return true
	}
}

func backoffDelay(base, max time.Duration, exponent int) time.Duration {
	delay := float64(base) * math.Pow(2, float64(exponent))
	if delay > float64(max) {
		delay = float64(max)
	}
	return time.Duration(delay)
}

func jitterDelay(delay time.Duration) time.Duration {
	if delay <= 0 {
		return 0
	}
	factor := 0.9 + (retryJitter() * 0.2)
	return time.Duration(float64(delay) * factor)
}

func compactError(err error) string {
	if err == nil {
		return ""
	}
	line := strings.Split(strings.TrimSpace(err.Error()), "\n")[0]
	return line
}
