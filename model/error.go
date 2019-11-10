package model

import (
	"errors"
)

// ErrInvalidPostID for invalid post id error
var ErrInvalidPostID = errors.New("invalid post id")

// ErrInvalidParameter for invalid parameter error
var ErrInvalidParameter = errors.New("invalid parameter")

// ErrInvalidRoute for invalid route error
var ErrInvalidRoute = errors.New("no route was found matching the URL and request method")
