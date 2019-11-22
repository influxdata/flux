// Package ast declares the types used to represent the syntax tree for Flux source code.
package ast

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/influxdata/flux/ast/internal/fbast"
	"github.com/influxdata/flux/internal/parser"
)

// Position represents a specific location in the source
type Position struct {
	Line   int `json:"line"`   // Line is the line in the source marked by this position
	Column int `json:"column"` // Column is the column in the source marked by this position
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

func (p Position) Less(o Position) bool {
	if p.Line == o.Line {
		return p.Column < o.Column
	}
	return p.Line < o.Line
}

func (p Position) IsValid() bool {
	return p.Line > 0 && p.Column > 0
}

func (p Position) FromBuf(buf *fbast.Position) {
	p.Line = int(buf.Line())
	p.Column = int(buf.Column())
}

// SourceLocation represents the location of a node in the AST
type SourceLocation struct {
	File   string   `json:"file,omitempty"`
	Start  Position `json:"start"`            // Start is the location in the source the node starts
	End    Position `json:"end"`              // End is the location in the source the node ends
	Source string   `json:"source,omitempty"` // Source is optional raw source
}

func (l SourceLocation) String() string {
	if l.File != "" {
		return fmt.Sprintf("%s|%v-%v", l.File, l.Start, l.End)
	}
	return fmt.Sprintf("%v-%v", l.Start, l.End)
}

func (l SourceLocation) Less(o SourceLocation) bool {
	if l.Start == o.Start {
		return l.End.Less(o.End)
	}
	return l.Start.Less(o.Start)
}

func (l SourceLocation) IsValid() bool {
	return l.Start.IsValid() && l.End.IsValid()
}

func (l *SourceLocation) Copy() *SourceLocation {
	if l == nil {
		return nil
	}
	nl := *l
	return &nl
}

func (l SourceLocation) FromBuf(buf *fbast.SourceLocation) {
	l.File = string(buf.File())
	l.Start.FromBuf(buf.Start(nil))
	l.End.FromBuf(buf.End(nil))
	l.Source = string(buf.Source())
}

// Node represents a node in the InfluxDB abstract syntax tree.
type Node interface {
	node()
	Type() string // Type property is a string that contains the variant type of the node
	Location() SourceLocation
	Errs() []Error
	Copy() Node

	// All node must support json marshalling
	json.Marshaler
}

func (*Package) node()           {}
func (*File) node()              {}
func (*PackageClause) node()     {}
func (*ImportDeclaration) node() {}
func (*Block) node()             {}

func (*BadStatement) node()        {}
func (*ExpressionStatement) node() {}
func (*ReturnStatement) node()     {}
func (*OptionStatement) node()     {}
func (*BuiltinStatement) node()    {}
func (*TestStatement) node()       {}
func (*VariableAssignment) node()  {}
func (*MemberAssignment) node()    {}

func (*StringExpression) node()      {}
func (*ParenExpression) node()       {}
func (*ArrayExpression) node()       {}
func (*FunctionExpression) node()    {}
func (*BinaryExpression) node()      {}
func (*CallExpression) node()        {}
func (*ConditionalExpression) node() {}
func (*LogicalExpression) node()     {}
func (*MemberExpression) node()      {}
func (*IndexExpression) node()       {}
func (*PipeExpression) node()        {}
func (*ObjectExpression) node()      {}
func (*UnaryExpression) node()       {}

func (*Property) node()   {}
func (*Identifier) node() {}

func (*TextPart) node()         {}
func (*InterpolatedPart) node() {}

func (*BooleanLiteral) node()         {}
func (*DateTimeLiteral) node()        {}
func (*DurationLiteral) node()        {}
func (*FloatLiteral) node()           {}
func (*IntegerLiteral) node()         {}
func (*PipeLiteral) node()            {}
func (*RegexpLiteral) node()          {}
func (*StringLiteral) node()          {}
func (*UnsignedIntegerLiteral) node() {}

// BaseNode holds the attributes every expression or statement should have
type BaseNode struct {
	Loc    *SourceLocation `json:"location,omitempty"`
	Errors []Error         `json:"errors,omitempty"`
}

// Location is the source location of the Node
func (b BaseNode) Location() SourceLocation {
	if b.Loc == nil {
		return SourceLocation{}
	}
	return *b.Loc
}

func (b BaseNode) Errs() []Error {
	return b.Errors
}

func (b BaseNode) Copy() BaseNode {
	// Note b is already shallow copy because of the non pointer receiver
	b.Loc = b.Loc.Copy()
	if len(b.Errors) > 0 {
		cpy := make([]Error, len(b.Errors))
		copy(cpy, b.Errors)
		b.Errors = cpy
	}
	return b
}

func (b BaseNode) FromBuf(buf *fbast.BaseNode) {
	b.Location().FromBuf(buf.Loc(nil))
	b.Errors = make([]Error, buf.ErrorsLength())
	for i := 0; i < buf.ErrorsLength(); i++ {
		b.Errors[i] = Error{string(buf.Errors(i))}
	}
}

// Error represents an error in the AST construction.
// The node that this is attached to is not valid.
type Error struct {
	Msg string `json:"msg"`
}

func (e Error) Error() string {
	return e.Msg
}

// Package represents a complete package source tree
type Package struct {
	BaseNode
	Path    string  `json:"path,omitempty"`
	Package string  `json:"package"`
	Files   []*File `json:"files"`
}

// Type is the abstract type
func (*Package) Type() string { return "Package" }

func (p *Package) Copy() Node {
	if p == nil {
		return p
	}
	np := new(Package)
	*np = *p
	np.BaseNode = p.BaseNode.Copy()

	if len(p.Files) > 0 {
		np.Files = make([]*File, len(p.Files))
		for i, f := range p.Files {
			np.Files[i] = f.Copy().(*File)
		}
	}
	return np
}

func (p Package) FromBuf(buf *fbast.Package) {
	p.BaseNode.FromBuf(buf.BaseNode(nil))
	p.Path = string(buf.Path())
	p.Package = string(buf.Package())
	p.Files = make([]*File, buf.FilesLength())
	for i := 0; i < buf.FilesLength(); i++ {
		fbf := new(fbast.File)
		if !buf.Files(fbf, i) {
			p.BaseNode.Errors = append(p.BaseNode.Errors,
				Error{fmt.Sprintf("Encountered error in deserializing Package.Files[%d]", i)})
		} else {
			p.Files[i] = File{}.FromBuf(fbf)
		}
	}
}

// File represents a source from a single file
type File struct {
	BaseNode
	Name    string               `json:"name,omitempty"` // name of the file
	Package *PackageClause       `json:"package"`
	Imports []*ImportDeclaration `json:"imports"`
	Body    []Statement          `json:"body"`
}

// Type is the abstract type
func (*File) Type() string { return "File" }

func (f *File) Copy() Node {
	if f == nil {
		return f
	}
	nf := new(File)
	*nf = *f
	nf.BaseNode = f.BaseNode.Copy()

	nf.Package = f.Package.Copy().(*PackageClause)

	if len(f.Imports) > 0 {
		nf.Imports = make([]*ImportDeclaration, len(f.Imports))
		for i, s := range f.Imports {
			nf.Imports[i] = s.Copy().(*ImportDeclaration)
		}
	}

	if len(f.Body) > 0 {
		nf.Body = make([]Statement, len(f.Body))
		for i, s := range f.Body {
			nf.Body[i] = s.Copy().(Statement)
		}
	}
	return nf
}

func (f File) FromBuf(buf *fbast.File) *File {
	f.BaseNode.FromBuf(buf.BaseNode(nil))
	f.Name = string(buf.Name())
	f.Package = PackageClause{}.FromBuf(buf.Package(nil))
	f.Imports = make([]*ImportDeclaration, buf.ImportsLength())
	for i := 0; i < buf.ImportsLength(); i++ {
		fbd := new(fbast.ImportDeclaration)
		if !buf.Imports(fbd, i) {
			f.BaseNode.Errors = append(f.BaseNode.Errors,
				Error{fmt.Sprintf("Encountered error in deserializing File.Imports[%d]", i)})
		} else {
			f.Imports[i] = ImportDeclaration{}.FromBuf(fbd)
		}
	}
	var err []Error
	f.Body, err = statementArrayFromBuf(buf.BodyLength(), buf.Body, "File.Body")
	if len(err) > 0 {
		f.BaseNode.Errors = append(f.BaseNode.Errors, err...)
	}
	return &f
}

// PackageClause defines the current package identifier.
type PackageClause struct {
	BaseNode
	Name *Identifier `json:"name"`
}

// Type is the abstract type
func (*PackageClause) Type() string { return "PackageClause" }

func (c *PackageClause) Copy() Node {
	if c == nil {
		return c
	}
	nc := new(PackageClause)
	*nc = *c
	nc.BaseNode = c.BaseNode.Copy()

	nc.Name = c.Name.Copy().(*Identifier)
	return nc
}

func (c PackageClause) FromBuf(buf *fbast.PackageClause) *PackageClause {
	c.BaseNode.FromBuf(buf.BaseNode(nil))
	c.Name = Identifier{}.FromBuf(buf.Name(nil))
	return &c
}

// ImportDeclaration declares a single import
type ImportDeclaration struct {
	BaseNode
	As   *Identifier    `json:"as"`
	Path *StringLiteral `json:"path"`
}

// Type is the abstract type
func (*ImportDeclaration) Type() string { return "ImportDeclaration" }

func (d *ImportDeclaration) Copy() Node {
	if d == nil {
		return d
	}
	nd := new(ImportDeclaration)
	*nd = *d
	nd.BaseNode = d.BaseNode.Copy()

	return nd
}

func (d ImportDeclaration) FromBuf(buf *fbast.ImportDeclaration) *ImportDeclaration {
	d.BaseNode.FromBuf(buf.BaseNode(nil))
	d.As = Identifier{}.FromBuf(buf.As(nil))
	d.Path = StringLiteral{}.FromBuf(buf.Path(nil))
	return &d
}

// Block is a set of statements
type Block struct {
	BaseNode
	Body []Statement `json:"body"`
}

// Type is the abstract type
func (*Block) Type() string { return "Block" }

func (s *Block) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(Block)
	*ns = *s
	ns.BaseNode = s.BaseNode.Copy()

	if len(s.Body) > 0 {
		ns.Body = make([]Statement, len(s.Body))
		for i, stmt := range s.Body {
			ns.Body[i] = stmt.Copy().(Statement)
		}
	}
	return ns
}

func (s Block) FromBuf(buf *fbast.Block) *Block {
	s.BaseNode.FromBuf(buf.BaseNode(nil))
	var err []Error
	s.Body, err = statementArrayFromBuf(buf.BodyLength(), buf.Body, "Block.Body")
	if len(err) > 0 {
		s.BaseNode.Errors = append(s.BaseNode.Errors, err...)
	}
	return &s
}

// Statement Perhaps we don't even want statements nor expression statements
type Statement interface {
	Node
	stmt()
}

func (*BadStatement) stmt()        {}
func (*VariableAssignment) stmt()  {}
func (*MemberAssignment) stmt()    {}
func (*ExpressionStatement) stmt() {}
func (*ReturnStatement) stmt()     {}
func (*OptionStatement) stmt()     {}
func (*BuiltinStatement) stmt()    {}
func (*TestStatement) stmt()       {}

type Assignment interface {
	Statement
	assignment()
}

func (*VariableAssignment) assignment() {}
func (*MemberAssignment) assignment()   {}

// BadStatement is a placeholder for statements for which no correct statement nodes
// can be created.
type BadStatement struct {
	BaseNode
	Text string `json:"text"`
}

// Type is the abstract type.
func (*BadStatement) Type() string { return "BadStatement" }

func (s *BadStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := *s
	ns.BaseNode = s.BaseNode.Copy()
	return &ns
}

func (s BadStatement) FromBuf(buf *fbast.BadStatement) *BadStatement {
	s.BaseNode.FromBuf(buf.BaseNode(nil))
	s.Text = string(buf.Text())
	return &s
}

// ExpressionStatement may consist of an expression that does not return a value and is executed solely for its side-effects.
type ExpressionStatement struct {
	BaseNode
	Expression Expression `json:"expression"`
}

// Type is the abstract type
func (*ExpressionStatement) Type() string { return "ExpressionStatement" }

func (s *ExpressionStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(ExpressionStatement)
	*ns = *s
	ns.BaseNode = s.BaseNode.Copy()

	if s.Expression != nil {
		ns.Expression = s.Expression.Copy().(Expression)
	}

	return ns
}

func (s ExpressionStatement) FromBuf(buf *fbast.ExpressionStatement) *ExpressionStatement {
	s.BaseNode.FromBuf(buf.BaseNode(nil))
	s.Expression = exprFromBuf("ExpressionStatement.Expression", s.BaseNode, buf.Expression, buf.ExpressionType())
	return &s
}

// ReturnStatement defines an Expression to return
type ReturnStatement struct {
	BaseNode
	Argument Expression `json:"argument"`
}

// Type is the abstract type
func (*ReturnStatement) Type() string { return "ReturnStatement" }

func (s *ReturnStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(ReturnStatement)
	*ns = *s
	ns.BaseNode = s.BaseNode.Copy()

	if s.Argument != nil {
		ns.Argument = s.Argument.Copy().(Expression)
	}

	return ns
}

func (s ReturnStatement) FromBuf(buf *fbast.ReturnStatement) *ReturnStatement {
	s.BaseNode.FromBuf(buf.BaseNode(nil))
	s.Argument = exprFromBuf("ReturnStatement.Argument", s.BaseNode, buf.Argument, buf.ArgumentType())
	return &s
}

// OptionStatement syntactically is a single variable declaration
type OptionStatement struct {
	BaseNode
	Assignment Assignment `json:"assignment"`
}

// Type is the abstract type
func (*OptionStatement) Type() string { return "OptionStatement" }

// Copy returns a deep copy of an OptionStatement Node
func (s *OptionStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(OptionStatement)
	*ns = *s
	ns.BaseNode = s.BaseNode.Copy()

	if s.Assignment != nil {
		ns.Assignment = s.Assignment.Copy().(Assignment)
	}

	return ns
}

func (s OptionStatement) FromBuf(buf *fbast.OptionStatement) *OptionStatement {
	s.BaseNode.FromBuf(buf.BaseNode(nil))
	s.Assignment = assignmentFromBuf("OptionStatement.Assignment", s.BaseNode, buf.Assignment, buf.AssignmentType())
	return &s
}

// BuiltinStatement declares a builtin identifier and its type
type BuiltinStatement struct {
	BaseNode
	ID *Identifier `json:"id"`
	// TODO(nathanielc): Add type expression here
	// Type TypeExpression
}

// Type is the abstract type
func (*BuiltinStatement) Type() string { return "BuiltinStatement" }

// Copy returns a deep copy of an BuiltinStatement Node
func (s *BuiltinStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(BuiltinStatement)
	*ns = *s
	ns.BaseNode = s.BaseNode.Copy()

	ns.ID = s.ID.Copy().(*Identifier)

	return ns
}

func (s BuiltinStatement) FromBuf(buf *fbast.BuiltinStatement) *BuiltinStatement {
	s.BaseNode.FromBuf(buf.BaseNode(nil))
	s.ID = Identifier{}.FromBuf(buf.Id(nil))
	return &s
}

// TestStatement declares a Flux test case
type TestStatement struct {
	BaseNode
	Assignment *VariableAssignment `json:"assignment"`
}

// Type is the abstract type
func (*TestStatement) Type() string { return "TestStatement" }

// Copy returns a deep copy of a TestStatement Node
func (s *TestStatement) Copy() Node {
	if s == nil {
		return s
	}
	ns := new(TestStatement)
	*ns = *s
	ns.BaseNode = s.BaseNode.Copy()

	if s.Assignment != nil {
		ns.Assignment = s.Assignment.Copy().(*VariableAssignment)
	}

	return ns
}

func (s TestStatement) FromBuf(buf *fbast.TestStatement) *TestStatement {
	s.BaseNode.FromBuf(buf.BaseNode(nil))
	s.Assignment = assignmentFromBuf("TestStatement.Assignment",
		s.BaseNode, buf.Assignment, buf.AssignmentType()).(*VariableAssignment)
	return &s
}

// VariableAssignment represents the declaration of a variable
type VariableAssignment struct {
	BaseNode
	ID   *Identifier `json:"id"`
	Init Expression  `json:"init"`
}

// Type is the abstract type
func (*VariableAssignment) Type() string { return "VariableAssignment" }

func (d *VariableAssignment) Copy() Node {
	if d == nil {
		return d
	}
	nd := new(VariableAssignment)
	*nd = *d
	nd.BaseNode = d.BaseNode.Copy()

	if d.Init != nil {
		nd.Init = d.Init.Copy().(Expression)
	}

	return nd
}

func (d VariableAssignment) FromBuf(buf *fbast.VariableAssignment) *VariableAssignment {
	d.BaseNode.FromBuf(buf.BaseNode(nil))
	d.ID = Identifier{}.FromBuf(buf.Id(nil))
	d.Init = exprFromBuf("VariableAssignment.Init", d.BaseNode, buf.Init_, buf.Init_type())
	return &d
}

type MemberAssignment struct {
	BaseNode
	Member *MemberExpression `json:"member"`
	Init   Expression        `json:"init"`
}

func (*MemberAssignment) Type() string { return "MemberAssignment" }

func (a *MemberAssignment) Copy() Node {
	if a == nil {
		return a
	}
	na := new(MemberAssignment)
	*na = *a
	na.BaseNode = a.BaseNode.Copy()

	if a.Member != nil {
		na.Member = a.Member.Copy().(*MemberExpression)
	}
	if a.Init != nil {
		na.Init = a.Init.Copy().(Expression)
	}

	return na
}

func (a MemberAssignment) FromBuf(buf *fbast.MemberAssignment) *MemberAssignment {
	a.BaseNode.FromBuf(buf.BaseNode(nil))
	a.Member = MemberExpression{}.FromBuf(buf.Member(nil))
	a.Init = exprFromBuf("MemberAssignment.Init", a.BaseNode, buf.Init_, buf.Init_type())
	return &a
}

// Expression represents an action that can be performed by InfluxDB that can be evaluated to a value.
type Expression interface {
	Node
	expression()
}

func (*StringExpression) expression()       {}
func (*ParenExpression) expression()        {}
func (*ArrayExpression) expression()        {}
func (*FunctionExpression) expression()     {}
func (*BinaryExpression) expression()       {}
func (*BooleanLiteral) expression()         {}
func (*CallExpression) expression()         {}
func (*ConditionalExpression) expression()  {}
func (*DateTimeLiteral) expression()        {}
func (*DurationLiteral) expression()        {}
func (*FloatLiteral) expression()           {}
func (*Identifier) expression()             {}
func (*IntegerLiteral) expression()         {}
func (*LogicalExpression) expression()      {}
func (*MemberExpression) expression()       {}
func (*IndexExpression) expression()        {}
func (*ObjectExpression) expression()       {}
func (*PipeExpression) expression()         {}
func (*PipeLiteral) expression()            {}
func (*RegexpLiteral) expression()          {}
func (*StringLiteral) expression()          {}
func (*UnaryExpression) expression()        {}
func (*UnsignedIntegerLiteral) expression() {}

type StringExpression struct {
	BaseNode

	Parts []StringExpressionPart `json:"parts"`
}

func (*StringExpression) Type() string { return "StringExpression" }

func (e *StringExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(StringExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if len(e.Parts) > 0 {
		ne.Parts = make([]StringExpressionPart, len(e.Parts))
		for i, p := range e.Parts {
			ne.Parts[i] = p.Copy().(StringExpressionPart)
		}
	}

	return ne
}

func (e StringExpression) FromBuf(buf *fbast.StringExpression) *StringExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.Parts = make([]StringExpressionPart, buf.PartsLength())
	for i := 0; i < buf.PartsLength(); i++ {
		fbp := new(fbast.StringExpressionPart)
		if !buf.Parts(fbp, i) {
			e.BaseNode.Errors = append(e.BaseNode.Errors,
				Error{fmt.Sprintf("Encountered error in deserializing StringExpression.Parts[%d]", i)})
		} else if fbp.TextValue() != nil {
			e.Parts[i] = TextPart{}.FromBuf(fbp)
		} else {
			e.Parts[i] = InterpolatedPart{}.FromBuf(fbp)
		}
	}
	return &e
}

type StringExpressionPart interface {
	Node
	stringPart()
}

func (*TextPart) stringPart()         {}
func (*InterpolatedPart) stringPart() {}

type TextPart struct {
	BaseNode
	Value string `json:"value"`
}

func (*TextPart) Type() string { return "TextPart" }

func (p *TextPart) Copy() Node {
	if p == nil {
		return p
	}
	np := new(TextPart)
	*np = *p
	np.BaseNode = p.BaseNode.Copy()
	return np
}

func (p TextPart) FromBuf(buf *fbast.StringExpressionPart) *TextPart {
	p.BaseNode.FromBuf(buf.BaseNode(nil))
	p.Value = string(buf.TextValue())
	return &p
}

type InterpolatedPart struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (*InterpolatedPart) Type() string { return "InterpolatedPart" }

func (p *InterpolatedPart) Copy() Node {
	if p == nil {
		return p
	}
	np := new(InterpolatedPart)
	*np = *p
	np.BaseNode = p.BaseNode.Copy()

	if p.Expression != nil {
		np.Expression = p.Expression.Copy().(Expression)
	}
	return np
}

func (p InterpolatedPart) FromBuf(buf *fbast.StringExpressionPart) *InterpolatedPart {
	p.BaseNode.FromBuf(buf.BaseNode(nil))
	p.Expression = exprFromBuf("InterpolatedPart.Expression", p.BaseNode,
		buf.InterpolatedExpression, buf.InterpolatedExpressionType())
	return &p
}

// ParenExpression represents an expressions that is wrapped in parentheses in the source code.
// It has no semantic meaning, rather it only communicates information about the syntax of the source code.
type ParenExpression struct {
	BaseNode
	Expression Expression `json:"expression"`
}

func (*ParenExpression) Type() string { return "ParenExpression" }

func (e *ParenExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(ParenExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if e.Expression != nil {
		ne.Expression = e.Expression.Copy().(Expression)
	}
	return ne
}

func (e ParenExpression) FromBuf(buf *fbast.ParenExpression) *ParenExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.Expression = exprFromBuf("ParenExpression.Expression", e.BaseNode, buf.Expression, buf.ExpressionType())
	return &e
}

// CallExpression represents a function call
type CallExpression struct {
	BaseNode
	Callee    Expression   `json:"callee"`
	Arguments []Expression `json:"arguments,omitempty"`
}

// Type is the abstract type
func (*CallExpression) Type() string { return "CallExpression" }

func (e *CallExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(CallExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if e.Callee != nil {
		ne.Callee = e.Callee.Copy().(Expression)
	}

	if len(e.Arguments) > 0 {
		ne.Arguments = make([]Expression, len(e.Arguments))
		for i, arg := range e.Arguments {
			ne.Arguments[i] = arg.Copy().(Expression)
		}
	}

	return ne
}

func (e CallExpression) FromBuf(buf *fbast.CallExpression) *CallExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.Callee = exprFromBuf("CallExpression.Callee", e.BaseNode, buf.Callee, buf.CalleeType())
	e.Arguments = make([]Expression, 0)
	arg := buf.Arguments(nil)
	if arg != nil {
		e.Arguments = []Expression{ObjectExpression{}.FromBuf(arg)}
	}
	return &e
}

type PipeExpression struct {
	BaseNode
	Argument Expression      `json:"argument"`
	Call     *CallExpression `json:"call"`
}

// Type is the abstract type
func (*PipeExpression) Type() string { return "PipeExpression" }

func (e *PipeExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(PipeExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if e.Argument != nil {
		ne.Argument = e.Argument.Copy().(Expression)
	}
	ne.Call = e.Call.Copy().(*CallExpression)

	return ne
}

func (e PipeExpression) FromBuf(buf *fbast.PipeExpression) *PipeExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.Argument = exprFromBuf("PipeExpression.Argument", e.BaseNode, buf.Argument, buf.ArgumentType())
	e.Call = CallExpression{}.FromBuf(buf.Call(nil))
	return &e
}

// MemberExpression represents calling a property of a CallExpression
type MemberExpression struct {
	BaseNode
	Object   Expression  `json:"object"`
	Property PropertyKey `json:"property"`
}

// Type is the abstract type
func (*MemberExpression) Type() string { return "MemberExpression" }

func (e *MemberExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(MemberExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if e.Object != nil {
		ne.Object = e.Object.Copy().(Expression)
	}
	if e.Property != nil {
		ne.Property = e.Property.Copy().(PropertyKey)
	}

	return ne
}

func (e MemberExpression) FromBuf(buf *fbast.MemberExpression) *MemberExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.Object = exprFromBuf("MemberExpression.Object", e.BaseNode, buf.Object, buf.ObjectType())
	e.Property = propertyKeyFromBuf("MemberExpression.Property", e.BaseNode, buf.Property, buf.PropertyType())
	return &e
}

// IndexExpression represents indexing into an array
type IndexExpression struct {
	BaseNode
	Array Expression `json:"array"`
	Index Expression `json:"index"`
}

func (*IndexExpression) Type() string { return "IndexExpression" }

func (e *IndexExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(IndexExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if e.Array != nil {
		ne.Array = e.Array.Copy().(Expression)
	}
	if e.Index != nil {
		ne.Index = e.Index.Copy().(Expression)
	}
	return ne
}

func (e IndexExpression) FromBuf(buf *fbast.IndexExpression) *IndexExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.Array = exprFromBuf("IndexExpression.Array", e.BaseNode, buf.Array, buf.ArrayType())
	e.Index = exprFromBuf("IndexExpression.Index", e.BaseNode, buf.Index, buf.IndexType())
	return &e
}

type FunctionExpression struct {
	BaseNode
	Params []*Property `json:"params"`
	Body   Node        `json:"body"`
}

// Type is the abstract type
func (*FunctionExpression) Type() string { return "FunctionExpression" }

func (e *FunctionExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(FunctionExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if len(e.Params) > 0 {
		ne.Params = make([]*Property, len(e.Params))
		for i, param := range e.Params {
			ne.Params[i] = param.Copy().(*Property)
		}
	}

	if e.Body != nil {
		ne.Body = e.Body.Copy()
	}

	return ne
}

func (e FunctionExpression) FromBuf(buf *fbast.FunctionExpression) *FunctionExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.Params = make([]*Property, buf.ParamsLength())
	for i := 0; i < buf.ParamsLength(); i++ {
		fbp := new(fbast.Property)
		if !buf.Params(fbp, i) {
			e.BaseNode.Errors = append(e.BaseNode.Errors,
				Error{fmt.Sprintf("Encountered error in deserializing FunctionExpression.Params[%d]", i)})
		} else {
			e.Params[i] = Property{}.FromBuf(fbp)
		}
	}
	t := new(flatbuffers.Table)
	if !buf.Body(t) {
		e.BaseNode.Errors = append(e.BaseNode.Errors,
			Error{"Encountered error in deserializing FunctionExpression.Body"})
	} else {
		switch buf.BodyType() {
		case fbast.ExpressionOrBlockBlock:
			b := new(fbast.Block)
			b.Init(t.Bytes, t.Pos)
			e.Body = Block{}.FromBuf(b)
		case fbast.ExpressionOrBlockWrappedExpression:
			we := new(fbast.WrappedExpression)
			we.Init(t.Bytes, t.Pos)
			e.Body = exprFromBuf("FunctionExpression.Body", e.BaseNode, we.Expr, we.ExprType())
		default:
			e.BaseNode.Errors = append(e.BaseNode.Errors,
				Error{"Encountered error in deserializing FunctionExpression.Body"})
		}
	}
	return &e
}

// OperatorKind are Equality and Arithmatic operators.
// Result of evaluating an equality operator is always of type Boolean based on whether the
// comparison is true
// Arithmetic operators take numerical values (either literals or variables) as their operands
//  and return a single numerical value.
type OperatorKind int

const (
	opBegin OperatorKind = iota
	MultiplicationOperator
	DivisionOperator
	ModuloOperator
	PowerOperator
	AdditionOperator
	SubtractionOperator
	LessThanEqualOperator
	LessThanOperator
	GreaterThanEqualOperator
	GreaterThanOperator
	StartsWithOperator
	InOperator
	NotOperator
	ExistsOperator
	NotEmptyOperator
	EmptyOperator
	EqualOperator
	NotEqualOperator
	RegexpMatchOperator
	NotRegexpMatchOperator
	opEnd
)

func (o OperatorKind) String() string {
	return OperatorTokens[o]
}

// OperatorLookup converts the operators to OperatorKind
func OperatorLookup(op string) OperatorKind {
	return operators[op]
}

func (o OperatorKind) MarshalText() ([]byte, error) {
	text, ok := OperatorTokens[o]
	if !ok {
		return nil, fmt.Errorf("unknown operator %d", int(o))
	}
	return []byte(text), nil
}
func (o *OperatorKind) UnmarshalText(data []byte) error {
	var ok bool
	*o, ok = operators[string(data)]
	if !ok {
		return fmt.Errorf("unknown operator %q", string(data))
	}
	return nil
}

// BinaryExpression use binary operators act on two operands in an expression.
// BinaryExpression includes relational and arithmatic operators
type BinaryExpression struct {
	BaseNode
	Operator OperatorKind `json:"operator"`
	Left     Expression   `json:"left"`
	Right    Expression   `json:"right"`
}

// Type is the abstract type
func (*BinaryExpression) Type() string { return "BinaryExpression" }

func (e *BinaryExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(BinaryExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if e.Left != nil {
		ne.Left = e.Left.Copy().(Expression)
	}
	if e.Right != nil {
		ne.Right = e.Right.Copy().(Expression)
	}

	return ne
}

func (e BinaryExpression) FromBuf(buf *fbast.BinaryExpression) *BinaryExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.Operator = OperatorLookup(fbast.EnumNamesOperator[buf.Operator()])
	e.Left = exprFromBuf("BinaryExpression.Left", e.BaseNode, buf.Left, buf.LeftType())
	e.Right = exprFromBuf("BinaryExpression.Right", e.BaseNode, buf.Right, buf.RightType())
	return &e
}

// UnaryExpression use operators act on a single operand in an expression.
type UnaryExpression struct {
	BaseNode
	Operator OperatorKind `json:"operator"`
	Argument Expression   `json:"argument"`
}

// Type is the abstract type
func (*UnaryExpression) Type() string { return "UnaryExpression" }

func (e *UnaryExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(UnaryExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if e.Argument != nil {
		ne.Argument = e.Argument.Copy().(Expression)
	}

	return ne
}

func (e UnaryExpression) FromBuf(buf *fbast.UnaryExpression) *UnaryExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.Operator = OperatorLookup(fbast.EnumNamesOperator[buf.Operator()])
	e.Argument = exprFromBuf("UnaryExpression.Argument", e.BaseNode, buf.Argument, buf.ArgumentType())
	return &e
}

// LogicalOperatorKind are used with boolean (logical) values
type LogicalOperatorKind int

const (
	logOpBegin LogicalOperatorKind = iota
	AndOperator
	OrOperator
	logOpEnd
)

func (o LogicalOperatorKind) String() string {
	return LogicalOperatorTokens[o]
}

// LogicalOperatorLookup converts the operators to LogicalOperatorKind
func LogicalOperatorLookup(op string) LogicalOperatorKind {
	return logOperators[op]
}

func (o LogicalOperatorKind) MarshalText() ([]byte, error) {
	text, ok := LogicalOperatorTokens[o]
	if !ok {
		return nil, fmt.Errorf("unknown logical operator %d", int(o))
	}
	return []byte(text), nil
}
func (o *LogicalOperatorKind) UnmarshalText(data []byte) error {
	var ok bool
	*o, ok = logOperators[string(data)]
	if !ok {
		return fmt.Errorf("unknown logical operator %q", string(data))
	}
	return nil
}

// LogicalExpression represent the rule conditions that collectively evaluate to either true or false.
// `or` expressions compute the disjunction of two boolean expressions and return boolean values.
// `and`` expressions compute the conjunction of two boolean expressions and return boolean values.
type LogicalExpression struct {
	BaseNode
	Operator LogicalOperatorKind `json:"operator"`
	Left     Expression          `json:"left"`
	Right    Expression          `json:"right"`
}

// Type is the abstract type
func (*LogicalExpression) Type() string { return "LogicalExpression" }

func (e *LogicalExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(LogicalExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if e.Left != nil {
		ne.Left = e.Left.Copy().(Expression)
	}
	if e.Right != nil {
		ne.Right = e.Right.Copy().(Expression)
	}

	return ne
}

func (e LogicalExpression) FromBuf(buf *fbast.LogicalExpression) *LogicalExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.Operator = LogicalOperatorLookup(fbast.EnumNamesLogicalOperator[buf.Operator()])
	e.Left = exprFromBuf("LogicalExpression.Left", e.BaseNode, buf.Left, buf.LeftType())
	e.Right = exprFromBuf("LogicalExpression.Right", e.BaseNode, buf.Right, buf.RightType())
	return &e
}

// ArrayExpression is used to create and directly specify the elements of an array object
type ArrayExpression struct {
	BaseNode
	Elements []Expression `json:"elements"`
}

// Type is the abstract type
func (*ArrayExpression) Type() string { return "ArrayExpression" }

func (e *ArrayExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(ArrayExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if len(e.Elements) > 0 {
		ne.Elements = make([]Expression, len(e.Elements))
		for i, el := range e.Elements {
			ne.Elements[i] = el.Copy().(Expression)
		}
	}

	return ne
}

func (e ArrayExpression) FromBuf(buf *fbast.ArrayExpression) *ArrayExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	var err []Error
	e.Elements, err = exprArrayFromBuf(buf.ElementsLength(), buf.Elements, "ArrayExpression.Elements")
	if len(err) > 0 {
		e.BaseNode.Errors = append(e.BaseNode.Errors, err...)
	}
	return &e
}

// ObjectExpression allows the declaration of an anonymous object within a declaration.
type ObjectExpression struct {
	BaseNode
	With       *Identifier `json:"with,omitempty"`
	Properties []*Property `json:"properties"`
}

// Type is the abstract type
func (*ObjectExpression) Type() string { return "ObjectExpression" }

func (e *ObjectExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(ObjectExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if len(e.Properties) > 0 {
		ne.Properties = make([]*Property, len(e.Properties))
		for i, p := range e.Properties {
			ne.Properties[i] = p.Copy().(*Property)
		}
	}

	return ne
}

func (e ObjectExpression) FromBuf(buf *fbast.ObjectExpression) *ObjectExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.With = Identifier{}.FromBuf(buf.With(nil))
	e.Properties = make([]*Property, buf.PropertiesLength())
	for i := 0; i < buf.PropertiesLength(); i++ {
		fbp := new(fbast.Property)
		if !buf.Properties(fbp, i) {
			e.BaseNode.Errors = append(e.BaseNode.Errors,
				Error{fmt.Sprintf("Encountered error in deserializing ObjectExpression.Properties[%d]", i)})
		} else {
			e.Properties[i] = Property{}.FromBuf(fbp)
		}
	}
	return &e
}

// ConditionalExpression selects one of two expressions, `Alternate` or `Consequent`
// depending on a third, boolean, expression, `Test`.
type ConditionalExpression struct {
	BaseNode
	Test       Expression `json:"test"`
	Consequent Expression `json:"consequent"`
	Alternate  Expression `json:"alternate"`
}

// Type is the abstract type
func (*ConditionalExpression) Type() string { return "ConditionalExpression" }

func (e *ConditionalExpression) Copy() Node {
	if e == nil {
		return e
	}
	ne := new(ConditionalExpression)
	*ne = *e
	ne.BaseNode = e.BaseNode.Copy()

	if e.Test != nil {
		ne.Test = e.Test.Copy().(Expression)
	}
	if e.Alternate != nil {
		ne.Alternate = e.Alternate.Copy().(Expression)
	}
	if e.Consequent != nil {
		ne.Consequent = e.Consequent.Copy().(Expression)
	}
	return ne
}

func (e ConditionalExpression) FromBuf(buf *fbast.ConditionalExpression) *ConditionalExpression {
	e.BaseNode.FromBuf(buf.BaseNode(nil))
	e.Test = exprFromBuf("ConditionalExpression.Test", e.BaseNode, buf.Test, buf.TestType())
	e.Consequent = exprFromBuf("ConditionalExpression.Consequent", e.BaseNode, buf.Consequent, buf.ConsequentType())
	e.Alternate = exprFromBuf("ConditionalExpression.Alternate", e.BaseNode, buf.Alternate, buf.AlternateType())
	return &e
}

// PropertyKey represents an object key
type PropertyKey interface {
	Node
	Key() string
}

// Property is the value associated with a key.
// A property's key can be either an identifier or string literal.
type Property struct {
	BaseNode
	Key   PropertyKey `json:"key"`
	Value Expression  `json:"value"`
}

func (p *Property) Copy() Node {
	if p == nil {
		return p
	}
	np := new(Property)
	*np = *p
	np.BaseNode = p.BaseNode.Copy()

	if p.Value != nil {
		np.Value = p.Value.Copy().(Expression)
	}

	return np
}

// Type is the abstract type
func (*Property) Type() string { return "Property" }

func (p Property) FromBuf(buf *fbast.Property) *Property {
	p.BaseNode.FromBuf(buf.BaseNode(nil))
	// deserialize key
	p.Key = propertyKeyFromBuf("Property.Key", p.BaseNode, buf.Key, buf.KeyType())
	// deserialize value
	p.Value = exprFromBuf("Property.Value", p.BaseNode, buf.Value, buf.ValueType())
	return &p
}

// Identifier represents a name that identifies a unique Node
type Identifier struct {
	BaseNode
	Name string `json:"name"`
}

// Identifiers are valid object keys
func (i *Identifier) Key() string {
	return i.Name
}

// Type is the abstract type
func (*Identifier) Type() string { return "Identifier" }

func (i *Identifier) Copy() Node {
	if i == nil {
		return i
	}
	ni := new(Identifier)
	*ni = *i
	ni.BaseNode = i.BaseNode.Copy()

	return ni
}

func (i Identifier) FromBuf(buf *fbast.Identifier) *Identifier {
	i.BaseNode.FromBuf(buf.BaseNode(nil))
	i.Name = string(buf.Name())
	return &i
}

// Literal is the lexical form for a literal expression which defines
// boolean, string, integer, number, duration, datetime or field values.
// Literals must be coerced explicitly.
type Literal interface {
	Expression
	literal()
}

func (*BooleanLiteral) literal()         {}
func (*DateTimeLiteral) literal()        {}
func (*DurationLiteral) literal()        {}
func (*FloatLiteral) literal()           {}
func (*IntegerLiteral) literal()         {}
func (*PipeLiteral) literal()            {}
func (*RegexpLiteral) literal()          {}
func (*StringLiteral) literal()          {}
func (*UnsignedIntegerLiteral) literal() {}

// PipeLiteral represents an specialized literal value, indicating the left hand value of a pipe expression.
type PipeLiteral struct {
	BaseNode
}

// Type is the abstract type
func (*PipeLiteral) Type() string { return "PipeLiteral" }

func (p *PipeLiteral) Copy() Node {
	if p == nil {
		return p
	}
	np := new(PipeLiteral)
	*np = *p
	np.BaseNode = p.BaseNode.Copy()
	return np
}

func (p PipeLiteral) FromBuf(buf *fbast.PipeLiteral) *PipeLiteral {
	p.BaseNode.FromBuf(buf.BaseNode(nil))
	return &p
}

// StringLiteral expressions begin and end with double quote marks.
type StringLiteral struct {
	BaseNode
	// Value is the unescaped value of the string literal
	Value string `json:"value"`
}

// StringLiterals are valid object keys
func (l *StringLiteral) Key() string {
	return l.Value
}

func (*StringLiteral) Type() string { return "StringLiteral" }

func (l *StringLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(StringLiteral)
	*nl = *l
	nl.BaseNode = l.BaseNode.Copy()
	return nl
}

func (l StringLiteral) FromBuf(buf *fbast.StringLiteral) *StringLiteral {
	l.BaseNode.FromBuf(buf.BaseNode(nil))
	l.Value = string(buf.Value())
	return &l
}

// BooleanLiteral represent boolean values
type BooleanLiteral struct {
	BaseNode
	Value bool `json:"value"`
}

// Type is the abstract type
func (*BooleanLiteral) Type() string { return "BooleanLiteral" }

func (l *BooleanLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(BooleanLiteral)
	*nl = *l
	nl.BaseNode = l.BaseNode.Copy()
	return nl
}

func (l BooleanLiteral) FromBuf(buf *fbast.BooleanLiteral) *BooleanLiteral {
	l.BaseNode.FromBuf(buf.BaseNode(nil))
	l.Value = buf.Value()
	return &l
}

// FloatLiteral  represent floating point numbers according to the double representations defined by the IEEE-754-1985
type FloatLiteral struct {
	BaseNode
	Value float64 `json:"value"`
}

// Type is the abstract type
func (*FloatLiteral) Type() string { return "FloatLiteral" }

func (l *FloatLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(FloatLiteral)
	*nl = *l
	nl.BaseNode = l.BaseNode.Copy()
	return nl
}

func (l FloatLiteral) FromBuf(buf *fbast.FloatLiteral) *FloatLiteral {
	l.BaseNode.FromBuf(buf.BaseNode(nil))
	l.Value = buf.Value()
	return &l
}

// IntegerLiteral represent integer numbers.
type IntegerLiteral struct {
	BaseNode
	Value int64 `json:"value"`
}

// Type is the abstract type
func (*IntegerLiteral) Type() string { return "IntegerLiteral" }

func (l *IntegerLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(IntegerLiteral)
	*nl = *l
	nl.BaseNode = l.BaseNode.Copy()
	return nl
}

func (l IntegerLiteral) FromBuf(buf *fbast.IntegerLiteral) *IntegerLiteral {
	l.BaseNode.FromBuf(buf.BaseNode(nil))
	l.Value = buf.Value()
	return &l
}

// UnsignedIntegerLiteral represent integer numbers.
type UnsignedIntegerLiteral struct {
	BaseNode
	Value uint64 `json:"value"`
}

// Type is the abstract type
func (*UnsignedIntegerLiteral) Type() string { return "UnsignedIntegerLiteral" }

func (l *UnsignedIntegerLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(UnsignedIntegerLiteral)
	*nl = *l
	nl.BaseNode = l.BaseNode.Copy()
	return nl
}

func (l UnsignedIntegerLiteral) FromBuf(buf *fbast.UnsignedIntegerLiteral) *UnsignedIntegerLiteral {
	l.BaseNode.FromBuf(buf.BaseNode(nil))
	l.Value = buf.Value()
	return &l
}

// RegexpLiteral expressions begin and end with `/` and are regular expressions with syntax accepted by RE2
type RegexpLiteral struct {
	BaseNode
	Value *regexp.Regexp `json:"value"`
}

// Type is the abstract type
func (*RegexpLiteral) Type() string { return "RegexpLiteral" }

func (l *RegexpLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(RegexpLiteral)
	*nl = *l
	nl.BaseNode = l.BaseNode.Copy()

	if l.Value != nil {
		nl.Value = l.Value
	}
	return nl
}

func (l RegexpLiteral) FromBuf(buf *fbast.RegexpLiteral) *RegexpLiteral {
	l.BaseNode.FromBuf(buf.BaseNode(nil))
	var err error
	if l.Value, err = parser.ParseRegexp(string(buf.Value())); err != nil {
		l.BaseNode.Errors = append(l.BaseNode.Errors, Error{err.Error()})
	}
	return &l
}

// Duration is a pair consisting of length of time and the unit of time measured.
// It is the atomic unit from which all duration literals are composed.
type Duration struct {
	Magnitude int64  `json:"magnitude"`
	Unit      string `json:"unit"`
}

// toDuration returns a time.Duration corresponding to Duration.  It is an approximation, as months, etc
// can't be properly figured out without knowing the time from when.
// This may have to be modified to also accept a time.Time to make this exact.
func toDuration(l Duration) (time.Duration, error) {
	// TODO: This is temporary code until we have proper duration type that takes different months, DST, etc into account
	var dur time.Duration
	var err error
	mag := l.Magnitude
	unit := l.Unit

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

func (d Duration) FromBuf(buf *fbast.Duration) Duration {
	d.Magnitude = buf.Magnitude()
	d.Unit = fbast.EnumNamesTimeUnit[buf.Unit()]
	return d
}

const (
	NanosecondUnit  = "ns"
	MicrosecondUnit = "us"
	MillisecondUnit = "ms"
	SecondUnit      = "s"
	MinuteUnit      = "m"
	HourUnit        = "h"
	DayUnit         = "d"
	WeekUnit        = "w"
	MonthUnit       = "mo"
	YearUnit        = "y"
)

// DurationLiteral represents the elapsed time between two instants as an
// int64 nanosecond count with syntax of golang's time.Duration
// TODO: this may be better as a class initialization
type DurationLiteral struct {
	BaseNode
	Values []Duration `json:"values"`
}

// Type is the abstract type
func (*DurationLiteral) Type() string { return "DurationLiteral" }

func (l *DurationLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(DurationLiteral)
	*nl = *l
	nl.BaseNode = l.BaseNode.Copy()

	if len(l.Values) > 0 {
		nl.Values = make([]Duration, len(l.Values))
		copy(nl.Values, l.Values)
	}
	return nl
}

func (l DurationLiteral) FromBuf(buf *fbast.DurationLiteral) *DurationLiteral {
	l.BaseNode.FromBuf(buf.BaseNode(nil))
	l.Values = make([]Duration, buf.ValuesLength())
	for i := 0; i < buf.ValuesLength(); i++ {
		d := new(fbast.Duration)
		if !buf.Values(d, i) {
			l.BaseNode.Errors = append(l.BaseNode.Errors,
				Error{fmt.Sprintf("Encountered error in deserializing DurationLiteral.Values[%d]", i)})
		} else {
			l.Values[i] = Duration{}.FromBuf(d)
		}
	}
	return &l
}

// Duration gives you a DurationLiteral from a time.Duration.
// Currently this is an approximation, but since we accept time, it can be made exact.
// TODO: makes this exact and not an approximation.
// currently the time.Time is ignored
func DurationFrom(l *DurationLiteral, _ time.Time) (time.Duration, error) {
	var d time.Duration
	for i := range l.Values {
		tempD, err := toDuration(l.Values[i])
		if err != nil {
			return 0, err
		}
		d += tempD
	}
	return d, nil
}

// TODO: we need a "duration from" that takes a time and a durationliteral, and gives an exact time.Duration instead of an approximation

// DateTimeLiteral represents an instant in time with nanosecond precision using
// the syntax of golang's RFC3339 Nanosecond variant
// TODO: this may be better as a class initialization
type DateTimeLiteral struct {
	BaseNode
	Value time.Time `json:"value"`
}

// Type is the abstract type
func (*DateTimeLiteral) Type() string { return "DateTimeLiteral" }

func (l *DateTimeLiteral) Copy() Node {
	if l == nil {
		return l
	}
	nl := new(DateTimeLiteral)
	*nl = *l
	nl.BaseNode = l.BaseNode.Copy()
	return nl
}

func (l DateTimeLiteral) FromBuf(buf *fbast.DateTimeLiteral) *DateTimeLiteral {
	l.BaseNode.FromBuf(buf.BaseNode(nil))
	var err error
	if l.Value, err = parser.ParseTime(string(buf.Value())); err != nil {
		l.BaseNode.Errors = append(l.BaseNode.Errors, Error{err.Error()})
	}
	return &l
}

// OperatorTokens converts OperatorKind to string
var OperatorTokens = map[OperatorKind]string{
	MultiplicationOperator:   "*",
	DivisionOperator:         "/",
	ModuloOperator:           "%",
	PowerOperator:            "^",
	AdditionOperator:         "+",
	SubtractionOperator:      "-",
	LessThanEqualOperator:    "<=",
	LessThanOperator:         "<",
	GreaterThanOperator:      ">",
	GreaterThanEqualOperator: ">=",
	InOperator:               "in",
	NotOperator:              "not",
	ExistsOperator:           "exists",
	NotEmptyOperator:         "not empty",
	EmptyOperator:            "empty",
	StartsWithOperator:       "startswith",
	EqualOperator:            "==",
	NotEqualOperator:         "!=",
	RegexpMatchOperator:      "=~",
	NotRegexpMatchOperator:   "!~",
}

// LogicalOperatorTokens converts LogicalOperatorKind to string
var LogicalOperatorTokens = map[LogicalOperatorKind]string{
	AndOperator: "and",
	OrOperator:  "or",
}

var operators map[string]OperatorKind
var logOperators map[string]LogicalOperatorKind

func init() {
	operators = make(map[string]OperatorKind)
	for op := opBegin + 1; op < opEnd; op++ {
		operators[OperatorTokens[op]] = op
	}

	logOperators = make(map[string]LogicalOperatorKind)
	for op := logOpBegin + 1; op < logOpEnd; op++ {
		logOperators[LogicalOperatorTokens[op]] = op
	}
}
