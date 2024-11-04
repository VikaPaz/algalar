package models

import "errors"

var (
	ErrLoadEnvFailed      = errors.New("failed to load environment")
	ErrConnectionDBFailed = errors.New("failed to connect to database")
	ErrServerFailed       = errors.New("failed to connect to server")
	ErrClientFailed       = errors.New("failed to create client")
)
