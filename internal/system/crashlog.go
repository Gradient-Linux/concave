package system

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/workspace"
)

var (
	crashNow     = time.Now
	crashOpen    = os.OpenFile
	crashStderr  = os.Stderr
	crashExit    = os.Exit
	crashStackFn = debug.Stack
)

// InstallCrashHandler must be deferred at the top of main().
func InstallCrashHandler(version string, args []string) {
	recovered := recover()
	if recovered == nil {
		return
	}

	path, err := appendCrashReport(version, args, recovered)
	if err != nil {
		path = fmt.Sprintf("/tmp/concave-crash-%d.log", os.Getpid())
	}

	fmt.Fprintln(crashStderr, "concave crashed unexpectedly.")
	fmt.Fprintf(crashStderr, "Crash details written to: %s\n", path)
	fmt.Fprintln(crashStderr, "Please report this at: https://github.com/Gradient-Linux/concave/issues")
	crashExit(ExitPanic)
}

func appendCrashReport(version string, args []string, recovered any) (string, error) {
	primaryPath := filepath.Join(workspace.Root(), "logs", "concave.log")
	if err := os.MkdirAll(filepath.Dir(primaryPath), 0o755); err == nil {
		if err := writeCrashReport(primaryPath, buildCrashReport(version, args, recovered)); err == nil {
			return primaryPath, nil
		}
	}

	fallbackPath := filepath.Join("/tmp", fmt.Sprintf("concave-crash-%d.log", os.Getpid()))
	if err := writeCrashReport(fallbackPath, buildCrashReport(version, args, recovered)); err != nil {
		return fallbackPath, err
	}
	return fallbackPath, nil
}

func writeCrashReport(path, report string) error {
	file, err := crashOpen(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(report); err != nil {
		return err
	}
	return nil
}

func buildCrashReport(version string, args []string, recovered any) string {
	var buf bytes.Buffer
	command := "concave"
	if len(args) > 1 {
		command += " " + filepath.Base(args[1])
	}

	fmt.Fprintln(&buf, "=== concave crash report ===")
	fmt.Fprintf(&buf, "Time:     %s\n", crashNow().UTC().Format(time.RFC3339))
	fmt.Fprintf(&buf, "Version:  %s\n", version)
	fmt.Fprintf(&buf, "Command:  %s\n", command)
	fmt.Fprintf(&buf, "OS:       %s/%s (%s)\n", runtime.GOOS, runtime.GOARCH, distroName())
	fmt.Fprintf(&buf, "Panic:    %v\n\n", recovered)
	buf.WriteString(sanitizeStack(string(crashStackFn())))
	if !strings.HasSuffix(buf.String(), "\n") {
		buf.WriteByte('\n')
	}
	fmt.Fprintln(&buf, "\nPlease report this at: https://github.com/Gradient-Linux/concave/issues")
	fmt.Fprintln(&buf, "=== end of crash report ===")
	fmt.Fprintln(&buf)
	return buf.String()
}

func distroName() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "unknown distro"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
		}
	}
	return "unknown distro"
}

func sanitizeStack(stack string) string {
	lines := strings.Split(strings.TrimSpace(stack), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, ".go:") {
			fields := strings.Fields(trimmed)
			if len(fields) > 0 {
				fields[0] = filepath.Base(fields[0])
				lines[i] = "\t" + strings.Join(fields, " ")
			}
		}
	}
	return strings.Join(lines, "\n") + "\n"
}
