package tasks_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb/tasks"
	"github.com/influxdata/flux/values"
)

func TestLastSuccess(t *testing.T) {
	t.Run("lastSuccess test", func(t *testing.T) {
		fluxArg := values.NewObjectWithValues(map[string]values.Value{"orTime": values.NewTime(100)})

		_, got := tasks.LastSuccessFunction.Call(dependenciestest.Default().Inject(context.Background()), fluxArg)

		if got == nil {
			t.Fatal("lastSuccess should produce an error")
		}

		want := &flux.Error{
			Code: codes.Unimplemented,
			Msg:  "This function is not yet implemented.",
		}

		if !cmp.Equal(want, got) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(want, got))
		}
	})
}
