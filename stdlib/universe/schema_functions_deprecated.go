package universe

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
)

type deprecatedSchemaMutationTransformation struct {
	execute.ExecutionNode
	d        execute.Dataset
	cache    execute.TableBuilderCache
	ctx      context.Context
	mutators []SchemaMutator
}

func createDeprecatedSchemaMutationTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)

	t, err := NewDeprecatedSchemaMutationTransformation(a.Context(), spec, d, cache)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

func NewDeprecatedSchemaMutationTransformation(ctx context.Context, spec plan.ProcedureSpec, d execute.Dataset, cache execute.TableBuilderCache) (execute.Transformation, error) {
	s, ok := spec.(*SchemaMutationProcedureSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	mutators := make([]SchemaMutator, len(s.Mutations))
	for i, mutation := range s.Mutations {
		m, err := mutation.Mutator()
		if err != nil {
			return nil, err
		}
		mutators[i] = m
	}

	return &deprecatedSchemaMutationTransformation{
		d:        d,
		cache:    cache,
		mutators: mutators,
		ctx:      ctx,
	}, nil
}

func (t *deprecatedSchemaMutationTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	ctx := NewBuilderContext(tbl)
	for _, m := range t.mutators {
		err := m.Mutate(t.ctx, ctx)
		if err != nil {
			return err
		}
	}

	builder, created := t.cache.TableBuilder(ctx.Key())
	if created {
		for _, c := range ctx.Cols() {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	} else {
		// We are appending to an existing table, due to dropping columns in the group key.
		// Make sure that tables are compatible.
		if len(ctx.Cols()) != len(builder.Cols()) {
			key := builder.Key().String()
			return errors.New(codes.Invalid, "requested operation merges tables with different numbers of columns for group key "+key)
		}
		for i, cm := range ctx.Cols() {
			bcm := builder.Cols()[i]
			if cm != bcm {
				key := builder.Key().String()
				return errors.New(codes.Invalid, "requested operation merges tables with different schemas for group key "+key)
			}
		}
	}

	return tbl.Do(func(cr flux.ColReader) error {
		for i := 0; i < cr.Len(); i++ {
			if err := execute.AppendMappedRecordWithNulls(i, cr, builder, ctx.ColMap()); err != nil {
				return err
			}
		}
		return nil
	})
}

func (t *deprecatedSchemaMutationTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *deprecatedSchemaMutationTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *deprecatedSchemaMutationTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *deprecatedSchemaMutationTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
