package valuestest

import (
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/runtime"
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
	var sc *ComparableScope = nil
	for {
		if s != nil {
			sc = &ComparableScope{
				Values: make(map[string]values.Value),
				Child:  sc,
			}
			s.LocalRange(func(k string, v values.Value) {
				sc.Values[k] = v
			})
			s = s.Pop()
		} else {
			break
		}
	}
	return sc
})

// Scope returns a scope that contains the prelude.
func Scope() values.Scope {
	return runtime.Prelude()
}
