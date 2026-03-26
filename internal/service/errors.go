package service

import "fmt"

// NotEnabledError is returned when a database backend is not enabled in config.
type NotEnabledError struct {
	Backend string
}

func (e *NotEnabledError) Error() string {
	return fmt.Sprintf("%s not enabled", e.Backend)
}

// ErrNotEnabled is a convenience constructor for NotEnabledError.
func ErrNotEnabled(backend string) error {
	return &NotEnabledError{Backend: backend}
}

// IsNotEnabled checks whether the error signals a disabled backend.
func IsNotEnabled(err error) bool {
	_, ok := err.(*NotEnabledError)
	return ok
}
