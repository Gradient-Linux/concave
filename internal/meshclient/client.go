package meshclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

// DefaultSocketPath is the mesh daemon socket path.
const DefaultSocketPath = "/run/gradient/mesh.sock"

// ErrUnavailable is returned when the mesh socket is not reachable.
var ErrUnavailable = errors.New("mesh unavailable")

// NodeVisibility matches the mesh daemon visibility values.
type NodeVisibility string

// NodeInfo is the fleet-visible snapshot of one node.
type NodeInfo struct {
	Hostname        string         `json:"hostname"`
	MachineID       string         `json:"machine_id"`
	GradientVersion string         `json:"gradient_version"`
	Visibility      NodeVisibility `json:"visibility"`
	InstalledSuites []string       `json:"installed_suites"`
	ResolverRunning bool           `json:"resolver_running"`
	LastSeen        time.Time      `json:"last_seen"`
	Address         string         `json:"address"`
}

type request struct {
	Action string `json:"action"`
}

type response struct {
	Self  *NodeInfo  `json:"self,omitempty"`
	Peers []NodeInfo `json:"peers,omitempty"`
	Error string     `json:"error,omitempty"`
}

// QuerySelf queries the local mesh node snapshot.
func QuerySelf(socketPath string) (NodeInfo, error) {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	resp, err := query(socketPath, request{Action: "self"})
	if err != nil {
		return NodeInfo{}, err
	}
	if resp.Self == nil {
		return NodeInfo{}, nil
	}
	return *resp.Self, nil
}

// QueryFleet queries the visible fleet snapshot.
func QueryFleet(socketPath string) ([]NodeInfo, error) {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	resp, err := query(socketPath, request{Action: "fleet"})
	if err != nil {
		return nil, err
	}
	if resp.Peers == nil {
		return []NodeInfo{}, nil
	}
	return resp.Peers, nil
}

func query(socketPath string, req request) (response, error) {
	conn, err := net.DialTimeout("unix", socketPath, 200*time.Millisecond)
	if err != nil {
		return response{}, mapUnavailable(err)
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return response{}, fmt.Errorf("encode mesh request: %w", err)
	}
	var resp response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return response{}, fmt.Errorf("decode mesh response: %w", err)
	}
	if resp.Error != "" {
		return response{}, errors.New(resp.Error)
	}
	return resp, nil
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
