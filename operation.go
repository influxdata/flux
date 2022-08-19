package flux

// OperationSpec specifies an operation as part of a query.
type OperationSpec interface {
	// Kind returns the kind of the operation.
	Kind() OperationKind
}

// OperationKind denotes the kind of operations.
type OperationKind string
