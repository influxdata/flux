package semantic

import "fmt"

func runChecks(pkg *Package) error {
	// check for options declared below package block
	if err := checkOptionDecs(pkg); err != nil {
		return err
	}
	// check for variable reassignments
	if err := checkVarDecs(pkg); err != nil {
		return err
	}
	return nil
}

func checkOptionDecs(pkg *Package) error {
	var stmt *OptionStatement
	optionFn := func(opt *OptionStatement) {
		stmt = opt
	}
	visitor := optionDecVisitor{
		optionFn:     optionFn,
		packageBlock: true,
	}
	Walk(NewScopedVisitor(visitor), pkg)
	if stmt != nil {
		name, err := optionName(stmt)
		if err != nil {
			return err
		}
		return fmt.Errorf("option %q declared below package block at %v", name, stmt.Location())
	}
	return nil
}

func optionName(opt *OptionStatement) (string, error) {
	switch n := opt.Assignment.(type) {
	case *NativeVariableAssignment:
		return n.Identifier.Name, nil
	case *MemberAssignment:
		obj := n.Member.Object
		id, ok := obj.(*IdentifierExpression)
		if !ok {
			return "", fmt.Errorf("unsupported option qualifier %T", obj)
		}
		return id.Name + "." + n.Member.Property, nil
	default:
		return "", fmt.Errorf("unsupported assignment %T", n)
	}
}

type optionDecVisitor struct {
	optionFn     func(*OptionStatement)
	packageBlock bool
}

func (v optionDecVisitor) Visit(node Node) Visitor {
	n, ok := node.(*OptionStatement)
	if ok && !v.packageBlock {
		v.optionFn(n)
		return nil
	}
	return v
}

func (v optionDecVisitor) Nest() NestingVisitor {
	v.packageBlock = false
	return v
}

func (v optionDecVisitor) Done(node Node) {}

func checkVarDecs(pkg *Package) error {
	var node *NativeVariableAssignment
	visitor := varDecVisitor{
		names: make(map[string]bool, 8),
		varFn: func(n *NativeVariableAssignment) {
			node = n
		},
	}
	Walk(NewScopedVisitor(visitor), pkg)
	if node != nil {
		name := node.Identifier.Name
		return fmt.Errorf("var %q redeclared at %v", name, node.Location())
	}
	return nil
}

type varDecVisitor struct {
	names  map[string]bool
	varFn  func(*NativeVariableAssignment)
	option bool
}

func (v varDecVisitor) Visit(node Node) Visitor {
	switch n := node.(type) {
	case *OptionStatement:
		v.option = true
	case *NativeVariableAssignment:
		name := n.Identifier.Name
		if v.option {
			v.option = false
		} else if v.names[name] {
			v.varFn(n)
			return nil
		}
		v.names[name] = true
	case *FunctionParameter:
		v.names[n.Key.Name] = true
	}
	return v
}

func (v varDecVisitor) Nest() NestingVisitor {
	v.names = make(map[string]bool)
	return v
}

func (v varDecVisitor) Done(node Node) {}
