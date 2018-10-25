package array

// StringRef is a reference to a String.
type StringRef interface {
	BaseRef
	Value(i int) string
	StringSlice(start, stop int) StringRef
}

// String represents an abstraction over a string array.
type String interface {
	StringRef

	// Free will release the memory for this array. After Free is called,
	// the array should no longer be used.
	Free()
}

// StringBuilder represents an abstraction over building a string array.
type StringBuilder interface {
	BaseBuilder
	Append(v string)
	AppendValues(v []string, valid ...[]bool)

	// BuildStringArray will construct the array.
	BuildStringArray() String
}
