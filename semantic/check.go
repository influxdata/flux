package semantic

import "fmt"

func runChecks(pkg *Package) error {
	// check for options declared below package block
	if err := checkOptionAssignments(pkg); err != nil {
		return err
	}
	// check for dependencies among options
	if err := checkOptionDependencies(pkg); err != nil {
		return err
	}
	// check for variable reassignments
	if err := checkVarAssignments(pkg); err != nil {
		return err
	}
	return nil
}

func checkOptionAssignments(pkg *Package) error {
	var stmt *OptionStatement
	optionFn := func(opt *OptionStatement) {
		stmt = opt
	}
	visitor := optionAssignmentVisitor{
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

// This visitor finds options declared below the package block.
// Any such option is passed to optionFn.
type optionAssignmentVisitor struct {
	optionFn     func(*OptionStatement)
	packageBlock bool
}

func (v optionAssignmentVisitor) Visit(node Node) Visitor {
	n, ok := node.(*OptionStatement)
	if ok && !v.packageBlock {
		v.optionFn(n)
		return nil
	}
	return v
}

func (v optionAssignmentVisitor) Nest() NestingVisitor {
	v.packageBlock = false
	return v
}

func (v optionAssignmentVisitor) Done(node Node) {}

func checkVarAssignments(pkg *Package) error {
	var node *NativeVariableAssignment
	visitor := varAssignmentVisitor{
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

// This visitor finds variable reassignments within a package.
// Any such reassignment is passed to varFn.
type varAssignmentVisitor struct {
	names  map[string]bool
	varFn  func(*NativeVariableAssignment)
	option bool
}

func (v varAssignmentVisitor) Visit(node Node) Visitor {
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

func (v varAssignmentVisitor) Nest() NestingVisitor {
	v.names = make(map[string]bool)
	return v
}

func (v varAssignmentVisitor) Done(node Node) {}

func checkOptionDependencies(pkg *Package) error {
	var options optionStmtVisitor
	Walk(&options, pkg)

	var ref *IdentifierExpression

	visitor := optionDependencyVisitor{
		option: make(map[string]bool, len(options)),
		shadow: make(map[string]bool, len(options)),
		ref: func(n *IdentifierExpression) {
			ref = n
		},
	}

	for _, dec := range options {
		// option name
		name := dec.Identifier.Name
		visitor.option[name] = true

		// check for dependencies among options
		Walk(NewScopedVisitor(&visitor), dec.Init)

		if ref != nil {
			return fmt.Errorf("option dependency: option %q depends on option %q defined in the same package at %v", name, ref.Name, ref.Location())
		}
	}
	return nil
}

type optionStmtVisitor []*NativeVariableAssignment

func (v *optionStmtVisitor) Visit(node Node) Visitor {
	if stmt, ok := node.(Statement); ok {
		if opt, ok := stmt.(*OptionStatement); ok {
			if n, ok := opt.Assignment.(*NativeVariableAssignment); ok {
				*v = append(*v, n)
			}
		}
		return nil
	}
	return v
}

func (v *optionStmtVisitor) Done(node Node) {}

// This visitor checks for dependencies among options in the same package block.
// Any reference to another option made within an option statement is passed to ref.
type optionDependencyVisitor struct {
	option map[string]bool
	shadow map[string]bool
	ref    func(*IdentifierExpression)
}

func (v optionDependencyVisitor) Visit(node Node) Visitor {
	switch n := node.(type) {
	case *NativeVariableAssignment:
		// var declarations shadow options
		v.shadow[n.Identifier.Name] = true
	case *FunctionParameter:
		// function params shadow options
		v.shadow[n.Key.Name] = true
	case *IdentifierExpression:
		if v.option[n.Name] && !v.shadow[n.Name] {
			v.ref(n)
			return nil
		}
	}
	return v
}

func (v optionDependencyVisitor) Nest() NestingVisitor {
	shadows := make(map[string]bool, len(v.shadow))
	for k, v := range v.shadow {
		shadows[k] = v
	}
	v.shadow = shadows
	return v
}

func (v optionDependencyVisitor) Done(node Node) {}
