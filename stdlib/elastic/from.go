package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/olivere/elastic"
	"github.com/olivere/elastic/config"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const FromKind = "elastic.from"

type FromOpSpec struct {
	DataSourceName string `json:"dataSourceName,omitempty"`
	Query          string `json:"query,omitempty"`
}

func init() {
	fromSignature := semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"dataSourceName": semantic.String,
			"query":          semantic.String,
		},
		Required: semantic.LabelSet{"dataSourceName", "query"},
		Return:   flux.TableObjectType,
	}
	flux.RegisterPackageValue("elastic", "from", flux.FunctionValue(FromKind, createFromOpSpec, fromSignature))
	flux.RegisterOpSpec(FromKind, newFromOp)
	plan.RegisterProcedureSpec(FromKind, newFromProcedure, FromKind)
	execute.RegisterSource(FromKind, createFromSource)
}

func createFromOpSpec(args flux.Arguments, administration *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromOpSpec)

	if dataSourceName, err := args.GetRequiredString("dataSourceName"); err != nil {
		return nil, err
	} else {
		spec.DataSourceName = dataSourceName
	}

	if query, err := args.GetRequiredString("query"); err != nil {
		return nil, err
	} else {
		spec.Query = query
	}

	return spec, nil
}

func newFromOp() flux.OperationSpec {
	return new(FromOpSpec)
}

func (s *FromOpSpec) Kind() flux.OperationKind {
	return FromKind
}

type FromProcedureSpec struct {
	plan.DefaultCost
	DataSourceName string
	Query          string
}

func newFromProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FromProcedureSpec{
		DataSourceName: spec.DataSourceName,
		Query:          spec.Query,
	}, nil
}

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return FromKind
}

func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromProcedureSpec)
	ns.DataSourceName = s.DataSourceName
	ns.Query = s.Query
	return ns
}

func createFromSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*FromProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", prSpec)
	}

	ElasticIterator := ElasticIterator{id: dsid, spec: spec, administration: a}

	return execute.CreateSourceFromDecoder(&ElasticIterator, dsid, a)
}

type ElasticIterator struct {
	id             execute.DatasetID
	administration execute.Administration
	spec           *FromProcedureSpec
	client         *elastic.Client
	searchResult   *elastic.SearchResult
}

func (c *ElasticIterator) Connect() error {
	cfg, err := config.Parse(c.spec.DataSourceName)
	if err != nil {
		return err
	}
	client, err := elastic.NewClientFromConfig(cfg)
	if err != nil {
		return err
	}
	if _, _, err = client.Ping(cfg.URL).Do(context.TODO()); err != nil {
		return err
	}
	c.client = client

	return nil
}

func (c *ElasticIterator) Fetch() (bool, error) {
	searchResult, err := c.client.Search().
		Query(elastic.NewSimpleQueryStringQuery(c.spec.Query)).
		Do(context.TODO())

	if err != nil {
		return false, err
	}
	c.searchResult = searchResult

	return false, nil
}

func (c *ElasticIterator) Decode() (flux.Table, error) {
	groupKey := execute.NewGroupKey(nil, nil)
	builder := execute.NewColListTableBuilder(groupKey, c.administration.Allocator())

	firstRow := true
	for _, hit := range c.searchResult.Hits.Hits {
		item := make(map[string]interface{})
		err := json.Unmarshal(*hit.Source, &item)
		if err != nil {
			return nil, err
		}

		if firstRow {
			for name, value := range item {
				var dataType flux.ColType
				switch value.(type) {
				case bool:
					dataType = flux.TBool
				case int64:
					dataType = flux.TInt
				case uint64:
					dataType = flux.TUInt
				case float64:
					dataType = flux.TFloat
				case string:
					dataType = flux.TString
				case []uint8:
					// Hack for MySQL, might need to work with charset? TODO
					dataType = flux.TString
				case time.Time:
					dataType = flux.TTime
				default:
					fmt.Println(name, reflect.TypeOf(value))
					execute.PanicUnknownType(flux.TInvalid)
				}

				_, err := builder.AddCol(flux.ColMeta{Label: name, Type: dataType})
				if err != nil {
					return nil, err
				}
			}
			firstRow = false
		}
		var j = 0
		for _, value := range item {
			switch value.(type) {
			case bool:
				if err := builder.AppendBool(j, value.(bool)); err != nil {
					return nil, err
				}
			case int64:
				if err := builder.AppendInt(j, value.(int64)); err != nil {
					return nil, err
				}
			case uint64:
				if err := builder.AppendUInt(j, value.(uint64)); err != nil {
					return nil, err
				}
			case float64:
				if err := builder.AppendFloat(j, value.(float64)); err != nil {
					return nil, err
				}
			case string:
				if err := builder.AppendString(j, value.(string)); err != nil {
					return nil, err
				}
			case []uint8:
				// Hack for MySQL, might need to work with charset? #TODO
				if err := builder.AppendString(j, string(value.([]uint8))); err != nil {
					return nil, err
				}
			case time.Time:
				if err := builder.AppendTime(j, values.ConvertTime(value.(time.Time))); err != nil {
					return nil, err
				}
			default:
				execute.PanicUnknownType(flux.TInvalid)
			}
			j++
		}

	}

	return builder.Table()
}

func (c *ElasticIterator) Close() error {
	var err error
	_, err = c.client.Flush().Do(context.TODO())
	c.client = nil
	return err
}
