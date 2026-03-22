package system

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Gradient-Linux/concave/internal/workspace"
)

var ErrLockHeld = errors.New("another concave operation is in progress")

var (
	lockNow  = time.Now
	lockOpen = os.OpenFile
	lockPID  = os.Getpid
)

// Lock acquires an exclusive advisory lock for mutating concave operations.
func Lock(subcommand string) (func(), error) {
	if err := workspace.EnsureLayout(); err != nil {
		return nil, err
	}

	path := workspace.ConfigPath(".concave.lock")
	file, err := lockOpen(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		info := readLockInfo(file)
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			Debug("lock already held", "subcommand", subcommand, "holder", info)
			return nil, fmt.Errorf("%w.\nHeld by: %s\n\nWait for it to finish, or if it is stuck:\n  kill %d\nThe lock will be released automatically.", ErrLockHeld, info.summary(), info.PID)
		}
		return nil, fmt.Errorf("acquire lock: %w", err)
	}

	info := lockInfo{PID: lockPID(), Subcommand: subcommand, Started: lockNow().UTC()}
	if err := writeLockInfo(file, info); err != nil {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
		return nil, err
	}

	Debug("lock acquired", "subcommand", subcommand, "pid", info.PID)
	unlocked := false
	return func() {
		if unlocked {
			return
		}
		unlocked = true
		Debug("lock released", "subcommand", subcommand, "pid", info.PID)
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		_ = file.Close()
		_ = os.Remove(path)
	}, nil
}

type lockInfo struct {
	PID        int
	Subcommand string
	Started    time.Time
}

func (i lockInfo) summary() string {
	if i.PID == 0 || i.Subcommand == "" || i.Started.IsZero() {
		return "concave operation (details unavailable)"
	}
	return fmt.Sprintf("concave %s (PID %d, started %s ago)", i.Subcommand, i.PID, humanSince(i.Started))
}

func readLockInfo(file *os.File) lockInfo {
	if _, err := file.Seek(0, 0); err != nil {
		return lockInfo{}
	}
	data, err := os.ReadFile(file.Name())
	if err != nil {
		return lockInfo{}
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 3 {
		return lockInfo{}
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(lines[0]))
	started, _ := time.Parse(time.RFC3339, strings.TrimSpace(lines[2]))
	return lockInfo{
		PID:        pid,
		Subcommand: strings.TrimSpace(lines[1]),
		Started:    started,
	}
}

func writeLockInfo(file *os.File, info lockInfo) error {
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("truncate lock file: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("seek lock file: %w", err)
	}
	if _, err := fmt.Fprintf(file, "%d\n%s\n%s\n", info.PID, info.Subcommand, info.Started.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("write lock file: %w", err)
	}
	return file.Sync()
}

func humanSince(started time.Time) string {
	if started.IsZero() {
		return "unknown time"
	}
	elapsed := time.Since(started).Round(time.Second)
	switch {
	case elapsed < time.Minute:
		return fmt.Sprintf("%d seconds", int(elapsed.Seconds()))
	case elapsed < time.Hour:
		return fmt.Sprintf("%d minutes", int(elapsed.Minutes()))
	default:
		return fmt.Sprintf("%d hours", int(elapsed.Hours()))
	}
}
