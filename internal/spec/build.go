package spec

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/operation"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/opentracing/opentracing-go"
)

type ider struct {
	id     *int
	lookup map[*flux.TableObject]operation.NodeID
}

func (i *ider) nextID() int {
	next := *i.id
	*i.id++
	return next
}

func (i *ider) get(t *flux.TableObject) (operation.NodeID, bool) {
	tableID, ok := i.lookup[t]
	return tableID, ok
}

func (i *ider) set(t *flux.TableObject, id int) operation.NodeID {
	opID := operation.NodeID(fmt.Sprintf("%s%d", t.Kind, id))
	i.lookup[t] = opID
	return opID
}

func (i *ider) ID(t *flux.TableObject) operation.NodeID {
	tableID, ok := i.get(t)
	if !ok {
		tableID = i.set(t, i.nextID())
	}
	return tableID
}

// FromEvaluation produces a flux.Spec from an array of side-effects.
//
// The `skipYields` param can be used to adjust the spec (ie to omit the yields).
// This is useful for situations like table functions (tableFind, etc) where the
// consumer of the spec can only accept one (1) result, presumably coming from
// the terminal node in the plan.
// In keeping with the "one result" requirement, when `skipYields` is true
// FromEvaluation will produce an error for inputs producing > 1 result.
func FromEvaluation(ctx context.Context, ses []interpreter.SideEffect, now time.Time, skipYields bool) (*operation.Spec, error) {
	var nextNodeID *int
	if value := ctx.Value(plan.NextPlanNodeIDKey); value != nil {
		nextNodeID = value.(*int)
	} else {
		nextNodeID = new(int)
	}
	ider := &ider{
		id:     nextNodeID,
		lookup: make(map[*flux.TableObject]operation.NodeID),
	}

	spec := &operation.Spec{Now: now}
	seen := make(map[*flux.TableObject]bool)
	objs := make([]*flux.TableObject, 0, len(ses))
	resultCount := 0

	for _, se := range ses {
		if op, ok := se.Value.(*flux.TableObject); ok {
			s, cctx := opentracing.StartSpanFromContext(ctx, "toSpec")
			s.SetTag("opKind", op.Kind)
			if se.Node != nil {
				s.SetTag("loc", se.Node.Location().String())
			}

			if !isDuplicateTableObject(cctx, op, objs) {
				// Don't bother incrementing the count if this is a yield which
				// will be skipped anyway.
				if !(skipYields && op.Kind == "yield") {
					resultCount += 1
				}
				buildSpecWithTrace(cctx, op, ider, spec, seen, skipYields)
				objs = append(objs, op)
			}

			s.Finish()
		}
	}

	if len(spec.Operations) == 0 {
		return nil,
			errors.New(codes.Invalid,
				"this Flux script returns no streaming data. "+
					"Consider adding a \"yield\" or invoking streaming functions directly, without performing an assignment")
	}
	// When skipYields is true, we're likely running a sub-program (ie. tableFind).
	// In this case we ignore any yields but we also have an extra requirement:
	// there can only be 1 result. Since side-effects automatically produce
	// results, when there is more than 1 non-yield side-effect, we error.
	if skipYields && resultCount > 1 {
		return nil,
			errors.Newf(
				codes.Invalid,
				"expected exactly 1 result from table stream, found %d", resultCount,
			)
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

func buildSpecWithTrace(ctx context.Context, t *flux.TableObject, ider *ider, spec *operation.Spec, visited map[*flux.TableObject]bool, skipYields bool) {
	s, cctx := opentracing.StartSpanFromContext(ctx, "buildSpec")
	s.SetTag("opKind", t.Kind)
	buildSpec(cctx, t, ider, spec, visited, skipYields)
	s.Finish()
}

func buildSpec(ctx context.Context, t *flux.TableObject, ider *ider, spec *operation.Spec, visited map[*flux.TableObject]bool, skipYields bool) {
	// Check if this table object has already been used
	// and mark the spec as having a conflict if it has.
	if t.Owned {
		spec.HasConflict = true
	}

	// Traverse graph upwards to first unvisited node.
	// Note: parents are sorted based on parameter name, so the visit order is consistent.
	var parents []*flux.TableObject
	if skipYields {
		parents = getNonYieldParents(make([]*flux.TableObject, 0), t)
	} else {
		parents = t.Parents
	}

	for _, p := range parents {
		if !visited[p] {
			// recurse up parents
			buildSpec(ctx, p, ider, spec, visited, skipYields)
		}
	}

	// Assign ID to table object after visiting all ancestors.
	tableID := ider.ID(t)

	if !(skipYields && t.Kind == "yield") {
		// Link table object to all parents after assigning ID.
		for _, p := range parents {
			spec.Edges = append(spec.Edges, operation.Edge{
				Parent: ider.ID(p),
				Child:  tableID,
			})
		}
		op := &operation.Node{
			ID:   ider.ID(t),
			Spec: t.Spec,
			Source: operation.NodeSource{
				Stack: t.Source.Stack,
			},
		}
		spec.Operations = append(spec.Operations, op)
	}

	visited[t] = true
	t.Owned = true
}

// getNonYieldParents builds an array of parents, skipping any yields found along the way.
func getNonYieldParents(acc []*flux.TableObject, to *flux.TableObject) []*flux.TableObject {
	for _, p := range to.Parents {
		if p.Kind == "yield" {
			acc = getNonYieldParents(acc, p)
		} else {
			acc = append(acc, p)
		}
	}
	return acc
}

// FromTableObject returns a spec from a TableObject.
func FromTableObject(ctx context.Context, to *flux.TableObject, now time.Time) (*operation.Spec, error) {
	return FromEvaluation(ctx, []interpreter.SideEffect{{Value: to}}, now, true)
}

// FromScript returns a spec from a script expressed as a raw string.
// This is duplicate logic for what happens when a flux.Program runs.
// This function is used in tests that compare flux.Specs (e.g. in planner tests).
func FromScript(ctx context.Context, runtime flux.Runtime, now time.Time, script string) (*operation.Spec, error) {
	s, _ := opentracing.StartSpanFromContext(ctx, "parse")
	astPkg, err := runtime.Parse(ctx, script)
	if err != nil {
		return nil, err
	}
	s.Finish()

	deps := execute.NewExecutionDependencies(nil, &now, nil)
	ctx = deps.Inject(ctx)

	s, cctx := opentracing.StartSpanFromContext(ctx, "eval")
	sideEffects, scope, err := runtime.Eval(cctx, astPkg, nil, flux.SetNowOption(now))
	if err != nil {
		return nil, err
	}
	s.Finish()

	s, cctx = opentracing.StartSpanFromContext(ctx, "compile")
	defer s.Finish()
	nowOpt, ok := scope.Lookup(interpreter.NowOption)
	if !ok {
		return nil, fmt.Errorf("%q option not set", interpreter.NowOption)
	}
	nowTime, err := nowOpt.Function().Call(ctx, nil)
	if err != nil {
		return nil, err
	}

	return FromEvaluation(cctx, sideEffects, nowTime.Time().Time(), false)
}
