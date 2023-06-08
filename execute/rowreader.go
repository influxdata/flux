package execute

import (
	"io"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/values"
)

type RowReader interface {
	Next() bool
	GetNextRow() ([]values.Value, error)
	ColumnNames() []string
	ColumnTypes() []flux.ColType
	SetColumns([]interface{})
	io.Closer
}
