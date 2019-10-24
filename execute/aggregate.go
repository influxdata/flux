package execute

import (
	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

type aggregateTransformation struct {
	d     Dataset
	cache TableBuilderCache
	agg   Aggregate

	config AggregateConfig
}

type AggregateConfig struct {
	plan.DefaultCost
	Columns []string `json:"columns"`
}

var DefaultAggregateConfig = AggregateConfig{
	Columns: []string{DefaultValueColLabel},
}

// AggregateSignature returns a function signature common to all aggregate functions,
// with any additional arguments.
func AggregateSignature(args map[string]semantic.PolyType, required []string) semantic.FunctionPolySignature {
	if args == nil {
		args = make(map[string]semantic.PolyType)
	}
	args["column"] = semantic.String
	return flux.FunctionSignature(args, required)
}

func (c AggregateConfig) Copy() AggregateConfig {
	nc := c
	if c.Columns != nil {
		nc.Columns = make([]string, len(c.Columns))
		copy(nc.Columns, c.Columns)
	}
	return nc
}

func (c *AggregateConfig) ReadArgs(args flux.Arguments) error {
	if col, ok, err := args.GetString("column"); err != nil {
		return err
	} else if ok {
		c.Columns = []string{col}
	} else {
		c.Columns = DefaultAggregateConfig.Columns
	}
	return nil
}

func NewAggregateTransformation(d Dataset, c TableBuilderCache, agg Aggregate, config AggregateConfig) *aggregateTransformation {
	return &aggregateTransformation{
		d:      d,
		cache:  c,
		agg:    agg,
		config: config,
	}
}

func NewAggregateTransformationAndDataset(id DatasetID, mode AccumulationMode, agg Aggregate, config AggregateConfig, a *memory.Allocator) (*aggregateTransformation, Dataset) {
	cache := NewTableBuilderCache(a)
	d := NewDataset(id, mode, cache)
	return NewAggregateTransformation(d, cache, agg, config), d
}

func (t *aggregateTransformation) RetractTable(id DatasetID, key flux.GroupKey) error {
	//TODO(nathanielc): Store intermediate state for retractions
	return t.d.RetractTable(key)
}

func (t *aggregateTransformation) Process(id DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "aggregate found duplicate table with key: %v", tbl.Key())
	}

	if err := AddTableKeyCols(tbl.Key(), builder); err != nil {
		return err
	}

	builderColMap := make([]int, len(t.config.Columns))
	tableColMap := make([]int, len(t.config.Columns))
	aggregates := make([]ValueFunc, len(t.config.Columns))
	cols := tbl.Cols()
	for j, label := range t.config.Columns {
		idx := -1
		for bj, bc := range cols {
			if bc.Label == label {
				idx = bj
				break
			}
		}
		if idx < 0 {
			return errors.Newf(codes.FailedPrecondition, "column %q does not exist", label)
		}
		c := cols[idx]
		if tbl.Key().HasCol(c.Label) {
			return errors.New(codes.FailedPrecondition, "cannot aggregate columns that are part of the group key")
		}
		var vf ValueFunc
		switch c.Type {
		case flux.TBool:
			vf = t.agg.NewBoolAgg()
		case flux.TInt:
			vf = t.agg.NewIntAgg()
		case flux.TUInt:
			vf = t.agg.NewUIntAgg()
		case flux.TFloat:
			vf = t.agg.NewFloatAgg()
		case flux.TString:
			vf = t.agg.NewStringAgg()
		}
		if vf == nil {
			return errors.Newf(codes.FailedPrecondition, "unsupported aggregate column type %v", c.Type)
		}
		aggregates[j] = vf

		var err error
		builderColMap[j], err = builder.AddCol(flux.ColMeta{
			Label: c.Label,
			Type:  vf.Type(),
		})
		if err != nil {
			return err
		}
		tableColMap[j] = idx
	}

	if err := tbl.Do(func(cr flux.ColReader) error {
		for j := range t.config.Columns {
			vf := aggregates[j]

			tj := tableColMap[j]
			c := tbl.Cols()[tj]

			switch c.Type {
			case flux.TBool:
				vf.(DoBoolAgg).DoBool(cr.Bools(tj))
			case flux.TInt:
				vf.(DoIntAgg).DoInt(cr.Ints(tj))
			case flux.TUInt:
				vf.(DoUIntAgg).DoUInt(cr.UInts(tj))
			case flux.TFloat:
				vf.(DoFloatAgg).DoFloat(cr.Floats(tj))
			case flux.TString:
				vf.(DoStringAgg).DoString(cr.Strings(tj))
			default:
				return errors.Newf(codes.Invalid, "unsupported aggregate type %v", c.Type)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	for j, vf := range aggregates {
		bj := builderColMap[j]

		// If the value is null, append a null to the column.
		if vf.IsNull() {
			if err := builder.AppendNil(bj); err != nil {
				return err
			}
			continue
		}

		// Append aggregated value
		switch vf.Type() {
		case flux.TBool:
			v := vf.(BoolValueFunc).ValueBool()
			if err := builder.AppendBool(bj, v); err != nil {
				return err
			}
		case flux.TInt:
			v := vf.(IntValueFunc).ValueInt()
			if err := builder.AppendInt(bj, v); err != nil {
				return err
			}
		case flux.TUInt:
			v := vf.(UIntValueFunc).ValueUInt()
			if err := builder.AppendUInt(bj, v); err != nil {
				return err
			}
		case flux.TFloat:
			v := vf.(FloatValueFunc).ValueFloat()
			if err := builder.AppendFloat(bj, v); err != nil {
				return err
			}
		case flux.TString:
			v := vf.(StringValueFunc).ValueString()
			if err := builder.AppendString(bj, v); err != nil {
				return err
			}
		}
	}

	return AppendKeyValues(tbl.Key(), builder)
}

func (t *aggregateTransformation) UpdateWatermark(id DatasetID, mark Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *aggregateTransformation) UpdateProcessingTime(id DatasetID, pt Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *aggregateTransformation) Finish(id DatasetID, err error) {
	t.d.Finish(err)
}

type Aggregate interface {
	NewBoolAgg() DoBoolAgg
	NewIntAgg() DoIntAgg
	NewUIntAgg() DoUIntAgg
	NewFloatAgg() DoFloatAgg
	NewStringAgg() DoStringAgg
}

type ValueFunc interface {
	Type() flux.ColType
	IsNull() bool
}
type DoBoolAgg interface {
	ValueFunc
	DoBool(*array.Boolean)
}
type DoFloatAgg interface {
	ValueFunc
	DoFloat(*array.Float64)
}
type DoIntAgg interface {
	ValueFunc
	DoInt(*array.Int64)
}
type DoUIntAgg interface {
	ValueFunc
	DoUInt(*array.Uint64)
}
type DoStringAgg interface {
	ValueFunc
	DoString(*array.Binary)
}

type BoolValueFunc interface {
	ValueBool() bool
}
type FloatValueFunc interface {
	ValueFloat() float64
}
type IntValueFunc interface {
	ValueInt() int64
}
type UIntValueFunc interface {
	ValueUInt() uint64
}
type StringValueFunc interface {
	ValueString() string
}
