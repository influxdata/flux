package spec

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/values"
)

func unWrapSideEffects(ses []interpreter.SideEffect) []values.Value {
	vs := make([]values.Value, len(ses))
	for i, se := range ses {
		vs[i] = se.Value
	}
	return vs
}

func FromValues(evalValues []values.Value, now time.Time) (*flux.Spec, error) {
	ider := &ider{
		id:     0,
		lookup: make(map[*flux.TableObject]flux.OperationID),
	}

	spec := &flux.Spec{Now: now}
	seen := make(map[*flux.TableObject]bool)
	objs := make([]*flux.TableObject, 0, len(evalValues))

	for _, call := range evalValues {
		if op, ok := call.(*flux.TableObject); ok {
			dup := false
			for _, tableObject := range objs {
				if op.Equal(tableObject) {
					dup = true
					break
				}
			}
			if !dup {
				buildSpec(op, ider, spec, seen)
				objs = append(objs, op)
			}
		}
	}

	if len(spec.Operations) == 0 {
		return nil,
			fmt.Errorf("this Flux script returns no streaming data. " +
				"Consider adding a \"yield\" or invoking streaming functions directly, without performing an assignment")
	}

	return spec, nil
}

func FromEvaluation(ses []interpreter.SideEffect, now time.Time) (*flux.Spec, error) {
	vs := unWrapSideEffects(ses)
	return FromValues(vs, now)
}

// FromTableObject returns a spec from a TableObject.
func FromTableObject(to *flux.TableObject, now time.Time) (*flux.Spec, error) {
	return FromValues([]values.Value{to}, now)
}

// FromScript returns a spec from a script expressed as a raw string.
// This is duplicate logic for what happens when a flux.Program runs.
// This function is used in tests that compare flux.Specs (e.g. in planner tests).
func FromScript(ctx context.Context, deps dependencies.Interface, now time.Time, script string) (*flux.Spec, error) {
	astPkg, err := flux.Parse(script)
	if err != nil {
		return nil, err
	}
	sideEffects, scope, err := flux.EvalAST(ctx, deps, astPkg, flux.SetNowOption(now))
	if err != nil {
		return nil, err
	}
	nowOpt, ok := scope.Lookup(flux.NowOption)
	if !ok {
		return nil, fmt.Errorf("%q option not set", flux.NowOption)
	}
	nowTime, err := nowOpt.Function().Call(ctx, deps, nil)
	if err != nil {
		return nil, err
	}

	return FromEvaluation(sideEffects, nowTime.Time().Time())
}

type ider struct {
	id     int
	lookup map[*flux.TableObject]flux.OperationID
}

func (i *ider) nextID() int {
	next := i.id
	i.id++
	return next
}

func (i *ider) get(t *flux.TableObject) (flux.OperationID, bool) {
	tableID, ok := i.lookup[t]
	return tableID, ok
}

func (i *ider) set(t *flux.TableObject, id int) flux.OperationID {
	opID := flux.OperationID(fmt.Sprintf("%s%d", t.Kind, id))
	i.lookup[t] = opID
	return opID
}

func (i *ider) ID(t *flux.TableObject) flux.OperationID {
	tableID, ok := i.get(t)
	if !ok {
		tableID = i.set(t, i.nextID())
	}
	return tableID
}

func buildSpec(t *flux.TableObject, ider flux.IDer, spec *flux.Spec, visited map[*flux.TableObject]bool) {
	// Traverse graph upwards to first unvisited node.
	// Note: parents are sorted based on parameter name, so the visit order is consistent.
	t.Parents.Range(func(i int, v values.Value) {
		p := v.(*flux.TableObject)
		if !visited[p] {
			// rescurse up parents
			buildSpec(p, ider, spec, visited)
		}
	})

	// Assign ID to table object after visiting all ancestors.
	tableID := ider.ID(t)

	// Link table object to all parents after assigning ID.
	t.Parents.Range(func(i int, v values.Value) {
		p := v.(*flux.TableObject)
		spec.Edges = append(spec.Edges, flux.Edge{
			Parent: ider.ID(p),
			Child:  tableID,
		})
	})

	visited[t] = true
	spec.Operations = append(spec.Operations, t.Operation(ider))
}
