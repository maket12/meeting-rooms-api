package errs

type OutErr struct {
	Code    int
	CodeStr string
	Message string
	Reason  error
}

func NewOutError(code int, codeStr, msg string, reason error) *OutErr {
	return &OutErr{
		Code:    code,
		CodeStr: codeStr,
		Message: msg,
		Reason:  reason,
	}
}
