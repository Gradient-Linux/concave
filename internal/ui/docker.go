package ui

import (
	"strings"
	"time"
)

const dockerProgressMinInterval = 1500 * time.Millisecond

// DockerPullReporter returns a throttled reporter for docker pull output.
func DockerPullReporter(prefix string) func(string) {
	var (
		lastLine string
		lastAt   time.Time
	)
	prefix = strings.TrimSpace(prefix)

	return func(line string) {
		line = strings.TrimSpace(line)
		if line == "" {
			return
		}

		line = compactDockerProgress(line)
		now := time.Now()
		if line == lastLine && now.Sub(lastAt) < dockerProgressMinInterval {
			return
		}

		lastLine = line
		lastAt = now
		if prefix != "" {
			ProgressMessage(prefix + " — " + line)
		} else {
			ProgressMessage(line)
		}
		Info("Pull", line)
	}
}

func compactDockerProgress(line string) string {
	line = strings.TrimSpace(line)
	if strings.Contains(line, "\r") {
		parts := strings.Split(line, "\r")
		line = parts[len(parts)-1]
	}
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	if len(fields[0]) > 16 && strings.IndexByte(fields[0], ':') == -1 {
		fields[0] = fields[0][:12]
	}
	return strings.Join(fields, " ")
}
