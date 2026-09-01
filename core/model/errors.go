package model

import "fmt"

// ErrorCode identifies a specific, known application error condition. It is
// the "code" field of the API response envelope (see adapter/http).
//
// 0 always means success. Non-zero codes are grouped by resource in bands
// of 1000: codes 1000-1999 belong to the first resource, 2000-2999 to the
// second, and so on — e.g. 1001 is the first error for resource 1, 1002 the
// second. Codes below 1000 are reserved for generic, resource-agnostic
// errors.
type ErrorCode int

const (
	// ErrCodeNone means the request succeeded.
	ErrCodeNone ErrorCode = 0

	// ErrCodeUnknown is the fallback for errors that don't map to a more
	// specific, registered code.
	ErrCodeUnknown ErrorCode = 1

	// ErrCodeInvalidRequest means the request itself was malformed (e.g.
	// invalid query parameters) and couldn't be processed.
	ErrCodeInvalidRequest ErrorCode = 2

	// Vehicle errors (resource 1).
	ErrCodeVehicleNotFound ErrorCode = 1001
)

// errorMessages maps each ErrorCode to the human-readable, English message
// returned in the API response's "message" field.
var errorMessages = map[ErrorCode]string{
	ErrCodeNone:            "success",
	ErrCodeUnknown:         "an unexpected error occurred",
	ErrCodeInvalidRequest:  "invalid request",
	ErrCodeVehicleNotFound: "vehicle not found",
}

// MessageForCode returns the registered message for code, falling back to
// ErrCodeUnknown's message if code isn't registered.
func MessageForCode(code ErrorCode) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return errorMessages[ErrCodeUnknown]
}

// Error is an application error identified by a stable ErrorCode. Its
// message is always the registered, humanized description for Code —
// callers should wrap an underlying cause (e.g. a database error) via
// NewError rather than invent ad hoc messages, so API responses stay
// consistent and never leak internal details.
type Error struct {
	Code ErrorCode
	Err  error
}

// NewError builds an *Error for code, optionally wrapping cause for
// logging/debugging. cause is never exposed in Error() beyond wrapping and
// is not part of the registered, user-facing message.
func NewError(code ErrorCode, cause error) *Error {
	return &Error{Code: code, Err: cause}
}

func (e *Error) Error() string {
	msg := MessageForCode(e.Code)
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

func (e *Error) Unwrap() error {
	return e.Err
}
