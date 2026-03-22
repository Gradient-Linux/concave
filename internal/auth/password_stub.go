//go:build !linux || !cgo

package auth

import (
	"errors"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrRateLimited        = errors.New("too many attempts. try again in 60s")
)

// Authenticate is unsupported without Linux PAM.
func Authenticate(username, password string) (Role, error) {
	_ = username
	_ = password
	return 0, errors.New("pam authentication requires linux with cgo enabled")
}

// IsRateLimited reports whether an auth error is a lockout.
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}
