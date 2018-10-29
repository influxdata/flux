package semantic

func Annotate(node Node) *Annotator {
	annotator := &Annotator{
		annotations: make(map[Node]annotation),
	}
	Walk(annotator, node)
	return annotator
}

type Annotator struct {
	f           fresher
	annotations map[Node]annotation
}

type annotation struct {
	Var  Tvar
	Type PolyType
	Err  error
}

func (v *Annotator) Visit(node Node) Visitor {
	switch n := node.(type) {
	case *FunctionBlock,
		*FunctionParameter,
		Expression:
		v.annotations[n] = annotation{
			Var: v.f.Fresh(),
		}
	}
	return v
}
func (v *Annotator) Done(node Node) {}
