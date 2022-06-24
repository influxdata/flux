package execute

import (
	"io"

	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

type RowReader interface {
	Next() bool
	GetNextRow() ([]values.Value, error)
	ColumnNames() []string
	ColumnTypes() []flux.ColType
	SetColumns([]interface{})
	io.Closer
}
