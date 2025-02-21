package server

import "errors"

var (
	ErrNoAvailableMethods = errors.New("no available methods")
	ErrServiceNotFound    = errors.New("service not found")
	ErrMethodNotFound     = errors.New("method not found")
)