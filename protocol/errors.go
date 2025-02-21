package protocol

import "errors"

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrInvalidMagic   = errors.New("invalid magic number")
)