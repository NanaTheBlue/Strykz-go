package authapi

import "errors"

var (
	ErrUserLen          = errors.New("Invalid Username Length")
	ErrInvChar          = errors.New("Invalid Characters in Username")
	ERRPassWordLenShort = errors.New("password length is to short")
	ERRPassWordLenLong  = errors.New("password length is to long")
	ERRPassWordMatch    = errors.New("passwords dont match")
	ERRInvEmailF        = errors.New("invalid email format")
)
