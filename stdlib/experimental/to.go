package experimental

import (
	"context"
	"fmt"
	"sort"
	"time"

	lp "github.com/influxdata/line-protocol"
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/codes"
	"github.com/mvn-trinhnguyen2-dn/flux/execute"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/errors"
	"github.com/mvn-trinhnguyen2-dn/flux/plan"
	"github.com/mvn-trinhnguyen2-dn/flux/runtime"
	"github.com/mvn-trinhnguyen2-dn/flux/stdlib/influxdata/influxdb"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

const ToKind = "experimental-to"

func init() {
	toSignature := runtime.MustLookupBuiltinType("experimental", "to")
	runtime.RegisterPackageValue("experimental", "to", flux.MustValue(flux.FunctionValueWithSideEffect("to", createToOpSpec, toSignature)))
	plan.RegisterProcedureSpecWithSideEffect(ToKind, newToProcedure, ToKind)
	execute.RegisterTransformation(ToKind, createToTransformation)
}

// ToOpSpec is the flux.OperationSpec for the `to` flux function.
type ToOpSpec struct {
	Org    influxdb.NameOrID
	Bucket influxdb.NameOrID
	Host   string
	Token  string
}

// ReadArgs reads the args from flux.Arguments into the op spec
func (s *ToOpSpec) ReadArgs(args flux.Arguments) error {
	if b, ok, err := influxdb.GetNameOrID(args, "bucket", "bucketID"); err != nil {
		return err
	} else if !ok {
		return errors.New(codes.Invalid, "must specify bucket or bucketID")
	} else {
		s.Bucket = b
	}

	if o, ok, err := influxdb.GetNameOrID(args, "org", "orgID"); err != nil {
		return err
	} else if ok {
		s.Org = o
	}

	if host, ok, err := args.GetString("host"); err != nil {
		return err
	} else if ok {
		s.Host = host
	}

	if token, ok, err := args.GetString("token"); err != nil {
		return err
	} else if ok {
		s.Token = token
	}
	return nil
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

// ToProcedureSpec is the procedure spec for the `to` flux function.
type ToProcedureSpec struct {
	plan.DefaultCost
	Config influxdb.Config
}

// Kind returns the kind for the procedure spec for the `to` flux function.
func (o *ToProcedureSpec) Kind() plan.ProcedureKind {
	return ToKind
}

// Copy clones the procedure spec for `to` flux function.
func (o *ToProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *o
	return &ns
}

func newToProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ToOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &ToProcedureSpec{
		Config: influxdb.Config{
			Org:    spec.Org,
			Bucket: spec.Bucket,
			Host:   spec.Host,
			Token:  spec.Token,
		},
	}, nil
}

func createToTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ToProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)

	t, err := NewToTransformation(a.Context(), d, cache, s)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

// ToTransformation is the transformation for the `to` flux function.
type ToTransformation struct {
	execute.ExecutionNode
	ctx    context.Context
	d      execute.Dataset
	cache  execute.TableBuilderCache
	writer influxdb.Writer
}

// RetractTable retracts the table for the transformation for the `to` flux function.
func (t *ToTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

// NewToTransformation returns a new *ToTransformation with the appropriate fields set.
func NewToTransformation(ctx context.Context, d execute.Dataset, cache execute.TableBuilderCache, s *ToProcedureSpec) (*ToTransformation, error) {
	provider := influxdb.GetProvider(ctx)
	writer, err := provider.WriterFor(ctx, s.Config)
	if err != nil {
		return nil, err
	}
	return &ToTransformation{
		ctx:    ctx,
		d:      d,
		cache:  cache,
		writer: writer,
	}, nil
}

// Process does the actual work for the ToTransformation.
func (t *ToTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	return t.writeTable(tbl)
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
	writeErr := t.writer.Close()
	if err == nil {
		err = writeErr
	}
	t.d.Finish(err)
}

const (
	defaultFieldColLabel       = influxdb.DefaultFieldColLabel
	defaultMeasurementColLabel = influxdb.DefaultMeasurementColLabel
	defaultTimeColLabel        = execute.DefaultTimeColLabel
	defaultStartColLabel       = execute.DefaultStartColLabel
	defaultStopColLabel        = execute.DefaultStopColLabel
)

type LabelAndOffset struct {
	Label  string
	Offset int
}

// tablePointsMetadata stores state needed to write the points from one table.
type tablePointsMetadata struct {
	// Name is the measurement name for this table.
	Name string
	// Tags holds the tags in the table excluding the measurement.
	Tags []*lp.Tag
	// The column offset in the input table where the _time column is stored
	TimestampOffset int
	// The labels and offsets of all the fields in the table
	Fields []LabelAndOffset
}

func getTablePointsMetadata(tbl flux.Table) (md tablePointsMetadata, err error) {
	// Find measurement, tags
	foundMeasurement := false
	md.Tags = make([]*lp.Tag, 0, len(tbl.Key().Cols()))
	isTag := make(map[string]bool)
	for j, col := range tbl.Key().Cols() {
		switch col.Label {
		case defaultStartColLabel:
			continue
		case defaultStopColLabel:
			continue
		case defaultFieldColLabel:
			return md, errors.Newf(codes.FailedPrecondition, "found column %q in the group key; experimental.to() expects pivoted data", col.Label)
		case defaultMeasurementColLabel:
			foundMeasurement = true
			if col.Type != flux.TString {
				return md, errors.Newf(codes.FailedPrecondition, "group key column %q has type %v; type %v is required", col.Label, col.Type, flux.TString)
			}
			md.Name = tbl.Key().ValueString(j)
		default:
			if col.Type != flux.TString {
				return md, errors.Newf(codes.FailedPrecondition, "group key column %q has type %v; type %v is required", col.Label, col.Type, flux.TString)
			}
			isTag[col.Label] = true
			md.Tags = append(md.Tags, &lp.Tag{
				Key:   col.Label,
				Value: tbl.Key().ValueString(j),
			})
		}
	}
	sort.SliceStable(md.Tags, func(i, j int) bool {
		return md.Tags[i].Key < md.Tags[j].Key
	})
	if !foundMeasurement {
		return md, errors.Newf(codes.FailedPrecondition, "required column %q not in group key", defaultMeasurementColLabel)
	}

	// Find the time column as it is required.
	md.TimestampOffset = execute.ColIdx(defaultTimeColLabel, tbl.Cols())
	if md.TimestampOffset < 0 {
		return md, errors.Newf(codes.FailedPrecondition, "input table is missing required column %q", defaultTimeColLabel)
	} else if col := tbl.Cols()[md.TimestampOffset]; col.Type != flux.TTime {
		return md, errors.Newf(codes.FailedPrecondition, "column %q has type %s; type %s is required", defaultTimeColLabel, col.Type, flux.TTime)
	}

	// Loop over all of the remaining columns to find the fields.
	// By this point, we know all of the tags and we can exclude the time
	// column from the list of fields so we can allocate an appropriate size.
	md.Fields = make([]LabelAndOffset, 0, len(tbl.Cols())-len(md.Tags)-1)
	for j, col := range tbl.Cols() {
		switch col.Label {
		case defaultStartColLabel, defaultStopColLabel, defaultMeasurementColLabel, defaultTimeColLabel:
			continue
		default:
			if !isTag[col.Label] {
				md.Fields = append(md.Fields, LabelAndOffset{
					Label:  col.Label,
					Offset: j,
				})
			}
		}
	}
	return md, nil
}

func (t *ToTransformation) writeTable(tbl flux.Table) error {
	builder, new := t.cache.TableBuilder(tbl.Key())
	if new {
		if err := execute.AddTableCols(tbl, builder); err != nil {
			return err
		}
	}

	tmd, err := getTablePointsMetadata(tbl)
	if err != nil {
		return err
	}

	return tbl.Do(func(cr flux.ColReader) error {
		if cr.Len() == 0 {
			// Nothing to do
			return nil
		}

		metrics := make([]lp.Metric, 0, cr.Len())
		for i := 0; i < cr.Len(); i++ {
			timestamp := cr.Times(tmd.TimestampOffset).Value(i)
			metric := &influxdb.RowMetric{
				NameStr: tmd.Name,
				TS:      time.Unix(0, timestamp),
				Tags:    tmd.Tags,
				Fields:  make([]*lp.Field, 0, len(tmd.Fields)),
			}
			for _, lao := range tmd.Fields {
				fieldVal := execute.ValueForRow(cr, i, lao.Offset)

				// Skip this iteration if field value is null
				if fieldVal.IsNull() {
					continue
				}

				metric.Fields = append(metric.Fields, &lp.Field{
					Key:   lao.Label,
					Value: values.Unwrap(fieldVal),
				})
			}

			if len(metric.Fields) > 0 {
				metrics = append(metrics, metric)
			}

			if err := execute.AppendRecord(i, cr, builder); err != nil {
				return err
			}
		}
		return t.writer.Write(metrics...)
	})
}
