package array

// String represents an abstraction over a string array.
type String interface {
	Base
	Value(i int) string
	StringSlice(start, stop int) String
}
