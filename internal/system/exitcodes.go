package system

import "fmt"

const (
	ExitSuccess   = 0
	ExitUserError = 1
	ExitDocker    = 2
	ExitPanic     = 3
	ExitSIGINT    = 130
	ExitSIGTERM   = 143
)

type codedError struct {
	code int
	err  error
}

func (e codedError) Error() string {
	if e.err == nil {
		return fmt.Sprintf("exit %d", e.code)
	}
	return e.err.Error()
}

func (e codedError) Unwrap() error {
	return e.err
}

func (e codedError) ExitCode() int {
	return e.code
}

// WithExitCode wraps an error so callers can preserve a stable process exit code.
func WithExitCode(err error, code int) error {
	if err == nil {
		return nil
	}
	if current, ok := err.(interface{ ExitCode() int }); ok && current.ExitCode() == code {
		return err
	}
	return codedError{code: code, err: err}
}

// NewExitError creates an error with a fixed process exit code.
func NewExitError(code int, format string, args ...any) error {
	return codedError{code: code, err: fmt.Errorf(format, args...)}
}
