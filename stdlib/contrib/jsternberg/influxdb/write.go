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

const WriteKind = pkgpath + ".write"

func init() {
	runtime.RegisterPackageValue(pkgpath, "write", flux.MustValue(flux.FunctionValueWithSideEffect(
		"write",
		createWriteOpSpec,
		runtime.MustLookupBuiltinType(pkgpath, "write"),
	)))
	plan.RegisterProcedureSpec(WriteKind, newWriteProcedure, WriteKind)
	execute.RegisterTransformation(WriteKind, createWriteTransformation)
}

type WriteOpSpec struct {
	Org    string
	Bucket string
	Host   string
	Token  string
}

func createWriteOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(WriteOpSpec)

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

func (s *WriteOpSpec) Kind() flux.OperationKind {
	return WriteKind
}

type WriteProcedureSpec struct {
	plan.DefaultCost

	Org         influxdb.NameOrID
	Bucket      influxdb.NameOrID
	Measurement string
	Host        string
	Token       string
}

func newWriteProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*WriteOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &WriteProcedureSpec{
		Org:    influxdb.NameOrID{Name: spec.Org},
		Bucket: influxdb.NameOrID{Name: spec.Bucket},
		Host:   spec.Host,
		Token:  spec.Token,
	}, nil
}

func (s *WriteProcedureSpec) Kind() plan.ProcedureKind {
	return WriteKind
}

func (s *WriteProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(WriteProcedureSpec)
	*ns = *s
	return ns
}

func (s *WriteProcedureSpec) SetOrg(org *influxdb.NameOrID) { s.Org = *org }
func (s *WriteProcedureSpec) SetHost(host *string)          { s.Host = *host }
func (s *WriteProcedureSpec) SetToken(token *string)        { s.Token = *token }
func (s *WriteProcedureSpec) GetOrg() *influxdb.NameOrID {
	if s.Org.Name != "" || s.Org.ID != "" {
		return &s.Org
	}
	return nil
}
func (s *WriteProcedureSpec) GetHost() *string {
	if s.Host != "" {
		return &s.Host
	}
	return nil
}
func (s *WriteProcedureSpec) GetToken() *string {
	if s.Token != "" {
		return &s.Token
	}
	return nil
}

func (s *WriteProcedureSpec) PostPhysicalValidate(id plan.NodeID) error {
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

func createWriteTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*WriteProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewWriteTransformation(a.Context(), s, id)
}

type writeTransformation struct {
	ctx    context.Context
	d      *execute.PassthroughDataset
	name   string
	writer influxdb.PointsWriter
}

func NewWriteTransformation(ctx context.Context, spec *WriteProcedureSpec, id execute.DatasetID) (execute.Transformation, execute.Dataset, error) {
	wp := influxdb.GetPointsWriter(ctx)
	pw, err := wp.WriterFor(ctx, spec.Org, spec.Bucket, spec.Host, spec.Token)
	if err != nil {
		return nil, nil, err
	}
	t := &writeTransformation{
		ctx:    ctx,
		d:      execute.NewPassthroughDataset(id),
		name:   spec.Measurement,
		writer: pw,
	}
	return t, t.d, nil
}

func (t *writeTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *writeTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	tidx, fields, err := t.getReadSchema(tbl)
	if err != nil {
		return err
	} else if len(fields) == 0 {
		return errors.Newf(codes.FailedPrecondition, "at least one field is requireed")
	}

	// Retrieve the tags for the written value.
	tags, err := t.getTags(tbl.Key())
	if err != nil {
		return err
	}

	// Write the table.
	outTable, err := t.writeTable(tbl, tidx, tags, fields)
	if err != nil {
		return err
	}
	return t.d.Process(outTable)
}

func (t *writeTransformation) getReadSchema(tbl flux.Table) (time int, fields []int, err error) {
	time = -1
	key, cols := tbl.Key(), tbl.Cols()

	fields = make([]int, 0, len(cols)-len(key.Cols()))
	for i, col := range cols {
		if key.HasCol(col.Label) {
			continue
		} else if col.Label == "_time" || col.Label == "time" {
			if i == -1 {
				time = i
				if typ := cols[time].Type; typ != flux.TTime {
					return -1, nil, errors.Newf(codes.FailedPrecondition, "%s column is of type %s but %s is required", col.Label, typ, flux.TTime)
				}
			}
			continue
		}
		fields = append(fields, i)
	}

	if time < 0 {
		return -1, nil, errors.New(codes.FailedPrecondition, "_time column not found")
	}
	return time, fields, nil
}

func (t *writeTransformation) getTags(key flux.GroupKey) ([]*lp.Tag, error) {
	tags := make([]*lp.Tag, 0, len(key.Cols()))
	for i, c := range key.Cols() {
		if c.Type != flux.TString {
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

func (t *writeTransformation) writeTable(tbl flux.Table, tidx int, tags []*lp.Tag, fields []int) (flux.Table, error) {
	cols := tbl.Cols()
	p := point{
		tags:   tags,
		fields: make([]*lp.Field, len(fields)),
	}
	for i := range fields {
		p.fields[i] = &lp.Field{}
	}
	return table.StreamWithContext(t.ctx, tbl.Key(), tbl.Cols(), func(ctx context.Context, w *table.StreamWriter) error {
		return tbl.Do(func(cr flux.ColReader) error {
			ts := cr.Times(tidx)
			vs := make([]arrowutil.Values, len(fields))
			for i, j := range fields {
				vs[i] = arrowutil.AsValues(table.Values(cr, j))
			}

			// Read each of the values and write the point.
			for i, sz := 0, cr.Len(); i < sz; i++ {
				if ts.IsNull(i) {
					continue
				}
				ncols := 0

				p.name = t.name
				for j, idx := range fields {
					if vs[j].IsNull(i) {
						continue
					}
					p.fields[ncols].Key = cols[idx].Label
					p.fields[ncols].Value = vs[j].Value(i)
					ncols++
				}
				p.time = time.Unix(0, ts.Value(i))

				np := p
				np.fields = p.fields[:ncols]
				if err := t.writer.WritePoint(&np); err != nil {
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

func (t *writeTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *writeTransformation) UpdateProcessingTime(id execute.DatasetID, ts execute.Time) error {
	return t.d.UpdateProcessingTime(ts)
}

func (t *writeTransformation) Finish(id execute.DatasetID, err error) {
	if werr := t.writer.Close(); werr != nil && err == nil {
		err = werr
	}
	t.d.Finish(err)
}
