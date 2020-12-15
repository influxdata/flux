package errors

import "errors"

// Is is a wrapper around the errors.Is function.
func Is(err, target error) bool {
	return errors.Is(err, target)
}
