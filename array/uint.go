package array

// UIntRef is a reference to a UInt.
type UIntRef interface {
	BaseRef
	Value(i int) uint64
	UIntSlice(start, stop int) UIntRef

	// Uint64Values will return the underlying slice for the UInt array. It is the size
	// of the array and null values will be present, but the data at null indexes will be invalid.
	Uint64Values() []uint64
}

// UInt represents an abstraction over an unsigned array.
type UInt interface {
	UIntRef

	// Free will release the memory for this array. After Free is called,
	// the array should no longer be used.
	Free()
}

// UIntBuilder represents an abstraction over building a uint array.
type UIntBuilder interface {
	BaseBuilder
	Append(v uint64)
	AppendValues(v []uint64, valid ...[]bool)

	// BuildUIntArray will construct the array.
	BuildUIntArray() UInt
}
