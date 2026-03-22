package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Session holds a cached UI session token.
type Session struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	Role      Role      `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SessionPath returns the client session cache path.
func SessionPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.Getenv("HOME"), ".config", "concave", "session.json")
	}
	return filepath.Join(configDir, "concave", "session.json")
}

// LoadSession loads a cached session and rejects expired entries.
func LoadSession() (Session, error) {
	path := SessionPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return Session{}, err
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if session.Token == "" || session.ExpiresAt.IsZero() || time.Now().After(session.ExpiresAt) {
		return Session{}, fmt.Errorf("session expired")
	}
	return session, nil
}

// SaveSession writes a cached session with mode 0600.
func SaveSession(s Session) error {
	path := SessionPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

// ClearSession removes the cached session if present.
func ClearSession() error {
	err := os.Remove(SessionPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
