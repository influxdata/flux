package influxdb

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const WideToKind = "wide-to"

func init() {
	wideToSignature := runtime.MustLookupBuiltinType("influxdata/influxdb", "wideTo")
	runtime.RegisterPackageValue("influxdata/influxdb", "wideTo", flux.MustValue(flux.FunctionValueWithSideEffect("to", createWideToOpSpec, wideToSignature)))
	plan.RegisterProcedureSpecWithSideEffect(WideToKind, newWideToProcedure, WideToKind)
	execute.RegisterTransformation(WideToKind, createWideToTransformation)
}

// WideToOpSpec is the flux.OperationSpec for the `to` flux function.
type WideToOpSpec struct {
	Org    NameOrID
	Bucket NameOrID
	Host   string
	Token  string
}

// ReadArgs reads the args from flux.Arguments into the op spec
func (s *WideToOpSpec) ReadArgs(args flux.Arguments) error {
	if b, ok, err := GetNameOrID(args, "bucket", "bucketID"); err != nil {
		return err
	} else if !ok {
		return errors.New(codes.Invalid, "must specify bucket or bucketID")
	} else {
		s.Bucket = b
	}

	if o, ok, err := GetNameOrID(args, "org", "orgID"); err != nil {
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

func createWideToOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	s := &WideToOpSpec{}
	if err := s.ReadArgs(args); err != nil {
		return nil, err
	}
	return s, nil
}

// Kind returns the kind for the WideToOpSpec function.
func (WideToOpSpec) Kind() flux.OperationKind {
	return WideToKind
}

// WideToProcedureSpec is the procedure spec for the `to` flux function.
type WideToProcedureSpec struct {
	plan.DefaultCost
	Config Config
}

// Kind returns the kind for the procedure spec for the `to` flux function.
func (o *WideToProcedureSpec) Kind() plan.ProcedureKind {
	return WideToKind
}

// Copy clones the procedure spec for `to` flux function.
func (o *WideToProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *o
	return &ns
}

func newWideToProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*WideToOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &WideToProcedureSpec{
		Config: Config{
			Org:    spec.Org,
			Bucket: spec.Bucket,
			Host:   spec.Host,
			Token:  spec.Token,
		},
	}, nil
}

func createWideToTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*WideToProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)

	t, err := NewWideToTransformation(a.Context(), d, cache, s)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

// WideToTransformation is the transformation for the `to` flux function.
type WideToTransformation struct {
	execute.ExecutionNode
	ctx    context.Context
	d      execute.Dataset
	cache  execute.TableBuilderCache
	writer Writer
}

// RetractTable retracts the table for the transformation for the `to` flux function.
func (t *WideToTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

// NewWideToTransformation returns a new *WideToTransformation with the appropriate fields set.
func NewWideToTransformation(ctx context.Context, d execute.Dataset, cache execute.TableBuilderCache, s *WideToProcedureSpec) (*WideToTransformation, error) {
	provider := GetProvider(ctx)
	writer, err := provider.WriterFor(ctx, s.Config)
	if err != nil {
		return nil, err
	}
	return &WideToTransformation{
		ctx:    ctx,
		d:      d,
		cache:  cache,
		writer: writer,
	}, nil
}

// Process does the actual work for the WideToTransformation.
func (t *WideToTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	return t.writeTable(tbl)
}

// UpdateWatermark updates the watermark for the transformation for the `to` flux function.
func (t *WideToTransformation) UpdateWatermark(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateWatermark(pt)
}

// UpdateProcessingTime updates the processing time for the transformation for the `to` flux function.
func (t *WideToTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

// Finish is called after the `to` flux function's transformation is done processing.
func (t *WideToTransformation) Finish(id execute.DatasetID, err error) {
	writeErr := t.writer.Close()
	if err == nil {
		err = writeErr
	}
	t.d.Finish(err)
}

const (
	defaultWideToFieldColLabel       = DefaultFieldColLabel
	defaultWideToMeasurementColLabel = DefaultMeasurementColLabel
	defaultWideToTimeColLabel        = execute.DefaultTimeColLabel
	defaultWideToStartColLabel       = execute.DefaultStartColLabel
	defaultWideToStopColLabel        = execute.DefaultStopColLabel
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
	Tags []*influxdb.Tag
	// The column offset in the input table where the _time column is stored
	TimestampOffset int
	// The labels and offsets of all the fields in the table
	Fields []LabelAndOffset
}

func getTablePointsMetadata(tbl flux.Table) (md tablePointsMetadata, err error) {
	// Find measurement, tags
	foundMeasurement := false
	md.Tags = make([]*influxdb.Tag, 0, len(tbl.Key().Cols()))
	isTag := make(map[string]bool)
	for j, col := range tbl.Key().Cols() {
		switch col.Label {
		case defaultWideToStartColLabel:
			continue
		case defaultWideToStopColLabel:
			continue
		case defaultWideToFieldColLabel:
			return md, errors.Newf(codes.FailedPrecondition, "found column %q in the group key; wideTo() expects pivoted data", col.Label)
		case defaultWideToMeasurementColLabel:
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

			value := tbl.Key().ValueString(j)
			if value == "" {
				// Skip tag value if it is empty.
				continue
			}
			md.Tags = append(md.Tags, &influxdb.Tag{
				Key:   col.Label,
				Value: value,
			})
		}
	}
	sort.SliceStable(md.Tags, func(i, j int) bool {
		return md.Tags[i].Key < md.Tags[j].Key
	})
	if !foundMeasurement {
		return md, errors.Newf(codes.FailedPrecondition, "required column %q not in group key", defaultWideToMeasurementColLabel)
	}

	// Find the time column as it is required.
	md.TimestampOffset = execute.ColIdx(defaultWideToTimeColLabel, tbl.Cols())
	if md.TimestampOffset < 0 {
		return md, errors.Newf(codes.FailedPrecondition, "input table is missing required column %q", defaultWideToTimeColLabel)
	} else if col := tbl.Cols()[md.TimestampOffset]; col.Type != flux.TTime {
		return md, errors.Newf(codes.FailedPrecondition, "column %q has type %s; type %s is required", defaultWideToTimeColLabel, col.Type, flux.TTime)
	}

	// Loop over all of the remaining columns to find the fields.
	// By this point, we know all of the tags and we can exclude the time
	// column from the list of fields so we can allocate an appropriate size.
	md.Fields = make([]LabelAndOffset, 0, len(tbl.Cols())-len(md.Tags)-1)
	for j, col := range tbl.Cols() {
		switch col.Label {
		case defaultWideToStartColLabel, defaultWideToStopColLabel, defaultWideToMeasurementColLabel, defaultWideToTimeColLabel:
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

func (t *WideToTransformation) writeTable(tbl flux.Table) error {
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

		metrics := make([]influxdb.Metric, 0, cr.Len())
		for i := 0; i < cr.Len(); i++ {
			timestamp := cr.Times(tmd.TimestampOffset).Value(i)
			metric := &RowMetric{
				NameStr: tmd.Name,
				TS:      time.Unix(0, timestamp),
				Tags:    tmd.Tags,
				Fields:  make([]*influxdb.Field, 0, len(tmd.Fields)),
			}
			for _, lao := range tmd.Fields {
				fieldVal := execute.ValueForRow(cr, i, lao.Offset)

				// Skip this iteration if field value is null
				if fieldVal.IsNull() {
					continue
				}

				v := values.Unwrap(fieldVal)
				if fv, ok := v.(float64); ok {
					if math.IsNaN(fv) || math.IsInf(fv, 0) {
						// Cannot write NaN or Inf points.
						continue
					}
				}

				metric.Fields = append(metric.Fields, &influxdb.Field{
					Key:   lao.Label,
					Value: v,
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
