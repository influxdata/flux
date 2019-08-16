package bigtable

import (
	"cloud.google.com/go/bigtable"
	"context"
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
	"google.golang.org/api/option"
)

const FromBigtableKind = "fromBigtable"

type FromBigtableOpSpec struct {
	Token    string `json:"token,omitempty"`
	Project  string `json:"project,omitempty"`
	Instance string `json:"instance,omitempty"`
	Table    string `json:"table,omitempty"`
}

func init() {
	fromBigtableSignature := semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"token":      semantic.String,
			"connection": semantic.Object,
			"table":      semantic.String,
		},
		Required: semantic.LabelSet{"token", "connection", "table"},
		Return:   flux.TableObjectType,
	}
	flux.RegisterPackageValue("experimental/bigtable", "from", flux.FunctionValue(FromBigtableKind, createFromBigtableOpSpec, fromBigtableSignature))
	flux.RegisterOpSpec(FromBigtableKind, newFromBigtableOp)
	plan.RegisterProcedureSpec(FromBigtableKind, newFromBigtableProcedure, FromBigtableKind)
	plan.RegisterPhysicalRules(BigtableFilterRewriteRule{}, BigtableLimitRewriteRule{})
	execute.RegisterSource(FromBigtableKind, createFromBigtableSource)
}

func createFromBigtableOpSpec(args flux.Arguments, administration *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromBigtableOpSpec)

	if token, err := args.GetRequiredString("token"); err != nil {
		return nil, err
	} else {
		spec.Token = token
	}

	if connection, err := args.GetRequiredObject("connection"); err != nil {
		return nil, err
	} else {
		project, ok := connection.Get("project")
		if !ok {
			return nil, fmt.Errorf("invalid connection object")
		}
		instance, ok := connection.Get("instance")
		if !ok {
			return nil, fmt.Errorf("invalid connection object")
		}
		spec.Project = project.Str()
		spec.Instance = instance.Str()
	}

	if table, err := args.GetRequiredString("table"); err != nil {
		return nil, err
	} else {
		spec.Table = table
	}

	return spec, nil
}

func newFromBigtableOp() flux.OperationSpec {
	return new(FromBigtableOpSpec)
}

func (s *FromBigtableOpSpec) Kind() flux.OperationKind {
	return FromBigtableKind
}

type FromBigtableProcedureSpec struct {
	plan.DefaultCost
	Token    string
	Project  string
	Instance string
	Table    string

	// Used by BigtableFilterRewriteRule
	RowSet      bigtable.RowSet
	Filter      bigtable.Filter
	ReadOptions []bigtable.ReadOption
}

func newFromBigtableProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromBigtableOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FromBigtableProcedureSpec{
		Token:       spec.Token,
		Project:     spec.Project,
		Instance:    spec.Instance,
		Table:       spec.Table,
		Filter:      bigtable.PassAllFilter(),
		ReadOptions: make([]bigtable.ReadOption, 0),
	}, nil
}

func (s *FromBigtableProcedureSpec) Kind() plan.ProcedureKind {
	return FromBigtableKind
}

func (s *FromBigtableProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromBigtableProcedureSpec)
	ns.Token = s.Token
	ns.Project = s.Project
	ns.Instance = s.Instance
	ns.Table = s.Table
	ns.RowSet = s.RowSet
	ns.Filter = s.Filter
	ns.ReadOptions = make([]bigtable.ReadOption, 0)
	for _, v := range s.ReadOptions {
		ns.ReadOptions = append(ns.ReadOptions, v)
	}

	return ns
}

func createFromBigtableSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*FromBigtableProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", prSpec)
	}

	BigtableDecoder := BigtableDecoder{id: dsid, administration: a, spec: spec}

	return execute.CreateSourceFromDecoder(&BigtableDecoder, dsid, a)
}

type FamilyRow []bigtable.ReadItem

type BigtableDecoder struct {
	id             execute.DatasetID
	administration execute.Administration
	spec           *FromBigtableProcedureSpec

	client *bigtable.Client
	tbl    *bigtable.Table

	reader *execute.RowReader
}

var _ execute.SourceDecoder = (*BigtableDecoder)(nil)

func (c *BigtableDecoder) Connect(ctx context.Context) error {
	client, err := bigtable.NewClient(ctx, c.spec.Project, c.spec.Instance, option.WithCredentialsJSON([]byte(c.spec.Token)))
	if err != nil {
		return err
	}

	c.client = client
	c.tbl = client.Open(c.spec.Table)

	return nil
}

func (c *BigtableDecoder) Fetch(ctx context.Context) (bool, error) {
	// On the first Fetch, get all the data
	if c.reader == nil {
		r, err := NewBigtableRowReader(ctx, c)
		if err != nil {
			return false, err
		}
		c.reader = &r
	}

	// Every time we call Fetch, change which family we want to look at
	return (*c.reader).(*BigtableRowReader).nextFamily()
}

func (c *BigtableDecoder) Decode(ctx context.Context) (flux.Table, error) {
	familyCol := flux.ColMeta{Label: "family", Type: flux.TString}

	var groupKey flux.GroupKey
	if bigtableReader, ok := (*c.reader).(*BigtableRowReader); ok {
		groupKey = execute.NewGroupKey([]flux.ColMeta{familyCol}, []values.Value{values.NewString(bigtableReader.currentFamily())})
	} else {
		groupKey = execute.NewGroupKey(nil, nil)
	}
	builder := execute.NewColListTableBuilder(groupKey, c.administration.Allocator())

	// Every table will have a row key, timestamp, and family
	rowKeyIdx, err := builder.AddCol(flux.ColMeta{Label: "rowKey", Type: flux.TString})
	if err != nil {
		return nil, err
	}
	timeIdx, err := builder.AddCol(flux.ColMeta{Label: execute.DefaultTimeColLabel, Type: flux.TTime})
	if err != nil {
		return nil, err
	}
	familyIdx, err := builder.AddCol(familyCol)
	if err != nil {
		return nil, err
	}

	columns := map[string]int{"rowKey": rowKeyIdx, execute.DefaultTimeColLabel: timeIdx, "family": familyIdx}
	rowIndex := 0
	reader := *c.reader

	for reader.Next() {
		rowValues, err := reader.GetNextRow()
		if err != nil {
			return nil, err
		}

		for j, val := range rowValues {
			if j >= len(columns) {
				label := reader.ColumnNames()[j]
				idx, err := builder.AddCol(flux.ColMeta{Label: label, Type: flux.TString})
				if err != nil {
					return nil, err
				}
				columns[label] = idx
				if err := builder.SetValue(rowIndex, j, val); err != nil {
					return nil, err
				}
			} else {
				if err := builder.AppendValue(j, val); err != nil {
					return nil, err
				}
			}
		}
		rowIndex++
	}

	return builder.Table()
}

func (c *BigtableDecoder) Close() error {
	return c.client.Close()
}

type BigtableFilterRewriteRule struct{}

func (r BigtableFilterRewriteRule) Name() string {
	return "BigtableFilterRewriteRule"
}

func (r BigtableFilterRewriteRule) Pattern() plan.Pattern {
	return plan.Pat(universe.FilterKind, plan.Pat(FromBigtableKind))
}

func (r BigtableFilterRewriteRule) Rewrite(filter plan.Node) (plan.Node, bool, error) {
	query := filter.Predecessors()[0]

	node, changed := AddFilterToNode(query, filter)
	return node, changed, nil
}

type BigtableLimitRewriteRule struct{}

func (r BigtableLimitRewriteRule) Name() string {
	return "BigtableLimitRewriteRule"
}

func (r BigtableLimitRewriteRule) Pattern() plan.Pattern {
	return plan.Pat(universe.LimitKind, plan.Pat(FromBigtableKind))
}

func (r BigtableLimitRewriteRule) Rewrite(limit plan.Node) (plan.Node, bool, error) {
	query := limit.Predecessors()[0]

	node, changed := AddLimitToNode(query, limit)
	return node, changed, nil
}
