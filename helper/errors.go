package helper

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	CodeValidation       = "VALIDATION_ERROR"
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeUnsupportedMedia = "UNSUPPORTED_MEDIA_TYPE"
	CodeNotAcceptable    = "NOT_ACCEPTABLE"
	CodeTooManyRequests  = "TOO_MANY_REQUESTS"
	CodeInternal         = "INTERNAL_ERROR"
	CodeUnavailable      = "SERVICE_UNAVAILABLE"
)

// Struct (class) untuk error aplikasi
type AppError struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	cause   error
}

// Method untuk struct error aplikasi:
// func Error() string
//
//	mengembalikan string keterangan error lengkap
//	error Code, Message, Cause
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// func Unwrap() error
//
//	mengembalikan error Cause
func (e *AppError) Unwrap() error {
	return e.cause
}

// Helper
//   Mengembalikan AppError
//   Menggunakan Konstanta Error Code di atas

func BadRequest(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusBadRequest,
		Code:    CodeBadRequest,
		Message: message,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnauthorized,
		Code:    CodeUnauthorized,
		Message: message,
	}
}

func Forbidden(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusForbidden,
		Code:    CodeForbidden,
		Message: message,
	}
}

func NotFound(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusNotFound,
		Code:    CodeNotFound,
		Message: message,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusConflict,
		Code:    CodeConflict,
		Message: message,
	}
}

func Validation(fields map[string]string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnprocessableEntity,
		Code:    CodeValidation,
		Message: "validasi gagal",
		Fields:  fields,
	}
}

func NotAcceptable(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusNotAcceptable,
		Code:    CodeNotAcceptable,
		Message: message,
	}
}

func UnsuportedMediaType(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnsupportedMediaType,
		Code:    CodeUnsupportedMedia,
		Message: message,
	}
}

func TooManyRequest(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusTooManyRequests,
		Code:    CodeTooManyRequests,
		Message: message,
	}
}

func ServiceUnavailable(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusServiceUnavailable,
		Code:    CodeUnavailable,
		Message: message,
	}
}

// Internal server error
//
//	hanya digunakan untuk log internal
//	tidak mengembalikan pesan berisi informasi detail ke client
func Internal(cause error) *AppError {
	return &AppError{
		Status:  fiber.StatusInternalServerError,
		Code:    CodeInternal,
		Message: "something wrong happened!",
		cause:   cause,
	}
}
