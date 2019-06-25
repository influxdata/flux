package execute

import (
	"database/sql"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/values"
)

type RowReader interface {
	Next() bool
	GetNextRow() ([]values.Value, error)
	InitColumnNames([]string)
	InitColumnTypes([]*sql.ColumnType)
	ColumnNames() []string
	ColumnTypes() []flux.ColType
	SetColumns([]interface{})
}
