package execute

import "github.com/influxdata/flux"

func NewProcessMsg(tbl flux.Table) ProcessMsg {
	return &processMsg{table: tbl}
}
