package system

import "github.com/Gradient-Linux/concave/internal/logx"

// Logger is the package-level structured logger.
var Logger = logx.Logger()

// InitLogger configures the shared logger once at process start.
func InitLogger(verbose bool) {
	logx.Init(verbose)
	Logger = logx.Logger()
}

// Debug emits a structured debug message when verbose logging is enabled.
func Debug(msg string, args ...any) {
	logx.Debug(msg, args...)
}

// Debugf emits a formatted debug message when verbose logging is enabled.
func Debugf(format string, args ...any) {
	logx.Debugf(format, args...)
}
