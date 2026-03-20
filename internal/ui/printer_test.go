package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrinterFormatsStatusLines(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer ResetOutput()

	Pass("Docker", "running")
	Fail("GPU", "missing")
	Warn("Internet", "slow")
	Info("Pulling", "image")

	out := buf.String()
	for _, token := range []string{"✓", "✗", "⚠", "→"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected %q in output %q", token, out)
		}
	}
	if !strings.Contains(out, "Docker") || !strings.Contains(out, "running") {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestSpinnerStartStop(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer ResetOutput()

	spinner := NewSpinner("Pulling")
	spinner.Start()
	spinner.Stop("done")

	if !strings.Contains(buf.String(), "Pulling") {
		t.Fatalf("expected spinner output, got %q", buf.String())
	}
}
