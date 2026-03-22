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
}

// Line prints raw formatted output through the shared writer.
func Line(text string) {
	outputMu.Lock()
	defer outputMu.Unlock()
	_, _ = fmt.Fprintln(output, text)
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
	printProgress(label, current, total)
}

func printStatus(icon, label, detail string) {
	outputMu.Lock()
	defer outputMu.Unlock()
	_, _ = fmt.Fprintf(output, "  %s  %-20s %s\n", icon, label, detail)
}

func printProgress(label string, current, total int) {
	outputMu.Lock()
	defer outputMu.Unlock()
	ratio := float64(current) / float64(total)
	_, _ = fmt.Fprintf(output, "  →  %-20s %s %3d%%\n", label, gradientANSIBar(24, ratio), int(ratio*100))
}

func gradientANSIBar(width int, ratio float64) string {
	if width <= 0 {
		return ""
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	out := ""
	for idx := 0; idx < filled; idx++ {
		r := lerp(0x40, 0xF9, idx, max(1, filled))
		g := lerp(0x10, 0xD4, idx, max(1, filled))
		b := lerp(0x79, 0x4E, idx, max(1, filled))
		out += fmt.Sprintf("\x1b[38;2;%d;%d;%dm█", r, g, b)
	}
	if filled < width {
		out += "\x1b[38;2;107;114;128m" + repeat("░", width-filled)
	}
	return out + "\x1b[0m"
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
