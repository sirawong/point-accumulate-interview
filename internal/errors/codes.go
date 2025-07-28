package errors

import "errors"

var (
	ErrNotFound        = New("NOT_FOUND", "data not found")
	ErrInvalidArgument = New("INVALID_ARGUMENT", "invalid argument provided")
	ErrUnauthenticated = New("UNAUTHENTICATED", "authentication failed")
	ErrInternal        = New("INTERNAL_ERROR", "an internal errors occurred")
)

func New(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func (e *AppError) Wrap(err error) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Err:     err,
	}
}

func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: msg,
	}
}

func GetCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}

	return ErrInternal.Code
}
