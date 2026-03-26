package ui

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unicode/utf8"
	"unsafe"
)

type stickyProgressState struct {
	active  bool
	rows    int
	cols    int
	message string
	current int
	total   int
}

var stickyProgress stickyProgressState

func EndProgress() {
	outputMu.Lock()
	defer outputMu.Unlock()
	endProgressLocked()
}

func ProgressMessage(message string) {
	outputMu.Lock()
	defer outputMu.Unlock()
	if !stickyProgress.active {
		return
	}
	stickyProgress.message = strings.TrimSpace(message)
	renderProgressLocked()
}

func progressSupportedLocked() bool {
	file, ok := output.(*os.File)
	if !ok {
		return false
	}
	if term := strings.TrimSpace(os.Getenv("TERM")); term == "" || term == "dumb" {
		return false
	}
	info, err := file.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	rows, cols, ok := terminalSize(file)
	if !ok || rows < 3 || cols < 40 {
		return false
	}
	stickyProgress.rows = rows
	stickyProgress.cols = cols
	return true
}

func beginProgressLocked() bool {
	if stickyProgress.active {
		return true
	}
	if !progressSupportedLocked() {
		return false
	}
	clearScreenLocked()
	stickyProgress.active = true
	clearReservedProgressRowsLocked()
	_, _ = fmt.Fprintf(output, "\x1b[1;%dr", stickyProgress.rows-2)
	return true
}

func endProgressLocked() {
	if !stickyProgress.active {
		return
	}
	_, _ = fmt.Fprint(output, "\x1b[r")
	clearReservedProgressRowsLocked()
	_, _ = fmt.Fprintf(output, "\x1b[%d;1H", stickyProgress.rows-1)
	stickyProgress = stickyProgressState{}
}

func printStickyProgressLocked(message string, current, total int) {
	if total <= 0 {
		total = 1
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}

	if !beginProgressLocked() {
		printProgressLineLocked(message, current, total, 24)
		return
	}

	if file, ok := output.(*os.File); ok {
		if rows, cols, ok := terminalSize(file); ok {
			if rows != stickyProgress.rows {
				stickyProgress.rows = rows
				_, _ = fmt.Fprint(output, "\x1b[r")
				clearReservedProgressRowsLocked()
				_, _ = fmt.Fprintf(output, "\x1b[1;%dr", stickyProgress.rows-2)
			}
			stickyProgress.cols = cols
		}
	}

	stickyProgress.message = strings.TrimSpace(message)
	stickyProgress.current = current
	stickyProgress.total = total
	renderProgressLocked()
}

func renderProgressLocked() {
	if !stickyProgress.active {
		return
	}
	line := stickyProgressLine(stickyProgress.cols, stickyProgress.message, stickyProgress.current, stickyProgress.total)
	_, _ = fmt.Fprintf(output, "\x1b[s\x1b[%d;1H\x1b[2K%s\x1b[u", stickyProgress.rows, line)
}

func clearReservedProgressRowsLocked() {
	if stickyProgress.rows <= 1 {
		return
	}
	_, _ = fmt.Fprintf(output, "\x1b[s\x1b[%d;1H\x1b[2K\x1b[%d;1H\x1b[2K\x1b[u", stickyProgress.rows-1, stickyProgress.rows)
}

func clearScreenLocked() {
	_, _ = fmt.Fprint(output, "\x1b[2J\x1b[H")
}

func stickyProgressLine(cols int, message string, current, total int) string {
	if total <= 0 {
		total = 1
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}
	if cols <= 0 {
		cols = 80
	}

	ratio := float64(current) / float64(total)
	percent := int(ratio * 100)

	message = strings.TrimSpace(message)
	if message == "" {
		message = "Working..."
	}

	suffix := fmt.Sprintf("] %3d%%", percent)
	messageWidth := runeWidth(message)
	barWidth := cols - (2 + messageWidth + 2) - runeWidth(suffix)
	if barWidth < 5 {
		maxMessage := cols - runeWidth(suffix) - 9
		if maxMessage < 8 {
			maxMessage = 8
		}
		message = truncateRunes(message, maxMessage)
		messageWidth = runeWidth(message)
		barWidth = cols - (2 + messageWidth + 2) - runeWidth(suffix)
	}
	if barWidth < 5 {
		barWidth = 5
	}
	if barWidth > 200 {
		barWidth = 200
	}

	prefix := "  " + message + " ["
	return prefix + gradientProgressBar(barWidth, ratio) + suffix
}

func printProgressLineLocked(message string, current, total, width int) {
	if total <= 0 {
		total = 1
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}
	if width < 5 {
		width = 5
	}
	ratio := float64(current) / float64(total)
	_, _ = fmt.Fprintf(output, "  %s [%s] %3d%%\n", message, gradientProgressBar(width, ratio), int(ratio*100))
}

func gradientProgressBar(width int, ratio float64) string {
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
	denom := max(1, width-1)
	for idx := 0; idx < filled; idx++ {
		r := lerp(102, 255, idx, denom)
		g := lerp(0, 215, idx, denom)
		b := lerp(153, 0, idx, denom)
		out += fmt.Sprintf("\x1b[38;2;%d;%d;%dm━", r, g, b)
	}
	if filled < width {
		out += "\x1b[38;2;60;60;60m" + repeat("━", width-filled)
	}
	return out + "\x1b[0m"
}

func truncateRunes(value string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if utf8.RuneCountInString(value) <= maxRunes {
		return value
	}
	runes := []rune(value)
	return string(runes[:maxRunes])
}

func runeWidth(value string) int {
	return utf8.RuneCountInString(value)
}

type winsize struct {
	Rows uint16
	Cols uint16
	X    uint16
	Y    uint16
}

func terminalSize(file *os.File) (int, int, bool) {
	if file == nil {
		return 0, 0, false
	}
	ws := &winsize{}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	if errno != 0 || ws.Rows == 0 || ws.Cols == 0 {
		return 0, 0, false
	}
	return int(ws.Rows), int(ws.Cols), true
}

func resetProgressStateLocked() {
	stickyProgress = stickyProgressState{}
}
