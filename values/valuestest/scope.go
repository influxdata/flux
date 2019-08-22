package valuestest

import (
	"context"
	"errors"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// ScopeComparer checks that two scopes are equal in both nesting and contents.
// Functions cannot be compared for equality so only function types are checked.
var ScopeComparer = cmp.Comparer(func(l, r values.Scope) bool {
	for {
		if l == nil && r == nil {
			return true
		}
		if l == nil && r != nil || l != nil && r == nil {
			return false
		}
		equal := true
		l.LocalRange(func(k string, lv values.Value) {
			rv, ok := r.LocalLookup(k)
			if lv.PolyType().Nature() == semantic.Function {
				// only compare functions by type
				equal = equal && ok && lv.PolyType().Equal(rv.PolyType())
			} else {
				equal = equal && ok && lv.Equal(rv)
			}
		})
		if !equal {
			return false
		}
		r.LocalRange(func(k string, rv values.Value) {
			_, ok := l.LocalLookup(k)
			equal = equal && ok
		})
		if !equal {
			return false
		}

		l = l.Pop()
		r = r.Pop()
	}
})

// NowScope generates scope with the prelude + the now option.
func NowScope() values.Scope {
	scope := flux.Prelude()
	scope.Set("now", values.NewFunction(
		"now",
		semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
			Return: semantic.Time,
		}),
		func(ctx context.Context, deps dependencies.Interface, args values.Object) (values.Value, error) {
			//Functions are only compared by type so the function body here is not important
			return nil, errors.New("NowScope was called")
		},
		false,
	))
	return scope
}
