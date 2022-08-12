package gen

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/gen"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const TablesKind = "internal/gen.tables"

type Tag struct {
	Name        string `json:"name"`
	Cardinality int    `json:"cardinality"`
}

type TablesOpSpec struct {
	N     int     `json:"n"`
	Tags  []Tag   `json:"tags,omitempty"`
	Nulls float64 `json:"nulls,omitempty"`
	Seed  *int64  `json:"Seed,omitempty"`
}

func init() {
	tablesSignature := runtime.MustLookupBuiltinType("internal/gen", "tables")
	runtime.RegisterPackageValue("internal/gen", "tables", flux.MustValue(flux.FunctionValue(TablesKind, createTablesOpSpec, tablesSignature)))
	plan.RegisterProcedureSpec(TablesKind, newTablesProcedure, TablesKind)
	execute.RegisterSource(TablesKind, createTablesSource)
}

func createTablesOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(TablesOpSpec)

	if n, err := args.GetRequiredInt("n"); err != nil {
		return nil, err
	} else {
		spec.N = int(n)
	}

	if tags, ok := args.Get("tags"); ok {
		if nature := tags.Type().Nature(); nature != semantic.Array {
			return nil, errors.Newf(codes.Invalid, "expected array for %q, got %s", "tags", nature)
		}

		var err error
		tags.Array().Range(func(i int, v values.Value) {
			if err != nil {
				return
			}

			if v.Type().Nature() != semantic.Object {
				err = errors.Newf(codes.Invalid, "tag at index %d must be an object", i)
				return
			}

			var tag Tag
			if v, ok := v.Object().Get("name"); !ok {
				err = errors.Newf(codes.Invalid, "missing %q parameter in tag at index %d", "name", i)
				return
			} else if v.Type().Nature() != semantic.String {
				err = errors.Newf(codes.Invalid, "expected string for %q at index %d, got %s", "name", i, v.Type())
				return
			} else {
				tag.Name = v.Str()
			}

			if v, ok := v.Object().Get("cardinality"); !ok {
				err = errors.Newf(codes.Invalid, "missing %q parameter in tag at index %d", "cardinality", i)
				return
			} else if v.Type().Nature() != semantic.Int {
				err = errors.Newf(codes.Invalid, "expected int for %q at index %d, got %s", "cardinality", i, v.Type())
				return
			} else {
				tag.Cardinality = int(v.Int())
			}
			spec.Tags = append(spec.Tags, tag)
		})
	}

	if nulls, ok, err := args.GetFloat("nulls"); err != nil {
		return nil, err
	} else if ok {
		spec.Nulls = nulls
	}

	if seed, ok, err := args.GetInt("seed"); err != nil {
		return nil, err
	} else if ok {
		spec.Seed = &seed
	}

	return spec, nil
}

func (s *TablesOpSpec) Kind() flux.OperationKind {
	return TablesKind
}

type TablesProcedureSpec struct {
	plan.DefaultCost
	Schema gen.Schema
}

func newTablesProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*TablesOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	schema := gen.Schema{
		NumPoints: spec.N,
		Nulls:     spec.Nulls,
		Seed:      spec.Seed,
	}

	if len(spec.Tags) > 0 {
		schema.Tags = make([]gen.Tag, len(spec.Tags))
		for i, tag := range spec.Tags {
			schema.Tags[i] = gen.Tag{
				Name:        tag.Name,
				Cardinality: tag.Cardinality,
			}
		}
	}

	return &TablesProcedureSpec{Schema: schema}, nil
}

func (s *TablesProcedureSpec) Kind() plan.ProcedureKind {
	return TablesKind
}

func (s *TablesProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(TablesProcedureSpec)
	return ns
}

func createTablesSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*TablesProcedureSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", prSpec)
	}
	return &Source{
		id:     dsid,
		schema: spec.Schema,
		alloc:  a.Allocator(),
	}, nil
}

type Source struct {
	execute.ExecutionNode
	id execute.DatasetID
	ts execute.TransformationSet

	schema gen.Schema
	alloc  memory.Allocator
}

func (s *Source) AddTransformation(t execute.Transformation) {
	s.ts = append(s.ts, t)
}

func (s *Source) Run(ctx context.Context) {
	schema := s.schema
	schema.Alloc = s.alloc

	tables, err := gen.Input(ctx, schema)
	if err != nil {
		s.ts.Finish(s.id, err)
		return
	}

	err = tables.Do(func(table flux.Table) error {
		return s.ts.Process(s.id, table)
	})
	s.ts.Finish(s.id, err)
}
