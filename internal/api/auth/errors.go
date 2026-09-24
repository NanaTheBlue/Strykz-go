package authapi

import "errors"

var (
	ErrUserLen          = errors.New("invalid username length")
	ErrInvChar          = errors.New("invalid characters in username")
	ErrPasswordLenShort = errors.New("password length is too short")
	ErrPasswordLenLong  = errors.New("password length is too long")
	ErrPasswordMatch    = errors.New("passwords do not match")
	ErrInvalidEmail     = errors.New("invalid email format")
)
