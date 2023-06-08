package table

import (
	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/arrow"
	"github.com/InfluxCommunity/flux/execute/table"
)

type Chunk = table.Chunk

func ChunkFromBuffer(buf arrow.TableBuffer) Chunk {
	return table.ChunkFromBuffer(buf)
}

func ChunkFromReader(cr flux.ColReader) Chunk {
	return table.ChunkFromReader(cr)
}
