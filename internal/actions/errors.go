package actions

import (
	"errors"
	"fmt"
)

var (
	ErrCancelled    = errors.New("cancelled")
	ErrNotInstalled = errors.New("xray not installed")
)

// ActionError represents a structured error with a hint.
type ActionError struct {
	Message string
	Hint    string
	Err     error
}

func (e *ActionError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s\n%s", e.Message, e.Hint)
	}
	return e.Message
}

func (e *ActionError) Unwrap() error {
	return e.Err
}

func NewActionError(message, hint string) *ActionError {
	return &ActionError{Message: message, Hint: hint}
}

func WrapError(err error, message, hint string) *ActionError {
	return &ActionError{Message: message, Hint: hint, Err: err}
}

func NotInstalledError() *ActionError {
	return &ActionError{
		Message: "xray is not installed",
		Hint:    "Run 'nhserver install' first",
		Err:     ErrNotInstalled,
	}
}
