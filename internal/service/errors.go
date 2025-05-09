package service

import "errors"

var (
	ErrShortCodeTaken = errors.New("short code already taken")
)
