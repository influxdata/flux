package universe

import (
	"fmt"
	"sort"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

const ReduceKind = "reduce"

type ReduceOpSpec struct {
	Fn          *semantic.FunctionExpression `json:"fn"`
	ReducerType semantic.Type                `json:"reducer_type"`
	Identity    map[string]string            `json:"identity"`
}

func init() {
	reduceSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"fn": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{
					"r":           semantic.Tvar(1),
					"accumulator": semantic.Tvar(2),
				},
				Required: semantic.LabelSet{"r", "accumulator"},
				Return:   semantic.Tvar(2),
			}),
			"identity": semantic.Tvar(2),
		},
		[]string{"fn", "identity"},
	)

	flux.RegisterPackageValue("universe", ReduceKind, flux.FunctionValue(ReduceKind, createReduceOpSpec, reduceSignature))
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
		spec.ReducerType = o.Type()
		spec.Identity = make(map[string]string, o.Len())
		var haderr error
		o.Range(func(name string, v values.Value) {
			stringer, ok := v.(values.ValueStringer)
			if !ok {
				haderr = errors.New("ne contains unencodable type")
				return
			}
			spec.Identity[name] = stringer.String()
		})
		if haderr != nil {
			return nil, haderr
		}
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
	Fn          *semantic.FunctionExpression
	ReducerType semantic.Type
	Identity    map[string]string
}

func newReduceProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ReduceOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &ReduceProcedureSpec{
		Fn:          spec.Fn,
		ReducerType: spec.ReducerType,
		Identity:    spec.Identity,
	}, nil
}

func (s *ReduceProcedureSpec) Kind() plan.ProcedureKind {
	return ReduceKind
}
func (s *ReduceProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(ReduceProcedureSpec)
	*ns = *s
	ns.Fn = s.Fn.Copy().(*semantic.FunctionExpression)
	ns.ReducerType = s.ReducerType
	for k, v := range s.Identity {
		ns.Identity[k] = v
	}
	return ns
}

func createReduceTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ReduceProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t, err := NewReduceTransformation(d, cache, s)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type reduceTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	fn             *execute.RowReduceFn
	neutralElement map[string]values.Value
}

func NewReduceTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ReduceProcedureSpec) (*reduceTransformation, error) {
	fn, err := execute.NewRowReduceFn(spec.Fn)
	if err != nil {
		return nil, err
	}

	ne := make(map[string]values.Value)
	for k, v := range spec.ReducerType.Properties() {
		newVal, err := values.NewFromString(v.Nature(), spec.Identity[k])
		if err != nil {
			return nil, err
		}
		ne[k] = newVal
	}

	return &reduceTransformation{
		d:              d,
		cache:          cache,
		fn:             fn,
		neutralElement: ne,
	}, nil
}

func (t *reduceTransformation) Process(id execute.DatasetID, tbl flux.Table) error {

	var reducer values.Object = values.NewObjectWithValues(t.neutralElement)
	// t.fn.Prepare will pre-compile the function given the specific columns of the input table,
	// plus the reducer type.  For a given set of type values, we will cache the compiled function to
	// avoid costly recompilation.
	cols := tbl.Cols()
	err := t.fn.Prepare(cols, map[string]semantic.Type{"accumulator": reducer.Type()})
	if err != nil {
		// TODO(nathanielc): Should we not fail the query for failed compilation?
		return err
	}

	// tbl.Do is our method for iterating over a table.
	// it takes a function parameter that operated on a column reader type.
	// internally, large data tables will be broken into parts, and each separate part
	// will be processed by this function.
	if err := tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			// the RowReduce function type takes a row of values, and an accumulator value, and
			// computes a new accumulator result.
			m, err := t.fn.Eval(i, cr, map[string]values.Value{"accumulator": reducer})
			if err != nil {
				return errors.Wrap(err, "failed to evaluate reduce function")
			}
			reducer = m
		}
		return nil
	}); err != nil {
		return err
	}

	gkb := execute.NewGroupKeyBuilder(tbl.Key())
	typ := reducer.Type().Properties()
	var typeKeys []string
	for k := range typ {
		typeKeys = append(typeKeys, k)
	}
	// go maps have unsorted keys, so we need to extract the keys and sort them
	sort.Strings(typeKeys)
	for _, k := range typeKeys {
		if tbl.Key().HasCol(k) {
			val, _ := reducer.Get(k)
			gkb.SetKeyValue(k, val)
		}
	}

	// Find the output table that we will write to.  For reduce, we will write rows with the same
	// table group key
	tblKey, err := gkb.Build()
	if err != nil {
		return err
	}
	builder, created := t.cache.TableBuilder(tblKey)
	if created {
		// add the key columns to the table.
		if err := execute.AddTableKeyCols(tblKey, builder); err != nil {
			return err
		}

		// add table columns for each key in the reducer type map
		for _, k := range typeKeys {
			if tblKey.HasCol(k) {
				continue
			}
			if _, err := builder.AddCol(flux.ColMeta{
				Label: k,
				Type:  flux.ColumnType(typ[k]),
			}); err != nil {
				return err
			}
		}

		for j, c := range builder.Cols() {
			v, ok := reducer.Get(c.Label)
			if !ok {
				if idx := execute.ColIdx(c.Label, tbl.Key().Cols()); idx >= 0 {
					v = tbl.Key().Value(idx)
				} else {
					// This should be unreachable
					return fmt.Errorf("could not find value for column %q", c.Label)
				}
			}
			if err := builder.AppendValue(j, v); err != nil {
				return err
			}
		}
	} else {
		return errors.New("two reducers writing result to the same table")
	}
	return nil
}

// advanced stream processing functions, most transformations use this boiler-plate.
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
