package domain

import "errors"

var (
	ErrInvalidSignal   = errors.New("invalid trading signal")
	ErrExecutionFailed = errors.New("trade execution failed")
)
