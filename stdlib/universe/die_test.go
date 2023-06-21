package universe_test

import (
	"context"
	"testing"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/dependencies/dependenciestest"
	"github.com/InfluxCommunity/flux/dependency"
	"github.com/InfluxCommunity/flux/stdlib/universe"
	"github.com/InfluxCommunity/flux/values"
	"github.com/google/go-cmp/cmp"
)

func TestDie(t *testing.T) {
	t.Run("die test", func(t *testing.T) {
		dieFn := universe.Die()

		fluxArg := values.NewObjectWithValues(map[string]values.Value{"msg": values.NewString("this is an error message")})

		ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
		defer deps.Finish()
		_, got := dieFn.Call(ctx, fluxArg)

		if got == nil {
			t.Fatal("this function should produce an error")
		}

		want := &flux.Error{
			Code: codes.Invalid,
			Msg:  "this is an error message",
		}

		if !cmp.Equal(want, got) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(want, got))
		}
	})
}
