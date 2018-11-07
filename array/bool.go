package array

// BooleanRef is a reference to a Boolean.
type BooleanRef interface {
	BaseRef
	Value(i int) bool
	BooleanSlice(start, stop int) BooleanRef
}

// Boolean represents an abstraction over a bool array.
type Boolean interface {
	BooleanRef

	// Free will release the memory for this array. After Free is called,
	// the array should no longer be used.
	Free()
}

// BooleanBuilder represents an abstraction over building a bool array.
type BooleanBuilder interface {
	BaseBuilder
	Append(v bool)
	AppendValues(v []bool, valid ...[]bool)

	// BuildBooleanArray will construct the array.
	BuildBooleanArray() Boolean
}
