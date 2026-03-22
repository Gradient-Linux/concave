package system

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestExitCodeValues(t *testing.T) {
	if ExitSuccess != 0 || ExitUserError != 1 || ExitDocker != 2 || ExitPanic != 3 || ExitSIGINT != 130 || ExitSIGTERM != 143 {
		t.Fatalf("unexpected exit code values: %d %d %d %d %d %d", ExitSuccess, ExitUserError, ExitDocker, ExitPanic, ExitSIGINT, ExitSIGTERM)
	}
}

func TestWithSignalHandlerStopCancelsWithoutExit(t *testing.T) {
	oldNotify, oldStop, oldExit := signalNotify, signalStop, signalExit
	t.Cleanup(func() {
		signalNotify, signalStop, signalExit = oldNotify, oldStop, oldExit
	})

	var captured chan<- os.Signal
	var cleanupCalled atomic.Bool
	var exitCode atomic.Int64

	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) { captured = c }
	signalStop = func(c chan<- os.Signal) {
		if c != captured {
			t.Fatalf("unexpected signal channel")
		}
	}
	signalExit = func(code int) { exitCode.Store(int64(code)) }

	ctx, stop := WithSignalHandler(context.Background(), func() { cleanupCalled.Store(true) })
	stop()

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("context was not canceled by stop()")
	}
	if cleanupCalled.Load() {
		t.Fatal("cleanup should not run when stop() is called without a signal")
	}
	if exitCode.Load() != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode.Load())
	}
}

func TestWithSignalHandlerInterruptRunsCleanupAndExit(t *testing.T) {
	oldNotify, oldStop, oldExit := signalNotify, signalStop, signalExit
	t.Cleanup(func() {
		signalNotify, signalStop, signalExit = oldNotify, oldStop, oldExit
	})

	var captured chan<- os.Signal
	var cleanupCalled atomic.Bool
	done := make(chan struct{})
	exitCode := int32(0)

	signalNotify = func(c chan<- os.Signal, _ ...os.Signal) { captured = c }
	signalStop = func(chan<- os.Signal) {}
	signalExit = func(code int) {
		atomic.StoreInt32(&exitCode, int32(code))
		close(done)
	}

	_, stop := WithSignalHandler(context.Background(), func() { cleanupCalled.Store(true) })
	defer stop()
	captured <- os.Interrupt

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("signal exit path did not run")
	}
	if !cleanupCalled.Load() {
		t.Fatal("cleanup should run on interrupt")
	}
	if got := atomic.LoadInt32(&exitCode); int(got) != ExitSIGINT {
		t.Fatalf("exit code = %d, want %d", got, ExitSIGINT)
	}
}

func TestBuildCrashReportSanitizesArgs(t *testing.T) {
	oldNow, oldStack := crashNow, crashStackFn
	t.Cleanup(func() {
		crashNow, crashStackFn = oldNow, oldStack
	})

	crashNow = func() time.Time { return time.Unix(1_700_000_000, 0).UTC() }
	crashStackFn = func() []byte { return []byte("goroutine 1 [running]:\n/tmp/secret/path/main.go:10 +0x1\n") }

	report := buildCrashReport("v0.1.0", []string{"concave", "install", "--token=secret"}, errors.New("boom"))
	if !strings.Contains(report, "Command:  concave install") {
		t.Fatalf("report missing sanitized command: %q", report)
	}
	if strings.Contains(report, "--token=secret") {
		t.Fatalf("report leaked full arguments: %q", report)
	}
	if strings.Contains(report, "/tmp/secret/path/") {
		t.Fatalf("report leaked full stack path: %q", report)
	}
}

func TestAppendCrashReportFallsBackToTmp(t *testing.T) {
	oldOpen := crashOpen
	t.Cleanup(func() { crashOpen = oldOpen })

	home := t.TempDir()
	t.Setenv("HOME", home)

	crashOpen = func(path string, flag int, perm os.FileMode) (*os.File, error) {
		if strings.Contains(path, filepath.Join("gradient", "logs", "concave.log")) {
			return nil, os.ErrPermission
		}
		return os.OpenFile(path, flag, perm)
	}

	path, err := appendCrashReport("v0.1.0", []string{"concave", "install"}, "boom")
	if err != nil {
		t.Fatalf("appendCrashReport() error = %v", err)
	}
	if !strings.HasPrefix(path, filepath.Join("/tmp", "concave-crash-")) {
		t.Fatalf("fallback path = %q", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if !strings.Contains(string(data), "concave crash report") {
		t.Fatalf("unexpected crash log: %q", string(data))
	}
	_ = os.Remove(path)
}

func TestLockReportsHolderInfo(t *testing.T) {
	oldNow := lockNow
	t.Cleanup(func() { lockNow = oldNow })

	home := t.TempDir()
	t.Setenv("HOME", home)
	lockNow = func() time.Time { return time.Unix(1_700_000_000, 0).UTC() }

	unlock, err := Lock("install")
	if err != nil {
		t.Fatalf("Lock() error = %v", err)
	}
	defer unlock()

	_, err = Lock("update")
	if err == nil {
		t.Fatal("expected second lock attempt to fail")
	}
	if !strings.Contains(err.Error(), "Held by:") || !strings.Contains(err.Error(), "concave install") {
		t.Fatalf("unexpected lock error: %v", err)
	}
}
