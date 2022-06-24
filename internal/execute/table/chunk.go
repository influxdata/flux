package table

import (
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/arrow"
	"github.com/mvn-trinhnguyen2-dn/flux/execute/table"
)

type Chunk = table.Chunk

func ChunkFromBuffer(buf arrow.TableBuffer) Chunk {
	return table.ChunkFromBuffer(buf)
}

func ChunkFromReader(cr flux.ColReader) Chunk {
	return table.ChunkFromReader(cr)
}
