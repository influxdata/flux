package compiler

import (
	"context"

	"github.com/influxdata/flux/semantic"
)

// staticCast will coerce the src type into the dest type and return the
// register where this value is.
//
// This function is mostly only used when the src and the dest are
// records. If they are records, we have to coerce them into the same
// structure.
//
// This does NOT resolve recursive structures. Recursive structures
// should have already been resolved in another location.
func (c *compiledFn) staticCast(dest, src semantic.MonoType, reg int) int {
	if src.Nature() != semantic.Object {
		return reg
	}

	n, _ := src.NumProperties()
	labels := make([]int, n)
	for i := 0; i < n; i++ {
		prop, _ := src.RecordProperty(i)
		labels[i], _ = findPropertyIndex(prop.Name(), dest)
	}

	e := &staticCastEvaluator{
		evaluator: evaluator{
			t:   dest,
			ret: c.scope.Push(NewRecord(dest)),
		},
		in:     reg,
		labels: labels,
	}
	c.body = append(c.body, e)
	return e.ret
}

type staticCastEvaluator struct {
	evaluator
	in     int
	labels []int
}

func (s *staticCastEvaluator) Eval(ctx context.Context, scope []Value, origin int) error {
	r, ret := scope[s.in], scope[s.ret]
	for i, v := range s.labels {
		if v < 0 {
			continue
		}
		ret.Set(v, r.Get(i))
	}
	return nil
}
