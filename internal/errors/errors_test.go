package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

func TestErrorString(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		want string
	}{
		{
			name: "normal message",
			err:  errors.New(codes.Invalid, "expected message"),
			want: "expected message",
		},
		{
			name: "wrapped error",
			err:  errors.Wrap(stderrors.New("wrapped error"), codes.Invalid, "expected message"),
			want: "expected message: wrapped error",
		},
		{
			name: "wrapped error with no message",
			err:  errors.Wrap(stderrors.New("wrapped error"), codes.Invalid),
			want: "wrapped error",
		},
		{
			name: "code only",
			err:  errors.New(codes.NotFound),
			want: "not found",
		},
		{
			name: "formatted message",
			err:  errors.Newf(codes.Invalid, "formatted message %q", "string"),
			want: `formatted message "string"`,
		},
		{
			name: "wrapped error with formatted message",
			err:  errors.Wrapf(stderrors.New("wrapped error"), codes.Invalid, "formatted message %q", "string"),
			want: `formatted message "string": wrapped error`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := tt.err.Error(), tt.want; got != want {
				t.Errorf("unexpected error message -want/+got:\n\t- %s\n\t+ %s", want, got)
			}
		})
	}
}

func TestError(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		want *flux.Error
	}{
		{
			name: "code only",
			err:  errors.New(codes.NotFound),
			want: &flux.Error{
				Code: codes.NotFound,
			},
		},
		{
			name: "code and message",
			err:  errors.New(codes.NotFound, "source is missing"),
			want: &flux.Error{
				Code: codes.NotFound,
				Msg:  "source is missing",
			},
		},
		{
			name: "code and formatted message",
			err:  errors.Newf(codes.Invalid, "number must be greater than zero, was %d", -1),
			want: &flux.Error{
				Code: codes.Invalid,
				Msg:  "number must be greater than zero, was -1",
			},
		},
		{
			name: "code, message, and wrapped error",
			err:  errors.Wrap(stderrors.New("expected error"), codes.Unknown, "unknown error from external system"),
			want: &flux.Error{
				Code: codes.Unknown,
				Msg:  "unknown error from external system",
				Err:  stderrors.New("expected error"),
			},
		},
		{
			name: "code, formatted message, and wrapped error",
			err:  errors.Wrapf(stderrors.New("expected error"), codes.Unknown, "unknown error from %q", "influxdb"),
			want: &flux.Error{
				Code: codes.Unknown,
				Msg:  `unknown error from "influxdb"`,
				Err:  stderrors.New("expected error"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, want := tt.err.(*flux.Error), tt.want
			if got.Code != want.Code {
				t.Errorf("unexpected error code -want/+got:\n\t- %v\n\t+ %v", got.Code, want.Code)
			}
			if got.Msg != want.Msg {
				t.Errorf("unexpected error message -want/+got:\n\t- %v\n\t+ %v", got.Msg, want.Msg)
			}
			if got, want := errorString(got.Err), errorString(want.Err); got != want {
				t.Errorf("unexpected error message -want/+got:\n\t- %v\n\t+ %v", got, want)
			}
		})
	}
}

func errorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return "<nil>"
}
