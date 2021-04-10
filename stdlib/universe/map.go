package universe

import (
	"context"
	"sort"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/execkit"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const MapKind = "map"

type MapOpSpec struct {
	Fn       interpreter.ResolvedFunction `json:"fn"`
	MergeKey bool                         `json:"mergeKey"`
}

func init() {
	execkit.RegisterTransformation(&MapProcedureSpec{})
}

type MapProcedureSpec struct {
	plan.DefaultCost
	Fn       interpreter.ResolvedFunction `json:"fn"`
	MergeKey bool
}

func (s *MapProcedureSpec) CreateTransformation(id execute.DatasetID, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	t, err := NewMapTransformation(a.Context(), s, a.Allocator())
	if err != nil {
		return nil, nil, err
	}
	return execkit.NewGroupTransformation(id, t, a.Allocator())
}

func (s *MapProcedureSpec) ReadArgs(args flux.Arguments, a *flux.Administration) error {
	if err := a.AddParentFromArgs(args); err != nil {
		return err
	}

	if f, err := args.GetRequiredFunction("fn"); err != nil {
		return err
	} else {
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return err
		}
		s.Fn = fn
	}

	if m, ok, err := args.GetBool("mergeKey"); err != nil {
		return err
	} else if ok {
		s.MergeKey = m
	} else {
		// deprecated parameter: default is now false.
		s.MergeKey = false
	}
	return nil
}

func (s *MapProcedureSpec) Kind() plan.ProcedureKind {
	return MapKind
}
func (s *MapProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(MapProcedureSpec)
	*ns = *s
	ns.Fn = s.Fn.Copy()
	return ns
}

type mapTransformation struct {
	execute.ExecutionNode
	ctx      context.Context
	fn       *execute.RowMapFn
	mergeKey bool
	mem      memory.Allocator
}

func NewMapTransformation(ctx context.Context, spec *MapProcedureSpec, mem memory.Allocator) (*mapTransformation, error) {
	fn := execute.NewRowMapFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
	return &mapTransformation{
		fn:       fn,
		ctx:      ctx,
		mergeKey: spec.MergeKey,
		mem:      mem,
	}, nil
}

func (t *mapTransformation) Process(view table.View, d *execkit.Dataset, mem memory.Allocator) error {
	// Prepare the functions for the column types.
	cols := view.Cols()
	fn, err := t.fn.Prepare(cols)
	if err != nil {
		// TODO(nathanielc): Should we not fail the query for failed compilation?
		return err
	}

	// TODO(jsternberg): Use type inference. The original algorithm
	// didn't really use type inference at all so I removed its usage
	// in favor of the real returned type.

	buf := view.Buffer()
	cache := table.BuilderCache{
		New: func(key flux.GroupKey) table.Builder {
			return table.NewArrowBuilder(key, t.mem)
		},
		Tables: execute.NewRandomAccessGroupLookup(),
	}

	var on map[string]bool
	l := view.Len()
	for i := 0; i < l; i++ {
		m, err := fn.Eval(t.ctx, i, &buf)
		if err != nil {
			return errors.Wrap(err, codes.Invalid, "failed to evaluate map function")
		}

		// If we haven't determined the columns to group on, do that now.
		if on == nil {
			var err error
			on, err = t.groupOn(view.Key(), m.Type())
			if err != nil {
				return err
			}
		}

		key := groupKeyForObject(i, &buf, m, on)
		builder, created := table.GetArrowBuilder(key, &cache)
		if created {
			if err := t.createSchema(fn, builder, m); err != nil {
				return err
			}
		}

		for j, c := range builder.Cols() {
			v, ok := m.Get(c.Label)
			if !ok {
				if idx := execute.ColIdx(c.Label, view.Key().Cols()); t.mergeKey && idx >= 0 {
					v = view.Key().Value(idx)
				} else {
					// This should be unreachable
					return errors.Newf(codes.Internal, "could not find value for column %q", c.Label)
				}
			}
			b := builder.Builders[j]
			if !v.IsNull() && c.Type.String() != v.Type().Nature().String() {
				return errors.Newf(codes.Internal, "column %s:%s is not of type %v",
					c.Label, c.Type, v.Type(),
				)
			}
			if err := arrow.AppendValue(b, v); err != nil {
				return err
			}
		}
	}

	// Send all of the tables that we constructed downstream.
	return cache.ForEach(func(key flux.GroupKey, builder table.Builder) error {
		tbl, err := builder.Table()
		if err != nil {
			return err
		}
		return tbl.Do(func(cr flux.ColReader) error {
			view := table.ViewFromReader(cr)
			view.Retain()
			return d.Process(view)
		})
	})
}

func (t *mapTransformation) groupOn(key flux.GroupKey, m semantic.MonoType) (map[string]bool, error) {
	on := make(map[string]bool, len(key.Cols()))
	for _, c := range key.Cols() {
		if t.mergeKey {
			on[c.Label] = true
			continue
		}

		// If the label isn't included in the properties,
		// then it wasn't returned by the eval.
		n, err := m.NumProperties()
		if err != nil {
			return nil, err
		}

		for i := 0; i < n; i++ {
			Record, err := m.RecordProperty(i)
			if err != nil {
				return nil, err
			}

			if Record.Name() == c.Label {
				on[c.Label] = true
				break
			}
		}
	}
	return on, nil
}

// createSchema will create the schema for a table based on the object.
// This should only be called when a table is created anew.
//
// TODO(jsternberg): I am pretty sure this method and its usage don't
// match with the spec, but it is a faithful reproduction of the current
// map behavior. When we get around to rewriting portions of map, this
// should be rewritten to use the inferred type from type inference
// and it should be capable of consolidating schemas from non-uniform
// tables.
func (t *mapTransformation) createSchema(fn *execute.RowMapPreparedFn, b *table.ArrowBuilder, m values.Object) error {
	if t.mergeKey {
		for _, col := range b.Key().Cols() {
			if _, err := b.AddCol(col); err != nil {
				return err
			}
		}
	}

	returnType := fn.Type()

	numProps, err := returnType.NumProperties()
	if err != nil {
		return err
	}

	props := make(map[string]semantic.Nature, numProps)
	// Deduplicate the properties in the return type.
	// Scan properties in reverse order to ensure we only
	// add visible properties to the list.
	for i := numProps - 1; i >= 0; i-- {
		prop, err := returnType.RecordProperty(i)
		if err != nil {
			return err
		}
		typ, err := prop.TypeOf()
		if err != nil {
			return err
		}
		props[prop.Name()] = typ.Nature()
	}

	// Add columns from function in sorted order.
	n, err := m.Type().NumProperties()
	if err != nil {
		return err
	}

	keys := make([]string, 0, n)
	for i := 0; i < n; i++ {
		Record, err := m.Type().RecordProperty(i)
		if err != nil {
			return err
		}
		keys = append(keys, Record.Name())
	}
	sort.Strings(keys)

	for _, k := range keys {
		if t.mergeKey && b.Key().HasCol(k) {
			continue
		}

		v, ok := m.Get(k)
		if !ok {
			continue
		}

		nature := v.Type().Nature()

		if kind, ok := props[k]; ok && kind != semantic.Invalid {
			nature = kind
		}
		if nature == semantic.Invalid {
			continue
		}
		ty := execute.ConvertFromKind(nature)
		if ty == flux.TInvalid {
			return errors.Newf(codes.Invalid, `map object property "%s" is %v type which is not supported in a flux table`, k, nature)
		}
		if _, err := b.AddCol(flux.ColMeta{
			Label: k,
			Type:  ty,
		}); err != nil {
			return err
		}
	}
	return nil
}

func groupKeyForObject(i int, cr flux.ColReader, obj values.Object, on map[string]bool) flux.GroupKey {
	cols := make([]flux.ColMeta, 0, len(on))
	vs := make([]values.Value, 0, len(on))
	for j, c := range cr.Cols() {
		if !on[c.Label] {
			continue
		}
		v, ok := obj.Get(c.Label)
		if ok {
			vs = append(vs, v)
			cols = append(cols, flux.ColMeta{
				Label: c.Label,
				Type:  flux.ColumnType(v.Type()),
			})
		} else {
			vs = append(vs, execute.ValueForRow(cr, i, j))
			cols = append(cols, c)
		}
	}
	return execute.NewGroupKey(cols, vs)
}
