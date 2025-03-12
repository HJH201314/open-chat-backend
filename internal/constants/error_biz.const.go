package constants

type BizError struct {
	Code int
	Msg  string
}

var (
	ErrNoPermission = BizError{Code: 10001, Msg: "no permission"}
	ErrInvalidInput = BizError{Code: 10002, Msg: "invalid \"input\""}
)
