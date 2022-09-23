package universe

import (
	"context"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	vectorizedMapKind = "vectorizedMap"
)

func init() {
	plan.RegisterPhysicalRules(vectorizeMapRule{})
	execute.RegisterTransformation(vectorizedMapKind, createVectorizedMapTransformation)
}

type vectorizedMapProcedureSpec struct {
	plan.DefaultCost
	Fn interpreter.ResolvedFunction
}

func (v *vectorizedMapProcedureSpec) Kind() plan.ProcedureKind {
	return vectorizedMapKind
}

func (v *vectorizedMapProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(vectorizedMapProcedureSpec)
	*ns = *v
	ns.Fn = v.Fn.Copy()
	return ns
}

type vectorizeMapRule struct{}

func (v vectorizeMapRule) Name() string {
	return "vectorizeMapRule"
}

func (v vectorizeMapRule) Pattern() plan.Pattern {
	return plan.MultiSuccessor(MapKind)
}

func (v vectorizeMapRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	mapSpec := node.ProcedureSpec().(*MapProcedureSpec)
	if mapSpec.Fn.Fn.Vectorized == nil {
		return node, false, nil
	}

	return plan.ReplacePhysicalNodes(ctx, node, node, vectorizedMapKind, &vectorizedMapProcedureSpec{
		Fn: interpreter.ResolvedFunction{
			Fn:    mapSpec.Fn.Fn.Vectorized,
			Scope: mapSpec.Fn.Scope,
		},
	}), true, nil
}

func createVectorizedMapTransformation(
	id execute.DatasetID,
	mode execute.AccumulationMode,
	spec plan.ProcedureSpec,
	a execute.Administration,
) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*vectorizedMapProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	tr := &mapTransformation{
		ctx: a.Context(),
		fn: &mapVectorFunc{
			fn: execute.NewVectorMapFn(s.Fn.Fn, compiler.ToScope(s.Fn.Scope)),
		},
	}
	return execute.NewGroupTransformation(id, tr, a.Allocator())
}

type mapVectorFunc struct {
	fn *execute.VectorMapFn
}

func (m *mapVectorFunc) Prepare(ctx context.Context, cols []flux.ColMeta) (mapPreparedFunc, error) {
	fn, err := m.fn.Prepare(ctx, cols)
	if err != nil {
		return nil, err
	}
	return &mapVectorPreparedFunc{
		fn: fn,
	}, nil
}

type mapVectorPreparedFunc struct {
	fn *execute.VectorMapPreparedFn
}

func (m *mapVectorPreparedFunc) createSchema(record values.Object) ([]flux.ColMeta, error) {
	returnType := m.fn.Type()

	numProps, err := returnType.NumProperties()
	if err != nil {
		return nil, err
	}

	props := make(map[string]semantic.Nature, numProps)
	// Deduplicate the properties in the return type.
	// Scan properties in reverse order to ensure we only
	// add visible properties to the list.
	for i := numProps - 1; i >= 0; i-- {
		prop, err := returnType.RecordProperty(i)
		if err != nil {
			return nil, err
		}
		typ, err := prop.TypeOf()
		if err != nil {
			return nil, err
		}
		elemTyp, err := typ.ElemType()
		if err != nil {
			return nil, err
		}
		props[prop.Name()] = elemTyp.Nature()
	}

	// Add columns from function in sorted order.
	n, err := record.Type().NumProperties()
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, n)
	for i := 0; i < n; i++ {
		prop, err := record.Type().RecordProperty(i)
		if err != nil {
			return nil, err
		}
		keys = append(keys, prop.Name())
	}

	cols := make([]flux.ColMeta, 0, len(keys))
	for _, k := range keys {
		v, ok := record.Get(k)
		if !ok {
			continue
		}

		nature := semantic.Invalid
		if !v.IsNull() {
			elemType, err := v.Type().ElemType()
			if err != nil {
				return nil, err
			}
			nature = elemType.Nature()
		}

		if kind, ok := props[k]; ok && kind != semantic.Invalid {
			nature = kind
		}
		if nature == semantic.Invalid {
			continue
		}
		ty := execute.ConvertFromKind(nature)
		if ty == flux.TInvalid {
			return nil, errors.Newf(codes.Invalid, `map object property "%s" is %v type which is not supported in a flux table`, k, nature)
		}
		cols = append(cols, flux.ColMeta{
			Label: k,
			Type:  ty,
		})
	}
	return cols, nil
}

func (m *mapVectorPreparedFunc) Eval(ctx context.Context, chunk table.Chunk, mem memory.Allocator) ([]flux.ColMeta, []array.Array, error) {
	res, err := m.fn.Eval(ctx, chunk)
	if err != nil {
		return nil, nil, err
	}
	defer res.Release()

	cols, err := m.createSchema(res)
	if err != nil {
		return nil, nil, err
	}

	var n int
	arrs := make([]array.Array, len(cols))
	repeaters := make([]bool, len(cols))
	for i, col := range cols {
		v, ok := res.Get(col.Label)
		if !ok || v.IsNull() {
			continue
		}

		vec := v.Vector()
		if vec.IsRepeat() {
			repeaters[i] = true
			continue
		}

		arr := vec.Arr()
		arr.Retain()
		if n == 0 {
			n = arr.Len()
		}
		arrs[i] = arr
	}

	for i, col := range cols {
		if arrs[i] == nil {
			if repeaters[i] {
				b := arrow.NewBuilder(col.Type, mem)
				b.Resize(n)
				v, ok := res.Get(col.Label)
				if !ok || v.IsNull() {
					continue
				}

				val := v.Vector().(*values.VectorRepeatValue).Value()
				for i := 0; i < n; i++ {
					// FIXME: Add an `arrow.Fill(b, val, n)` to utils?
					//  Assume this is slower than needed since
					//  we're type switching for every iteration.
					//  Arrow might even include a canonical way to fill.
					err := arrow.AppendValue(b, val)
					if err != nil {
						b.Release()
						return nil, nil, err
					}
				}
				arrs[i] = b.NewArray()
				b.Release()
			} else {
				arrs[i] = arrow.Nulls(col.Type, n, mem)
			}
		}
	}
	return cols, arrs, nil
}
