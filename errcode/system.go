package errcode

import "errors"

type ErrMsg struct {
	Error string `json:"error"`
}

func (err *ErrMsg) Err() error {
	return errors.New(err.Error)
}

var (
	ErrDataNotExists    = errors.New("data not exists")
	ErrSystemExecFailed = errors.New("program execution error")
)
