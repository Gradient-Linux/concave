package logx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type noopHandler struct{}

func (noopHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (noopHandler) Handle(context.Context, slog.Record) error { return nil }
func (noopHandler) WithAttrs([]slog.Attr) slog.Handler        { return noopHandler{} }
func (noopHandler) WithGroup(string) slog.Handler             { return noopHandler{} }

var logger = slog.New(noopHandler{})

func Init(verbose bool) {
	if !verbose {
		logger = slog.New(noopHandler{})
		return
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Debugf(format string, args ...any) {
	logger.Debug(fmt.Sprintf(format, args...))
}

func Logger() *slog.Logger {
	return logger
}
