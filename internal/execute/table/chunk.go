package table

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute/table"
)

type Chunk = table.Chunk

func ChunkFromBuffer(buf arrow.TableBuffer) Chunk {
	return table.ChunkFromBuffer(buf)
}

func ChunkFromReader(cr flux.ColReader) Chunk {
	return table.ChunkFromReader(cr)
}
