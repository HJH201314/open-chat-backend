// Package constants ## THIS FILE WAS GENERATED, DO NOT MODIFY. ##
package constants

type BizError struct {
	Msg      string
	HttpCode int
	BizCode  int
}

// Error 实现 error 接口
func (e BizError) Error() string {
	return e.Msg
}

var (
	BizErrNoPermission = BizError{HttpCode: 400, BizCode: 10001, Msg: "no permission"}
	BizErrNoRecord     = BizError{HttpCode: 400, BizCode: 10002, Msg: "no record"}
	BizErrOutdated     = BizError{HttpCode: 400, BizCode: 10003, Msg: "outdated"}
)
