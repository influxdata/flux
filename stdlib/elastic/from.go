package elastic

import (
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const FromKind = "elastic.from"

type FromOpSpec struct {
	URL      string        `json:"url,omitempty"`
	QueryDSL values.Object `json:"queryDSL,omitempty"`
	User     string        `json:"user,omitempty"`
	Password string        `json:"password,omitempty"`
}

func init() {
	fromSignature := semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"url":      semantic.String,
			"queryDSL": semantic.Object,
			"user":     semantic.String,
			"password": semantic.String,
		},
		Required: semantic.LabelSet{"queryDSL"},
		Return:   flux.TableObjectType,
	}
	flux.RegisterPackageValue("elastic", "from", flux.FunctionValue(FromKind, createFromOpSpec, fromSignature))
	flux.RegisterOpSpec(FromKind, newFromOpSpec)
	plan.RegisterProcedureSpec(FromKind, newFromProcedureSpec, FromKind)
	execute.RegisterSource(FromKind, createFromSource)
}

func createFromOpSpec(args flux.Arguments, _ *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromOpSpec)

	if url, _, err := args.GetString("url"); err != nil {
		return nil, err
	} else {
		spec.URL = url
	}

	if query, err := args.GetRequiredObject("queryDSL"); err != nil {
		return nil, err
	} else {
		spec.QueryDSL = query
	}

	if user, _, err := args.GetString("user"); err != nil {
		return nil, err
	} else {
		spec.User = user
	}

	if password, _, err := args.GetString("password"); err != nil {
		return nil, err
	} else {
		spec.Password = password
	}

	return spec, nil
}

func newFromOpSpec() flux.OperationSpec {
	return new(FromOpSpec)
}

func (s *FromOpSpec) Kind() flux.OperationKind {
	return FromKind
}

type FromProcedureSpec struct {
	plan.DefaultCost
	URL      string
	QueryDSL values.Object
	User     string
	Password string
}

func newFromProcedureSpec(qs flux.OperationSpec, _ plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FromProcedureSpec{
		URL:      spec.URL,
		QueryDSL: spec.QueryDSL,
		User:     spec.User,
		Password: spec.Password,
	}, nil
}

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return FromKind
}

func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromProcedureSpec)
	ns.URL = s.URL
	ns.QueryDSL = s.QueryDSL
	ns.User = s.User
	ns.Password = s.Password
	return ns
}

func createFromSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*FromProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", prSpec)
	}

	ElasticIterator := Iterator{id: dsid, spec: spec, administration: a}

	return execute.CreateSourceFromDecoder(&ElasticIterator, dsid, a)
}

type Iterator struct {
	id             execute.DatasetID
	administration execute.Administration
	spec           *FromProcedureSpec
	client         *Client
	result         *map[string]interface{}
}

func (c *Iterator) Connect() error {
	client, err := NewClient(c.spec.URL, c.spec.User, c.spec.Password)
	if err != nil {
		return err
	}

	if err = client.Ping(); err != nil {
		return err
	}
	c.client = client

	return nil
}

func (c *Iterator) Fetch() (bool, error) {

	searchResult, err := c.client.Query(c.spec.QueryDSL)

	if err != nil {
		return false, err
	}
	c.result = searchResult

	return false, nil
}

func (c *Iterator) Decode() (flux.Table, error) {
	//groupKey := execute.NewGroupKey(nil, nil)
	//builder := execute.NewColListTableBuilder(groupKey, c.administration.Allocator())
	//
	//firstRow := true
	//for _, hit := range c.result.Hits.Hits {
	//	item := make(map[string]interface{})
	//	err := json.Unmarshal(*hit.Source, &item)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	if firstRow {
	//		for name, value := range item {
	//			var dataType flux.ColType
	//			switch value.(type) {
	//			case bool:
	//				dataType = flux.TBool
	//			case int64:
	//				dataType = flux.TInt
	//			case uint64:
	//				dataType = flux.TUInt
	//			case float64:
	//				dataType = flux.TFloat
	//			case string:
	//				dataType = flux.TString
	//			case []uint8:
	//				// Hack for MySQL, might need to work with charset? TODO
	//				dataType = flux.TString
	//			case time.Time:
	//				dataType = flux.TTime
	//			default:
	//				fmt.Println(name, reflect.TypeOf(value))
	//				execute.PanicUnknownType(flux.TInvalid)
	//			}
	//
	//			_, err := builder.AddCol(flux.ColMeta{Label: name, Type: dataType})
	//			if err != nil {
	//				return nil, err
	//			}
	//		}
	//		firstRow = false
	//	}
	//	var j = 0
	//	for _, value := range item {
	//		switch value.(type) {
	//		case bool:
	//			if err := builder.AppendBool(j, value.(bool)); err != nil {
	//				return nil, err
	//			}
	//		case int64:
	//			if err := builder.AppendInt(j, value.(int64)); err != nil {
	//				return nil, err
	//			}
	//		case uint64:
	//			if err := builder.AppendUInt(j, value.(uint64)); err != nil {
	//				return nil, err
	//			}
	//		case float64:
	//			if err := builder.AppendFloat(j, value.(float64)); err != nil {
	//				return nil, err
	//			}
	//		case string:
	//			if err := builder.AppendString(j, value.(string)); err != nil {
	//				return nil, err
	//			}
	//		case []uint8:
	//			// Hack for MySQL, might need to work with charset? #TODO
	//			if err := builder.AppendString(j, string(value.([]uint8))); err != nil {
	//				return nil, err
	//			}
	//		case time.Time:
	//			if err := builder.AppendTime(j, values.ConvertTime(value.(time.Time))); err != nil {
	//				return nil, err
	//			}
	//		default:
	//			execute.PanicUnknownType(flux.TInvalid)
	//		}
	//		j++
	//	}
	//
	//}
	//
	//return builder.Table()
}

func (c *Iterator) Close() error {
	return nil
}
