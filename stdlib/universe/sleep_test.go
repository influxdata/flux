package universe

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/flux/values"
)

func TestSleep(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		myval := values.NewString("myvalue")
		args := values.NewObjectWithValues(
			map[string]values.Value{
				"v":        myval,
				"duration": values.NewDuration(values.Duration(time.Microsecond)),
			},
		)
		v, err := sleepFunc.Call(ctx, args)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if want, got := myval, v; !want.Equal(got) {
			t.Fatalf("unexpected value -want/+got:\n\t- %#v\n\t+ %#v", want, got)
		}
	})

	t.Run("Interrupted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
		defer cancel()

		myval := values.NewString("myvalue")
		args := values.NewObjectWithValues(
			map[string]values.Value{
				"v":        myval,
				"duration": values.NewDuration(values.Duration(200 * time.Millisecond)),
			},
		)
		_, err := sleepFunc.Call(ctx, args)
		if want, got := context.DeadlineExceeded, err; want != got {
			t.Fatalf("unexpected error -want/+got:\n\t- %v\n\t+ %v", want, got)
		}
	})
}
