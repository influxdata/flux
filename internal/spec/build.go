package spec

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/interpreter"
<<<<<<< HEAD
=======
	"github.com/influxdata/flux/lang/execdeps"
	"github.com/influxdata/flux/values"
>>>>>>> master
	"github.com/opentracing/opentracing-go"
)

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

func FromEvaluation(ctx context.Context, ses []interpreter.SideEffect, now time.Time) (*flux.Spec, error) {
	ider := &ider{
		id:     0,
		lookup: make(map[*flux.TableObject]flux.OperationID),
	}

	spec := &flux.Spec{Now: now}
	seen := make(map[*flux.TableObject]bool)
	objs := make([]*flux.TableObject, 0, len(ses))

	for _, se := range ses {
		if op, ok := se.Value.(*flux.TableObject); ok {
			s, cctx := opentracing.StartSpanFromContext(ctx, "toSpec")
			s.SetTag("opKind", op.Kind)
			if se.Node != nil {
				s.SetTag("loc", se.Node.Location().String())
			}

			if !isDuplicateTableObject(cctx, op, objs) {
				buildSpecWithTrace(cctx, op, ider, spec, seen)
				objs = append(objs, op)
			}
			s.Finish()
		}
	}

	if len(spec.Operations) == 0 {
		return nil,
			fmt.Errorf("this Flux script returns no streaming data. " +
				"Consider adding a \"yield\" or invoking streaming functions directly, without performing an assignment")
	}

	return spec, nil
}

func isDuplicateTableObject(ctx context.Context, op *flux.TableObject, objs []*flux.TableObject) bool {
	s, _ := opentracing.StartSpanFromContext(ctx, "isDuplicate")
	defer s.Finish()

	for _, tableObject := range objs {
		if op == tableObject {
			return true
		}
	}
	return false
}

func buildSpecWithTrace(ctx context.Context, t *flux.TableObject, ider flux.IDer, spec *flux.Spec, visited map[*flux.TableObject]bool) {
	s, _ := opentracing.StartSpanFromContext(ctx, "buildSpec")
	s.SetTag("opKind", t.Kind)
	buildSpec(t, ider, spec, visited)
	s.Finish()
}

func buildSpec(t *flux.TableObject, ider flux.IDer, spec *flux.Spec, visited map[*flux.TableObject]bool) {
	// Traverse graph upwards to first unvisited node.
	// Note: parents are sorted based on parameter name, so the visit order is consistent.
	for _, p := range t.Parents {
		if !visited[p] {
			// recurse up parents
			buildSpec(p, ider, spec, visited)
		}
	}

	// Assign ID to table object after visiting all ancestors.
	tableID := ider.ID(t)

	// Link table object to all parents after assigning ID.
	for _, p := range t.Parents {
		spec.Edges = append(spec.Edges, flux.Edge{
			Parent: ider.ID(p),
			Child:  tableID,
		})
	}

	visited[t] = true
	spec.Operations = append(spec.Operations, t.Operation(ider))
}

// FromTableObject returns a spec from a TableObject.
func FromTableObject(ctx context.Context, to *flux.TableObject, now time.Time) (*flux.Spec, error) {
	return FromEvaluation(ctx, []interpreter.SideEffect{{Value: to}}, now)
}

// FromScript returns a spec from a script expressed as a raw string.
// This is duplicate logic for what happens when a flux.Program runs.
// This function is used in tests that compare flux.Specs (e.g. in planner tests).
func FromScript(ctx context.Context, runtime flux.Runtime, now time.Time, script string) (*flux.Spec, error) {
	s, _ := opentracing.StartSpanFromContext(ctx, "parse")
	astPkg, err := runtime.Parse(script)
	if err != nil {
		return nil, err
	}
	s.Finish()

	deps := execdeps.NewExecutionDependencies(nil, &now, nil)
	ctx = deps.Inject(ctx)

	s, cctx := opentracing.StartSpanFromContext(ctx, "eval")
<<<<<<< HEAD
	sideEffects, scope, err := runtime.Eval(cctx, astPkg, flux.SetNowOption(now))
=======
	sideEffects, _, err := flux.EvalAST(cctx, astPkg)
>>>>>>> master
	if err != nil {
		return nil, err
	}
	s.Finish()

	s, cctx = opentracing.StartSpanFromContext(ctx, "compile")
	defer s.Finish()
	return FromEvaluation(cctx, sideEffects, *deps.Now)
}
