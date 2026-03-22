package system

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	signalNotify = signal.Notify
	signalStop   = signal.Stop
	signalExit   = os.Exit
)

// WithSignalHandler wraps a context with SIGINT/SIGTERM cancellation.
func WithSignalHandler(parent context.Context, cleanup func()) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	ctx, cancel := context.WithCancel(parent)
	signals := make(chan os.Signal, 1)
	done := make(chan struct{})
	var once sync.Once

	stop := func() {
		once.Do(func() {
			signalStop(signals)
			cancel()
			close(done)
		})
	}

	signalNotify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-signals:
			if sig == nil {
				return
			}
			Debug("signal received", "signal", sig.String())
			cancel()
			if cleanup != nil {
				Debug("signal cleanup started", "signal", sig.String())
				cleanup()
				Debug("signal cleanup finished", "signal", sig.String())
			}
			signalStop(signals)
			switch sig {
			case os.Interrupt:
				signalExit(ExitSIGINT)
			default:
				signalExit(ExitSIGTERM)
			}
		case <-done:
			return
		}
	}()

	return ctx, stop
}
