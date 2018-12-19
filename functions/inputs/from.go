package inputs

import (
	"github.com/influxdata/flux"
)

const FromKind = "from"

type FromOpSpec struct {
	Bucket   string `json:"bucket,omitempty"`
	BucketID string `json:"bucketID,omitempty"`
}

func init() {
	flux.RegisterOpSpec(FromKind, newFromOp)
}

func newFromOp() flux.OperationSpec {
	return new(FromOpSpec)
}

func (s *FromOpSpec) Kind() flux.OperationKind {
	return FromKind
}
