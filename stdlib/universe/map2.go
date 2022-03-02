package universe

import (
	"context"
	"sort"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/mutable"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

//go:generate -command tmpl ../../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@../../array/types.tmpldata -o map2.gen.go map2.gen.go.tmpl

type mapTransformation2 struct {
	ctx context.Context
	fn  mapFunc
}

func newMapTransformation2(ctx context.Context, id execute.DatasetID, spec *MapProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	var fn mapFunc
	if spec.Fn.Fn.Vectorized != nil {
		fn = &mapVectorFunc{
			fn: execute.NewVectorMapFn(
				spec.Fn.Fn.Vectorized,
				compiler.ToScope(spec.Fn.Scope),
			),
		}
	} else {
		fn = &mapRowFunc{
			fn: execute.NewRowMapFn(
				spec.Fn.Fn,
				compiler.ToScope(spec.Fn.Scope),
			),
		}
	}
	tr := &mapTransformation2{
		ctx: ctx,
		fn:  fn,
	}
	return execute.NewGroupTransformation(id, tr, mem)
}

func (m *mapTransformation2) Process(chunk table.Chunk, d *execute.TransportDataset, mem memory.Allocator) error {
	// The current version of map just silently drops
	// empty tables so let's just do that.
	if chunk.Len() == 0 {
		return nil
	}

	// Prepare the compiled function for the set of columns.
	cols := chunk.Cols()
	fn, err := m.fn.Prepare(cols)
	if err != nil {
		return err
	}

	// Execute function.
	cols, arrs, err := fn.Eval(m.ctx, chunk, mem)
	if err != nil {
		return err
	}
	return m.regroup(cols, chunk.Key(), arrs, d, mem)
}

// regroup will take the mapped output columns and regroup them into new group keys
// depending on the content of the columns.
func (m *mapTransformation2) regroup(cols []flux.ColMeta, key flux.GroupKey, arrs []array.Interface, d *execute.TransportDataset, mem memory.Allocator) error {
	// Determine which columns are part of the group key.
	keyIndices, keyCols := m.determineKeyColumns(cols, key)

	// Determine which of these key columns are not homogenous
	// and require us to regroup.
	regroupCols := m.regroupWith(keyIndices, arrs)
	if len(regroupCols) == 0 {
		// None of the columns are heterogeneous so
		// we can use the array as-is without regrouping.
		// Construct the values from the first row
		// and send it.
		key := m.makeKey(keyIndices, keyCols, cols, arrs, 0)
		return m.processTable(d, key, cols, arrs)
	}

	// This will require a regroup because one of the group key
	// columns is not homogenous. Since this is the case, we
	// will reconstruct the buffers and so we can defer releasing
	// the ones we have created.
	defer func() {
		for _, arr := range arrs {
			arr.Release()
		}
	}()

	// Determine which order the rows would be in if we sorted them.
	rowIndices := m.sort(arrs, arrs[0].Len(), regroupCols, mem)
	defer rowIndices.Release()

	// Regroup the values using the sorted row indices.
	return m.regroupSorted(d, regroupCols, keyIndices, keyCols, cols, rowIndices, arrs, mem)
}

// sort will use the given columns to create an index of sorted rows using the input arrays.
// This returns a set of indices mapping the ordered values to their original location
// in the array.
func (m *mapTransformation2) sort(arrs []array.Interface, n int, cols []int, mem memory.Allocator) *array.Int {
	// Construct the indices.
	indices := mutable.NewInt64Array(mem)
	indices.Resize(n)

	// Retrieve the raw slice and initialize the offsets.
	offsets := indices.Int64Values()
	for i := range offsets {
		offsets[i] = int64(i)
	}

	// Sort the offsets by using the comparison method.
	sort.SliceStable(offsets, func(i, j int) bool {
		i, j = int(offsets[i]), int(offsets[j])
		for _, col := range cols {
			arr := arrs[col]
			if cmp := arrowutil.Compare(arr, arr, i, j); cmp != 0 {
				return cmp < 0
			}
		}
		return false
	})

	// Return the now sorted indices.
	return indices.NewInt64Array()
}

// regroupSorted takes the sorted indices and regroups the columns into separate group keys.
func (m *mapTransformation2) regroupSorted(d *execute.TransportDataset, regroupCols, keyIndices []int, keyCols, cols []flux.ColMeta, rowIndices *array.Int, arrs []array.Interface, mem memory.Allocator) error {
	first, n := 0, arrs[0].Len()
	for first < n {
		// Use the first row to construct a key.
		key := m.makeKey(keyIndices, keyCols, cols, arrs, first)

		// Determine the last row that matches the same key.
		last := first + 1
		x := rowIndices.Value(first)
	OUTER:
		for last < n {
			for _, col := range regroupCols {
				arr := arrs[col]
				y := rowIndices.Value(last)
				if arrowutil.Compare(arr, arr, int(x), int(y)) != 0 {
					break OUTER
				}
			}
			// All the regroup columns were equivalent.
			last++
		}

		// Copy over the values by index.
		indices := arrow.IntSlice(rowIndices, first, last)
		vals := make([]array.Interface, len(cols))
		for j, col := range cols {
			b := arrow.NewBuilder(col.Type, mem)
			b.Resize(last - first)
			arrowutil.CopyByIndexTo(b, arrs[j], indices)
			vals[j] = b.NewArray()
		}
		indices.Release()

		if err := m.processTable(d, key, cols, vals); err != nil {
			return err
		}
		first = last
	}
	return nil
}

// determineKeyColumns determines which columns should be part of the group key.
// If a column previously existed in the group key and does not exist in the output,
// it will not be returned here.
//
// This returns the index of the key column in the list of columns along with a
// template for the key columns.
func (m *mapTransformation2) determineKeyColumns(cols []flux.ColMeta, key flux.GroupKey) ([]int, []flux.ColMeta) {
	indices := make([]int, 0, len(key.Cols()))
	keyCols := make([]flux.ColMeta, 0, len(key.Cols()))
	for i, col := range cols {
		if key.HasCol(col.Label) {
			indices = append(indices, i)
			keyCols = append(keyCols, col)
		}
	}
	return indices, keyCols
}

// regroupWith determines which columns will need to be used to regroup.
// A column needs to be regrouped if it was part of the group key and the values
// are not a single constant value.
//
// If the group key columns are all constants, then they would all end up in
// the same group key and we don't need to regroup. That is represented by returning
// an empty slice.
func (m *mapTransformation2) regroupWith(keyIndices []int, arrs []array.Interface) []int {
	regroup := make([]int, 0, len(keyIndices))
	for _, idx := range keyIndices {
		if !arrowutil.IsConstant(arrs[idx]) {
			regroup = append(regroup, idx)
		}
	}
	return regroup
}

// makeKey will construct a group key using the given values in the row.
func (m *mapTransformation2) makeKey(keyIndices []int, keyCols []flux.ColMeta, cols []flux.ColMeta, arrs []array.Interface, row int) flux.GroupKey {
	buffer := arrow.TableBuffer{
		Columns: cols,
		Values:  arrs,
	}
	vals := make([]values.Value, len(keyCols))
	for i, idx := range keyIndices {
		vals[i] = execute.ValueForRow(&buffer, row, idx)
	}
	return execute.NewGroupKey(keyCols, vals)
}

// processTable is a utility function for creating a table chunk and sending it through the transport.
func (m *mapTransformation2) processTable(d *execute.TransportDataset, key flux.GroupKey, cols []flux.ColMeta, arrs []array.Interface) error {
	buffer := arrow.TableBuffer{
		GroupKey: key,
		Columns:  cols,
		Values:   arrs,
	}
	chunk := table.ChunkFromBuffer(buffer)
	return d.Process(chunk)
}

func (m *mapTransformation2) Close() error {
	return nil
}

type mapFunc interface {
	Prepare(cols []flux.ColMeta) (mapPreparedFunc, error)
}

type mapPreparedFunc interface {
	Eval(ctx context.Context, chunk table.Chunk, mem memory.Allocator) ([]flux.ColMeta, []array.Interface, error)
}

type mapRowFunc struct {
	fn *execute.RowMapFn
}

func (m *mapRowFunc) Prepare(cols []flux.ColMeta) (mapPreparedFunc, error) {
	fn, err := m.fn.Prepare(cols)
	if err != nil {
		return nil, err
	}
	return &mapRowPreparedFunc{
		fn: fn,
	}, nil
}

type mapRowPreparedFunc struct {
	fn *execute.RowMapPreparedFn
}

func (m *mapRowPreparedFunc) initialize(cols []flux.ColMeta, mem memory.Allocator) []array.Builder {
	builders := make([]array.Builder, len(cols))
	for i, col := range cols {
		builders[i] = arrow.NewBuilder(col.Type, mem)
	}
	return builders
}

func (m *mapRowPreparedFunc) createSchema(record values.Object) ([]flux.ColMeta, error) {
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

func (m *mapRowPreparedFunc) Eval(ctx context.Context, chunk table.Chunk, mem memory.Allocator) ([]flux.ColMeta, []array.Interface, error) {
	var (
		cols     []flux.ColMeta
		builders []array.Builder
	)

	buffer := chunk.Buffer()
	for i, n := 0, chunk.Len(); i < n; i++ {
		res, err := m.fn.Eval(ctx, i, &buffer)
		if err != nil {
			return nil, nil, errors.Wrap(err, codes.Invalid, "failed to evaluate map function")
		}

		if i == 0 {
			cols, err = m.createSchema(res)
			if err != nil {
				return nil, nil, err
			}

			builders = m.initialize(cols, mem)
			for _, b := range builders {
				b.Resize(n)
			}
		}

		for i, col := range cols {
			v, _ := res.Get(col.Label)
			if err := arrow.AppendValue(builders[i], v); err != nil {
				return nil, nil, err
			}
		}
	}

	arrs := make([]array.Interface, len(builders))
	for i, b := range builders {
		arrs[i] = b.NewArray()
	}
	return cols, arrs, nil
}

type mapVectorFunc struct {
	fn *execute.VectorMapFn
}

func (m *mapVectorFunc) Prepare(cols []flux.ColMeta) (mapPreparedFunc, error) {
	fn, err := m.fn.Prepare(cols)
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

func (m *mapVectorPreparedFunc) Eval(ctx context.Context, chunk table.Chunk, mem memory.Allocator) ([]flux.ColMeta, []array.Interface, error) {
	ret := m.fn.Type()
	n, err := ret.NumProperties()
	if err != nil {
		return nil, nil, err
	}

	arr, err := m.fn.Eval(ctx, chunk)
	if err != nil {
		return nil, nil, err
	}

	nulls := 0
	cols := make([]flux.ColMeta, n)
	for i := range cols {
		// This array was null so we will filter it out later.
		if arr[i] == nil {
			nulls++
			continue
		}

		prop, err := ret.RecordProperty(i)
		if err != nil {
			return nil, nil, err
		}

		typ, err := prop.TypeOf()
		if err != nil {
			return nil, nil, err
		}

		if typ.Nature() != semantic.Vector {
			return nil, nil, errors.Newf(codes.Internal, "column %s is not a vector", prop.Name())
		}

		elem, err := typ.ElemType()
		if err != nil {
			return nil, nil, err
		}

		cols[i] = flux.ColMeta{
			Label: prop.Name(),
			Type:  flux.ColumnType(elem),
		}
		if cols[i].Type == flux.TInvalid {
			return nil, nil, errors.Newf(codes.FailedPrecondition, "column %s is not a basic type, is of type %s", prop.Name(), elem)
		}
	}

	if nulls > 0 {
		newArrs := make([]array.Interface, 0, len(arr)-nulls)
		newCols := make([]flux.ColMeta, 0, cap(newArrs))
		for i := range arr {
			if arr[i] != nil {
				newArrs = append(newArrs, arr[i])
				newCols = append(newCols, cols[i])
			}
		}
		cols, arr = newCols, newArrs
	}
	return cols, arr, nil
}
