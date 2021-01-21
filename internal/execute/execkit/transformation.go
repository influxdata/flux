package execkit

import (
	"github.com/influxdata/flux/execute"
)

// Transformation is a method of transforming a Table or Tables into another Table.
// The Transformation is kept at a bare-minimum to keep it simple.
// It contains one method which tells it to process the next message received from an upstream.
// The Message can then be typecast into the proper underlying message type.
//
// It is recommended to use one of the Transformation types that implement a specific type
// of transformation.
//
// For backwards compatibility, Transformation also implements execute.Transformation.
type Transformation interface {
	execute.Transformation
	execute.Transport
}
