package universe

import (
	"context"
	"sort"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const ReduceKind = "reduce"

type ReduceOpSpec struct {
	Fn       interpreter.ResolvedFunction `json:"fn"`
	Identity values.Object                `json:"identity"`
}

func init() {
	reduceSignature := runtime.MustLookupBuiltinType("universe", "reduce")

	runtime.RegisterPackageValue("universe", ReduceKind, flux.MustValue(flux.FunctionValue(ReduceKind, createReduceOpSpec, reduceSignature)))
	flux.RegisterOpSpec(ReduceKind, newReduceOp)
	plan.RegisterProcedureSpec(ReduceKind, newReduceProcedure, ReduceKind)
	execute.RegisterTransformation(ReduceKind, createReduceTransformation)
}

func createReduceOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(ReduceOpSpec)

	if f, err := args.GetRequiredFunction("fn"); err != nil {
		return nil, err
	} else {
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return nil, err
		}
		spec.Fn = fn
	}

	if o, err := args.GetRequiredObject("identity"); err != nil {
		return nil, err
	} else {
		spec.Identity = o
	}

	return spec, nil
}

func newReduceOp() flux.OperationSpec {
	return new(ReduceOpSpec)
}

func (s *ReduceOpSpec) Kind() flux.OperationKind {
	return ReduceKind
}

type ReduceProcedureSpec struct {
	plan.DefaultCost
	Fn       interpreter.ResolvedFunction
	Identity values.Object
}

func newReduceProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ReduceOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	} else if n := spec.Identity.Type().Nature(); n != semantic.Object {
		return nil, errors.Newf(codes.Invalid, "identity must be an object, got %s", n)
	}

	return &ReduceProcedureSpec{
		Fn:       spec.Fn,
		Identity: spec.Identity,
	}, nil
}

func (s *ReduceProcedureSpec) Kind() plan.ProcedureKind {
	return ReduceKind
}
func (s *ReduceProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(ReduceProcedureSpec)
	*ns = *s
	ns.Fn = s.Fn.Copy()
	return ns
}

func createReduceTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ReduceProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t, err := NewReduceTransformation(a.Context(), s, d, cache)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type reduceTransformation struct {
	d        execute.Dataset
	cache    execute.TableBuilderCache
	ctx      context.Context
	fn       *execute.RowReduceFn
	identity values.Object
}

func NewReduceTransformation(ctx context.Context, spec *ReduceProcedureSpec, d execute.Dataset, cache execute.TableBuilderCache) (*reduceTransformation, error) {
	fn := execute.NewRowReduceFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
	return &reduceTransformation{
		d:        d,
		cache:    cache,
		ctx:      ctx,
		fn:       fn,
		identity: spec.Identity,
	}, nil
}

func (t *reduceTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// Prepare the function with the column types list.
	cols := tbl.Cols()
	fn, err := t.fn.Prepare(cols, map[string]semantic.MonoType{"accumulator": t.identity.Type()})
	if err != nil {
		return err
	}

	// Start the reduce operation with the neutral element as the accumulator.
	const accumulatorParamName = "accumulator"
	params := map[string]values.Value{accumulatorParamName: t.identity}
	if err := tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			// the RowReduce function type takes a row of values, and an accumulator value, and
			// computes a new accumulator result.
			m, err := fn.Eval(t.ctx, i, cr, params)
			if err != nil {
				return errors.Wrap(err, codes.Inherit, "failed to evaluate reduce function")
			}
			params[accumulatorParamName] = m
		}
		return nil
	}); err != nil {
		return err
	}

	// Compute the group key by replacing columns from the reducer if needed.
	m := params[accumulatorParamName].Object()
	key := t.computeGroupKey(tbl.Key(), m)

	builder, created := t.cache.TableBuilder(key)
	if !created {
		return errors.New(codes.FailedPrecondition, "two reducers writing result to the same table")
	}

	// Add the key columns to the table.
	if err := execute.AddTableKeyCols(key, builder); err != nil {
		return err
	}

	// Add remaining columns from the object if they're not in the key.
	columns := make([]string, 0, m.Len())
	m.Range(func(name string, v values.Value) {
		if key.HasCol(name) {
			return
		}
		columns = append(columns, name)
	})
	sort.Strings(columns)

	for _, label := range columns {
		v, _ := m.Get(label)
		if _, err := builder.AddCol(flux.ColMeta{
			Label: label,
			Type:  flux.ColumnType(v.Type()),
		}); err != nil {
			return err
		}
	}

	// Append a value for each column.
	for j, c := range builder.Cols() {
		v, ok := m.Get(c.Label)
		if !ok {
			v = key.LabelValue(c.Label)
		}

		if err := builder.AppendValue(j, v); err != nil {
			return err
		}
	}
	return nil
}

func (t *reduceTransformation) computeGroupKey(key flux.GroupKey, v values.Object) flux.GroupKey {
	replace := false
	v.Range(func(name string, v values.Value) {
		if key.HasCol(name) {
			replace = true
		}
	})

	if !replace {
		return key
	}

	// Copy over the values and replace any in the group key.
	vs := make([]values.Value, len(key.Values()))
	copy(vs, key.Values())
	v.Range(func(name string, v values.Value) {
		if idx := execute.ColIdx(name, key.Cols()); idx >= 0 {
			vs[idx] = v
		}
	})
	return execute.NewGroupKey(key.Cols(), vs)
}

func (t *reduceTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *reduceTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *reduceTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *reduceTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
