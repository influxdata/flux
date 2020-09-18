package influxdb

import (
	"context"
	"sort"
	"time"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	lp "github.com/influxdata/line-protocol"
)

const ToKind = pkgpath + ".to"

func init() {
	runtime.RegisterPackageValue(pkgpath, "to", flux.MustValue(flux.FunctionValueWithSideEffect(
		"to",
		createToOpSpec,
		runtime.MustLookupBuiltinType(pkgpath, "to"),
	)))
	plan.RegisterProcedureSpec(ToKind, newToProcedure, ToKind)
	execute.RegisterTransformation(ToKind, createToTransformation)
}

type ToOpSpec struct {
	Org    string
	Bucket string
	Host   string
	Token  string
}

func createToOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(ToOpSpec)

	bucket, err := args.GetRequiredString("bucket")
	if err != nil {
		return nil, err
	}
	spec.Bucket = bucket

	if org, ok, err := args.GetString("org"); err != nil {
		return nil, err
	} else if ok {
		spec.Org = org
	}

	if host, ok, err := args.GetString("host"); err != nil {
		return nil, err
	} else if ok {
		spec.Host = host
	}

	if token, ok, err := args.GetString("token"); err != nil {
		return nil, err
	} else if ok {
		spec.Token = token
	}
	return spec, nil
}

func (s *ToOpSpec) Kind() flux.OperationKind {
	return ToKind
}

type ToProcedureSpec struct {
	plan.DefaultCost

	Org         influxdb.NameOrID
	Bucket      influxdb.NameOrID
	Measurement string
	Host        string
	Token       string
}

func newToProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ToOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &ToProcedureSpec{
		Org:    influxdb.NameOrID{Name: spec.Org},
		Bucket: influxdb.NameOrID{Name: spec.Bucket},
		Host:   spec.Host,
		Token:  spec.Token,
	}, nil
}

func (s *ToProcedureSpec) Kind() plan.ProcedureKind {
	return ToKind
}

func (s *ToProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(ToProcedureSpec)
	*ns = *s
	return ns
}

func (s *ToProcedureSpec) SetOrg(org *influxdb.NameOrID) { s.Org = *org }
func (s *ToProcedureSpec) SetHost(host *string)          { s.Host = *host }
func (s *ToProcedureSpec) SetToken(token *string)        { s.Token = *token }
func (s *ToProcedureSpec) GetOrg() *influxdb.NameOrID {
	if s.Org.Name != "" || s.Org.ID != "" {
		return &s.Org
	}
	return nil
}
func (s *ToProcedureSpec) GetHost() *string {
	if s.Host != "" {
		return &s.Host
	}
	return nil
}
func (s *ToProcedureSpec) GetToken() *string {
	if s.Token != "" {
		return &s.Token
	}
	return nil
}

func (s *ToProcedureSpec) PostPhysicalValidate(id plan.NodeID) error {
	// This condition should never be met.
	// Customized planner rules within each binary should have
	// filled in either a default host or registered a from procedure
	// for when no host is specified.
	// We mark this as an internal error because it is a programming
	// error if this one ever gets hit.
	if s.Host == "" {
		return errors.New(codes.Internal, "to requires a remote host to be specified")
	}
	return nil
}

func createToTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ToProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewToTransformation(a.Context(), s, id)
}

type toTransformation struct {
	ctx    context.Context
	d      *execute.PassthroughDataset
	writer influxdb.PointsWriter
}

func NewToTransformation(ctx context.Context, spec *ToProcedureSpec, id execute.DatasetID) (execute.Transformation, execute.Dataset, error) {
	wp := influxdb.GetPointsWriter(ctx)
	pw, err := wp.WriterFor(ctx, spec.Org, spec.Bucket, spec.Host, spec.Token)
	if err != nil {
		return nil, nil, err
	}
	t := &toTransformation{
		ctx:    ctx,
		d:      execute.NewPassthroughDataset(id),
		writer: pw,
	}
	return t, t.d, nil
}

func (t *toTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *toTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	midx, fidx, tidx, vidx, err := t.getReadSchema(tbl)
	if err != nil {
		return err
	}

	// Retrieve the tags for the written value.
	tags, err := t.getTags(tbl.Key())
	if err != nil {
		return err
	}

	// Write the table.
	outTable, err := t.writeTable(tbl, tags, midx, fidx, tidx, vidx)
	if err != nil {
		return err
	}
	return t.d.Process(outTable)
}

func (t *toTransformation) getReadSchema(tbl flux.Table) (measurement, field, time, value int, err error) {
	cols := tbl.Cols()

	// Validate that we have the columns we need.
	measurement = execute.ColIdx("_measurement", cols)
	if measurement < 0 {
		return -1, -1, -1, -1, errors.New(codes.FailedPrecondition, "_measurement column not found")
	} else if typ := cols[measurement].Type; typ != flux.TString {
		return -1, -1, -1, -1, errors.Newf(codes.FailedPrecondition, "_measurement column is of type %s but %s is required", typ, flux.TString)
	}

	field = execute.ColIdx("_field", cols)
	if field < 0 {
		return -1, -1, -1, -1, errors.New(codes.FailedPrecondition, "_field column not found")
	} else if typ := cols[field].Type; typ != flux.TString {
		return -1, -1, -1, -1, errors.Newf(codes.FailedPrecondition, "_field column is of type %s but %s is required", typ, flux.TString)
	}

	time = execute.ColIdx("_time", cols)
	if time < 0 {
		return -1, -1, -1, -1, errors.New(codes.FailedPrecondition, "_time column not found")
	} else if typ := cols[time].Type; typ != flux.TTime {
		return -1, -1, -1, -1, errors.Newf(codes.FailedPrecondition, "_time column is of type %s but %s is required", typ, flux.TTime)
	}

	value = execute.ColIdx("_value", cols)
	if value < 0 {
		return -1, -1, -1, -1, errors.New(codes.FailedPrecondition, "_time column not found")
	}
	return measurement, field, time, value, nil
}

func (t *toTransformation) getTags(key flux.GroupKey) ([]*lp.Tag, error) {
	tags := make([]*lp.Tag, 0, len(key.Cols()))
	for i, c := range key.Cols() {
		// Explicitly ignore _measurement, _field, _start, and _stop.
		if c.Label == "_measurement" || c.Label == "_field" || c.Label == "_start" || c.Label == "_stop" {
			continue
		} else if c.Type != flux.TString {
			return nil, errors.Newf(codes.FailedPrecondition, "key %q is not of type %s, but %s", c.Label, flux.TString, c.Type)
		}
		tags = append(tags, &lp.Tag{
			Key:   c.Label,
			Value: key.ValueString(i),
		})
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})
	return tags, nil
}

func (t *toTransformation) writeTable(tbl flux.Table, tags []*lp.Tag, midx, fidx, tidx, vidx int) (flux.Table, error) {
	p := point{
		tags:   tags,
		fields: []*lp.Field{{}},
	}
	return table.StreamWithContext(t.ctx, tbl.Key(), tbl.Cols(), func(ctx context.Context, w *table.StreamWriter) error {
		return tbl.Do(func(cr flux.ColReader) error {
			var (
				ms = cr.Strings(midx)
				fs = cr.Strings(fidx)
				ts = cr.Times(tidx)
				vs = arrowutil.AsValues(table.Values(cr, vidx))
			)

			// Read each of the values and write the point.
			for i, sz := 0, cr.Len(); i < sz; i++ {
				if ts.IsNull(i) || vs.IsNull(i) {
					continue
				}

				p.name = ms.ValueString(i)
				p.fields[0].Key = fs.ValueString(i)
				p.fields[0].Value = vs.Value(i)
				p.time = time.Unix(0, ts.Value(i))
				if err := t.writer.WritePoint(&p); err != nil {
					return err
				}
			}

			// Retain each of the columns and pass this on.
			values := make([]array.Interface, len(cr.Cols()))
			for i := range cr.Cols() {
				values[i] = table.Values(cr, i)
				values[i].Retain()
			}
			return w.Write(values)
		})
	})
}

func (t *toTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *toTransformation) UpdateProcessingTime(id execute.DatasetID, ts execute.Time) error {
	return t.d.UpdateProcessingTime(ts)
}

func (t *toTransformation) Finish(id execute.DatasetID, err error) {
	if werr := t.writer.Close(); werr != nil && err == nil {
		err = werr
	}
	t.d.Finish(err)
}

type point struct {
	time   time.Time
	name   string
	tags   []*lp.Tag
	fields []*lp.Field
}

func (p *point) Time() time.Time        { return p.time }
func (p *point) Name() string           { return p.name }
func (p *point) TagList() []*lp.Tag     { return p.tags }
func (p *point) FieldList() []*lp.Field { return p.fields }
