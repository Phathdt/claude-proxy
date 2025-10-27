package errors

import "net/http"

// AppError represents application-level errors with HTTP context
type AppError interface {
	error
	StatusCode() int
	ErrorCode() string
	Message() string
	Details() string
}

// BaseAppError implements AppError interface
type BaseAppError struct {
	Code       string `json:"code"`
	Msg        string `json:"message"`
	Detail     string `json:"details,omitempty"`
	HttpStatus int    `json:"-"`
}

func (e *BaseAppError) Error() string {
	if e.Detail != "" {
		return e.Msg + ": " + e.Detail
	}
	return e.Msg
}

func (e *BaseAppError) StatusCode() int {
	return e.HttpStatus
}

func (e *BaseAppError) ErrorCode() string {
	return e.Code
}

func (e *BaseAppError) Message() string {
	return e.Msg
}

func (e *BaseAppError) Details() string {
	return e.Detail
}

// Common error constructors
func NewBadRequestError(code, message, details string) AppError {
	return &BaseAppError{
		Code:       code,
		Msg:        message,
		Detail:     details,
		HttpStatus: http.StatusBadRequest,
	}
}

func NewNotFoundError(code, message, details string) AppError {
	return &BaseAppError{
		Code:       code,
		Msg:        message,
		Detail:     details,
		HttpStatus: http.StatusNotFound,
	}
}

func NewConflictError(code, message, details string) AppError {
	return &BaseAppError{
		Code:       code,
		Msg:        message,
		Detail:     details,
		HttpStatus: http.StatusConflict,
	}
}

func NewInternalError(code, message, details string) AppError {
	return &BaseAppError{
		Code:       code,
		Msg:        message,
		Detail:     details,
		HttpStatus: http.StatusInternalServerError,
	}
}

func NewValidationError(details string) AppError {
	return &BaseAppError{
		Code:       "VALIDATION_ERROR",
		Msg:        "Invalid request format or missing required fields",
		Detail:     details,
		HttpStatus: http.StatusBadRequest,
	}
}

func NewParseError(details string) AppError {
	return &BaseAppError{
		Code:       "PARSE_ERROR",
		Msg:        "Invalid parameter format",
		Detail:     details,
		HttpStatus: http.StatusBadRequest,
	}
}

func NewDatabaseError(operation, details string) AppError {
	return &BaseAppError{
		Code:       "DATABASE_ERROR",
		Msg:        "Database operation failed",
		Detail:     "Operation: " + operation + ". " + details,
		HttpStatus: http.StatusInternalServerError,
	}
}

func NewUnauthorizedError(details string) AppError {
	return &BaseAppError{
		Code:       "UNAUTHORIZED",
		Msg:        "Unauthorized access",
		Detail:     details,
		HttpStatus: http.StatusUnauthorized,
	}
}

func NewServiceUnavailableError(details string) AppError {
	return &BaseAppError{
		Code:       "SERVICE_UNAVAILABLE",
		Msg:        "Service temporarily unavailable",
		Detail:     details,
		HttpStatus: http.StatusServiceUnavailable,
	}
}

func NewInternalServerError(details string) AppError {
	return &BaseAppError{
		Code:       "INTERNAL_SERVER_ERROR",
		Msg:        "Internal server error",
		Detail:     details,
		HttpStatus: http.StatusInternalServerError,
	}
}

func NewRequestTimeoutError(details string) AppError {
	return &BaseAppError{
		Code:       "REQUEST_TIMEOUT",
		Msg:        "Request timeout",
		Detail:     details,
		HttpStatus: http.StatusRequestTimeout,
	}
}
