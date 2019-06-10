// Package spec provides functions for building a flux.Spec from different sources (e.g., string, AST).
// It is intended for internal use only.
package spec

import (
	"errors"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// FromScript returns a spec from a script expressed as a raw string.
func FromScript(script string, now time.Time) (*flux.Spec, error) {
	astPkg, err := flux.Parse(script)
	if err != nil {
		return nil, err
	}
	return FromAST(astPkg, now)
}

// FromAST returns a spec from an AST.
func FromAST(astPkg *ast.Package, now time.Time) (*flux.Spec, error) {
	semPkg, err := semantic.New(astPkg)
	if err != nil {
		return nil, err
	}
	return FromSemantic(semPkg, now)
}

// FromSemantic returns a spec from a semantic graph.
func FromSemantic(semPkg *semantic.Package, now time.Time) (*flux.Spec, error) {
	sideEffects, _, now, err := flux.EvalWithNow(semPkg, now)
	if err != nil {
		return nil, err
	}
	return FromSideEffects(sideEffects, now)
}

// FromSideEffects returns a spec from the side effects of evaluation.
func FromSideEffects(sideEffects []interpreter.SideEffect, now time.Time) (*flux.Spec, error) {
	spec, err := toSpec(sideEffects, now)
	if err != nil {
		return nil, err
	}
	if len(spec.Operations) == 0 {
		return nil,
			errors.New("this Flux script returns no streaming data. " +
				"Consider adding a \"yield\" or invoking streaming functions directly, without performing an assignment")
	}
	return spec, nil
}

// FromTableObject returns a spec from a TableObject.
func FromTableObject(to *flux.TableObject, now time.Time) *flux.Spec {
	ider := &ider{
		id:     0,
		lookup: make(map[*flux.TableObject]flux.OperationID),
	}
	spec := &flux.Spec{
		Now: now,
	}
	visited := make(map[*flux.TableObject]bool)
	buildSpec(to, ider, spec, visited)
	return spec
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

func toSpec(functionCalls []interpreter.SideEffect, now time.Time) (*flux.Spec, error) {
	ider := &ider{
		id:     0,
		lookup: make(map[*flux.TableObject]flux.OperationID),
	}

	spec := &flux.Spec{Now: now}
	seen := make(map[*flux.TableObject]bool)
	objs := make([]*flux.TableObject, 0, len(functionCalls))

	for _, call := range functionCalls {
		if op, ok := call.Value.(*flux.TableObject); ok {
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

	return spec, nil
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
