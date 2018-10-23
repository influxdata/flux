package array

// Base defines the base interface for working with any array type.
// All array types share this common interface.
type Base interface {
	IsNull(i int) bool
	IsValid(i int) bool
	Len() int
	NullN() int
	Slice(start, stop int) Base
}
