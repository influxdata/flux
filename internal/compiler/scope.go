package compiler

type Scope struct {
	data *scope
}

func NewScope() Scope {
	return Scope{
		data: &scope{},
	}
}

func (s *Scope) Values() []Value {
	return s.data.values
}

type scope struct {
	names  map[string]int
	values []Value
}

// Declare will push a default Value onto the stack to
// be referenced by index later. The default value is null
// and is expected to be assigned by another location.
func (s Scope) Declare() int {
	return s.Push(Value{})
}

// Push will push an unnamed Value onto the stack to be
// referenced by index later.
//
// The Value is the default value that will be initialized
// on the stack. This can be used to create a buffered
// Value or to define the real value in the case of literals.
func (s Scope) Push(value Value) int {
	s.data.values = append(s.data.values, value)
	return len(s.data.values) - 1
}

// Define will define a named variable on the stack
// and push a default value similar to Push.
func (s Scope) Define(name string, register int) {
	if s.data.names == nil {
		s.data.names = make(map[string]int)
	}
	s.data.names[name] = register
}

// Set will set the default value for a register.
func (s Scope) Set(reg int, value Value) {
	s.data.values[reg] = value
}

// Get retrieves the index for a named variable created
// using Define.
func (s Scope) Get(name string) int {
	reg, ok := s.data.names[name]
	if !ok {
		panic("variable not found")
	}
	return reg
}
