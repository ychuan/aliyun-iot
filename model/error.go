package model

import (
	"errors"
)

var ErrInvalidURI = errors.New("invalid URI")
var ErrNotFound = errors.New("not found")
var ErrInvalidParameter = errors.New("invalid parameter")
