package array

// FloatRef is a reference to a Float.
type FloatRef interface {
	BaseRef
	Value(i int) float64
	FloatSlice(start, stop int) FloatRef

	// Float64Values will return the underlying slice for the Float array. It is the size
	// of the array and null values will be present, but the data at null indexes will be invalid.
	Float64Values() []float64
}

// Float represents an abstraction over a float array.
type Float interface {
	FloatRef

	// Free will release the memory for this array. After Free is called,
	// the array should no longer be used.
	Free()
}

// FloatBuilder represents an abstraction over building a float array.
type FloatBuilder interface {
	BaseBuilder
	Append(v float64)
	AppendValues(v []float64, valid ...[]bool)

	// BuildFloatArray will construct the array.
	BuildFloatArray() Float
}
