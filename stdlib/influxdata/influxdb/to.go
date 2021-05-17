package influxdb

import (
	"context"
	"fmt"
	"sort"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/compiler"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb/internal"
	"github.com/influxdata/flux/values"
	lp "github.com/influxdata/line-protocol"
	"github.com/opentracing/opentracing-go"
)

// ToKind is the kind for the `to` flux function
const ToKind = "to"

func init() {
	toSignature := runtime.MustLookupBuiltinType("influxdata/influxdb", "to")
	runtime.RegisterPackageValue("influxdata/influxdb", ToKind, flux.MustValue(flux.FunctionValueWithSideEffect(ToKind, createToOpSpec, toSignature)))
	plan.RegisterProcedureSpecWithSideEffect(ToKind, newToProcedure, ToKind)
	execute.RegisterTransformation(ToKind, createToTransformation)
}

const (
	// TODO(jlapacik) remove this once we have execute.DefaultFieldColLabel
	defaultFieldColLabel       = "_field"
	defaultMeasurementColLabel = "_measurement"
	toOp                       = "influxdata/influxdb/to"
)

func createToTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ToProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}

	var (
		cache = execute.NewTableBuilderCache(a.Allocator())
		d     = execute.NewDataset(id, mode, cache)
		deps  = influxdb.GetProvider(a.Context())
	)

	t, err := NewToTransformation(a.Context(), d, cache, s, deps)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

// ToTransformation is the transformation for the `to` flux function.
type ToTransformation struct {
	execute.ExecutionNode
	ctx                context.Context
	bucket             NameOrID
	org                NameOrID
	d                  execute.Dataset
	fn                 *execute.RowMapFn
	cache              execute.TableBuilderCache
	spec               *ToOpSpec
	implicitTagColumns bool
	writer             influxdb.Writer
	span               opentracing.Span
}

// RetractTable retracts the table for the transformation for the `to` flux function.
func (t *ToTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

// NewToTransformation returns a new *ToTransformation with the appropriate fields set.
func NewToTransformation(ctx context.Context, d execute.Dataset, cache execute.TableBuilderCache, spec *ToProcedureSpec, deps influxdb.Provider) (*ToTransformation, error) {
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
		return nil, err
	}

	return &ToTransformation{
		ctx:                ctx,
		bucket:             bucket,
		org:                org,
		d:                  d,
		fn:                 fn,
		cache:              cache,
		spec:               spec.Spec,
		implicitTagColumns: spec.Spec.TagColumns == nil,
		writer:             writer,
		span:               span,
	}, nil
}

// Process does the actual work for the ToTransformation.
func (t *ToTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// If no tag columns are specified, by default we exclude
	// _field and _value from being tag columns.
	if t.implicitTagColumns {
		excludeColumns := map[string]bool{
			execute.DefaultValueColLabel: true,
			defaultFieldColLabel:         true,
			defaultMeasurementColLabel:   true,
		}

		// If a field function is specified then we exclude any column that
		// is referenced in the function expression from being a tag column.
		if t.spec.FieldFn.Fn != nil {
			recordParam := t.spec.FieldFn.Fn.Parameters.List[0].Key.Name
			exprNode := t.spec.FieldFn.Fn.Block
			colVisitor := newFieldFunctionVisitor(recordParam, tbl.Cols())

			// Walk the field function expression and record which columns
			// are referenced. None of these columns will be used as tag columns.
			semantic.Walk(colVisitor, exprNode)
			for k, v := range colVisitor.captured {
				excludeColumns[k] = v
			}
		}

		addTagsFromTable(t.spec, tbl, excludeColumns)
	}
	return writeTableToAPI(t.ctx, t, tbl)
}

// UpdateWatermark updates the watermark for the transformation for the `to` flux function.
func (t *ToTransformation) UpdateWatermark(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateWatermark(pt)
}

// UpdateProcessingTime updates the processing time for the transformation for the `to` flux function.
func (t *ToTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

// Finish is called after the `to` flux function's transformation is done processing.
func (t *ToTransformation) Finish(id execute.DatasetID, err error) {
	defer t.span.Finish()

	if err != nil {
		t.d.Finish(err)
		return
	}

	err = t.writer.Close()
	t.d.Finish(err)
}

func writeTableToAPI(ctx context.Context, t *ToTransformation, tbl flux.Table) (err error) {
	spec := t.spec

	builder, isNew := t.cache.TableBuilder(tbl.Key())
	if isNew {
		if err := execute.AddTableCols(tbl, builder); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("to() found duplicate table with group key: %v", tbl.Key())
	}

	// cache tag columns
	columns := tbl.Cols()
	isTag := make([]bool, len(columns))
	for i, col := range columns {
		tagIdx := sort.SearchStrings(spec.TagColumns, col.Label)
		isTag[i] = tagIdx < len(spec.TagColumns) && spec.TagColumns[tagIdx] == col.Label
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
		if fn, err = t.fn.Prepare(columns); err != nil {
			return err
		}
	}

	var fieldValues values.Object
	return tbl.Do(func(er flux.ColReader) error {
		var metrics []lp.Metric
		metrics = make([]lp.Metric, 0, er.Len())

	outer:
		for i := 0; i < er.Len(); i++ {
			metric := &internal.RowMetric{
				Tags: make([]*lp.Tag, 0, len(spec.TagColumns)),
			}

			// gather the timestamp, tags and measurement name
			for j, col := range er.Cols() {
				switch {
				case col.Label == spec.MeasurementColumn:
					metric.NameStr = er.Strings(j).ValueString(i)
				case col.Label == timeColLabel:
					valueTime := execute.ValueForRow(er, i, j)
					if valueTime.IsNull() {
						// skip rows with null timestamp
						continue outer
					}
					metric.TS = valueTime.Time().Time()
				case isTag[j]:
					if col.Type != flux.TString {
						return errors.New(codes.Invalid, "invalid type for tag column")
					}

					metric.Tags = append(metric.Tags, &lp.Tag{
						Key:   col.Label,
						Value: er.Strings(j).ValueString(i),
					})
				}
			}

			if metric.TS.IsZero() {
				return errors.New(codes.Invalid, "timestamp missing from block")
			}

			if fn == nil {
				if fieldValues, err = defaultFieldMapping(er, i); err != nil {
					return err
				}
			} else if fieldValues, err = fn.Eval(t.ctx, i, er); err != nil {
				return err
			}

			metric.Fields = make([]*lp.Field, 0, fieldValues.Len())

			var err error

			fieldValues.Range(func(k string, v values.Value) {
				if !v.IsNull() {
					field := &lp.Field{Key: k}

					switch v.Type().Nature() {
					case semantic.Float:
						field.Value = v.Float()
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

			if err := execute.AppendRecord(i, er, builder); err != nil {
				return err
			}
		}

		// only write if we have any metrics to write
		if len(metrics) > 0 {
			err = t.writer.Write(metrics...)
		}

		return err
	})
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
			if obj.Name == v.rowParam && v.columns[member.Property] {
				v.captured[member.Property] = true
			}
		}
	}
	v.visited[node] = true
	return v
}

func (v *fieldFunctionVisitor) Done(semantic.Node) {}

func addTagsFromTable(spec *ToOpSpec, table flux.Table, exclude map[string]bool) {
	if cap(spec.TagColumns) < len(table.Cols()) {
		spec.TagColumns = make([]string, 0, len(table.Cols()))
	} else {
		spec.TagColumns = spec.TagColumns[:0]
	}

	for _, column := range table.Cols() {
		if column.Type == flux.TString && !exclude[column.Label] {
			spec.TagColumns = append(spec.TagColumns, column.Label)
		}
	}
	sort.Strings(spec.TagColumns)
}

func defaultFieldMapping(er flux.ColReader, row int) (values.Object, error) {
	fieldColumnIdx := execute.ColIdx(defaultFieldColLabel, er.Cols())
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

/////////////////////
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
		o.MeasurementColumn = defaultMeasurementColLabel
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
