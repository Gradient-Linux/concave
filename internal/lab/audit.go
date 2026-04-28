package lab

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Gradient-Linux/concave/internal/workspace"
)

// AuditEvent is the record written for each operator action on an env.
//
// The audit log is a Gradient-specific touch on top of the raw lab lifecycle:
// every launch, extend, archive, destroy, and restore is appended here as a
// JSON line. This is the same format the rest of concave will converge on in
// Phase 0.3 (unified audit log).
type AuditEvent struct {
	Time    time.Time `json:"time"`
	Actor   string    `json:"actor"`
	Action  string    `json:"action"`
	EnvID   string    `json:"env_id,omitempty"`
	Driver  string    `json:"driver,omitempty"`
	Details any       `json:"details,omitempty"`
	PeerID  string    `json:"peer_id,omitempty"`
	Error   string    `json:"error,omitempty"`
}

// AuditWriter appends AuditEvents to a sink. Implementations are expected to
// be concurrency-safe.
type AuditWriter interface {
	Write(event AuditEvent) error
}

// FileAuditWriter is a JSONL file writer.
type FileAuditWriter struct {
	mu   sync.Mutex
	path string
}

// NewFileAuditWriter targets ~/gradient/config/lab-audit.jsonl by default.
func NewFileAuditWriter() *FileAuditWriter {
	return &FileAuditWriter{path: workspace.ConfigPath("lab-audit.jsonl")}
}

// NewFileAuditWriterAt targets an explicit path (useful for tests).
func NewFileAuditWriterAt(path string) *FileAuditWriter {
	return &FileAuditWriter{path: path}
}

// Write appends the event as one JSON line.
func (w *FileAuditWriter) Write(event AuditEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(w.path), 0o700); err != nil {
		return fmt.Errorf("ensure audit dir: %w", err)
	}
	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer f.Close()
	return writeAuditLine(f, event)
}

func writeAuditLine(w io.Writer, event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode audit: %w", err)
	}
	if _, err := w.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write audit: %w", err)
	}
	return nil
}

// DiscardAuditWriter silently drops events; used when no audit path has been
// configured.
type DiscardAuditWriter struct{}

// Write implements AuditWriter.
func (DiscardAuditWriter) Write(AuditEvent) error { return nil }

// MemoryAuditWriter is a test-only in-memory writer.
type MemoryAuditWriter struct {
	mu     sync.Mutex
	Events []AuditEvent
}

// Write implements AuditWriter.
func (m *MemoryAuditWriter) Write(event AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = append(m.Events, event)
	return nil
}

// Snapshot returns a defensive copy of the recorded events.
func (m *MemoryAuditWriter) Snapshot() []AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]AuditEvent, len(m.Events))
	copy(out, m.Events)
	return out
}

// localPeerID resolves a Gradient-specific peer identifier used to stamp
// archive manifests. Prefers /etc/machine-id, falls back to hostname.
func localPeerID() string {
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		if id := trimLine(string(data)); id != "" {
			return id
		}
	}
	if host, err := os.Hostname(); err == nil && host != "" {
		return host
	}
	return "local"
}

func trimLine(s string) string {
	for i, r := range s {
		if r == '\n' || r == '\r' {
			return s[:i]
		}
	}
	return s
}
