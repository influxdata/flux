package flux

import (
	"github.com/influxdata/flux/interpreter"
)

// Operation denotes a single operation in a query.
type Operation struct {
	ID     OperationID     `json:"id"`
	Spec   OperationSpec   `json:"spec"`
	Source OperationSource `json:"source"`
}

// OperationSpec specifies an operation as part of a query.
type OperationSpec interface {
	// Kind returns the kind of the operation.
	Kind() OperationKind
}

// OperationSource specifies the source location that created
// an operation.
type OperationSource struct {
	Stack []interpreter.StackEntry `json:"stack"`
}

// OperationID is a unique ID within a query for the operation.
type OperationID string

// OperationKind denotes the kind of operations.
type OperationKind string
