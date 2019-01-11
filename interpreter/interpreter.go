package interpreter

import (
	"fmt"
	"regexp"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

type Interpreter struct {
	types       map[semantic.Node]semantic.Type
	polyTypes   map[semantic.Node]semantic.PolyType
	sideEffects []values.Value
	pkg         string
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		types:     make(map[semantic.Node]semantic.Type),
		polyTypes: make(map[semantic.Node]semantic.PolyType),
	}
}

// Eval evaluates the expressions composing a Flux package and returns any side effects that occured.
func (itrp *Interpreter) Eval(node semantic.Node, scope Scope, importer Importer) ([]values.Value, error) {
	var n = node
	for s := scope; s != nil; s = s.Pop() {
		extern := &semantic.Extern{
			Block: &semantic.ExternBlock{
				Node: n,
			},
		}
		s.LocalRange(func(k string, v values.Value) {
			extern.Assignments = append(extern.Assignments, &semantic.ExternalVariableAssignment{
				Identifier: &semantic.Identifier{Name: k},
				ExternType: v.PolyType(),
			})
		})
		n = extern
	}

	sol, err := semantic.InferTypes(n, importer)
	if err != nil {
		return nil, err
	}

	semantic.Walk(semantic.CreateVisitor(func(node semantic.Node) {
		if typ, err := sol.TypeOf(node); err == nil {
			itrp.types[node] = typ
		}
		if polyType, err := sol.PolyTypeOf(node); err == nil {
			itrp.polyTypes[node] = polyType
		}
	}), node)

	if err := itrp.doRoot(node, scope, importer); err != nil {
		return nil, err
	}
	return itrp.sideEffects, nil
}

func (itrp *Interpreter) doRoot(node semantic.Node, scope Scope, importer Importer) error {
	switch n := node.(type) {
	case *semantic.Package:
		return itrp.doPackage(n, scope, importer)
	case *semantic.File:
		return itrp.doFile(n, scope, importer)
	case *semantic.Extern:
		return itrp.doExtern(n, scope, importer)
	default:
		return fmt.Errorf("unsupported root node %T", node)
	}
}

func (itrp *Interpreter) doExtern(extern *semantic.Extern, scope Scope, importer Importer) error {
	// We do not care about the type declarations, they were only important for type inference.
	return itrp.doRoot(extern.Block.Node, scope, importer)
}

func (itrp *Interpreter) doPackage(pkg *semantic.Package, scope Scope, importer Importer) error {
	for _, file := range pkg.Files {
		if err := itrp.doFile(file, scope, importer); err != nil {
			return err
		}
	}
	return nil
}

func (itrp *Interpreter) doFile(file *semantic.File, scope Scope, importer Importer) error {
	if err := itrp.doPackageClause(file.Package); err != nil {
		return err
	}
	for _, i := range file.Imports {
		if err := itrp.doImport(i, scope, importer); err != nil {
			return err
		}
	}
	for _, stmt := range file.Body {
		val, err := itrp.doStatement(stmt, scope)
		if err != nil {
			return err
		}
		if _, ok := stmt.(*semantic.ExpressionStatement); ok {
			// Only in the main package are all unassigned package
			// level expressions coerced into producing side effects.
			if itrp.pkg == semantic.PackageMain {
				itrp.sideEffects = append(itrp.sideEffects, val)
			}
		}
	}
	return nil
}

func (itrp *Interpreter) doPackageClause(pkg *semantic.PackageClause) error {
	name := semantic.PackageMain
	if pkg != nil {
		name = pkg.Name.Name
	}
	if itrp.pkg == "" {
		itrp.pkg = name
	}
	if itrp.pkg != name {
		return fmt.Errorf("package name mismatch %q != %q", itrp.pkg, name)
	}
	return nil
}

func (itrp *Interpreter) doImport(dec *semantic.ImportDeclaration, scope Scope, importer Importer) error {
	path := dec.Path.Value
	pkg, ok := importer.ImportPackageObject(path)
	if !ok {
		return fmt.Errorf("invalid import path %s", path)
	}
	name := pkg.Name()
	if dec.As != nil {
		name = dec.As.Name
	}
	scope.Set(name, pkg)
	// Packages can import side effects
	itrp.sideEffects = append(itrp.sideEffects, pkg.SideEffects()...)
	return nil
}

// doStatement returns the resolved value of a top-level statement
func (itrp *Interpreter) doStatement(stmt semantic.Statement, scope Scope) (values.Value, error) {
	scope.SetReturn(values.InvalidValue)
	switch s := stmt.(type) {
	case *semantic.OptionStatement:
		return itrp.doOptionStatement(s, scope)
	case *semantic.NativeVariableAssignment:
		return itrp.doVariableAssignment(s, scope)
	case *semantic.MemberAssignment:
		return itrp.doMemberAssignment(s, scope)
	case *semantic.ExpressionStatement:
		v, err := itrp.doExpression(s.Expression, scope)
		if err != nil {
			return nil, err
		}
		scope.SetReturn(v)
		return v, nil
	case *semantic.ReturnStatement:
		v, err := itrp.doExpression(s.Argument, scope)
		if err != nil {
			return nil, err
		}
		scope.SetReturn(v)
	default:
		return nil, fmt.Errorf("unsupported statement type %T", stmt)
	}
	return nil, nil
}

func (itrp *Interpreter) doOptionStatement(s *semantic.OptionStatement, scope Scope) (values.Value, error) {
	return itrp.doAssignment(s.Assignment, scope)
}

func (itrp *Interpreter) doVariableAssignment(dec *semantic.NativeVariableAssignment, scope Scope) (values.Value, error) {
	value, err := itrp.doExpression(dec.Init, scope)
	if err != nil {
		return nil, err
	}
	scope.Set(dec.Identifier.Name, value)
	return value, nil
}

func (itrp *Interpreter) doMemberAssignment(a *semantic.MemberAssignment, scope Scope) (values.Value, error) {
	object, err := itrp.doExpression(a.Member.Object, scope)
	if err != nil {
		return nil, err
	}
	init, err := itrp.doExpression(a.Init, scope)
	if err != nil {
		return nil, err
	}
	object.Object().Set(a.Member.Property, init)
	return object, nil
}

func (itrp *Interpreter) doAssignment(a semantic.Assignment, scope Scope) (values.Value, error) {
	switch a := a.(type) {
	case *semantic.NativeVariableAssignment:
		return itrp.doVariableAssignment(a, scope)
	case *semantic.MemberAssignment:
		return itrp.doMemberAssignment(a, scope)
	default:
		return nil, fmt.Errorf("unsupported assignment %T", a)
	}
}

func (itrp *Interpreter) doExpression(expr semantic.Expression, scope Scope) (values.Value, error) {
	switch e := expr.(type) {
	case semantic.Literal:
		return itrp.doLiteral(e)
	case *semantic.ArrayExpression:
		return itrp.doArray(e, scope)
	case *semantic.IdentifierExpression:
		value, ok := scope.Lookup(e.Name)
		if !ok {
			return nil, fmt.Errorf("undefined identifier %q", e.Name)
		}
		return value, nil
	case *semantic.CallExpression:
		v, err := itrp.doCall(e, scope)
		if err != nil {
			// Determine function name
			return nil, errors.Wrapf(err, "error calling function %q", functionName(e))
		}
		return v, nil
	case *semantic.MemberExpression:
		obj, err := itrp.doExpression(e.Object, scope)
		if err != nil {
			return nil, err
		}
		if typ := obj.Type().Nature(); typ != semantic.Object {
			return nil, fmt.Errorf("cannot access property %q on value of type %s", e.Property, typ)
		}
		v, ok := obj.Object().Get(e.Property)
		if !ok {
			return nil, fmt.Errorf("object has no property %q", e.Property)
		}
		return v, nil
	case *semantic.IndexExpression:
		arr, err := itrp.doExpression(e.Array, scope)
		if err != nil {
			return nil, err
		}
		idx, err := itrp.doExpression(e.Index, scope)
		if err != nil {
			return nil, err
		}
		return arr.Array().Get(int(idx.Int())), nil
	case *semantic.ObjectExpression:
		return itrp.doObject(e, scope)
	case *semantic.UnaryExpression:
		v, err := itrp.doExpression(e.Argument, scope)
		if err != nil {
			return nil, err
		}
		switch e.Operator {
		case ast.NotOperator:
			if v.Type() != semantic.Bool {
				return nil, fmt.Errorf("operand to unary expression is not a boolean value, got %v", v.Type())
			}
			return values.NewBool(!v.Bool()), nil
		case ast.SubtractionOperator:
			switch t := v.Type(); t {
			case semantic.Int:
				return values.NewInt(-v.Int()), nil
			case semantic.Float:
				return values.NewFloat(-v.Float()), nil
			case semantic.Duration:
				return values.NewDuration(-v.Duration()), nil
			default:
				return nil, fmt.Errorf("operand to unary expression is not a number value, got %v", v.Type())
			}
		default:
			return nil, fmt.Errorf("unsupported operator %q to unary expression", e.Operator)
		}

	case *semantic.BinaryExpression:
		l, err := itrp.doExpression(e.Left, scope)
		if err != nil {
			return nil, err
		}

		r, err := itrp.doExpression(e.Right, scope)
		if err != nil {
			return nil, err
		}

		bf, err := values.LookupBinaryFunction(values.BinaryFuncSignature{
			Operator: e.Operator,
			Left:     l.Type(),
			Right:    r.Type(),
		})
		if err != nil {
			return nil, err
		}
		return bf(l, r), nil
	case *semantic.LogicalExpression:
		l, err := itrp.doExpression(e.Left, scope)
		if err != nil {
			return nil, err
		}
		if l.Type() != semantic.Bool {
			return nil, fmt.Errorf("left operand to logcial expression is not a boolean value, got %v", l.Type())
		}
		left := l.Bool()

		if e.Operator == ast.AndOperator && !left {
			// Early return
			return values.NewBool(false), nil
		} else if e.Operator == ast.OrOperator && left {
			// Early return
			return values.NewBool(true), nil
		}

		r, err := itrp.doExpression(e.Right, scope)
		if err != nil {
			return nil, err
		}
		if r.Type() != semantic.Bool {
			return nil, errors.New("right operand to logcial expression is not a boolean value")
		}
		right := r.Bool()

		switch e.Operator {
		case ast.AndOperator:
			return values.NewBool(left && right), nil
		case ast.OrOperator:
			return values.NewBool(left || right), nil
		default:
			return nil, fmt.Errorf("invalid logical operator %v", e.Operator)
		}
	case *semantic.FunctionExpression:
		// Capture type information
		types := make(map[semantic.Node]semantic.Type)
		polyTypes := make(map[semantic.Node]semantic.PolyType)
		semantic.Walk(semantic.CreateVisitor(func(node semantic.Node) {
			if typ, ok := itrp.types[node]; ok {
				types[node] = typ
			}
			if polyType, ok := itrp.polyTypes[node]; ok {
				polyTypes[node] = polyType
			}
		}), e)
		return function{
			e:         e,
			scope:     scope,
			types:     types,
			polyTypes: polyTypes,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported expression %T", expr)
	}
}

func (itrp *Interpreter) doArray(a *semantic.ArrayExpression, scope Scope) (values.Value, error) {
	elements := make([]values.Value, len(a.Elements))
	arrayType, ok := itrp.types[a]
	if !ok {
		return nil, fmt.Errorf("expecting array type")
	}
	elementType := arrayType.ElementType()
	for i, el := range a.Elements {
		v, err := itrp.doExpression(el, scope)
		if err != nil {
			return nil, err
		}
		elements[i] = v
	}
	return values.NewArrayWithBacking(elementType, elements), nil
}

func (itrp *Interpreter) doObject(m *semantic.ObjectExpression, scope Scope) (values.Value, error) {
	obj := values.NewObject()
	for _, p := range m.Properties {
		v, err := itrp.doExpression(p.Value, scope)
		if err != nil {
			return nil, err
		}
		if _, ok := obj.Get(p.Key.Key()); ok {
			return nil, fmt.Errorf("duplicate key in object: %q", p.Key.Key())
		}
		obj.Set(p.Key.Key(), v)
	}
	return obj, nil
}

func (itrp *Interpreter) doLiteral(lit semantic.Literal) (values.Value, error) {
	switch l := lit.(type) {
	case *semantic.DateTimeLiteral:
		return values.NewTime(values.Time(l.Value.UnixNano())), nil
	case *semantic.DurationLiteral:
		return values.NewDuration(values.Duration(l.Value)), nil
	case *semantic.FloatLiteral:
		return values.NewFloat(l.Value), nil
	case *semantic.IntegerLiteral:
		return values.NewInt(l.Value), nil
	case *semantic.UnsignedIntegerLiteral:
		return values.NewUInt(l.Value), nil
	case *semantic.StringLiteral:
		return values.NewString(l.Value), nil
	case *semantic.RegexpLiteral:
		return values.NewRegexp(l.Value), nil
	case *semantic.BooleanLiteral:
		return values.NewBool(l.Value), nil
	default:
		return nil, fmt.Errorf("unknown literal type %T", lit)
	}
}

func functionName(call *semantic.CallExpression) string {
	switch callee := call.Callee.(type) {
	case *semantic.IdentifierExpression:
		return callee.Name
	case *semantic.MemberExpression:
		return callee.Property
	default:
		return "<anonymous function>"
	}
}

func DoFunctionCall(f func(args Arguments) (values.Value, error), argsObj values.Object) (values.Value, error) {
	args := NewArguments(argsObj)
	v, err := f(args)
	if err != nil {
		return nil, err
	}
	if unused := args.listUnused(); len(unused) > 0 {
		return nil, fmt.Errorf("unused arguments %v", unused)
	}
	return v, nil
}

type functionType interface {
	Signature() semantic.FunctionPolySignature
}

func (itrp *Interpreter) doCall(call *semantic.CallExpression, scope Scope) (values.Value, error) {
	callee, err := itrp.doExpression(call.Callee, scope)
	if err != nil {
		return nil, err
	}
	ft := callee.PolyType()
	if ft.Nature() != semantic.Function {
		return nil, fmt.Errorf("cannot call function, value is of type %v", callee.Type())
	}
	f := callee.Function()
	sig := ft.(functionType).Signature()
	argObj, err := itrp.doArguments(call.Arguments, scope, sig.PipeArgument, call.Pipe)
	if err != nil {
		return nil, err
	}

	// Check if the function is an interpFunction and rebind it.
	if af, ok := f.(function); ok {
		semantic.Walk(semantic.CreateVisitor(func(node semantic.Node) {
			if typ, ok := af.TypeOf(node); ok {
				itrp.types[node] = typ
			}
			if polyType, ok := af.PolyTypeOf(node); ok {
				itrp.polyTypes[node] = polyType
			}
		}), af.e)
		af.itrp = itrp
		f = af
	}

	// Call the function
	value, err := f.Call(argObj)
	if err != nil {
		return nil, err
	}

	if f.HasSideEffect() {
		itrp.sideEffects = append(itrp.sideEffects, value)
	}

	return value, nil
}

func (itrp *Interpreter) doArguments(args *semantic.ObjectExpression, scope Scope, pipeArgument string, pipe semantic.Expression) (values.Object, error) {
	obj := values.NewObject()
	if pipe == nil && (args == nil || len(args.Properties) == 0) {
		return obj, nil
	}
	for _, p := range args.Properties {
		value, err := itrp.doExpression(p.Value, scope)
		if err != nil {
			return nil, err
		}
		if _, ok := obj.Get(p.Key.Key()); ok {
			return nil, fmt.Errorf("duplicate keyword parameter specified: %q", p.Key.Key())
		}

		obj.Set(p.Key.Key(), value)
	}
	if pipe != nil && pipeArgument == "" {
		return nil, errors.New("pipe parameter value provided to function with no pipe parameter defined")
	}
	if pipe != nil {
		value, err := itrp.doExpression(pipe, scope)
		if err != nil {
			return nil, err
		}
		obj.Set(pipeArgument, value)
	}
	return obj, nil
}

// Value represents any value that can be the result of evaluating any expression.
type Value interface {
	// Type reports the type of value
	Type() semantic.Type
	// Value returns the actual value represented.
	Value() interface{}
	// Property returns a new value which is a property of this value.
	Property(name string) (values.Value, error)
}

type function struct {
	e     *semantic.FunctionExpression
	scope Scope

	types     map[semantic.Node]semantic.Type
	polyTypes map[semantic.Node]semantic.PolyType

	itrp *Interpreter
}

func (f function) TypeOf(node semantic.Node) (semantic.Type, bool) {
	t, ok := f.types[node]
	return t, ok
}
func (f function) PolyTypeOf(node semantic.Node) (semantic.PolyType, bool) {
	p, ok := f.polyTypes[node]
	return p, ok
}
func (f function) Type() semantic.Type {
	if t, ok := f.TypeOf(f.e); ok {
		return t
	}
	return semantic.Invalid
}
func (f function) PolyType() semantic.PolyType {
	if t, ok := f.PolyTypeOf(f.e); ok {
		return t
	}
	return semantic.Invalid
}

func (f function) IsNull() bool {
	return false
}
func (f function) Str() string {
	panic(values.UnexpectedKind(semantic.Function, semantic.String))
}
func (f function) Int() int64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.Int))
}
func (f function) UInt() uint64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.UInt))
}
func (f function) Float() float64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.Float))
}
func (f function) Bool() bool {
	panic(values.UnexpectedKind(semantic.Function, semantic.Bool))
}
func (f function) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Function, semantic.Time))
}
func (f function) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Function, semantic.Duration))
}
func (f function) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Function, semantic.Regexp))
}
func (f function) Array() values.Array {
	panic(values.UnexpectedKind(semantic.Function, semantic.Array))
}
func (f function) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Function, semantic.Object))
}
func (f function) Function() values.Function {
	return f
}
func (f function) Equal(rhs values.Value) bool {
	if f.Type() != rhs.Type() {
		return false
	}
	v, ok := rhs.(function)
	return ok && f.e == v.e && f.scope == v.scope
}
func (f function) HasSideEffect() bool {
	// Function definitions do not produce side effects.
	// Only a function call expression can produce side effects.
	return false
}

func (f function) Call(argsObj values.Object) (values.Value, error) {
	args := newArguments(argsObj)
	v, err := f.doCall(args)
	if err != nil {
		return nil, err
	}
	if unused := args.listUnused(); len(unused) > 0 {
		return nil, fmt.Errorf("unused arguments %s", unused)
	}
	return v, nil
}
func (f function) doCall(args Arguments) (values.Value, error) {
	blockScope := f.scope.Nest(nil)
	if f.e.Block.Parameters != nil {
	PARAMETERS:
		for _, p := range f.e.Block.Parameters.List {
			if f.e.Defaults != nil {
				for _, d := range f.e.Defaults.Properties {
					if d.Key.Key() == p.Key.Name {
						v, ok := args.Get(p.Key.Name)
						if !ok {
							// Use default value
							var err error
							// evaluate default expressions outside the block scope
							v, err = f.itrp.doExpression(d.Value, f.scope)
							if err != nil {
								return nil, err
							}
						}
						blockScope.Set(p.Key.Name, v)
						continue PARAMETERS
					}
				}
			}
			v, err := args.GetRequired(p.Key.Name)
			if err != nil {
				return nil, err
			}
			blockScope.Set(p.Key.Name, v)
		}
	}
	switch n := f.e.Block.Body.(type) {
	case semantic.Expression:
		return f.itrp.doExpression(n, blockScope)
	case *semantic.Block:
		nested := blockScope.Nest(nil)
		for i, stmt := range n.Body {
			_, err := f.itrp.doStatement(stmt, nested)
			if err != nil {
				return nil, err
			}
			// Validate a return statement is the last statement
			if _, ok := stmt.(*semantic.ReturnStatement); ok {
				if i != len(n.Body)-1 {
					return nil, errors.New("return statement is not the last statement in the block")
				}
			}
		}
		// TODO(jlapacik): Return values should not be associated with variable scope.
		// This check should be performed during type inference, not here.
		v := nested.Return()
		if v.PolyType().Nature() == semantic.Invalid {
			return nil, errors.New("function has no return value")
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported function body type %T", f.e.Block.Body)
	}
}

// Resolver represents a value that can resolve itself
type Resolver interface {
	Resolve() (semantic.Node, error)
}

func ResolveFunction(f values.Function) (*semantic.FunctionExpression, error) {
	resolver, ok := f.(Resolver)
	if !ok {
		return nil, errors.New("function is not resolvable")
	}
	resolved, err := resolver.Resolve()
	if err != nil {
		return nil, err
	}
	fn, ok := resolved.(*semantic.FunctionExpression)
	if !ok {
		return nil, errors.New("resolved function is not a function")
	}
	return fn, nil
}

// Resolve rewrites the function resolving any identifiers not listed in the function params.
func (f function) Resolve() (semantic.Node, error) {
	n := f.e.Copy()
	node, err := f.resolveIdentifiers(n)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (f function) resolveIdentifiers(n semantic.Node) (semantic.Node, error) {
	switch n := n.(type) {
	case *semantic.IdentifierExpression:
		if f.e.Block.Parameters != nil {
			for _, p := range f.e.Block.Parameters.List {
				if n.Name == p.Key.Name {
					// Identifier is a parameter do not resolve
					return n, nil
				}
			}
		}
		v, ok := f.scope.Lookup(n.Name)
		if !ok {
			return nil, fmt.Errorf("name %q does not exist in scope", n.Name)
		}
		return resolveValue(v)
	case *semantic.Block:
		for i, s := range n.Body {
			node, err := f.resolveIdentifiers(s)
			if err != nil {
				return nil, err
			}
			n.Body[i] = node.(semantic.Statement)
		}
	case *semantic.OptionStatement:
		node, err := f.resolveIdentifiers(n.Assignment)
		if err != nil {
			return nil, err
		}
		n.Assignment = node.(semantic.Assignment)
	case *semantic.ExpressionStatement:
		node, err := f.resolveIdentifiers(n.Expression)
		if err != nil {
			return nil, err
		}
		n.Expression = node.(semantic.Expression)
	case *semantic.ReturnStatement:
		node, err := f.resolveIdentifiers(n.Argument)
		if err != nil {
			return nil, err
		}
		n.Argument = node.(semantic.Expression)
	case *semantic.NativeVariableAssignment:
		node, err := f.resolveIdentifiers(n.Init)
		if err != nil {
			return nil, err
		}
		n.Init = node.(semantic.Expression)
	case *semantic.CallExpression:
		node, err := f.resolveIdentifiers(n.Arguments)
		if err != nil {
			return nil, err
		}
		n.Arguments = node.(*semantic.ObjectExpression)
	case *semantic.FunctionExpression:
		node, err := f.resolveIdentifiers(n.Block.Body)
		if err != nil {
			return nil, err
		}
		n.Block.Body = node
	case *semantic.BinaryExpression:
		node, err := f.resolveIdentifiers(n.Left)
		if err != nil {
			return nil, err
		}
		n.Left = node.(semantic.Expression)

		node, err = f.resolveIdentifiers(n.Right)
		if err != nil {
			return nil, err
		}
		n.Right = node.(semantic.Expression)
	case *semantic.UnaryExpression:
		node, err := f.resolveIdentifiers(n.Argument)
		if err != nil {
			return nil, err
		}
		n.Argument = node.(semantic.Expression)
	case *semantic.LogicalExpression:
		node, err := f.resolveIdentifiers(n.Left)
		if err != nil {
			return nil, err
		}
		n.Left = node.(semantic.Expression)
		node, err = f.resolveIdentifiers(n.Right)
		if err != nil {
			return nil, err
		}
		n.Right = node.(semantic.Expression)
	case *semantic.ArrayExpression:
		for i, el := range n.Elements {
			node, err := f.resolveIdentifiers(el)
			if err != nil {
				return nil, err
			}
			n.Elements[i] = node.(semantic.Expression)
		}
	case *semantic.ObjectExpression:
		for i, p := range n.Properties {
			node, err := f.resolveIdentifiers(p)
			if err != nil {
				return nil, err
			}
			n.Properties[i] = node.(*semantic.Property)
		}
	case *semantic.ConditionalExpression:
		node, err := f.resolveIdentifiers(n.Test)
		if err != nil {
			return nil, err
		}
		n.Test = node.(semantic.Expression)

		node, err = f.resolveIdentifiers(n.Alternate)
		if err != nil {
			return nil, err
		}
		n.Alternate = node.(semantic.Expression)

		node, err = f.resolveIdentifiers(n.Consequent)
		if err != nil {
			return nil, err
		}
		n.Consequent = node.(semantic.Expression)
	case *semantic.Property:
		node, err := f.resolveIdentifiers(n.Value)
		if err != nil {
			return nil, err
		}
		n.Value = node.(semantic.Expression)
	}
	return n, nil
}

func resolveValue(v values.Value) (semantic.Node, error) {
	switch k := v.Type().Nature(); k {
	case semantic.String:
		return &semantic.StringLiteral{
			Value: v.Str(),
		}, nil
	case semantic.Int:
		return &semantic.IntegerLiteral{
			Value: v.Int(),
		}, nil
	case semantic.UInt:
		return &semantic.UnsignedIntegerLiteral{
			Value: v.UInt(),
		}, nil
	case semantic.Float:
		return &semantic.FloatLiteral{
			Value: v.Float(),
		}, nil
	case semantic.Bool:
		return &semantic.BooleanLiteral{
			Value: v.Bool(),
		}, nil
	case semantic.Time:
		return &semantic.DateTimeLiteral{
			Value: v.Time().Time(),
		}, nil
	case semantic.Regexp:
		return &semantic.RegexpLiteral{
			Value: v.Regexp(),
		}, nil
	case semantic.Duration:
		return &semantic.DurationLiteral{
			Value: v.Duration().Duration(),
		}, nil
	case semantic.Function:
		resolver, ok := v.Function().(Resolver)
		if !ok {
			return nil, fmt.Errorf("function is not resolvable %T", v.Function())
		}
		return resolver.Resolve()
	case semantic.Array:
		arr := v.Array()
		node := new(semantic.ArrayExpression)
		node.Elements = make([]semantic.Expression, arr.Len())
		var err error
		arr.Range(func(i int, el values.Value) {
			if err != nil {
				return
			}
			var n semantic.Node
			n, err = resolveValue(el)
			if err != nil {
				return
			}
			node.Elements[i] = n.(semantic.Expression)
		})
		if err != nil {
			return nil, err
		}
		return node, nil
	case semantic.Object:
		obj := v.Object()
		node := new(semantic.ObjectExpression)
		node.Properties = make([]*semantic.Property, 0, obj.Len())
		var err error
		obj.Range(func(k string, v values.Value) {
			if err != nil {
				return
			}
			var n semantic.Node
			n, err = resolveValue(v)
			if err != nil {
				return
			}
			node.Properties = append(node.Properties, &semantic.Property{
				Key:   &semantic.Identifier{Name: k},
				Value: n.(semantic.Expression),
			})
		})
		if err != nil {
			return nil, err
		}
		return node, nil
	default:
		return nil, fmt.Errorf("cannot resove value of type %v", k)
	}
}

func ToStringArray(a values.Array) ([]string, error) {
	if a.Type().ElementType() != semantic.String {
		return nil, fmt.Errorf("cannot convert array of %v to an array of strings", a.Type().ElementType())
	}
	strs := make([]string, a.Len())
	a.Range(func(i int, v values.Value) {
		strs[i] = v.Str()
	})
	return strs, nil
}
func ToFloatArray(a values.Array) ([]float64, error) {
	if a.Type().ElementType() != semantic.Float {
		return nil, fmt.Errorf("cannot convert array of %v to an array of floats", a.Type().ElementType())
	}
	vs := make([]float64, a.Len())
	a.Range(func(i int, v values.Value) {
		vs[i] = v.Float()
	})
	return vs, nil
}

// Arguments provides access to the keyword arguments passed to a function.
// semantic.The Get{Type} methods return three values: the typed value of the arg,
// whether the argument was specified and any errors about the argument type.
// semantic.The GetRequired{Type} methods return only two values, the typed value of the arg and any errors, a missing argument is considered an error in this case.
type Arguments interface {
	GetAll() []string
	Get(name string) (values.Value, bool)
	GetRequired(name string) (values.Value, error)

	GetString(name string) (string, bool, error)
	GetInt(name string) (int64, bool, error)
	GetFloat(name string) (float64, bool, error)
	GetBool(name string) (bool, bool, error)
	GetFunction(name string) (values.Function, bool, error)
	GetArray(name string, t semantic.Nature) (values.Array, bool, error)
	GetObject(name string) (values.Object, bool, error)

	GetRequiredString(name string) (string, error)
	GetRequiredInt(name string) (int64, error)
	GetRequiredFloat(name string) (float64, error)
	GetRequiredBool(name string) (bool, error)
	GetRequiredFunction(name string) (values.Function, error)
	GetRequiredArray(name string, t semantic.Nature) (values.Array, error)
	GetRequiredObject(name string) (values.Object, error)

	// listUnused returns the list of provided arguments that were not used by the function.
	listUnused() []string
}

type arguments struct {
	obj  values.Object
	used map[string]bool
}

func newArguments(obj values.Object) *arguments {
	if obj == nil {
		return new(arguments)
	}
	return &arguments{
		obj:  obj,
		used: make(map[string]bool, obj.Len()),
	}
}
func NewArguments(obj values.Object) Arguments {
	return newArguments(obj)
}

func (a *arguments) GetAll() []string {
	args := make([]string, 0, a.obj.Len())
	a.obj.Range(func(name string, v values.Value) {
		args = append(args, name)
	})
	return args
}

func (a *arguments) Get(name string) (values.Value, bool) {
	a.used[name] = true
	v, ok := a.obj.Get(name)
	return v, ok
}

func (a *arguments) GetRequired(name string) (values.Value, error) {
	a.used[name] = true
	v, ok := a.obj.Get(name)
	if !ok {
		return nil, fmt.Errorf("missing required keyword argument %q", name)
	}
	return v, nil
}

func (a *arguments) GetString(name string) (string, bool, error) {
	v, ok, err := a.get(name, semantic.String, false)
	if err != nil || !ok {
		return "", ok, err
	}
	return v.Str(), ok, nil
}
func (a *arguments) GetRequiredString(name string) (string, error) {
	v, _, err := a.get(name, semantic.String, true)
	if err != nil {
		return "", err
	}
	return v.Str(), nil
}
func (a *arguments) GetInt(name string) (int64, bool, error) {
	v, ok, err := a.get(name, semantic.Int, false)
	if err != nil || !ok {
		return 0, ok, err
	}
	return v.Int(), ok, nil
}
func (a *arguments) GetRequiredInt(name string) (int64, error) {
	v, _, err := a.get(name, semantic.Int, true)
	if err != nil {
		return 0, err
	}
	return v.Int(), nil
}
func (a *arguments) GetFloat(name string) (float64, bool, error) {
	v, ok, err := a.get(name, semantic.Float, false)
	if err != nil || !ok {
		return 0, ok, err
	}
	return v.Float(), ok, nil
}
func (a *arguments) GetRequiredFloat(name string) (float64, error) {
	v, _, err := a.get(name, semantic.Float, true)
	if err != nil {
		return 0, err
	}
	return v.Float(), nil
}
func (a *arguments) GetBool(name string) (bool, bool, error) {
	v, ok, err := a.get(name, semantic.Bool, false)
	if err != nil || !ok {
		return false, ok, err
	}
	return v.Bool(), ok, nil
}
func (a *arguments) GetRequiredBool(name string) (bool, error) {
	v, _, err := a.get(name, semantic.Bool, true)
	if err != nil {
		return false, err
	}
	return v.Bool(), nil
}

func (a *arguments) GetArray(name string, t semantic.Nature) (values.Array, bool, error) {
	v, ok, err := a.get(name, semantic.Array, false)
	if err != nil || !ok {
		return nil, ok, err
	}
	arr := v.Array()
	if arr.Type().ElementType() != t {
		return nil, true, fmt.Errorf("keyword argument %q should be of an array of type %v, but got an array of type %v", name, t, arr.Type())
	}
	return v.Array(), ok, nil
}
func (a *arguments) GetRequiredArray(name string, t semantic.Nature) (values.Array, error) {
	v, _, err := a.get(name, semantic.Array, true)
	if err != nil {
		return nil, err
	}
	arr := v.Array()
	if arr.Type().ElementType().Nature() != t {
		return nil, fmt.Errorf("keyword argument %q should be of an array of type %v, but got an array of type %v", name, t, arr.Type().ElementType().Nature())
	}
	return arr, nil
}
func (a *arguments) GetFunction(name string) (values.Function, bool, error) {
	v, ok, err := a.get(name, semantic.Function, false)
	if err != nil || !ok {
		return nil, ok, err
	}
	return v.Function(), ok, nil
}
func (a *arguments) GetRequiredFunction(name string) (values.Function, error) {
	v, _, err := a.get(name, semantic.Function, true)
	if err != nil {
		return nil, err
	}
	return v.Function(), nil
}

func (a *arguments) GetObject(name string) (values.Object, bool, error) {
	v, ok, err := a.get(name, semantic.Object, false)
	if err != nil || !ok {
		return nil, ok, err
	}
	return v.Object(), ok, nil
}
func (a *arguments) GetRequiredObject(name string) (values.Object, error) {
	v, _, err := a.get(name, semantic.Object, true)
	if err != nil {
		return nil, err
	}
	return v.Object(), nil
}

func (a *arguments) get(name string, kind semantic.Nature, required bool) (values.Value, bool, error) {
	a.used[name] = true
	v, ok := a.obj.Get(name)
	if !ok {
		if required {
			return nil, false, fmt.Errorf("missing required keyword argument %q", name)
		}
		return nil, false, nil
	}
	if v.PolyType().Nature() != kind {
		return nil, true, fmt.Errorf("keyword argument %q should be of kind %v, but got %v", name, kind, v.PolyType().Nature())
	}
	return v, true, nil
}

func (a *arguments) listUnused() []string {
	var unused []string
	if a.obj != nil {
		a.obj.Range(func(k string, v values.Value) {
			if !a.used[k] {
				unused = append(unused, k)
			}
		})
	}
	return unused
}
