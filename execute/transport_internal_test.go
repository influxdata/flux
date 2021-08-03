package execute

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/table"
)

func NewProcessMsg(tbl flux.Table) ProcessMsg {
	return &processMsg{table: tbl}
}

func NewProcessChunkMsg(chunk table.Chunk) ProcessChunkMsg {
	return &processChunkMsg{chunk: chunk}
}
