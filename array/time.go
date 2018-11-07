package array

import "github.com/influxdata/flux/values"

// TimeRef is a reference to a Time.
type TimeRef interface {
	BaseRef
	Value(i int) values.Time
	TimeSlice(start, stop int) TimeRef
}

// Time represents an abstraction over a time array.
type Time interface {
	TimeRef

	// Free will release the memory for this array. After Free is called,
	// the array should no longer be used.
	Free()
}

// TimeBuilder represents an abstraction over building a time array.
type TimeBuilder interface {
	BaseBuilder
	Append(v values.Time)
	AppendValues(v []values.Time, valid ...[]bool)

	// BuildTimeArray will construct the array.
	BuildTimeArray() Time
}
