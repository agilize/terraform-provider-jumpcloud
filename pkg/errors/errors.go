package errors

import (
	"errors"
	"fmt"
)

// Standard error types
var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists is returned when a resource already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput is returned when input is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrPermissionDenied is returned when permission is denied
	ErrPermissionDenied = errors.New("permission denied")

	// ErrAuthentication is returned when authentication fails
	ErrAuthentication = errors.New("authentication failed")

	// ErrInternal is returned when an internal error occurs
	ErrInternal = errors.New("internal error")

	// ErrTimeout is returned when a timeout occurs
	ErrTimeout = errors.New("timeout")
)

// Error represents a standard error with a code and wrapped error
type Error struct {
	// Code is a machine-readable error code
	Code string

	// Message is a human-readable message
	Message string

	// Err is the original error
	Err error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %v", e.Code, e.Err)
}

// Unwrap returns the wrapped error
func (e *Error) Unwrap() error {
	return e.Err
}

// Is reports whether err is an instance of target
func (e *Error) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// NewNotFoundError returns a new not found error
func NewNotFoundError(format string, args ...interface{}) *Error {
	return &Error{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf(format, args...),
		Err:     ErrNotFound,
	}
}

// NewAlreadyExistsError returns a new already exists error
func NewAlreadyExistsError(format string, args ...interface{}) *Error {
	return &Error{
		Code:    "ALREADY_EXISTS",
		Message: fmt.Sprintf(format, args...),
		Err:     ErrAlreadyExists,
	}
}

// NewInvalidInputError returns a new invalid input error
func NewInvalidInputError(format string, args ...interface{}) *Error {
	return &Error{
		Code:    "INVALID_INPUT",
		Message: fmt.Sprintf(format, args...),
		Err:     ErrInvalidInput,
	}
}

// NewPermissionDeniedError returns a new permission denied error
func NewPermissionDeniedError(format string, args ...interface{}) *Error {
	return &Error{
		Code:    "PERMISSION_DENIED",
		Message: fmt.Sprintf(format, args...),
		Err:     ErrPermissionDenied,
	}
}

// NewAuthenticationError returns a new authentication error
func NewAuthenticationError(format string, args ...interface{}) *Error {
	return &Error{
		Code:    "AUTHENTICATION_FAILED",
		Message: fmt.Sprintf(format, args...),
		Err:     ErrAuthentication,
	}
}

// NewInternalError returns a new internal error
func NewInternalError(format string, args ...interface{}) *Error {
	return &Error{
		Code:    "INTERNAL_ERROR",
		Message: fmt.Sprintf(format, args...),
		Err:     ErrInternal,
	}
}

// Wrapper functions for easier error checking
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

func IsPermissionDenied(err error) bool {
	return errors.Is(err, ErrPermissionDenied)
}

func IsAuthentication(err error) bool {
	return errors.Is(err, ErrAuthentication)
}

func IsInternal(err error) bool {
	return errors.Is(err, ErrInternal)
}

func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}
