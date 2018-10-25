package array

import "github.com/influxdata/flux/semantic"

// BaseRef defines the base interface for working with an array interface.
type BaseRef interface {
	// Type returns the type of values that this array contains.
	Type() semantic.Type

	// IsNull will return true if the given index is null.
	IsNull(i int) bool

	// IsValid will return true if the given index is a valid value (not null).
	IsValid(i int) bool

	// Len will return the length of this array.
	Len() int

	// NullN will return the number of null values in this array.
	NullN() int

	// Slice will return a slice of the array.
	Slice(start, stop int) BaseRef

	// Copy will retain a reference to the array.
	Copy() Base
}

// Base defines the base interface for working with any array type.
// All array types share this common interface.
type Base interface {
	BaseRef

	// Free will release the memory for this array. After Free is called,
	// the array should no longer be used.
	Free()
}

// BaseBuilder defines the base interface for building an array of a given array type.
// All builder types share this common interface.
type BaseBuilder interface {
	// Type returns the type of values that this builder accepts.
	Type() semantic.Type

	// Len returns the currently allocated length for the array builder.
	Len() int

	// Cap returns the current capacity of the underlying array.
	Cap() int

	// Reserve ensures there is enough space for appending n elements by checking
	// the capacity and calling resize if necessary.
	Reserve(n int)

	// AppendNull will append a null value to the array.
	AppendNull()

	// BuildArray will construct the array.
	BuildArray() Base

	// Free will release the memory used for this builder.
	Free()
}
