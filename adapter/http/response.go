package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// Response is the standard envelope every /api/v1 endpoint responds with.
// Code is model.ErrCodeNone (0) on success; any other value identifies a
// model.ErrorCode and Message is that code's registered, human-readable
// description. Data carries the payload on success and is omitted on
// error.
type Response[T any] struct {
	Code    model.ErrorCode `json:"code"`
	Message string          `json:"message"`
	Data    T               `json:"data,omitempty"`
}

// errorStatusCodes maps a model.ErrorCode to the HTTP status it is
// reported with. Codes without an entry fall back to 500, so a newly added
// error code that hasn't been mapped yet still fails safely rather than
// reporting success.
var errorStatusCodes = map[model.ErrorCode]int{
	model.ErrCodeUnknown:         http.StatusInternalServerError,
	model.ErrCodeInvalidRequest:  http.StatusBadRequest,
	model.ErrCodeVehicleNotFound: http.StatusNotFound,
}

func statusForCode(code model.ErrorCode) int {
	if status, ok := errorStatusCodes[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

func message(code model.ErrorCode, messages ...string) string {
	if len(messages) == 0 {
		return model.MessageForCode(code)
	}
	return fmt.Sprintf("%s; %s", model.MessageForCode(code), strings.Join(messages, "; "))
}

// OK writes a 200 response wrapping data in the standard success envelope.
func OK(c *gin.Context, data any) {
	respondOK(c, http.StatusOK, data)
}

// Created writes a 201 response wrapping data in the standard success
// envelope.
func Created[T any](c *gin.Context, data T) {
	respondOK(c, http.StatusCreated, data)
}

func respondOK[T any](c *gin.Context, status int, data T) {
	c.JSON(status, Response[T]{
		Code:    model.ErrCodeNone,
		Message: message(model.ErrCodeNone),
		Data:    data,
	})
}

// Fail writes an error response for code. Its HTTP status and message are
// both derived from code, so controllers never choose them ad hoc and the
// same code always reports the same way everywhere.
func Fail(c *gin.Context, code model.ErrorCode, messages ...string) {
	c.JSON(statusForCode(code), Response[any]{
		Code:    code,
		Message: message(code, messages...),
	})
}

// RespondError writes err as a standard error response. If err wraps a
// *model.Error (see errors.As), its Code drives the response; otherwise it
// falls back to model.ErrCodeUnknown, so unexpected errors always produce a
// valid envelope instead of leaking internal details to the client.
func RespondError(c *gin.Context, err error) {
	var merr *model.Error
	if errors.As(err, &merr) {
		Fail(c, merr.Code)
		return
	}
	Fail(c, model.ErrCodeUnknown)
}
