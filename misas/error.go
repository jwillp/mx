package misas

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
)

// ErrorKindBadLogic indicates a problem with the programming logic.
// Reserved for unexpected conditions, API misuse, or bugs.
var ErrorKindBadLogic ErrorKind = "bad_logic"

// ErrorKindInternal is a catch-all for unknown errors.
var ErrorKindInternal ErrorKind = "internal"

// ErrorKindInvalid is used when a request fails validation.
var ErrorKindInvalid ErrorKind = "invalid"

// ErrorKindNotFound is used when a requested resource was not found.
var ErrorKindNotFound ErrorKind = "not_found"

// ErrorKindConflict is used when a request cannot be completed due to a conflict
// in the current state of an entity (e.g., concurrent edits, preexisting entity etc.)
var ErrorKindConflict ErrorKind = "conflict"

// ErrorKindTimeout is used when a request times out.
// Can be more specific than standard Go timeout errors.
var ErrorKindTimeout ErrorKind = "timeout"

// ErrorKindUnauthenticated is used when a request fails due to missing authentication.
var ErrorKindUnauthenticated ErrorKind = "unauthenticated"

// ErrorKindUnauthorized is used when a request fails due to insufficient authorization.
var ErrorKindUnauthorized ErrorKind = "unauthorized"

// ErrorKindNotImplemented is used when a request type is not yet implemented.
var ErrorKindNotImplemented ErrorKind = "not_implemented"

// ErrorKind represents the type of error from technical perspective. It relates to how
// an error should be interpreted in a technical context.
type ErrorKind string

// ErrorCode represents the type of error from a domain/business-specific standpoint.
// It relates to how an error should be interpreted in a business/domain context.
type ErrorCode string

var ErrBadLogic = Error{kind: ErrorKindBadLogic}.WithMessage("a programming logic error occurred")
var ErrInternal = NewError(ErrorKindInternal).WithMessage("an internal error occurred")
var ErrInvalid = NewError(ErrorKindInvalid).WithMessage("request failed validation")
var ErrNotFound = NewError(ErrorKindNotFound).WithMessage("the requested resource was not found")
var ErrConflict = NewError(ErrorKindConflict).WithMessage("request could not be completed due to a conflict")
var ErrTimeout = NewError(ErrorKindTimeout).WithMessage("request timed out")
var ErrUnauthenticated = NewError(ErrorKindUnauthenticated).WithMessage("request requires authentication")
var ErrUnauthorized = NewError(ErrorKindUnauthorized).WithMessage("request requires authorization")
var ErrNotImplemented = NewError(ErrorKindNotImplemented).WithMessage("request not implemented yet")

func ErrorHasCode(err error, code ErrorCode) bool {
	var e Error
	if !errors.As(err, &e) {
		return false
	}

	return e.Code() == code
}

func ErrorHasKind(err error, kind ErrorKind) bool {
	var e Error
	if !errors.As(err, &e) {
		return false
	}

	return e.Kind() == kind
}

//nolint:errname
type Error struct {
	kind    ErrorKind
	code    ErrorCode
	message string
	cause   error
}

func NewError(kind ErrorKind) Error {
	if kind == "" {
		return ErrBadLogic.WithMessage("cannot create error with empty kind")
	}

	return Error{kind: kind}
}

// NewInternalErrorFrom converts a standard error into an Error with support for common
// stdlib error types like os.IsNotExist, os.IsTimeout, context.DeadlineExceeded etc.
func NewInternalErrorFrom(err error) Error {
	var me Error
	if errors.As(err, &me) {
		return me
	}

	if os.IsTimeout(err) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return ErrTimeout.WithCause(err).WithMessage(err.Error())
	}

	if err, ok := err.(net.Error); ok && err.Timeout() {
		return ErrTimeout.WithCause(err).WithMessage(err.Error())
	}

	if os.IsNotExist(err) {
		return ErrNotFound.WithCause(err).WithMessage(err.Error())
	}

	if os.IsExist(err) {
		return ErrConflict.WithCause(err).WithMessage(err.Error())
	}

	return ErrInternal.WithCause(err).WithMessage(err.Error())
}

// Kind represents the type of error from technical perspective. It relates to how
// an error should be interpreted in a technical context.
func (e Error) Kind() ErrorKind { return e.kind }
func (e Error) Code() ErrorCode { return e.code }
func (e Error) Cause() error    { return e.cause }

func (e Error) Error() string {
	msg := string(e.kind)
	if e.code != "" {
		msg += fmt.Sprintf("(%s)", e.code)
	}
	msg = fmt.Sprintf("[%s]", msg)
	if e.message != "" {
		msg += " " + e.message
	} else if e.cause != nil {
		msg = fmt.Sprintf("%s: %s", msg, e.cause.Error())
	}

	return msg
}

func (e Error) Is(target error) bool {
	targetErr, ok := target.(Error)
	if !ok {
		return false
	}

	if e.kind != targetErr.kind {
		return false
	}

	if e.code != "" && e.code != targetErr.code {
		return false
	}

	return true
}

func (e Error) MarshalJSON() ([]byte, error) {
	var cause string
	if e.cause != nil {
		cause = e.cause.Error()
	}

	return json.Marshal(struct {
		Message string `json:"message"`
		Kind    string `json:"kind"`
		Code    string `json:"code"`
		Cause   string `json:"cause"`
	}{
		Message: e.message,
		Kind:    string(e.kind),
		Code:    string(e.code),
		Cause:   cause,
	})
}

func (e Error) WithCause(err error) Error {
	e.cause = err
	return e
}

func (e Error) WithMessage(message string) Error {
	if message == "" && e.message != "" {
		return e // prevent error message obfuscation
	}
	e.message = message
	return e
}

func (e Error) WithAppendedMessage(message string) Error {
	if message == "" {
		return e
	}
	e.message = fmt.Sprintf("%s: %s", e.message, message)
	return e
}

func (e Error) WithPrependedMessage(message string) Error {
	if message == "" {
		return e
	}
	e.message = fmt.Sprintf("%s: %s", message, e.message)
	return e
}

func (e Error) WithCode(code ErrorCode) Error {
	e.code = code
	return e
}
