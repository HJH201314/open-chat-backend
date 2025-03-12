// Package constants ## THIS FILE WAS GENERATED, DO NOT MODIFY. ##
package constants

type BizError struct {
	Code int
	Msg  string
}

var (
	ErrNoPermission = BizError{Code: 10001, Msg: "no permission"}
)
