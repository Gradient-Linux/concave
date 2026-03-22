//go:build linux && cgo

package auth

import (
	"errors"
	"os/user"
	"testing"
	"time"
)

func restorePasswordGlobals(t *testing.T) {
	t.Helper()
	oldVerify := verifyPassword
	oldLimiter := rateLimiter
	oldLookupUser := lookupUser
	oldLookupGroupID := lookupGroupID
	oldLookupGroupIDs := lookupGroupIDs
	t.Cleanup(func() {
		verifyPassword = oldVerify
		rateLimiter = oldLimiter
		lookupUser = oldLookupUser
		lookupGroupID = oldLookupGroupID
		lookupGroupIDs = oldLookupGroupIDs
	})
}

func TestRateLimiterBlocksAfterFiveAttempts(t *testing.T) {
	limiter := newRateLimiter(5, time.Minute)
	for i := 0; i < 5; i++ {
		if !limiter.Allow("alice") {
			t.Fatalf("Allow(%d) = false too early", i)
		}
		limiter.Record("alice")
	}
	if limiter.Allow("alice") {
		t.Fatal("Allow() = true after five attempts, want false")
	}
}

func TestRateLimiterResetsAfterWindow(t *testing.T) {
	limiter := newRateLimiter(5, time.Second)
	now := time.Now().UTC()
	limiter.attempts["alice"] = []time.Time{
		now.Add(-3 * time.Second),
		now.Add(-2 * time.Second),
		now.Add(-1500 * time.Millisecond),
		now.Add(-1200 * time.Millisecond),
		now.Add(-1100 * time.Millisecond),
	}
	if !limiter.Allow("alice") {
		t.Fatal("Allow() = false, want true after pruning old attempts")
	}
}

func TestAuthenticateSameErrorForWrongPasswordAndMissingUser(t *testing.T) {
	restorePasswordGlobals(t)

	verifyPassword = func(username, password string) error {
		if username == "missing" {
			return nil
		}
		return errors.New("wrong password")
	}
	rateLimiter = newRateLimiter(5, time.Minute)
	lookupUser = func(username string) (*user.User, error) {
		if username == "missing" {
			return &user.User{Username: username}, nil
		}
		return &user.User{Username: username}, nil
	}
	lookupGroupIDs = func(u *user.User) ([]string, error) {
		if u.Username == "missing" {
			return []string{"10"}, nil
		}
		return []string{"10"}, nil
	}
	lookupGroupID = func(gid string) (*user.Group, error) {
		if gid == "10" {
			return &user.Group{Name: "docker"}, nil
		}
		return &user.Group{Name: "docker"}, nil
	}

	_, errWrong := Authenticate("alice", "wrong")
	_, errMissing := Authenticate("missing", "correct")
	if !errors.Is(errWrong, ErrInvalidCredentials) {
		t.Fatalf("Authenticate(wrong) error = %v, want ErrInvalidCredentials", errWrong)
	}
	if !errors.Is(errMissing, ErrInvalidCredentials) {
		t.Fatalf("Authenticate(missing) error = %v, want ErrInvalidCredentials", errMissing)
	}
}

func TestAuthenticateSuccessReturnsRole(t *testing.T) {
	restorePasswordGlobals(t)

	verifyPassword = func(username, password string) error { return nil }
	rateLimiter = newRateLimiter(5, time.Minute)
	lookupUser = func(username string) (*user.User, error) {
		return &user.User{Username: username}, nil
	}
	lookupGroupIDs = func(*user.User) ([]string, error) { return []string{"10"}, nil }
	lookupGroupID = func(gid string) (*user.Group, error) {
		return &user.Group{Name: "gradient-operator"}, nil
	}

	role, err := Authenticate("alice", "correct")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if role != RoleOperator {
		t.Fatalf("Authenticate() role = %v, want %v", role, RoleOperator)
	}
}
