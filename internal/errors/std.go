package errors

import "errors"

// Is is a wrapper around the errors.Is function.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As is a wrapper around the errors.As function.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
