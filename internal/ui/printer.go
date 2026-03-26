package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	outputMu sync.Mutex
	output   io.Writer = os.Stdout
)

// SetOutput directs UI output to the provided writer.
func SetOutput(w io.Writer) {
	outputMu.Lock()
	defer outputMu.Unlock()
	endProgressLocked()
	resetProgressStateLocked()
	output = w
}

// ResetOutput restores UI output to stdout.
func ResetOutput() {
	SetOutput(os.Stdout)
}

// Pass prints a successful status line.
func Pass(label, detail string) {
	printStatus("✓", label, detail)
}

// Fail prints a failed status line.
func Fail(label, detail string) {
	printStatus("✗", label, detail)
}

// Warn prints a warning status line.
func Warn(label, detail string) {
	printStatus("⚠", label, detail)
}

// Info prints an informational status line.
func Info(label, detail string) {
	printStatus("→", label, detail)
}

// Header prints a section title.
func Header(title string) {
	outputMu.Lock()
	defer outputMu.Unlock()
	_, _ = fmt.Fprintf(output, "%s\n", title)
	renderProgressLocked()
}

// Line prints raw formatted output through the shared writer.
func Line(text string) {
	outputMu.Lock()
	defer outputMu.Unlock()
	_, _ = fmt.Fprintln(output, text)
	renderProgressLocked()
}

// Progress prints a step-based progress bar using the project gradient.
func Progress(label string, current, total int) {
	if total <= 0 {
		total = 1
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}
	outputMu.Lock()
	defer outputMu.Unlock()
	printStickyProgressLocked(label, current, total)
}

func printStatus(icon, label, detail string) {
	outputMu.Lock()
	defer outputMu.Unlock()
	_, _ = fmt.Fprintf(output, "  %s  %-20s %s\n", icon, label, detail)
	renderProgressLocked()
}

func lerp(start, end, idx, count int) int {
	if count <= 1 {
		return start
	}
	return start + ((end-start)*idx)/(count-1)
}

func repeat(char string, count int) string {
	if count <= 0 {
		return ""
	}
	text := ""
	for i := 0; i < count; i++ {
		text += char
	}
	return text
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
