package universe_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
)

func TestDie(t *testing.T) {
	t.Run("die test", func(t *testing.T) {
		dieFn := universe.Die()

		fluxArg := values.NewObjectWithValues(map[string]values.Value{"msg": values.NewString("this is an error message")})

		_, got := dieFn.Call(dependenciestest.Default().Inject(context.Background()), fluxArg)

		if got == nil {
			t.Fatal("this function should produce an error")
		}

		want := &flux.Error{
			Code: codes.Internal,
			Msg:  "this is an error message",
		}

		if !cmp.Equal(want, got) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(want, got))
		}
	})
}
