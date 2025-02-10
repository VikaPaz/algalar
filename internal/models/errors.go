package models

import "errors"

var (
	ErrLoadEnvFailed              = errors.New("failed to load environment")
	ErrConnectionDBFailed         = errors.New("failed to connect to database")
	ErrServerFailed               = errors.New("failed to connect to server")
	ErrClientFailed               = errors.New("failed to create client")
	ErrNoContent                  = errors.New("failed to provide content: not exists")
	ErrAlreadyExists              = errors.New("failed to create: already exists")
	ErrInvalidRefreshToken        = errors.New("invalid refresh token")
	ErrLoginOrPassword            = errors.New("login and password are required")
	ErrDriverNotFound             = errors.New("driver not found")
	ErrInvalidInput               = errors.New("invalid input")
	ErrFailedToExecuteQuery       = errors.New("failed to execute query")
	ErrFailedToProcessRow         = errors.New("failed to process row while retrieving data")
	ErrRowsIterationError         = errors.New("error while iterating rows")
	ErrInvalidContext             = errors.New("invalid context")
	ErrUserUpdateFailed           = errors.New("failed to update user")
	ErrUnauthorizedRequest        = errors.New("unauthorized request")
	ErrFailedToFetchBreakages     = errors.New("failed to fetch breakages for the car")
	ErrFailedToEncodeResponse     = errors.New("failed to encode response")
	ErrInvalidRequestBody         = errors.New("invalid request body")
	ErrInvalidPointFormat         = errors.New("invalid point format, must contain exactly two coordinates")
	ErrFailedToCreateBreakage     = errors.New("failed to create breakage")
	ErrFailedToCreateNotification = errors.New("failed to create notification")
)
