package errors

import "errors"

var (
	ErrInvalidTermiiConfig   = errors.New("Invalid termii config")
	ErrTermiiSendFailed      = errors.New("SMS delivery failed")
	ErrInvalidPasswordLength = errors.New("Password must be at least 8 characters long")
	ErrInvalidPassword       = errors.New("Password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	ErrInvalidPhoneNumber    = errors.New("Invalid nigerian phone number")
	ErrUserAlreadyExists     = errors.New("User already exists")
	ErrPasswordMismatch      = errors.New("Password and confirm password do not match")
	ErrInvalidRequestBody    = errors.New("Invalid request body")
	ErrUnauthorized          = errors.New("Unauthorized")
	ErrMissingSession        = errors.New("Your session has expired. Please log in again.")
	ErrInvalidCredentials    = errors.New("Invalid email or password")
)
