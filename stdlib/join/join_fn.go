package join

import (
	"context"
	"sort"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// JoinFn handles the logic of calling the function in the `as` parameter of join.tables()
type JoinFn struct {
	fn       *execute.RowJoinFn
	prepared *execute.RowJoinPreparedFn
	args     values.Object
	schema   []flux.ColMeta
	ltyp     *semantic.MonoType
	rtyp     *semantic.MonoType
}

func NewJoinFn(fn interpreter.ResolvedFunction) *JoinFn {
	return &JoinFn{
		fn: execute.NewRowJoinFn(fn.Fn, compiler.ToScope(fn.Scope)),
	}
}

func (f *JoinFn) Prepare(ctx context.Context, lcols, rcols []flux.ColMeta) error {
	typ := f.Type()
	args, err := typ.SortedArguments()
	if err != nil {
		return err
	}

	if f.ltyp == nil {
		ltyp, err := getObjectType(args[0], lcols)
		if err != nil {
			return errors.Wrap(err, codes.Invalid, "error preparing left side of join")
		}
		f.ltyp = &ltyp
	}

	if f.rtyp == nil {
		rtyp, err := getObjectType(args[1], rcols)
		if err != nil {
			return errors.Wrap(err, codes.Invalid, "error preparing right side of join")
		}
		f.rtyp = &rtyp
	}

	in := semantic.NewObjectType([]semantic.PropertyType{
		{Key: []byte("l"), Value: *f.ltyp},
		{Key: []byte("r"), Value: *f.rtyp},
	})
	f.args = values.NewObject(in)
	prepared, err := f.fn.Prepare(
		ctx,
		lcols,
		map[string]semantic.MonoType{"r": *f.rtyp},
		false,
	)
	f.prepared = prepared
	if err != nil {
		return err
	}
	return nil
}

// Produces the output of the joinProduct, if there is any to be produced. In some cases,
// the specified join method might require that no output be produced if one side is empty.
// If the returned bool == true, that means this function returned a non-empty table chunk
// containing the joined data.
func (f *JoinFn) Eval(ctx context.Context, p *joinProduct, method string, mem memory.Allocator) ([]table.Chunk, bool, error) {
	// It's possible for either side to be empty, in which case we consult the join method
	// to determine what to do next. However, it shouldn't be possible for both left and right
	// to be empty.
	if p.left.len() == 0 && p.right.len() == 0 {
		return nil, false, errors.New(codes.Internal, "tried to join on an empty set")
	}

	// Check if either side is empty. If so, we may be able to exit the function early,
	// depending on the join method. If we can't exit early, then we create a default row
	// for the empty side, where all of the group key columns are populated, and everything
	// else is null.
	if p.left.nrows() < 1 {
		if method == "inner" || method == "left" {
			return nil, false, nil
		}
		groupKey := p.right[0].Key()
		defaultRow := defaultRow(groupKey, f.leftType())
		cols := colsFromObjectType(f.leftType())
		b := execute.NewChunkBuilder(cols, 1, mem)
		b.AppendRecord(defaultRow)
		c := b.Build(groupKey)
		p.left = append(p.left, c)
	} else if p.right.nrows() < 1 {
		if method == "inner" || method == "right" {
			return nil, false, nil
		}
		groupKey := p.left[0].Key()
		defaultRow := defaultRow(groupKey, f.rightType())
		cols := colsFromObjectType(f.rightType())
		b := execute.NewChunkBuilder(cols, 1, mem)
		b.AppendRecord(defaultRow)
		c := b.Build(groupKey)
		p.right = append(p.right, c)
	}
	c, err := f.crossProduct(ctx, p, mem)
	if err != nil {
		return nil, false, err
	}
	chunks := splitChunk(*c)
	p.Release()
	return chunks, true, nil
}

func (f *JoinFn) crossProduct(ctx context.Context, p *joinProduct, mem memory.Allocator) (*table.Chunk, error) {
	var builder *execute.ChunkBuilder
	for i := 0; i < p.left.nrows(); i++ {
		l := p.left.getRow(i, f.leftType())

		for j := 0; j < p.right.nrows(); j++ {
			r := p.right.getRow(j, f.rightType())
			joined, err := f.eval(ctx, l, r)
			if err != nil {
				return nil, err
			}

			// Make sure the group key is not modfied
			// TODO(sean): Potential optimization - determine whether or not it is
			// necessary to validate every row. There may be some cases where we can
			// know for sure if the group key is never going to change.
			err = validateGroupKey(joined, p.left[0].Key())
			if err != nil {
				return nil, err
			}
			if f.schema == nil {
				cols, err := f.createSchema(joined)
				if err != nil {
					return nil, err
				}
				f.schema = cols
			}
			if builder == nil {
				builder = execute.NewChunkBuilder(f.schema, p.left.nrows()*p.right.nrows(), mem)
			}
			builder.AppendRecord(joined)
		}
	}
	c := builder.Build(p.left[0].Key())
	return &c, nil
}

func (f *JoinFn) createSchema(record values.Object) ([]flux.ColMeta, error) {
	returnType := f.ReturnType()

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
		props[prop.Name()] = typ.Nature()
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
	sort.Strings(keys)

	cols := make([]flux.ColMeta, 0, len(keys))
	for _, k := range keys {
		v, ok := record.Get(k)
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
			return nil, errors.Newf(codes.Invalid, `map object property "%s" is %v type which is not supported in a flux table`, k, nature)
		}
		cols = append(cols, flux.ColMeta{
			Label: k,
			Type:  ty,
		})
	}
	return cols, nil
}

func (f *JoinFn) eval(ctx context.Context, l, r values.Object) (values.Object, error) {
	f.args.Set("l", l)
	f.args.Set("r", r)

	joined, err := f.prepared.Eval(ctx, f.args)
	if err != nil {
		return nil, err
	}
	obj := joined.Object()
	return obj, nil
}

func (f *JoinFn) Type() semantic.MonoType {
	return f.fn.Type()
}

func (f *JoinFn) ReturnType() semantic.MonoType {
	return f.fn.ReturnType()
}

func (f *JoinFn) leftType() semantic.MonoType {
	return *f.ltyp
}

func (f *JoinFn) rightType() semantic.MonoType {
	return *f.rtyp
}

func defaultRow(key flux.GroupKey, objType semantic.MonoType) values.Object {
	obj := values.NewObject(objType)
	obj.Range(func(name string, v values.Value) {
		val := key.LabelValue(name)
		if val != nil {
			obj.Set(name, val)
		} else {
			obj.Set(name, values.Null)
		}
	})
	return obj
}

func getObjectType(arg *semantic.Argument, cols []flux.ColMeta) (semantic.MonoType, error) {
	if err := checkCols(arg, cols); err != nil {
		return semantic.MonoType{}, err
	}

	t := make([]semantic.PropertyType, len(cols))
	for i, col := range cols {
		t[i] = semantic.PropertyType{
			Key:   []byte(col.Label),
			Value: flux.SemanticType(col.Type),
		}
	}
	return semantic.NewObjectType(t), nil
}

func checkCols(arg *semantic.Argument, cols []flux.ColMeta) error {
	argType, err := arg.TypeOf()
	if err != nil {
		return err
	}

	props, err := argType.SortedProperties()
	if err != nil {
		return err
	}

	for _, prop := range props {
		name := prop.Name()
		found := false
		if len(cols) == 0 {
			return errors.Newf(codes.Invalid, "cannot join on an empty table")
		}
		for _, column := range cols {
			if column.Label == name {
				found = true
				break
			}
		}
		if !found {
			return errors.Newf(codes.Invalid, "table is missing label %s", name)
		}
	}
	return nil
}

func colsFromObjectType(t semantic.MonoType) []flux.ColMeta {
	n, _ := t.NumProperties()
	cols := make([]flux.ColMeta, 0, n)
	for i := 0; i < n; i++ {
		prop, _ := t.RecordProperty(i)
		typ, _ := prop.TypeOf()
		col := flux.ColMeta{Label: prop.Name(), Type: flux.ColumnType(typ)}
		cols = append(cols, col)
	}
	return cols
}

// splits a chunk into a list of chunks, each with a max size of 1000 rows
func splitChunk(c table.Chunk) []table.Chunk {
	l := c.Len()
	bufferSize := table.BufferSize
	if l <= bufferSize {
		return []table.Chunk{c}
	}

	chunks := make([]table.Chunk, 0, int(l/bufferSize)+1)
	curSize := l
	var i, start, stop int
	for ; curSize > 0; i++ {
		if curSize > bufferSize {
			start = i * bufferSize
			stop = bufferSize
			curSize -= bufferSize
		} else {
			start = i * bufferSize
			stop = start + curSize
			curSize = 0
		}
		slice := getChunkSlice(c, start, stop)
		chunks = append(chunks, slice)
	}
	return chunks
}

func validateGroupKey(obj values.Object, key flux.GroupKey) error {
	for _, col := range key.Cols() {
		if v, ok := obj.Get(col.Label); !ok || !v.Equal(key.LabelValue(col.Label)) {
			return errors.Newf(
				codes.Invalid,
				"join cannot modify group key: output record has a missing or invalid value for column '%s:%s'",
				col.Label,
				col.Type,
			)
		}
	}
	return nil
}
