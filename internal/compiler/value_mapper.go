package compiler

type valueMapper interface {
	Map(in Value, scope []Value)
}

type basicValueMapper int

func (b basicValueMapper) Map(in Value, scope []Value) {
	scope[b] = in
}
