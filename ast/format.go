package ast

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Returns a valid query for a given AST rooted at node `n`.
func Format(n Node) string {
	f := &formatter{new(strings.Builder)}
	f.formatNode(n)
	return f.get()
}

type formatter struct {
	*strings.Builder
}

func (f *formatter) get() string {
	return f.String()
}

// strings.Builder's methods never returns a non-nil error.
func (f *formatter) writeString(s string) {
	_, err := f.WriteString(s)
	if err != nil {
		panic(err)
	}
}

func (f *formatter) writeRune(r rune) {
	_, err := f.WriteRune(r)
	if err != nil {
		panic(err)
	}
}

func (f *formatter) formatProgram(n *Program) {
	sep := '\n'
	for i, c := range n.Body {
		f.formatNode(c)
		if i < len(n.Body)-1 {
			f.writeRune(sep)
		}
	}
}

func (f *formatter) formatBlockStatement(n *BlockStatement) {
	sep := '\n'
	for i, c := range n.Body {
		f.formatNode(c)
		if i < len(n.Body)-1 {
			f.writeRune(sep)
		}
	}
}

func (f *formatter) formatExpressionStatement(n *ExpressionStatement) {
	f.formatNode(n.Expression)
}

func (f *formatter) formatReturnStatement(n *ReturnStatement) {
	f.writeString("return ")
	f.formatNode(n.Argument)
}

func (f *formatter) formatOptionStatement(n *OptionStatement) {
	f.writeString("option ")
	f.formatNode(n.Declaration)
}

func (f *formatter) formatVariableDeclaration(n *VariableDeclaration) {
	sep := ' '
	for i, c := range n.Declarations {
		f.formatNode(c)
		if i < len(n.Declarations)-1 {
			f.writeRune(sep)
		}
	}
}

func (f *formatter) formatVariableDeclarator(n *VariableDeclarator) {
	f.formatNode(n.ID)
	f.writeRune('=')
	f.formatNode(n.Init)
}

func (f *formatter) formatArrayExpression(n *ArrayExpression) {
	f.writeRune('[')

	sep := ','
	for i, c := range n.Elements {
		f.formatNode(c)
		if i < len(n.Elements)-1 {
			f.writeRune(sep)
		}
	}

	f.writeRune(']')
}

func (f *formatter) formatArrowFunctionExpression(n *ArrowFunctionExpression) {
	f.writeRune('(')

	sep := ','
	for i, c := range n.Params {
		// treat properties differently than in general case
		f.formatArrowFunctionArgument(c)
		if i < len(n.Params)-1 {
			f.writeRune(sep)
		}
	}

	f.writeString(")=>")
	f.formatNode(n.Body)
}

func (f *formatter) formatUnaryExpression(n *UnaryExpression) {
	f.writeString(n.Operator.String())
	f.formatNode(n.Argument)
}

func (f *formatter) formatBinaryExpression(n *BinaryExpression) {
	f.formatNode(n.Left)
	f.writeString(n.Operator.String())
	f.formatNode(n.Right)
}

func (f *formatter) formatLogicalExpression(n *LogicalExpression) {
	f.formatNode(n.Left)
	f.writeString(n.Operator.String())
	f.formatNode(n.Right)
}

func (f *formatter) formatCallExpression(n *CallExpression) {
	f.formatNode(n.Callee)
	f.writeRune('(')

	sep := ','
	for i, c := range n.Arguments {
		// treat ObjectExpression as argument in a special way
		// (an object as argument doesn't need braces)
		if oe, ok := c.(*ObjectExpression); ok {
			f.formatObjectExpressionAsFunctionArgument(oe)
		} else {
			f.formatNode(c)
		}

		if i < len(n.Arguments)-1 {
			f.writeRune(sep)
		}
	}

	f.writeRune(')')
}

func (f *formatter) formatPipeExpression(n *PipeExpression) {
	f.formatNode(n.Argument)
	f.writeString("|>")
	f.formatNode(n.Call)
}

func (f *formatter) formatConditionalExpression(n *ConditionalExpression) {
	f.formatNode(n.Test)
	f.writeRune('?')
	f.formatNode(n.Consequent)
	f.writeRune(':')
	f.formatNode(n.Alternate)
}

func (f *formatter) formatMemberExpression(n *MemberExpression) {
	f.formatNode(n.Object)
	f.writeRune('.')
	f.formatNode(n.Property)
}

func (f *formatter) formatIndexExpression(n *IndexExpression) {
	f.formatNode(n.Array)
	f.writeRune('[')
	f.formatNode(n.Index)
	f.writeRune(']')
}

func (f *formatter) formatObjectExpression(n *ObjectExpression) {
	f.formatObjectExpressionBraces(n, true)
}

func (f *formatter) formatObjectExpressionAsFunctionArgument(n *ObjectExpression) {
	f.formatObjectExpressionBraces(n, false)
}

func (f *formatter) formatObjectExpressionBraces(n *ObjectExpression, braces bool) {
	if braces {
		f.writeRune('{')
	}

	sep := ','
	for i, c := range n.Properties {
		f.formatNode(c)
		if i < len(n.Properties)-1 {
			f.writeRune(sep)
		}
	}

	if braces {
		f.writeRune('}')
	}
}

func (f *formatter) formatProperty(n *Property) {
	f.formatNode(n.Key)
	f.writeRune(':')
	f.formatNode(n.Value)
}

func (f *formatter) formatArrowFunctionArgument(n *Property) {
	// in this case we are not in a function declaration
	if n.Value == nil {
		f.formatNode(n.Key)
		return
	}

	// in a function declaration
	f.formatNode(n.Key)
	f.writeRune('=')
	f.formatNode(n.Value)
}

func (f *formatter) formatIdentifier(n *Identifier) {
	f.writeString(n.Name)
}

func (f *formatter) formatStringLiteral(n *StringLiteral) {
	f.writeRune('"')
	f.writeString(n.Value)
	f.writeRune('"')
}

func (f *formatter) formatBooleanLiteral(n *BooleanLiteral) {
	f.writeString(strconv.FormatBool(n.Value))
}

func (f *formatter) formatDateTimeLiteral(n *DateTimeLiteral) {
	f.writeString(n.Value.Format(time.RFC3339Nano))
}

func (f *formatter) formatDurationLiteral(n *DurationLiteral) {
	formatDuration := func(d Duration) {
		f.writeString(strconv.FormatInt(d.Magnitude, 10))
		f.writeString(d.Unit)
	}

	sep := ' '
	for i, d := range n.Values {
		formatDuration(d)
		if i < len(n.Values)-1 {
			f.writeRune(sep)
		}
	}
}

func (f *formatter) formatFloatLiteral(n *FloatLiteral) {
	sf := strconv.FormatFloat(n.Value, 'f', -1, 64)

	if !strings.Contains(sf, ".") {
		sf += ".0" // force to make it a float
	}

	f.writeString(sf)
}

func (f *formatter) formatIntegerLiteral(n *IntegerLiteral) {
	f.writeString(strconv.FormatInt(n.Value, 10))
}

func (f *formatter) formatUnsignedIntegerLiteral(n *UnsignedIntegerLiteral) {
	f.writeString(strconv.FormatUint(n.Value, 10))
}

func (f *formatter) formatPipeLiteral(_ *PipeLiteral) {
	f.writeString("<-")
}

func (f *formatter) formatRegexpLiteral(n *RegexpLiteral) {
	f.writeRune('/')
	f.writeString(n.Value.String())
	f.writeRune('/')
}

func (f *formatter) formatNode(n Node) {
	switch n := n.(type) {
	case *Program:
		f.formatProgram(n)
	case *BlockStatement:
		f.formatBlockStatement(n)
	case *OptionStatement:
		f.formatOptionStatement(n)
	case *ExpressionStatement:
		f.formatExpressionStatement(n)
	case *ReturnStatement:
		f.formatReturnStatement(n)
	case *VariableDeclaration:
		f.formatVariableDeclaration(n)
	case *VariableDeclarator:
		f.formatVariableDeclarator(n)
	case *CallExpression:
		f.formatCallExpression(n)
	case *PipeExpression:
		f.formatPipeExpression(n)
	case *MemberExpression:
		f.formatMemberExpression(n)
	case *IndexExpression:
		f.formatIndexExpression(n)
	case *BinaryExpression:
		f.formatBinaryExpression(n)
	case *UnaryExpression:
		f.formatUnaryExpression(n)
	case *LogicalExpression:
		f.formatLogicalExpression(n)
	case *ObjectExpression:
		f.formatObjectExpression(n)
	case *ConditionalExpression:
		f.formatConditionalExpression(n)
	case *ArrayExpression:
		f.formatArrayExpression(n)
	case *Identifier:
		f.formatIdentifier(n)
	case *PipeLiteral:
		f.formatPipeLiteral(n)
	case *StringLiteral:
		f.formatStringLiteral(n)
	case *BooleanLiteral:
		f.formatBooleanLiteral(n)
	case *FloatLiteral:
		f.formatFloatLiteral(n)
	case *IntegerLiteral:
		f.formatIntegerLiteral(n)
	case *UnsignedIntegerLiteral:
		f.formatUnsignedIntegerLiteral(n)
	case *RegexpLiteral:
		f.formatRegexpLiteral(n)
	case *DurationLiteral:
		f.formatDurationLiteral(n)
	case *DateTimeLiteral:
		f.formatDateTimeLiteral(n)
	case *ArrowFunctionExpression:
		f.formatArrowFunctionExpression(n)
	case *Property:
		f.formatProperty(n)
	default:
		// If we were able not to find the type, than this switch is wrong
		panic(fmt.Errorf("unknown type %q", n.Type()))
	}
}
