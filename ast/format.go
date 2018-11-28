package ast

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

//Produces a valid string query from a AST. Use:
func Format(node Node) string {
	fmtV := NewFormatVisitor()
	Walk(fmtV, node)
	return fmtV.Get()
}

// Delimiters express the pieces of string that must be appended by leaf nodes once the visiting recursion is over.
// As such, they are pushed down to leaf visitors that are responsible for writing the output.
type delimiters struct {
	l *strings.Builder
	r *strings.Builder
}

func emptyDelimiters() *delimiters {
	return &delimiters{new(strings.Builder), new(strings.Builder)}
}

func (d *delimiters) extendL(left string) *delimiters {
	d.l.WriteString(left)
	return d
}

func (d *delimiters) extendR(right string) *delimiters {
	d.r.WriteString(reverse(right))
	return d
}

func (d *delimiters) resetR() *delimiters {
	return &delimiters{l: d.l, r: new(strings.Builder)}
}

func (d *delimiters) resetL() *delimiters {
	return &delimiters{l: new(strings.Builder), r: d.r}
}

func (d *delimiters) getL() string {
	return d.l.String()
}

func (d *delimiters) getR() string {
	return reverse(d.r.String())
}

func reverse(s string) string {
	if utf8.RuneCountInString(s) == 1 {
		return s
	}

	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

type FormatVisitor struct {
	sb *strings.Builder // used to accumulate the resulting query
}

func NewFormatVisitor() *FormatVisitor {
	return &FormatVisitor{sb: new(strings.Builder)}
}

func (fv *FormatVisitor) Visit(node Node) Visitor {
	return fv.createVisitor(node, emptyDelimiters())
}

func (fv *FormatVisitor) Done(node Node) {}

func (fv *FormatVisitor) Get() string {
	return fv.sb.String()
}

func (fv *FormatVisitor) write(s string) {
	fv.sb.WriteString(s)
}

/*
A node is visited by its parent's visitor.
For example, the Key of a property is visited by a propertyVisitor.
This means it is useless to specify visitors for leaf nodes given that they have no children,
and so their Visit method won't be called.
That's why leaf nodes shouldn't return a visitor, but directly write to the builder.
*/
func (fv *FormatVisitor) createVisitor(child Node, d *delimiters) Visitor {
	var v Visitor
	switch n := child.(type) {
	case *Program:
		v = &programVisitor{FormatVisitor: fv, node: n, dels: d}
	case *BlockStatement:
		v = &blockStatementVisitor{FormatVisitor: fv, node: n, dels: d}
	case *OptionStatement:
		v = &optionStatementVisitor{FormatVisitor: fv, node: n, dels: d}
	case *ExpressionStatement:
		v = &expressionStatementVisitor{FormatVisitor: fv, node: n, dels: d}
	case *ReturnStatement:
		v = &returnStatementVisitor{FormatVisitor: fv, node: n, dels: d}
	case *VariableDeclaration:
		v = &variableDeclarationVisitor{FormatVisitor: fv, node: n, dels: d}
	case *VariableDeclarator:
		v = &variableDeclaratorVisitor{FormatVisitor: fv, node: n, dels: d}
	case *CallExpression:
		v = &callExpressionVisitor{FormatVisitor: fv, node: n, dels: d}
	case *PipeExpression:
		v = &pipeExpressionVisitor{FormatVisitor: fv, node: n, dels: d}
	case *MemberExpression:
		v = &memberExpressionVisitor{FormatVisitor: fv, node: n, dels: d}
	case *IndexExpression:
		v = &indexExpressionVisitor{FormatVisitor: fv, node: n, dels: d}
	case *BinaryExpression:
		v = &binaryExpressionVisitor{FormatVisitor: fv, node: n, dels: d}
	case *LogicalExpression:
		v = &logicalExpressionVisitor{FormatVisitor: fv, node: n, dels: d}
	case *UnaryExpression:
		v = &unaryExpressionVisitor{FormatVisitor: fv, node: n, dels: d}
	case *ObjectExpression:
		v = &objectExpressionVisitor{FormatVisitor: fv, node: n, dels: d}
	case *ConditionalExpression:
		v = &conditionalExpressionVisitor{FormatVisitor: fv, node: n, dels: d}
	case *ArrayExpression:
		v = &arrayExpressionVisitor{FormatVisitor: fv, node: n, dels: d}
	case *ArrowFunctionExpression:
		v = &arrowFunctionExpressionVisitor{FormatVisitor: fv, node: n, dels: d}
	case *Property:
		v = &propertyVisitor{FormatVisitor: fv, node: n, dels: d}

		// ---- leaf nodes
	case *Identifier:
		writeIdentifier(n, d, fv)
	case *PipeLiteral:
		writePipeLiteral(n, d, fv)
	case *StringLiteral:
		writeStringLiteral(n, d, fv)
	case *BooleanLiteral:
		writeBooleanLiteral(n, d, fv)
	case *FloatLiteral:
		writeFloatLiteral(n, d, fv)
	case *IntegerLiteral:
		writeIntegerLiteral(n, d, fv)
	case *UnsignedIntegerLiteral:
		writeUnsignedIntegerLiteral(n, d, fv)
	case *RegexpLiteral:
		writeRegexpLiteral(n, d, fv)
	case *DurationLiteral:
		writeDurationLiteral(n, d, fv)
	case *DateTimeLiteral:
		writeDateTimeLiteral(n, d, fv)
	default:
		// If we were able not to find the type, than this switch is wrong
		panic(fmt.Errorf("unknown type %q", n.Type()))
	}

	return v
}

/*
This utility function returns proper delimiters for a child from a slice of children.
It calculates the delimiters from the current delimiters, a prefix, a suffix, and a separator, by
considering the position of the child among children.
*/
func getDelimitersForSlice(children interface{}, child interface{}, cur *delimiters, pref, suf, sep string) *delimiters {
	s := reflect.ValueOf(children)
	if s.Kind() != reflect.Slice {
		panic("children must be a slice type")
	}

	l := s.Len()

	// don't worry about empty slices, they won't be visited

	first := s.Index(0).Interface()
	last := s.Index(l - 1).Interface()

	if first == last {
		return cur.extendL(pref).extendR(suf)
	}

	var dels *delimiters

	if child == first {
		dels = cur.resetR().extendL(pref).extendR(sep)
	} else if child == last {
		dels = cur.resetL().extendR(suf)
	} else {
		dels = emptyDelimiters().extendR(sep)
	}

	return dels
}

/*
Every type below corresponds to a Node type in the AST.

A visitor visits its children, and not the parent node. So, for example,
the `arrowFunctionVisitor`'s `Visit` method is called once for every child of an `ArrowFunctionExpressions`,
and so for every piece of argument and for the body.
The visitor returns new specific visitors for every children and pushes down delimiters, in order to make
the leaf nodes write them to the global string builder at the right moment.
*/

type programVisitor struct {
	*FormatVisitor
	node *Program
	dels *delimiters
}

func (v *programVisitor) Visit(node Node) Visitor {
	dels := getDelimitersForSlice(v.node.Body, node, v.dels, "", "\n", "\n")
	return v.createVisitor(node, dels)
}

type blockStatementVisitor struct {
	*FormatVisitor
	node *BlockStatement
	dels *delimiters
}

func (v *blockStatementVisitor) Visit(node Node) Visitor {
	dels := getDelimitersForSlice(v.node.Body, node, v.dels, "", "", "\n")
	return v.createVisitor(node, dels)
}

type optionStatementVisitor struct {
	*FormatVisitor
	node *OptionStatement
	dels *delimiters
}

func (v *optionStatementVisitor) Visit(node Node) Visitor {
	return v.createVisitor(node, v.dels.extendL("option "))
}

type expressionStatementVisitor struct {
	*FormatVisitor
	node *ExpressionStatement
	dels *delimiters
}

func (v *expressionStatementVisitor) Visit(node Node) Visitor {
	return v.createVisitor(node, v.dels)
}

type returnStatementVisitor struct {
	*FormatVisitor
	node *ReturnStatement
	dels *delimiters
}

func (v *returnStatementVisitor) Visit(node Node) Visitor {
	return v.createVisitor(node, v.dels.extendL("return "))
}

type variableDeclarationVisitor struct {
	*FormatVisitor
	node *VariableDeclaration
	dels *delimiters
}

func (v *variableDeclarationVisitor) Visit(node Node) Visitor {
	dels := getDelimitersForSlice(v.node.Declarations, node, v.dels, "", "", "\n")
	return v.createVisitor(node, dels)
}

type variableDeclaratorVisitor struct {
	*FormatVisitor
	node *VariableDeclarator
	dels *delimiters
}

func (v *variableDeclaratorVisitor) Visit(node Node) Visitor {
	var dels *delimiters

	if node == v.node.ID {
		dels = v.dels.resetR().extendR("=")
	} else {
		dels = v.dels.resetL()
	}

	return v.createVisitor(node, dels)
}

type callExpressionVisitor struct {
	*FormatVisitor
	node *CallExpression
	dels *delimiters
}

func (v *callExpressionVisitor) Visit(node Node) Visitor {
	args := v.node.Arguments
	var dels *delimiters

	if node == v.node.Callee {
		// if there are no arguments, `Visit` won't be invoked by `Walk`,
		// so, we have to account for that.
		if len(args) > 0 {
			dels = v.dels.resetR()
		} else {
			dels = v.dels.extendR("()")
		}
	} else {
		dels = v.dels.resetL()
		dels = getDelimitersForSlice(args, node, dels, "(", ")", ",")

		// if the argument is an object expression, we must skip braces
		// when encoding it, so we have to create the visitor manually
		if node, ok := node.(*ObjectExpression); ok {
			return &objectExpressionVisitor{FormatVisitor: v.FormatVisitor, node: node, dels: dels, skipEnclosing: true}
		}
	}

	return v.createVisitor(node, dels)
}

type pipeExpressionVisitor struct {
	*FormatVisitor
	node *PipeExpression
	dels *delimiters
}

func (v *pipeExpressionVisitor) Visit(node Node) Visitor {
	var dels *delimiters

	if node == v.node.Argument {
		dels = v.dels.resetR().extendR("|>")
	} else {
		dels = v.dels.resetL()
	}

	return v.createVisitor(node, dels)
}

type memberExpressionVisitor struct {
	*FormatVisitor
	node *MemberExpression
	dels *delimiters
}

func (v *memberExpressionVisitor) Visit(node Node) Visitor {
	var dels *delimiters

	if node == v.node.Object {
		dels = v.dels.resetR().extendR(".")
	} else {
		dels = v.dels.resetL()
	}

	return v.createVisitor(node, dels)
}

type indexExpressionVisitor struct {
	*FormatVisitor
	node *IndexExpression
	dels *delimiters
}

func (v *indexExpressionVisitor) Visit(node Node) Visitor {
	var dels *delimiters

	if node == v.node.Array {
		dels = v.dels.resetR().extendR("[")
	} else {
		dels = v.dels.resetL().extendR("]")
	}

	return v.createVisitor(node, dels)
}

type binaryExpressionVisitor struct {
	*FormatVisitor
	node *BinaryExpression
	dels *delimiters
}

func (v *binaryExpressionVisitor) Visit(node Node) Visitor {
	var dels *delimiters

	if node == v.node.Left {
		dels = v.dels.resetR().extendR(v.node.Operator.String())
	} else {
		dels = v.dels.resetL()
	}

	return v.createVisitor(node, dels)
}

type logicalExpressionVisitor struct {
	*FormatVisitor
	node *LogicalExpression
	dels *delimiters
}

func (v *logicalExpressionVisitor) Visit(node Node) Visitor {
	var dels *delimiters

	if node == v.node.Left {
		dels = v.dels.resetR().extendR(v.node.Operator.String())
	} else {
		dels = v.dels.resetL()
	}

	return v.createVisitor(node, dels)
}

type unaryExpressionVisitor struct {
	*FormatVisitor
	node *UnaryExpression
	dels *delimiters
}

func (v *unaryExpressionVisitor) Visit(node Node) Visitor {
	return v.createVisitor(node, v.dels.extendL(v.node.Operator.String()))
}

type objectExpressionVisitor struct {
	*FormatVisitor
	node          *ObjectExpression
	dels          *delimiters
	skipEnclosing bool
}

func (v *objectExpressionVisitor) Visit(node Node) Visitor {
	var p, s string
	if !v.skipEnclosing {
		p = "{"
		s = "}"
	}

	dels := getDelimitersForSlice(v.node.Properties, node, v.dels, p, s, ",")
	return v.createVisitor(node, dels)
}

type conditionalExpressionVisitor struct {
	*FormatVisitor
	node *ConditionalExpression
	dels *delimiters
}

func (v *conditionalExpressionVisitor) Visit(node Node) Visitor {
	var dels *delimiters

	if node == v.node.Test {
		dels = v.dels.resetR().extendR("?")
	} else if node == v.node.Consequent {
		dels = emptyDelimiters().extendR(":")
	} else {
		dels = v.dels.resetL()
	}

	return v.createVisitor(node, dels)
}

type arrayExpressionVisitor struct {
	*FormatVisitor
	node *ArrayExpression
	dels *delimiters
}

func (v *arrayExpressionVisitor) Visit(node Node) Visitor {
	dels := getDelimitersForSlice(v.node.Elements, node, v.dels, "[", "]", ",")
	return v.createVisitor(node, dels)
}

type arrowFunctionExpressionVisitor struct {
	*FormatVisitor
	node *ArrowFunctionExpression
	dels *delimiters
}

func (v *arrowFunctionExpressionVisitor) Visit(node Node) Visitor {
	var dels *delimiters

	if node != v.node.Body {
		dels = v.dels.resetR()
		dels = getDelimitersForSlice(v.node.Params, node, dels, "(", ")=>", ",")

		// if the argument is a property, we must use "=" as separator
		// when encoding it, so we have to create the visitor manually
		if node, ok := node.(*Property); ok {
			return &propertyVisitor{FormatVisitor: v.FormatVisitor, node: node, dels: dels, useEqual: true}
		}
	} else {
		// if there are no params, `Visit` won't be invoked by `Walk`,
		// so, we have to account for that.
		if len(v.node.Params) > 0 {
			dels = v.dels.resetL()
		} else {
			dels = v.dels.extendL("()=>")
		}
	}

	return v.createVisitor(node, dels)
}

type propertyVisitor struct {
	*FormatVisitor
	node     *Property
	dels     *delimiters
	useEqual bool
}

func (v *propertyVisitor) Visit(node Node) Visitor {
	var dels *delimiters

	sep := ":"
	if v.useEqual {
		sep = "="
	}

	if node == v.node.Key {
		if v.node.Value == nil {
			// keep same delimiters
			dels = v.dels
		} else {
			dels = v.dels.resetR().extendR(sep)
		}
	} else if node == v.node.Value {
		dels = v.dels.resetL()
	}

	return v.createVisitor(node, dels)
}

// --------- Write functions for leaf nodes

func writeStringLiteral(node *StringLiteral, dels *delimiters, fv *FormatVisitor) {
	fv.write(dels.getL())
	fv.write("\"")
	fv.write(node.Value)
	fv.write("\"")
	fv.write(dels.getR())
}

func writePipeLiteral(_ *PipeLiteral, dels *delimiters, fv *FormatVisitor) {
	fv.write(dels.getL())
	fv.write("<-")
	fv.write(dels.getR())
}

func writeBooleanLiteral(node *BooleanLiteral, dels *delimiters, fv *FormatVisitor) {
	fv.write(dels.getL())
	fv.write(strconv.FormatBool(node.Value))
	fv.write(dels.getR())
}

func writeFloatLiteral(node *FloatLiteral, dels *delimiters, fv *FormatVisitor) {
	fv.write(dels.getL())
	conv := strconv.FormatFloat(node.Value, 'f', -1, 64)

	if !strings.Contains(conv, ".") {
		conv += ".0" // force to make it a float
	}

	fv.write(conv)
	fv.write(dels.getR())
}

func writeIntegerLiteral(node *IntegerLiteral, dels *delimiters, fv *FormatVisitor) {
	fv.write(dels.getL())
	fv.write(strconv.FormatInt(node.Value, 10))
	fv.write(dels.getR())
}

func writeUnsignedIntegerLiteral(node *UnsignedIntegerLiteral, dels *delimiters, fv *FormatVisitor) {
	fv.write(dels.getL())
	fv.write(strconv.FormatUint(node.Value, 10))
	fv.write(dels.getR())
}

func writeRegexpLiteral(node *RegexpLiteral, dels *delimiters, fv *FormatVisitor) {
	fv.write(dels.getL())
	fv.write("/")
	fv.write(node.Value.String())
	fv.write("/")
	fv.write(dels.getR())
}

func writeDurationLiteral(node *DurationLiteral, dels *delimiters, fv *FormatVisitor) {
	fv.write(dels.getL())
	for _, d := range node.Values {
		fv.write(strconv.FormatInt(d.Magnitude, 10))
		fv.write(d.Unit)
	}
	fv.write(dels.getR())
}

func writeDateTimeLiteral(node *DateTimeLiteral, dels *delimiters, fv *FormatVisitor) {
	fv.write(dels.getL())
	fv.write(node.Value.Format(time.RFC3339Nano))
	fv.write(dels.getR())
}

func writeIdentifier(node *Identifier, dels *delimiters, fv *FormatVisitor) {
	fv.write(dels.getL())
	fv.write(node.Name)
	fv.write(dels.getR())
}
