package iox

import (
	"context"

	stdarrow "github.com/apache/arrow/go/v7/arrow"
	arrowarray "github.com/apache/arrow/go/v7/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/dependencies/iox"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/function"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
)

const SqlKind = "experimental/iox.sql"

type SqlProcedureSpec struct {
	plan.DefaultCost
	Config iox.Config
	Query  string
}

func createSqlProcedureSpec(args *function.Arguments) (function.Source, error) {
	bucket, err := args.GetRequiredString("bucket")
	if err != nil {
		return nil, err
	}

	query, err := args.GetRequiredString("query")
	if err != nil {
		return nil, err
	}
	return &SqlProcedureSpec{
		Config: iox.Config{
			Bucket: influxdb.NameOrID{
				Name: bucket,
			},
		},
		Query: query,
	}, nil
}

func (s *SqlProcedureSpec) Kind() plan.ProcedureKind {
	return SqlKind
}

func (s *SqlProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	return &ns
}

func (s *SqlProcedureSpec) CreateSource(id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	ctx := a.Context()
	provider := iox.GetProvider(ctx)
	client, err := provider.ClientFor(ctx, s.Config)
	if err != nil {
		return nil, err
	}

	return &sqlSource{
		d:      execute.NewTransportDataset(id, a.Allocator()),
		client: client,
		query:  s.Query,
		mem:    a.Allocator(),
	}, nil
}

type sqlSource struct {
	execute.ExecutionNode
	d *execute.TransportDataset

	client iox.Client
	query  string
	mem    memory.Allocator
}

func (s *sqlSource) AddTransformation(t execute.Transformation) {
	s.d.AddTransformation(t)
}

func (s *sqlSource) Run(ctx context.Context) {
	err := s.run(ctx)
	s.d.Finish(err)
}

func (s *sqlSource) createSchema(schema *stdarrow.Schema) ([]flux.ColMeta, error) {
	fields := schema.Fields()
	cols := make([]flux.ColMeta, len(fields))
	for i, f := range fields {
		cols[i].Label = f.Name
		switch id := f.Type.ID(); id {
		case stdarrow.INT64:
			cols[i].Type = flux.TInt
		case stdarrow.UINT64:
			cols[i].Type = flux.TUInt
		case stdarrow.FLOAT64:
			cols[i].Type = flux.TFloat
		case stdarrow.STRING:
			cols[i].Type = flux.TString
		case stdarrow.BOOL:
			cols[i].Type = flux.TBool
		case stdarrow.TIMESTAMP:
			cols[i].Type = flux.TTime
		default:
			return nil, errors.Newf(codes.Internal, "unsupported arrow type %v", id)
		}
	}
	return cols, nil
}

func (s *sqlSource) run(ctx context.Context) error {
	// Note: query args are not actually supported yet, see
	// https://github.com/influxdata/influxdb_iox/issues/3718
	rr, err := s.client.Query(ctx, s.query, nil, s.mem)
	if err != nil {
		return err
	}
	defer rr.Release()

	cols, err := s.createSchema(rr.Schema())
	if err != nil {
		return err
	}
	key := execute.NewGroupKey(nil, nil)

	for rr.Next() {
		if err := s.produce(key, cols, rr.Record()); err != nil {
			return err
		}
	}
	return nil
}

func (s *sqlSource) produce(key flux.GroupKey, cols []flux.ColMeta, record stdarrow.Record) error {
	buffer := arrow.TableBuffer{
		GroupKey: key,
		Columns:  cols,
		Values:   make([]array.Array, len(cols)),
	}
	for i := range buffer.Columns {
		data := record.Column(i)
		switch id := data.DataType().ID(); id {
		case stdarrow.BOOL, stdarrow.INT64, stdarrow.UINT64, stdarrow.FLOAT64:
			// We can just use the data as-is.
			buffer.Values[i] = data
		case stdarrow.TIMESTAMP:
			// IOx returns time columns as Timestamp arrays, but they are really just
			// int64 arrays under the hood, so this is safe.
			// No need to retain here since calling NewInt64Data will bump the reference
			// count on the underlying data.
			rawData := data.(*arrowarray.Timestamp).Data()
			rawData.Reset(stdarrow.PrimitiveTypes.Int64, rawData.Len(), rawData.Buffers(), nil, data.NullN(), rawData.Offset())
			buffer.Values[i] = arrowarray.NewInt64Data(rawData)
		case stdarrow.STRING:
			// IOx returns string columns as String arrays, but Flux uses
			// Binary arrays. The underlying structure of the buffers is the same.
			binaryData := arrowarray.NewBinaryData(data.Data())
			buffer.Values[i] = array.NewStringFromBinaryArray(binaryData)
			binaryData.Release() // The String in data now owns this binary data.
		default:
			return errors.Newf(codes.FailedPrecondition, "unsupported arrow data type %v", id)
		}
		buffer.Values[i].Retain()
	}

	chunk := table.ChunkFromBuffer(buffer)
	return s.d.Process(chunk)
}
