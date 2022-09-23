package influxdb

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/apache/arrow/go/v7/arrow/bitutil"
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/opentracing/opentracing-go"
)

// ToKind is the kind for the `to` flux function
const ToKind = "to"

type (
	Tag    = influxdb.Tag
	Field  = influxdb.Field
	Metric = influxdb.Metric
	Writer = influxdb.Writer
)

func init() {
	toSignature := runtime.MustLookupBuiltinType("influxdata/influxdb", "to")
	runtime.RegisterPackageValue("influxdata/influxdb", ToKind, flux.MustValue(flux.FunctionValueWithSideEffect(ToKind, createToOpSpec, toSignature)))
	plan.RegisterProcedureSpecWithSideEffect(ToKind, newToProcedure, ToKind)
	execute.RegisterTransformation(ToKind, createToTransformation)
}

const (
	defaultToFieldColLabel       = DefaultFieldColLabel
	defaultToMeasurementColLabel = DefaultMeasurementColLabel
	toOp                         = "influxdata/influxdb/to"
)

func createToTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ToProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}

	deps := influxdb.GetProvider(a.Context())
	return NewToTransformation(a.Context(), id, s, deps, a.Allocator())
}

// toTransformation is the transformation for the `to` flux function.
type toTransformation struct {
	ctx                context.Context
	fn                 *execute.RowMapFn
	spec               *ToOpSpec
	implicitTagColumns bool
	tagColumns         []string
	writer             influxdb.Writer
	span               opentracing.Span
}

// NewToTransformation returns a new *ToTransformation with the appropriate fields set.
func NewToTransformation(ctx context.Context, id execute.DatasetID, spec *ToProcedureSpec, deps influxdb.Provider, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	var fn *execute.RowMapFn
	if spec.Spec.FieldFn.Fn != nil {
		fn = execute.NewRowMapFn(spec.Spec.FieldFn.Fn, compiler.ToScope(spec.Spec.FieldFn.Scope))
	}

	org := NameOrID{
		ID:   spec.Spec.OrgID,
		Name: spec.Spec.Org,
	}
	bucket := NameOrID{
		ID:   spec.Spec.BucketID,
		Name: spec.Spec.Bucket,
	}

	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "ToTransformation.Process")

	conf := influxdb.Config{
		Org:    org,
		Bucket: bucket,
		Host:   spec.Spec.Host,
		Token:  spec.Spec.Token,
	}
	writer, err := deps.WriterFor(ctx, conf)
	if err != nil {
		return nil, nil, err
	}

	return execute.NewNarrowTransformation(id, &toTransformation{
		ctx:                ctx,
		fn:                 fn,
		spec:               spec.Spec,
		implicitTagColumns: spec.Spec.TagColumns == nil,
		tagColumns:         append([]string(nil), spec.Spec.TagColumns...),
		writer:             writer,
		span:               span,
	}, mem)
}

// Process does the actual work for the ToTransformation.
func (t *toTransformation) Process(chunk table.Chunk, d *execute.TransportDataset, mem memory.Allocator) error {
	// If no tag columns are specified, by default we exclude
	// _field and _value from being tag columns.
	if t.implicitTagColumns {
		excludeColumns := map[string]bool{
			execute.DefaultValueColLabel: true,
			defaultToFieldColLabel:       true,
			defaultToMeasurementColLabel: true,
		}

		// If a field function is specified then we exclude any column that
		// is referenced in the function expression from being a tag column.
		if t.spec.FieldFn.Fn != nil {
			recordParam := t.spec.FieldFn.Fn.Parameters.List[0].Key.Name.Name()
			exprNode := t.spec.FieldFn.Fn.Block
			colVisitor := newFieldFunctionVisitor(recordParam, chunk.Cols())

			// Walk the field function expression and record which columns
			// are referenced. None of these columns will be used as tag columns.
			semantic.Walk(colVisitor, exprNode)
			for k, v := range colVisitor.captured {
				excludeColumns[k] = v
			}
		}

		t.addTagsFromTable(chunk.Cols(), excludeColumns)
	}

	if err := t.writeTable(chunk); err != nil {
		return err
	}

	// Filter out rows with null times if they exist.
	filtered := t.filterNulls(chunk, mem)
	return d.Process(filtered)
}

func (t *toTransformation) addTagsFromTable(cols []flux.ColMeta, exclude map[string]bool) {
	if cap(t.tagColumns) < len(cols) {
		t.tagColumns = make([]string, 0, len(cols))
	} else {
		t.tagColumns = t.tagColumns[:0]
	}

	for _, column := range cols {
		if column.Type == flux.TString && !exclude[column.Label] {
			t.tagColumns = append(t.tagColumns, column.Label)
		}
	}
	sort.Strings(t.tagColumns)
}

func (t *toTransformation) writeTable(chunk table.Chunk) (err error) {
	spec := t.spec

	// cache tag columns
	columns := chunk.Cols()
	isTag := make([]bool, len(columns))
	for i, col := range columns {
		tagIdx := sort.SearchStrings(t.tagColumns, col.Label)
		isTag[i] = tagIdx < len(t.tagColumns) && t.tagColumns[tagIdx] == col.Label
	}

	// do measurement
	measurementColLabel := spec.MeasurementColumn
	measurementColIdx := execute.ColIdx(measurementColLabel, columns)

	if measurementColIdx < 0 {
		return errors.Newf(codes.Invalid, "no column with label %s exists", measurementColLabel)
	} else if columns[measurementColIdx].Type != flux.TString {
		return errors.Newf(codes.Invalid, "column %s of type %s is not of type %s", measurementColLabel, columns[measurementColIdx].Type, flux.TString)
	}

	// do time
	timeColLabel := spec.TimeColumn
	timeColIdx := execute.ColIdx(timeColLabel, columns)

	if timeColIdx < 0 {
		return errors.New(codes.Invalid, "no time column detected")
	} else if columns[timeColIdx].Type != flux.TTime {
		return errors.Newf(codes.Invalid, "column %s of type %s is not of type %s", timeColLabel, columns[timeColIdx].Type, flux.TTime)
	}

	// prepare field function if applicable and record the number of values to write per row
	var fn *execute.RowMapPreparedFn
	if spec.FieldFn.Fn != nil {
		var err error
		if fn, err = t.fn.Prepare(t.ctx, columns); err != nil {
			return err
		}
	}

	var fieldValues values.Object
	metrics := make([]Metric, 0, chunk.Len())
	er := chunk.Buffer()

outer:
	for i := 0; i < chunk.Len(); i++ {
		metric := &RowMetric{
			Tags: make([]*Tag, 0, len(t.tagColumns)),
		}

		// gather the timestamp, tags and measurement name
		for j, col := range chunk.Cols() {
			switch {
			case col.Label == spec.MeasurementColumn:
				metric.NameStr = er.Strings(j).Value(i)
			case col.Label == timeColLabel:
				valueTime := execute.ValueForRow(&er, i, j)
				if valueTime.IsNull() {
					// skip rows with null timestamp
					continue outer
				}
				metric.TS = valueTime.Time().Time()
			case isTag[j]:
				if col.Type != flux.TString {
					return errors.New(codes.Invalid, "invalid type for tag column")
				}

				value := er.Strings(j).Value(i)
				if value == "" {
					// Skip tag value if it is empty.
					continue
				}

				metric.Tags = append(metric.Tags, &Tag{
					Key:   col.Label,
					Value: value,
				})
			}
		}

		if metric.TS.IsZero() {
			return errors.New(codes.Invalid, "timestamp missing from block")
		}

		if fn == nil {
			if fieldValues, err = defaultFieldMapping(&er, i); err != nil {
				return err
			}
		} else if fieldValues, err = fn.Eval(t.ctx, i, &er); err != nil {
			return err
		}

		metric.Fields = make([]*Field, 0, fieldValues.Len())

		var err error

		fieldValues.Range(func(k string, v values.Value) {
			if !v.IsNull() {
				field := &Field{Key: k}

				switch v.Type().Nature() {
				case semantic.Float:
					fv := v.Float()
					if math.IsNaN(fv) || math.IsInf(fv, 0) {
						// Cannot write NaN or Inf points.
						return
					}
					field.Value = fv
				case semantic.Int:
					field.Value = v.Int()
				case semantic.UInt:
					field.Value = v.UInt()
				case semantic.String:
					field.Value = v.Str()
				case semantic.Time:
					field.Value = int64(v.Time())
				case semantic.Bool:
					field.Value = v.Bool()
				default:
					if err == nil {
						err = fmt.Errorf("unsupported field type %v", v.Type())
					}

					return
				}

				metric.Fields = append(metric.Fields, field)
			}
		})

		if err != nil {
			return err
		}

		// drop metrics without any measurements
		if len(metric.Fields) > 0 {
			metrics = append(metrics, metric)
		}
	}

	// only write if we have any metrics to write
	if len(metrics) > 0 {
		err = t.writer.Write(metrics...)
	}

	return err
}

// filterNulls will filter out the rows where the time is null from the table chunk.
// If the table chunk does not have any rows where the time is null, it retains and
// returns the original table chunk.
//
// This filter is necessary because the `to()` function will only output the rows that
// are written and rows where the time is null are not written.
func (t *toTransformation) filterNulls(chunk table.Chunk, mem memory.Allocator) table.Chunk {
	idx := execute.ColIdx(t.spec.TimeColumn, chunk.Cols())
	ts := chunk.Ints(idx)
	if ts.NullN() == 0 {
		// If there are no null values, no filtering is needed.
		// Retain a copy of this table chunk and send it along.
		chunk.Retain()
		return chunk
	}

	bitset := memory.NewResizableBuffer(mem)
	bitset.Resize(ts.Len())

	for i, n := 0, ts.Len(); i < n; i++ {
		bitutil.SetBitTo(bitset.Buf(), i, ts.IsValid(i))
	}

	buffer := chunk.Buffer()
	buffer.Values = make([]array.Array, chunk.NCols())
	for j := range buffer.Values {
		arr := chunk.Values(j)
		buffer.Values[j] = arrowutil.Filter(arr, bitset.Bytes(), mem)
	}
	return table.ChunkFromBuffer(buffer)
}

func (t *toTransformation) Close() error {
	defer t.span.Finish()
	return t.writer.Close()
}

// fieldFunctionVisitor implements semantic.Visitor.
// fieldFunctionVisitor is used to walk the the field function expression
// of the `to` operation and to record all referenced columns. This visitor
// is only used when no tag columns are provided as input to the `to` func.
type fieldFunctionVisitor struct {
	columns  map[string]bool
	visited  map[semantic.Node]bool
	captured map[string]bool
	rowParam string
}

func newFieldFunctionVisitor(rowParam string, cols []flux.ColMeta) *fieldFunctionVisitor {
	columns := make(map[string]bool, len(cols))
	for _, col := range cols {
		columns[col.Label] = true
	}
	return &fieldFunctionVisitor{
		columns:  columns,
		visited:  make(map[semantic.Node]bool, len(cols)),
		captured: make(map[string]bool, len(cols)),
		rowParam: rowParam,
	}
}

// A field function is of the form `(r) => { Function Body }`, and it returns an object
// mapping field keys to values for each row r of the input. Visit records every column
// that is referenced in `Function Body`. These columns are either directly or indirectly
// used as value columns and as such need to be recorded so as not to be used as tag columns.
func (v *fieldFunctionVisitor) Visit(node semantic.Node) semantic.Visitor {
	if v.visited[node] {
		return v
	}
	if member, ok := node.(*semantic.MemberExpression); ok {
		if obj, ok := member.Object.(*semantic.IdentifierExpression); ok {
			if obj.Name.Name() == v.rowParam && v.columns[member.Property.Name()] {
				v.captured[member.Property.Name()] = true
			}
		}
	}
	v.visited[node] = true
	return v
}

func (v *fieldFunctionVisitor) Done(semantic.Node) {}

func defaultFieldMapping(er flux.ColReader, row int) (values.Object, error) {
	fieldColumnIdx := execute.ColIdx(defaultToFieldColLabel, er.Cols())
	valueColumnIdx := execute.ColIdx(execute.DefaultValueColLabel, er.Cols())

	if fieldColumnIdx < 0 {
		return nil, errors.New(codes.Invalid, "table has no _field column")
	}

	if valueColumnIdx < 0 {
		return nil, errors.New(codes.Invalid, "table has no _value column")
	}

	value := execute.ValueForRow(er, row, valueColumnIdx)
	field := execute.ValueForRow(er, row, fieldColumnIdx)
	props := []semantic.PropertyType{
		{
			Key:   []byte(field.Str()),
			Value: value.Type(),
		},
	}
	fieldValueMapping := values.NewObject(semantic.NewObjectType(props))
	fieldValueMapping.Set(field.Str(), value)
	return fieldValueMapping, nil
}

// ///////////////////
// from idpe query

// ToOpSpec is the flux.OperationSpec for the `to` flux function.
type ToOpSpec struct {
	Bucket            string                       `json:"bucket"`
	BucketID          string                       `json:"bucketID"`
	Org               string                       `json:"org"`
	OrgID             string                       `json:"orgID"`
	Host              string                       `json:"host"`
	Token             string                       `json:"token"`
	TimeColumn        string                       `json:"timeColumn"`
	MeasurementColumn string                       `json:"measurementColumn"`
	TagColumns        []string                     `json:"tagColumns"`
	FieldFn           interpreter.ResolvedFunction `json:"fieldFn"`
}

// ToProcedureSpec is the procedure spec for the `to` flux function.
type ToProcedureSpec struct {
	plan.DefaultCost
	Spec *ToOpSpec
}

func (o *ToProcedureSpec) PassThroughAttribute(attrKey string) bool {
	switch attrKey {
	case plan.ParallelRunKey, plan.CollationKey:
		return true
	}
	return false
}

// Kind returns the kind for the procedure spec for the `to` flux function.
func (o *ToProcedureSpec) Kind() plan.ProcedureKind {
	return ToKind
}

// Copy clones the procedure spec for `to` flux function.
func (o *ToProcedureSpec) Copy() plan.ProcedureSpec {
	s := o.Spec
	res := &ToProcedureSpec{
		Spec: &ToOpSpec{
			Bucket:            s.Bucket,
			BucketID:          s.BucketID,
			Org:               s.Org,
			OrgID:             s.OrgID,
			Host:              s.Host,
			Token:             s.Token,
			TimeColumn:        s.TimeColumn,
			MeasurementColumn: s.MeasurementColumn,
			TagColumns:        append([]string(nil), s.TagColumns...),
			FieldFn:           s.FieldFn.Copy(),
		},
	}
	return res
}

func newToProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ToOpSpec)
	if !ok && spec != nil {
		return nil, &flux.Error{
			Code: codes.Internal,
			Msg:  fmt.Sprintf("invalid spec type %T", qs),
		}
	}
	return &ToProcedureSpec{Spec: spec}, nil
}

// ReadArgs reads the args from flux.Arguments into the op spec
func (o *ToOpSpec) ReadArgs(args flux.Arguments) error {
	var err error
	var ok bool

	if o.Bucket, ok, _ = args.GetString("bucket"); !ok {
		if o.BucketID, err = args.GetRequiredString("bucketID"); err != nil {
			return err
		}
	} else if o.BucketID, ok, _ = args.GetString("bucketID"); ok {
		return &flux.Error{
			Code: codes.Invalid,
			Msg:  "cannot provide both `bucket` and `bucketID` parameters to the `to` function",
		}
	}

	if o.Org, ok, _ = args.GetString("org"); !ok {
		if o.OrgID, _, err = args.GetString("orgID"); err != nil {
			return err
		}
	} else if o.OrgID, ok, _ = args.GetString("orgID"); ok {
		return &flux.Error{
			Code: codes.Invalid,
			Msg:  "cannot provide both `org` and `orgID` parameters to the `to` function",
		}
	}

	if o.Host, ok, _ = args.GetString("host"); ok {
		if o.Token, err = args.GetRequiredString("token"); err != nil {
			return err
		}
	}

	if o.TimeColumn, ok, _ = args.GetString("timeColumn"); !ok {
		o.TimeColumn = execute.DefaultTimeColLabel
	}

	if o.MeasurementColumn, ok, _ = args.GetString("measurementColumn"); !ok {
		o.MeasurementColumn = defaultToMeasurementColLabel
	}

	if tags, ok, _ := args.GetArray("tagColumns", semantic.String); ok {
		o.TagColumns = make([]string, tags.Len())
		tags.Sort(func(i, j values.Value) bool {
			return i.Str() < j.Str()
		})
		tags.Range(func(i int, v values.Value) {
			o.TagColumns[i] = v.Str()
		})
	}

	if fieldFn, ok, _ := args.GetFunction("fieldFn"); ok {
		if o.FieldFn, err = interpreter.ResolveFunction(fieldFn); err != nil {
			return err
		}
	}

	return err
}

func createToOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	s := &ToOpSpec{}
	if err := s.ReadArgs(args); err != nil {
		return nil, err
	}
	return s, nil
}

// Kind returns the kind for the ToOpSpec function.
func (ToOpSpec) Kind() flux.OperationKind {
	return ToKind
}
