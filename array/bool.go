package array

// Boolean represents an abstraction over a bool array.
type Boolean interface {
	Base
	Value(i int) bool
	BooleanSlice(start, stop int) Boolean
}
