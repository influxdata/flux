package execute

import (
	"io"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/values"
)

type RowReader interface {
	Next() bool
	GetNextRow() ([]values.Value, error)
	ColumnNames() []string
	ColumnTypes() []flux.ColType
	SetColumns([]interface{})
	io.Closer
}
