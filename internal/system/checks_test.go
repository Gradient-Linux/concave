package system

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

type mockRunner struct {
	outputs map[string][]byte
	errors  map[string]error
}

func (m *mockRunner) Run(name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	if err, ok := m.errors[key]; ok {
		return nil, err
	}
	if out, ok := m.outputs[key]; ok {
		return out, nil
	}
	return nil, errors.New("unexpected command: " + key)
}

type stubConn struct{}

func (stubConn) Read([]byte) (int, error)         { return 0, nil }
func (stubConn) Write(b []byte) (int, error)      { return len(b), nil }
func (stubConn) Close() error                     { return nil }
func (stubConn) LocalAddr() net.Addr              { return nil }
func (stubConn) RemoteAddr() net.Addr             { return nil }
func (stubConn) SetDeadline(time.Time) error      { return nil }
func (stubConn) SetReadDeadline(time.Time) error  { return nil }
func (stubConn) SetWriteDeadline(time.Time) error { return nil }

func TestDockerRunning(t *testing.T) {
	previous := runner
	runner = &mockRunner{outputs: map[string][]byte{"docker info": []byte("ok")}}
	defer func() { runner = previous }()

	ok, err := DockerRunning()
	if err != nil || !ok {
		t.Fatalf("DockerRunning() = %v, %v", ok, err)
	}
}

func TestUserInDockerGroup(t *testing.T) {
	previous := runner
	runner = &mockRunner{outputs: map[string][]byte{"id -nG": []byte("wheel docker audio")}}
	defer func() { runner = previous }()

	ok, err := UserInDockerGroup()
	if err != nil || !ok {
		t.Fatalf("UserInDockerGroup() = %v, %v", ok, err)
	}
}

func TestInternetReachable(t *testing.T) {
	previous := dialContext
	dialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		return stubConn{}, nil
	}
	defer func() { dialContext = previous }()

	ok, err := InternetReachable()
	if err != nil || !ok {
		t.Fatalf("InternetReachable() = %v, %v", ok, err)
	}
}

func TestOpenURLFallsBackToGio(t *testing.T) {
	previous := runner
	runner = &mockRunner{
		errors:  map[string]error{"xdg-open https://example.com": errors.New("missing")},
		outputs: map[string][]byte{"gio open https://example.com": []byte("ok")},
	}
	defer func() { runner = previous }()

	if err := OpenURL("https://example.com"); err != nil {
		t.Fatalf("OpenURL() error = %v", err)
	}
}
