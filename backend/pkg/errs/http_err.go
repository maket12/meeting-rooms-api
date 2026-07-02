package errs

import "errors"

var (
	ErrInvalidJSON       = errors.New("invalid json")
	ErrInvalidIdentifier = errors.New("invalid identifier format")
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidDate       = errors.New("invalid input: date is empty or invalid")
	ErrInvalidPage       = errors.New("invalid input: page/page_size is not integer")
)

type OutErr struct {
	Code    int
	Message string
	Reason  error
}

func NewOutError(code int, msg string, reason error) *OutErr {
	return &OutErr{
		Code:    code,
		Message: msg,
		Reason:  reason,
	}
}
