package resolverclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

// DefaultSocketPath is the resolver daemon socket path.
const DefaultSocketPath = "/run/gradient/resolver.sock"

// ErrUnavailable is returned when the resolver socket is not reachable.
var ErrUnavailable = errors.New("resolver unavailable")

// DriftTier mirrors the resolver daemon drift classification.
type DriftTier int

// PackageDiff describes one package divergence from baseline.
type PackageDiff struct {
	Name     string    `json:"name"`
	Baseline string    `json:"baseline"`
	Current  string    `json:"current"`
	Tier     DriftTier `json:"tier"`
	Reason   string    `json:"reason"`
}

// DriftReport is one drift report from the resolver daemon.
type DriftReport struct {
	Group     string        `json:"group"`
	User      string        `json:"user"`
	Timestamp time.Time     `json:"timestamp"`
	Diffs     []PackageDiff `json:"diffs"`
	Clean     bool          `json:"clean"`
}

// Status describes the current resolver daemon state.
type Status struct {
	Running       bool          `json:"running"`
	LastScan      time.Time     `json:"last_scan"`
	GroupReports  []DriftReport `json:"group_reports"`
	SnapshotCount int           `json:"snapshot_count"`
	SocketPath    string        `json:"socket_path"`
}

type statusRequest struct {
	Type string `json:"type"`
}

type driftRequest struct {
	Type  string `json:"type"`
	Group string `json:"group"`
}

type response[T any] struct {
	Type    string `json:"type"`
	Payload T      `json:"payload"`
	Error   string `json:"error"`
}

// QueryStatus queries the resolver socket for daemon status.
func QueryStatus(socketPath string) (Status, error) {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	var resp response[Status]
	if err := query(socketPath, statusRequest{Type: "status"}, &resp); err != nil {
		return Status{}, err
	}
	if resp.Error != "" {
		return Status{}, errors.New(resp.Error)
	}
	return resp.Payload, nil
}

// QueryDrift queries the resolver socket for group drift reports.
func QueryDrift(socketPath, group string) ([]DriftReport, error) {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	var resp response[[]DriftReport]
	if err := query(socketPath, driftRequest{Type: "drift", Group: group}, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	if resp.Payload == nil {
		return []DriftReport{}, nil
	}
	return resp.Payload, nil
}

func query(socketPath string, req any, out any) error {
	conn, err := net.DialTimeout("unix", socketPath, 200*time.Millisecond)
	if err != nil {
		return mapUnavailable(err)
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return fmt.Errorf("encode resolver request: %w", err)
	}
	if err := json.NewDecoder(conn).Decode(out); err != nil {
		return fmt.Errorf("decode resolver response: %w", err)
	}
	return nil
}

func mapUnavailable(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return ErrUnavailable
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return ErrUnavailable
	}
	return err
}
