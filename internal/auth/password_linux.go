//go:build linux && cgo

package auth

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Gradient-Linux/concave/internal/system"
	"github.com/msteinert/pam"
)

var (
	// ErrInvalidCredentials keeps auth failures generic.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrRateLimited is returned before PAM is consulted when lockout is active.
	ErrRateLimited = errors.New("too many attempts. try again in 60s")

	verifyPassword = func(username, password string) error {
		tx, err := pam.StartFunc("login", username, func(_ pam.Style, _ string) (string, error) {
			return password, nil
		})
		if err != nil {
			return err
		}
		if err := tx.Authenticate(0); err != nil {
			return err
		}
		return tx.AcctMgmt(0)
	}
	rateLimiter = newRateLimiter(5, 60*time.Second)
)

// Authenticate verifies username and password via PAM.
func Authenticate(username, password string) (Role, error) {
	if !rateLimiter.Allow(username) {
		system.Logger.Warn("auth rate limited", "user", username)
		return 0, ErrRateLimited
	}

	if err := verifyPassword(username, password); err != nil {
		rateLimiter.Record(username)
		system.Logger.Warn("auth failure", "user", username, "reason", "invalid_credentials")
		return 0, ErrInvalidCredentials
	}

	role, err := ResolveRole(username)
	if err != nil {
		rateLimiter.Record(username)
		system.Logger.Warn("auth failure", "user", username, "reason", "no_gradient_group")
		return 0, ErrInvalidCredentials
	}

	rateLimiter.Reset(username)
	system.Logger.Info("auth success", "user", username, "role", role.String())
	return role, nil
}

// RateLimiter tracks failed auth attempts per username.
type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	max      int
	window   time.Duration
}

func newRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts: map[string][]time.Time{},
		max:      max,
		window:   window,
	}
}

// Allow returns false when the account is rate limited.
func (r *RateLimiter) Allow(username string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prune(username)
	return len(r.attempts[username]) < r.max
}

// Record stores a failed attempt timestamp.
func (r *RateLimiter) Record(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prune(username)
	r.attempts[username] = append(r.attempts[username], time.Now().UTC())
}

// Reset clears the attempt window for a user after success.
func (r *RateLimiter) Reset(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.attempts, username)
}

func (r *RateLimiter) prune(username string) {
	cutoff := time.Now().UTC().Add(-r.window)
	attempts := r.attempts[username]
	filtered := attempts[:0]
	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			filtered = append(filtered, attempt)
		}
	}
	if len(filtered) == 0 {
		delete(r.attempts, username)
		return
	}
	r.attempts[username] = filtered
}

// IsRateLimited reports whether an auth error is a lockout.
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// WrapAuthError normalizes backend auth errors for HTTP/UI layers.
func WrapAuthError(err error) error {
	switch {
	case errors.Is(err, ErrRateLimited):
		return err
	case errors.Is(err, ErrInvalidCredentials):
		return err
	case err == nil:
		return nil
	default:
		return fmt.Errorf("authentication failed: %w", err)
	}
}
