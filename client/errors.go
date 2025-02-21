package client

import (
	"errors"
)

var (
	ErrShutdown = errors.New("connection is shut down")
)

type ErrorString string

func (e ErrorString) Error() string {
	return string(e)
}

func ErrorFromString(s string) error {
	if s == "" {
		return nil
	}
	return ErrorString(s)
}