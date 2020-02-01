package valuestest

import (
	"context"
	"errors"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// ComparableScope is a representation of a Scope
// that is easily compared with the cmp package.
type ComparableScope struct {
	Values map[string]values.Value
	Child  *ComparableScope
}

// ScopeTransformer converts a scope to a ComparableScope.
var ScopeTransformer = cmp.Transformer("Scope", func(s values.Scope) *ComparableScope {
	sc := &ComparableScope{
		Values: make(map[string]values.Value),
		Child:  nil,
	}

	for {
		s.LocalRange(func(k string, v values.Value) {
			sc.Values[k] = v
		})
		s = s.Pop()
		if s != nil {
			sc = &ComparableScope{
				Values: make(map[string]values.Value),
				Child:  sc,
			}
		} else {
			break
		}
	}
	return sc
})

// NowScope generates scope with the prelude + the now option.
func NowScope() values.Scope {
	scope := flux.Prelude()
	scope.SetOption("universe", "now", values.NewFunction(
		"now",
		semantic.MustLookupBuiltinType("universe", "now"),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			//Functions are only compared by type so the function body here is not important
			return nil, errors.New("NowScope was called")
		},
		false,
	))
	return scope
}
