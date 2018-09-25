package semantic

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/influxdata/flux/ast"
)

type Node interface {
	node()
	NodeType() string
	Copy() Node

	json.Marshaler
}

func (*Program) node() {}

func (*BlockStatement) node()              {}
func (*OptionStatement) node()             {}
func (*ExpressionStatement) node()         {}
func (*ReturnStatement) node()             {}
func (*NativeVariableDeclaration) node()   {}
func (*ExternalVariableDeclaration) node() {}

func (*ArrayExpression) node()       {}
func (*FunctionExpression) node()    {}
func (*BinaryExpression) node()      {}
func (*CallExpression) node()        {}
func (*ConditionalExpression) node() {}
func (*IdentifierExpression) node()  {}
func (*LogicalExpression) node()     {}
func (*MemberExpression) node()      {}
func (*ObjectExpression) node()      {}
func (*UnaryExpression) node()       {}

func (*Identifier) node()    {}
func (*Property) node()      {}
func (*FunctionParam) node() {}

func (*BooleanLiteral) node()         {}
func (*DateTimeLiteral) node()        {}
func (*DurationLiteral) node()        {}
func (*FloatLiteral) node()           {}
func (*IntegerLiteral) node()         {}
func (*StringLiteral) node()          {}
func (*RegexpLiteral) node()          {}
func (*UnsignedIntegerLiteral) node() {}

type Statement interface {
	Node
	stmt()
}

func (*BlockStatement) stmt()              {}
func (*OptionStatement) stmt()             {}
func (*ExpressionStatement) stmt()         {}
func (*ReturnStatement) stmt()             {}
func (*NativeVariableDeclaration) stmt()   {}
func (*ExternalVariableDeclaration) stmt() {}

type Expression interface {
	Node
	Type() Type
	expression()
}

func (*ArrayExpression) expression()        {}
func (*BinaryExpression) expression()       {}
func (*BooleanLiteral) expression()         {}
func (*CallExpression) expression()         {}
func (*ConditionalExpression) expression()  {}
func (*DateTimeLiteral) expression()        {}
func (*DurationLiteral) expression()        {}
func (*FloatLiteral) expression()           {}
func (*FunctionExpression) expression()     {}
func (*IdentifierExpression) expression()   {}
func (*IntegerLiteral) expression()         {}
func (*LogicalExpression) expression()      {}
func (*MemberExpression) expression()       {}
func (*ObjectExpression) expression()       {}
func (*RegexpLiteral) expression()          {}
func (*StringLiteral) expression()          {}
func (*UnaryExpression) expression()        {}
func (*UnsignedIntegerLiteral) expression() {}

type Literal interface {
	Expression
	literal()
}

func (*BooleanLiteral) literal()         {}
func (*DateTimeLiteral) literal()        {}
func (*DurationLiteral) literal()        {}
func (*FloatLiteral) literal()           {}
func (*IntegerLiteral) literal()         {}
func (*RegexpLiteral) literal()          {}
func (*StringLiteral) literal()          {}
func (*UnsignedIntegerLiteral) literal() {}

type Program struct {
	Body []Statement `json:"body"`
}

func (*Program) NodeType() string { return "Program" }

func (p *Program) Copy() Node {
	if p == nil {
		return p
	}
	np := new(Program)
	*np = *p

	if len(p.Body) > 0 {
		np.Body = make([]Statement, len(p.Body))
		for i, s := range p.Body {
			np.Body[i] = s.Copy().(Statement)
		}
	}

	return np
}

type BlockStatement struct {
	Body []Statement `json:"body"`
}

func (*BlockStatement) NodeType() string { return "BlockStatement" }

func (s *BlockStatement) ReturnStatement() *ReturnStatement {
	return s.Body[len(s.Body)-1].(*ReturnStatement)
}

func (s *BlockStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(BlockStatement)
	*ns = *s

	if len(s.Body) > 0 {
		ns.Body = make([]Statement, len(s.Body))
		for i, stmt := range s.Body {
			ns.Body[i] = stmt.Copy().(Statement)
		}
	}

	return ns
}

type OptionStatement struct {
	Declaration VariableDeclaration `json:"declaration"`
}

func (s *OptionStatement) NodeType() string { return "OptionStatement" }

func (s *OptionStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(OptionStatement)
	*ns = *s

	ns.Declaration = s.Declaration.Copy().(VariableDeclaration)

	return ns
}

type ExpressionStatement struct {
	Expression Expression `json:"expression"`
}

func (*ExpressionStatement) NodeType() string { return "ExpressionStatement" }

func (s *ExpressionStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(ExpressionStatement)
	*ns = *s

	ns.Expression = s.Expression.Copy().(Expression)

	return ns
}

type ReturnStatement struct {
	Argument Expression `json:"argument"`
}

func (*ReturnStatement) NodeType() string { return "ReturnStatement" }

func (s *ReturnStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(ReturnStatement)
	*ns = *s

	ns.Argument = s.Argument.Copy().(Expression)

	return ns
}

type VariableDeclaration interface {
	Statement
	ID() *Identifier
	InitType() Type
}

type NativeVariableDeclaration struct {
	Identifier *Identifier `json:"identifier"`
	Init       Expression  `json:"init"`
}

func (d *NativeVariableDeclaration) ID() *Identifier {
	return d.Identifier
}
func (d *NativeVariableDeclaration) InitType() Type {
	return d.Init.Type()
}

func (*NativeVariableDeclaration) NodeType() string { return "NativeVariableDeclaration" }

func (s *NativeVariableDeclaration) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(NativeVariableDeclaration)
	*ns = *s

	ns.Identifier = s.Identifier.Copy().(*Identifier)

	if s.Init != nil {
		ns.Init = s.Init.Copy().(Expression)
	}

	return ns
}

type ExternalVariableDeclaration struct {
	Identifier *Identifier `json:"identifier"`
	Type       Type        `json:"type"`
}

func NewExternalVariableDeclaration(name string, typ Type) *ExternalVariableDeclaration {
	return &ExternalVariableDeclaration{
		Identifier: &Identifier{Name: name},
		Type:       typ,
	}
}

func (d *ExternalVariableDeclaration) ID() *Identifier {
	return d.Identifier
}
func (d *ExternalVariableDeclaration) InitType() Type {
	return d.Type
}

func (*ExternalVariableDeclaration) NodeType() string { return "ExternalVariableDeclaration" }

func (s *ExternalVariableDeclaration) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(ExternalVariableDeclaration)
	*ns = *s

	ns.Identifier = s.Identifier.Copy().(*Identifier)

	return ns
}

type ArrayExpression struct {
	Elements []Expression `json:"elements"`
	typ      atomic.Value //    Type
}

func (*ArrayExpression) NodeType() string { return "ArrayExpression" }
func (e *ArrayExpression) Type() Type {
	t := e.typ.Load()
	if t != nil {
		return t.(Type)
	}
	typ := arrayTypeOf(e)
	e.typ.Store(typ)
	return typ
}

func (e *ArrayExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(ArrayExpression)
	*ne = *e

	if len(e.Elements) > 0 {
		ne.Elements = make([]Expression, len(e.Elements))
		for i, elem := range e.Elements {
			ne.Elements[i] = elem.Copy().(Expression)
		}
	}

	return ne
}

type FunctionExpression struct {
	Params []*FunctionParam `json:"params"`
	Body   Node             `json:"body"`
	typ    atomic.Value     //Type

	returnTypeVar TypeVar
}

func (*FunctionExpression) NodeType() string { return "ArrowFunctionExpression" }
func (e *FunctionExpression) Type() Type {
	t := e.typ.Load()
	if t != nil {
		return t.(Type)
	}
	typ := functionTypeOf(e)
	e.typ.Store(typ)
	return typ
}

func (e *FunctionExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(FunctionExpression)
	*ne = *e

	if len(e.Params) > 0 {
		ne.Params = make([]*FunctionParam, len(e.Params))
		for i, p := range e.Params {
			ne.Params[i] = p.Copy().(*FunctionParam)
		}
	}
	ne.Body = e.Body.Copy()

	return ne
}

type FunctionParam struct {
	Key         *Identifier `json:"key"`
	Default     Expression  `json:"default"`
	Piped       bool        `json:"piped,omitempty"`
	declaration VariableDeclaration
}

func (*FunctionParam) NodeType() string { return "FunctionParam" }

func (f *FunctionParam) Type() Type {
	if f.declaration == nil {
		if f.Default != nil {
			f.declaration = &NativeVariableDeclaration{
				Identifier: f.Key,
				Init:       f.Default,
			}
		} else {
			return Invalid
		}
	}
	return f.declaration.InitType()
}

func (p *FunctionParam) Copy() Node {
	if p == nil {
		return p
	}
	np := new(FunctionParam)
	*np = *p

	np.Key = p.Key.Copy().(*Identifier)
	if np.Default != nil {
		np.Default = p.Default.Copy().(Expression)
	}

	return np
}

type BinaryExpression struct {
	Operator ast.OperatorKind `json:"operator"`
	Left     Expression       `json:"left"`
	Right    Expression       `json:"right"`
}

func (*BinaryExpression) NodeType() string { return "BinaryExpression" }
func (e *BinaryExpression) Type() Type {
	return binaryTypesLookup[binarySignature{
		operator: e.Operator,
		left:     e.Left.Type().Kind(),
		right:    e.Right.Type().Kind(),
	}]
}

func (e *BinaryExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(BinaryExpression)
	*ne = *e

	ne.Left = e.Left.Copy().(Expression)
	ne.Right = e.Right.Copy().(Expression)

	return ne
}

type CallExpression struct {
	Callee    Expression        `json:"callee"`
	Arguments *ObjectExpression `json:"arguments"`
	pipe      Expression
}

func (*CallExpression) NodeType() string { return "CallExpression" }
func (e *CallExpression) Type() Type {
	return e.Callee.Type().ReturnType()
}

func (e *CallExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(CallExpression)
	*ne = *e

	ne.Callee = e.Callee.Copy().(Expression)
	ne.Arguments = e.Arguments.Copy().(*ObjectExpression)

	return ne
}

type ConditionalExpression struct {
	Test       Expression `json:"test"`
	Alternate  Expression `json:"alternate"`
	Consequent Expression `json:"consequent"`
}

func (*ConditionalExpression) NodeType() string { return "ConditionalExpression" }

func (e *ConditionalExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(ConditionalExpression)
	*ne = *e

	ne.Test = e.Test.Copy().(Expression)
	ne.Alternate = e.Alternate.Copy().(Expression)
	ne.Consequent = e.Consequent.Copy().(Expression)

	return ne
}

type LogicalExpression struct {
	Operator ast.LogicalOperatorKind `json:"operator"`
	Left     Expression              `json:"left"`
	Right    Expression              `json:"right"`
}

func (*LogicalExpression) NodeType() string { return "LogicalExpression" }
func (*LogicalExpression) Type() Type       { return Bool }

func (e *LogicalExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(LogicalExpression)
	*ne = *e

	ne.Left = e.Left.Copy().(Expression)
	ne.Right = e.Right.Copy().(Expression)

	return ne
}

type MemberExpression struct {
	Object   Expression `json:"object"`
	Property string     `json:"property"`
}

func (*MemberExpression) NodeType() string { return "MemberExpression" }

func (e *MemberExpression) Type() Type {
	t := e.Object.Type()
	if t.Kind() != Object {
		return Invalid
	}
	return e.Object.Type().PropertyType(e.Property)
}

func (e *MemberExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(MemberExpression)
	*ne = *e

	ne.Object = e.Object.Copy().(Expression)

	return ne
}

type ObjectExpression struct {
	Properties []*Property  `json:"properties"`
	typ        atomic.Value //Type
}

func (*ObjectExpression) NodeType() string { return "ObjectExpression" }
func (e *ObjectExpression) Type() Type {
	t := e.typ.Load()
	if t != nil {
		return t.(Type)
	}
	typ := objectTypeOf(e)
	e.typ.Store(typ)
	return typ
}

func (e *ObjectExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(ObjectExpression)
	*ne = *e

	if len(e.Properties) > 0 {
		ne.Properties = make([]*Property, len(e.Properties))
		for i, prop := range e.Properties {
			ne.Properties[i] = prop.Copy().(*Property)
		}
	}

	return ne
}

type UnaryExpression struct {
	Operator ast.OperatorKind `json:"operator"`
	Argument Expression       `json:"argument"`
}

func (*UnaryExpression) NodeType() string { return "UnaryExpression" }
func (e *UnaryExpression) Type() Type {
	return e.Argument.Type()
}

func (e *UnaryExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(UnaryExpression)
	*ne = *e

	ne.Argument = e.Argument.Copy().(Expression)

	return ne
}

type Property struct {
	Key   *Identifier `json:"key"`
	Value Expression  `json:"value"`
}

func (*Property) NodeType() string { return "Property" }

func (p *Property) Copy() Node {
	if p == nil {
		return p
	}
	np := new(Property)
	*np = *p

	np.Value = p.Value.Copy().(Expression)

	return np
}

type IdentifierExpression struct {
	Name string `json:"name"`
	// declaration is the node that declares this identifier
	declaration VariableDeclaration
}

func (*IdentifierExpression) NodeType() string { return "IdentifierExpression" }

func (e *IdentifierExpression) Type() Type {
	if e.declaration == nil {
		return Invalid
	}
	return e.declaration.InitType()
}

func (e *IdentifierExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(IdentifierExpression)
	*ne = *e

	if ne.declaration != nil {
		ne.declaration = e.declaration.Copy().(VariableDeclaration)
	}

	return ne
}

type Identifier struct {
	Name string `json:"name"`
}

func (*Identifier) NodeType() string { return "Identifier" }

func (i *Identifier) Copy() Node {
	if i == nil {
		return i
	}
	ni := new(Identifier)
	*ni = *i

	return ni
}

type BooleanLiteral struct {
	Value bool `json:"value"`
}

func (*BooleanLiteral) NodeType() string { return "BooleanLiteral" }
func (*BooleanLiteral) Type() Type       { return Bool }

func (l *BooleanLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(BooleanLiteral)
	*nl = *l

	return nl
}

type DateTimeLiteral struct {
	Value time.Time `json:"value"`
}

func (*DateTimeLiteral) NodeType() string { return "DateTimeLiteral" }
func (*DateTimeLiteral) Type() Type       { return Time }

func (l *DateTimeLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(DateTimeLiteral)
	*nl = *l

	return nl
}

type DurationLiteral struct {
	Value time.Duration `json:"value"`
}

func (*DurationLiteral) NodeType() string { return "DurationLiteral" }
func (*DurationLiteral) Type() Type       { return Duration }

func (l *DurationLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(DurationLiteral)
	*nl = *l

	return nl
}

type IntegerLiteral struct {
	Value int64 `json:"value"`
}

func (*IntegerLiteral) NodeType() string { return "IntegerLiteral" }
func (*IntegerLiteral) Type() Type       { return Int }

func (l *IntegerLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(IntegerLiteral)
	*nl = *l

	return nl
}

type FloatLiteral struct {
	Value float64 `json:"value"`
}

func (*FloatLiteral) NodeType() string { return "FloatLiteral" }
func (*FloatLiteral) Type() Type       { return Float }

func (l *FloatLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(FloatLiteral)
	*nl = *l

	return nl
}

type RegexpLiteral struct {
	Value *regexp.Regexp `json:"value"`
}

func (*RegexpLiteral) NodeType() string { return "RegexpLiteral" }
func (*RegexpLiteral) Type() Type       { return Regexp }

func (l *RegexpLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(RegexpLiteral)
	*nl = *l

	nl.Value = l.Value.Copy()

	return nl
}

type StringLiteral struct {
	Value string `json:"value"`
}

func (*StringLiteral) NodeType() string { return "StringLiteral" }
func (*StringLiteral) Type() Type       { return String }

func (l *StringLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(StringLiteral)
	*nl = *l

	return nl
}

type UnsignedIntegerLiteral struct {
	Value uint64 `json:"value"`
}

func (*UnsignedIntegerLiteral) NodeType() string { return "UnsignedIntegerLiteral" }
func (*UnsignedIntegerLiteral) Type() Type       { return UInt }

func (l *UnsignedIntegerLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(UnsignedIntegerLiteral)
	*nl = *l

	return nl
}

// New creates a semantic graph from the provided AST
func New(prog *ast.Program) (*Program, error) {
	return analyzeProgram(prog)
}

func analyzeProgram(prog *ast.Program) (*Program, error) {
	p := &Program{
		Body: make([]Statement, len(prog.Body)),
	}
	for i, s := range prog.Body {
		n, err := analyzeStatment(s)
		if err != nil {
			return nil, err
		}
		p.Body[i] = n
	}
	return p, nil
}

func analyzeNode(n ast.Node) (Node, error) {
	switch n := n.(type) {
	case ast.Statement:
		return analyzeStatment(n)
	case ast.Expression:
		return analyzeExpression(n)
	default:
		return nil, fmt.Errorf("unsupported node %T", n)
	}
}

func analyzeStatment(s ast.Statement) (Statement, error) {
	switch s := s.(type) {
	case *ast.BlockStatement:
		return analyzeBlockStatement(s)
	case *ast.OptionStatement:
		return analyzeOptionStatement(s)
	case *ast.ExpressionStatement:
		return analyzeExpressionStatement(s)
	case *ast.ReturnStatement:
		return analyzeReturnStatement(s)
	case *ast.VariableDeclaration:
		// Expect a single declaration
		if len(s.Declarations) != 1 {
			return nil, fmt.Errorf("only single variable declarations are supported, found %d declarations", len(s.Declarations))
		}
		return analyzeVariableDeclaration(s.Declarations[0])
	default:
		return nil, fmt.Errorf("unsupported statement %T", s)
	}
}

func analyzeBlockStatement(block *ast.BlockStatement) (*BlockStatement, error) {
	b := &BlockStatement{
		Body: make([]Statement, len(block.Body)),
	}
	for i, s := range block.Body {
		n, err := analyzeStatment(s)
		if err != nil {
			return nil, err
		}
		b.Body[i] = n
	}
	last := len(b.Body) - 1
	if _, ok := b.Body[last].(*ReturnStatement); !ok {
		return nil, errors.New("missing return statement in block")
	}
	return b, nil
}

func analyzeOptionStatement(option *ast.OptionStatement) (*OptionStatement, error) {
	declaration, err := analyzeVariableDeclaration(option.Declaration)
	if err != nil {
		return nil, err
	}
	return &OptionStatement{
		Declaration: declaration,
	}, nil
}

func analyzeExpressionStatement(expr *ast.ExpressionStatement) (*ExpressionStatement, error) {
	e, err := analyzeExpression(expr.Expression)
	if err != nil {
		return nil, err
	}
	return &ExpressionStatement{
		Expression: e,
	}, nil
}

func analyzeReturnStatement(ret *ast.ReturnStatement) (*ReturnStatement, error) {
	arg, err := analyzeExpression(ret.Argument)
	if err != nil {
		return nil, err
	}
	return &ReturnStatement{
		Argument: arg,
	}, nil
}

func analyzeVariableDeclaration(decl *ast.VariableDeclarator) (*NativeVariableDeclaration, error) {
	id, err := analyzeIdentifier(decl.ID)
	if err != nil {
		return nil, err
	}
	init, err := analyzeExpression(decl.Init)
	if err != nil {
		return nil, err
	}
	vd := &NativeVariableDeclaration{
		Identifier: id,
		Init:       init,
	}
	return vd, nil
}

func analyzeExpression(expr ast.Expression) (Expression, error) {
	switch expr := expr.(type) {
	case *ast.ArrowFunctionExpression:
		return analyzeArrowFunctionExpression(expr)
	case *ast.CallExpression:
		return analyzeCallExpression(expr)
	case *ast.MemberExpression:
		return analyzeMemberExpression(expr)
	case *ast.PipeExpression:
		return analyzePipeExpression(expr)
	case *ast.BinaryExpression:
		return analyzeBinaryExpression(expr)
	case *ast.UnaryExpression:
		return analyzeUnaryExpression(expr)
	case *ast.LogicalExpression:
		return analyzeLogicalExpression(expr)
	case *ast.ObjectExpression:
		return analyzeObjectExpression(expr)
	case *ast.ArrayExpression:
		return analyzeArrayExpression(expr)
	case *ast.Identifier:
		return analyzeIdentifierExpression(expr)
	case ast.Literal:
		return analyzeLiteral(expr)
	default:
		return nil, fmt.Errorf("unsupported expression %T", expr)
	}
}

func analyzeLiteral(lit ast.Literal) (Literal, error) {
	switch lit := lit.(type) {
	case *ast.StringLiteral:
		return analyzeStringLiteral(lit)
	case *ast.BooleanLiteral:
		return analyzeBooleanLiteral(lit)
	case *ast.FloatLiteral:
		return analyzeFloatLiteral(lit)
	case *ast.IntegerLiteral:
		return analyzeIntegerLiteral(lit)
	case *ast.UnsignedIntegerLiteral:
		return analyzeUnsignedIntegerLiteral(lit)
	case *ast.RegexpLiteral:
		return analyzeRegexpLiteral(lit)
	case *ast.DurationLiteral:
		return analyzeDurationLiteral(lit)
	case *ast.DateTimeLiteral:
		return analyzeDateTimeLiteral(lit)
	case *ast.PipeLiteral:
		return nil, errors.New("a pipe literal may only be used as a default value for an argument in a function definition")
	default:
		return nil, fmt.Errorf("unsupported literal %T", lit)
	}
}

func analyzeArrowFunctionExpression(arrow *ast.ArrowFunctionExpression) (*FunctionExpression, error) {
	f := &FunctionExpression{
		Params: make([]*FunctionParam, len(arrow.Params)),
	}
	pipedCount := 0
	for i, p := range arrow.Params {
		key, err := analyzeIdentifier(p.Key)
		if err != nil {
			return nil, err
		}

		var def Expression
		var piped bool
		if p.Value != nil {
			if _, ok := p.Value.(*ast.PipeLiteral); ok {
				// Special case the PipeLiteral
				piped = true
				pipedCount++
				if pipedCount > 1 {
					return nil, errors.New("only a single argument may be piped")
				}
			} else {
				d, err := analyzeExpression(p.Value)
				if err != nil {
					return nil, err
				}
				def = d
			}
		}

		declaration := &NativeVariableDeclaration{
			Identifier: key,
			Init:       def,
		}

		f.Params[i] = &FunctionParam{
			Key:         key,
			Default:     def,
			Piped:       piped,
			declaration: declaration,
		}

	}

	b, err := analyzeNode(arrow.Body)
	if err != nil {
		return nil, err
	}
	f.Body = b

	return f, nil
}

func analyzeCallExpression(call *ast.CallExpression) (*CallExpression, error) {
	callee, err := analyzeExpression(call.Callee)
	if err != nil {
		return nil, err
	}
	var args *ObjectExpression
	if l := len(call.Arguments); l > 1 {
		return nil, fmt.Errorf("arguments are not a single object expression %v", args)
	} else if l == 1 {
		obj, ok := call.Arguments[0].(*ast.ObjectExpression)
		if !ok {
			return nil, fmt.Errorf("arguments not an object expression")
		}
		var err error
		args, err = analyzeObjectExpression(obj)
		if err != nil {
			return nil, err
		}
	} else {
		args = new(ObjectExpression)
	}

	return &CallExpression{
		Callee:    callee,
		Arguments: args,
	}, nil
}

func analyzeMemberExpression(member *ast.MemberExpression) (*MemberExpression, error) {
	obj, err := analyzeExpression(member.Object)
	if err != nil {
		return nil, err
	}

	var propertyName string
	switch p := member.Property.(type) {
	case *ast.Identifier:
		propertyName = p.Name
	case *ast.StringLiteral:
		propertyName = p.Value
	case *ast.IntegerLiteral:
		propertyName = strconv.FormatInt(p.Value, 10)
	default:
		return nil, fmt.Errorf("unsupported member property expression of type %T", member.Property)
	}

	return &MemberExpression{
		Object:   obj,
		Property: propertyName,
	}, nil
}

func analyzePipeExpression(pipe *ast.PipeExpression) (*CallExpression, error) {
	call, err := analyzeCallExpression(pipe.Call)
	if err != nil {
		return nil, err
	}

	value, err := analyzeExpression(pipe.Argument)
	if err != nil {
		return nil, err
	}

	call.pipe = value
	return call, nil
}

func analyzeBinaryExpression(binary *ast.BinaryExpression) (*BinaryExpression, error) {
	left, err := analyzeExpression(binary.Left)
	if err != nil {
		return nil, err
	}
	right, err := analyzeExpression(binary.Right)
	if err != nil {
		return nil, err
	}
	return &BinaryExpression{
		Operator: binary.Operator,
		Left:     left,
		Right:    right,
	}, nil
}

func analyzeUnaryExpression(unary *ast.UnaryExpression) (*UnaryExpression, error) {
	arg, err := analyzeExpression(unary.Argument)
	if err != nil {
		return nil, err
	}
	return &UnaryExpression{
		Operator: unary.Operator,
		Argument: arg,
	}, nil
}
func analyzeLogicalExpression(logical *ast.LogicalExpression) (*LogicalExpression, error) {
	left, err := analyzeExpression(logical.Left)
	if err != nil {
		return nil, err
	}
	right, err := analyzeExpression(logical.Right)
	if err != nil {
		return nil, err
	}
	return &LogicalExpression{
		Operator: logical.Operator,
		Left:     left,
		Right:    right,
	}, nil
}
func analyzeObjectExpression(obj *ast.ObjectExpression) (*ObjectExpression, error) {
	o := &ObjectExpression{
		Properties: make([]*Property, len(obj.Properties)),
	}
	for i, p := range obj.Properties {
		n, err := analyzeProperty(p)
		if err != nil {
			return nil, err
		}
		o.Properties[i] = n
	}
	return o, nil
}
func analyzeArrayExpression(array *ast.ArrayExpression) (*ArrayExpression, error) {
	a := &ArrayExpression{
		Elements: make([]Expression, len(array.Elements)),
	}
	for i, e := range array.Elements {
		n, err := analyzeExpression(e)
		if err != nil {
			return nil, err
		}
		a.Elements[i] = n
	}
	return a, nil
}

func analyzeIdentifier(ident *ast.Identifier) (*Identifier, error) {
	return &Identifier{
		Name: ident.Name,
	}, nil
}

func analyzeIdentifierExpression(ident *ast.Identifier) (*IdentifierExpression, error) {
	return &IdentifierExpression{
		Name: ident.Name,
	}, nil
}

func analyzeProperty(property *ast.Property) (*Property, error) {
	key, err := analyzeIdentifier(property.Key)
	if err != nil {
		return nil, err
	}
	value, err := analyzeExpression(property.Value)
	if err != nil {
		return nil, err
	}
	return &Property{
		Key:   key,
		Value: value,
	}, nil
}

func analyzeDateTimeLiteral(lit *ast.DateTimeLiteral) (*DateTimeLiteral, error) {
	return &DateTimeLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeDurationLiteral(lit *ast.DurationLiteral) (*DurationLiteral, error) {
	var duration time.Duration
	for _, d := range lit.Values {
		dur, err := toDuration(d)
		if err != nil {
			return nil, err
		}
		duration += dur
	}
	return &DurationLiteral{
		Value: duration,
	}, nil
}
func analyzeFloatLiteral(lit *ast.FloatLiteral) (*FloatLiteral, error) {
	return &FloatLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeIntegerLiteral(lit *ast.IntegerLiteral) (*IntegerLiteral, error) {
	return &IntegerLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeUnsignedIntegerLiteral(lit *ast.UnsignedIntegerLiteral) (*UnsignedIntegerLiteral, error) {
	return &UnsignedIntegerLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeStringLiteral(lit *ast.StringLiteral) (*StringLiteral, error) {
	return &StringLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeBooleanLiteral(lit *ast.BooleanLiteral) (*BooleanLiteral, error) {
	return &BooleanLiteral{
		Value: lit.Value,
	}, nil
}
func analyzeRegexpLiteral(lit *ast.RegexpLiteral) (*RegexpLiteral, error) {
	return &RegexpLiteral{
		Value: lit.Value,
	}, nil
}
func toDuration(lit ast.Duration) (time.Duration, error) {
	// TODO: This is temporary code until we have proper duration type that takes different months, DST, etc into account
	var dur time.Duration
	var err error
	mag := lit.Magnitude
	unit := lit.Unit

	switch unit {
	case "y":
		mag *= 12
		unit = "mo"
		fallthrough
	case "mo":
		const weeksPerMonth = 365.25 / 12 / 7
		mag = int64(float64(mag) * weeksPerMonth)
		unit = "w"
		fallthrough
	case "w":
		mag *= 7
		unit = "d"
		fallthrough
	case "d":
		mag *= 24
		unit = "h"
		fallthrough
	default:
		// ParseDuration will handle h, m, s, ms, us, ns.
		dur, err = time.ParseDuration(strconv.FormatInt(mag, 10) + unit)
	}
	return dur, err
}
