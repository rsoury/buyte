package store

import (
	"errors"
	"strings"
)

func IsConnectionUnauthorized(err error) bool {
	return strings.Contains(err.Error(), "graphql: Not Authorized")
}
func IsConnectionInvalid(err error) bool {
	return strings.Contains(err.Error(), "graphql: One or more parameter values were invalid")
}

// ErrNotFound is a standard no found error
var ErrNotFound = errors.New("Not Found")
