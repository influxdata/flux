package experimental_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/runtime"
)

func TestObjectKeys(t *testing.T) {
	script := `
import "experimental"
import "internal/testutil"

o = {a: 1, b: 2, c: 3}
experimental.objectKeys(o: o) == ["a", "b", "c"] or testutil.fail()
`
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	if _, _, err := runtime.Eval(ctx, script); err != nil {
		t.Fatal("evaluation of objectKeys failed: ", err)
	}
}
