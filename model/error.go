package model

import (
	"errors"
)

var ErrInvalidPostID = errors.New("invalid post id")
var ErrInvalidParameter = errors.New("invalid parameter")
var ErrInvalidRoute = errors.New("no route was found matching the URL and request method")
