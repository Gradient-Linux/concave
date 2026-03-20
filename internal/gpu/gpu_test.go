package gpu

import (
	"errors"
	"strings"
	"testing"
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

func TestDetect(t *testing.T) {
	previous := runner
	runner = &mockRunner{outputs: map[string][]byte{"nvidia-smi ": []byte("ok")}}
	defer func() { runner = previous }()

	state, err := Detect()
	if err != nil || state != GPUStateNVIDIA {
		t.Fatalf("Detect() = %v, %v", state, err)
	}
}

func TestDriverBranchForCapability(t *testing.T) {
	cases := map[string]string{
		"7.5": "535",
		"8.0": "560",
		"8.6": "560",
		"8.9": "570",
		"9.0": "570",
	}
	for capability, want := range cases {
		got, err := DriverBranchForCapability(capability)
		if err != nil {
			t.Fatalf("DriverBranchForCapability(%q) error = %v", capability, err)
		}
		if got != want {
			t.Fatalf("DriverBranchForCapability(%q) = %q, want %q", capability, got, want)
		}
	}
}
