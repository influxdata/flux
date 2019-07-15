// +build go1.13

package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
)

func TestUnwrap(t *testing.T) {
	err := errors.Wrap(memory.LimitExceededError{
		Limit:     1024,
		Allocated: 896,
		Wanted:    1152,
	}, codes.ResourceExhausted)

	var limitExceededErr memory.LimitExceededError
	if stderrors.As(err, &limitExceededErr) {
		if got, want := limitExceededErr.Limit, int64(1024); got != want {
			t.Errorf("unexpected wrapped error's memory limit -want/+got\n\t- %d\n\t+ %d", got, want)
		}
		if got, want := limitExceededErr.Allocated, int64(896); got != want {
			t.Errorf("unexpected wrapped error's memory allocated -want/+got\n\t- %d\n\t+ %d", got, want)
		}
		if got, want := limitExceededErr.Wanted, int64(1152); got != want {
			t.Errorf("unexpected wrapped error's memory wanted -want/+got\n\t- %d\n\t+ %d", got, want)
		}
	} else {
		t.Error("expected unwrapping error to work")
	}
}
