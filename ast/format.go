package ast

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func Format(n Node) string {
	return formatNode(n)
}

func formatChildren(children interface{}, sep string) string {
	s := reflect.ValueOf(children)
	if s.Kind() != reflect.Slice {
		panic("children must be a slice type")
	}

	schildren := make([]string, s.Len())
	for i := 0; i < s.Len(); i++ {
		child := formatNode(s.Index(i).Interface().(Node))
		schildren[i] = child
	}

	return strings.Join(schildren, sep)
}

func formatProgram(n *Program) string {
	return formatChildren(n.Body, "\n")
}

func formatBlockStatement(n *BlockStatement) string {
	return formatChildren(n.Body, "\n")
}

func formatExpressionStatement(n *ExpressionStatement) string {
	return formatNode(n.Expression)
}

func formatReturnStatement(n *ReturnStatement) string {
	return formatNode(n.Argument)
}

func formatOptionStatement(n *OptionStatement) string {
	format := "option %s"
	decl := formatNode(n.Declaration)
	return fmt.Sprintf(format, decl)
}

func formatVariableDeclaration(n *VariableDeclaration) string {
	return formatChildren(n.Declarations, " ")
}

func formatVariableDeclarator(n *VariableDeclarator) string {
	format := "%s=%s"
	id := formatNode(n.ID)
	init := formatNode(n.Init)
	return fmt.Sprintf(format, id, init)
}

func formatArrayExpression(n *ArrayExpression) string {
	format := "[%s]"
	s := formatChildren(n.Elements, ",")
	return fmt.Sprintf(format, s)
}

func formatArrowFunctionExpression(n *ArrowFunctionExpression) string {
	format := "(%s)=>%s"

	// must treat properties differently than in general case
	// must specify the separator used in properties ("=" instead of ":")
	// cannot use formatChildren
	props := make([]string, len(n.Params))
	for i := 0; i < len(n.Params); i++ {
		child := formatPropertyWSeparator(n.Params[i], "=")
		props[i] = child
	}

	params := strings.Join(props, ",")
	body := formatNode(n.Body)
	return fmt.Sprintf(format, params, body)
}

func formatBinaryExpression(n *BinaryExpression) string {
	format := "%s%s%s"
	left := formatNode(n.Left)
	op := n.Operator.String()
	right := formatNode(n.Right)
	return fmt.Sprintf(format, left, op, right)
}

func formatCallExpression(n *CallExpression) string {
	format := "%s(%s)"
	callee := formatNode(n.Callee)
	args := formatChildren(n.Arguments, ",")

	// remove braces to arguments because it is a special
	// case for an ObjectExpression, if so
	l := len(args)
	if l > 1 && args[0] == '{' && args[l-1] == '}' {
		args = args[1 : l-1]
	}

	return fmt.Sprintf(format, callee, args)
}

func formatConditionalExpression(n *ConditionalExpression) string {
	format := "%s?%s:%s"
	test := formatNode(n.Test)
	cons := formatNode(n.Consequent)
	alt := formatNode(n.Alternate)
	return fmt.Sprintf(format, test, cons, alt)
}

func formatMemberExpression(n *MemberExpression) string {
	format := "%s.%s"
	o := formatNode(n.Object)
	p := formatNode(n.Property)
	return fmt.Sprintf(format, o, p)
}

func formatIndexExpression(n *IndexExpression) string {
	format := "%s[%s]"
	array := formatNode(n.Array)
	i := formatNode(n.Index)
	return fmt.Sprintf(format, array, i)
}

func formatObjectExpression(n *ObjectExpression) string {
	format := "{%s}"
	properties := formatChildren(n.Properties, ",")
	return fmt.Sprintf(format, properties)
}

func formatPipeExpression(n *PipeExpression) string {
	format := "%s|>%s"
	arg := formatNode(n.Argument)
	call := formatNode(n.Call)
	return fmt.Sprintf(format, arg, call)
}

func formatUnaryExpression(n *UnaryExpression) string {
	format := "%s%s"
	op := n.Operator.String()
	exp := formatNode(n.Argument)
	return fmt.Sprintf(format, op, exp)
}

func formatLogicalExpression(n *LogicalExpression) string {
	format := "%s%s%s"
	left := formatNode(n.Left)
	op := n.Operator.String()
	right := formatNode(n.Right)
	return fmt.Sprintf(format, left, op, right)
}

func formatIdentifier(n *Identifier) string {
	return n.Name
}

func formatBooleanLiteral(n *BooleanLiteral) string {
	return strconv.FormatBool(n.Value)
}

func formatDateTimeLiteral(n *DateTimeLiteral) string {
	return n.Value.Format(time.RFC3339Nano)
}

func formatDurationLiteral(n *DurationLiteral) string {
	formatDuration := func(d Duration) string {
		format := "%s%s"
		mag := strconv.FormatInt(d.Magnitude, 10)
		return fmt.Sprintf(format, mag, d.Unit)
	}

	ds := make([]string, len(n.Values))
	for _, d := range n.Values {
		child := formatDuration(d)
		ds = append(ds, child)
	}

	return strings.Join(ds, "")
}

func formatFloatLiteral(n *FloatLiteral) string {
	conv := strconv.FormatFloat(n.Value, 'f', -1, 64)

	if !strings.Contains(conv, ".") {
		conv += ".0" // force to make it a float
	}

	return conv
}

func formatIntegerLiteral(n *IntegerLiteral) string {
	return strconv.FormatInt(n.Value, 10)
}

func formatPipeLiteral(_ *PipeLiteral) string {
	return "<-"
}

func formatRegexpLiteral(n *RegexpLiteral) string {
	format := "/%s/"
	return fmt.Sprintf(format, n.Value.String())
}

func formatStringLiteral(n *StringLiteral) string {
	format := "\"%s\""
	return fmt.Sprintf(format, n.Value)
}

func formatUnsignedIntegerLiteral(n *UnsignedIntegerLiteral) string {
	return strconv.FormatUint(n.Value, 10)
}

func formatProperty(n *Property) string {
	return formatPropertyWSeparator(n, ":")
}

func formatPropertyWSeparator(n *Property, sep string) string {
	if n.Value == nil {
		return formatNode(n.Key)
	}

	format := "%s%s%s"
	k := formatNode(n.Key)
	v := formatNode(n.Value)
	return fmt.Sprintf(format, k, sep, v)
}

func formatNode(n Node) string {
	var result string
	switch n := n.(type) {
	case *Program:
		result = formatProgram(n)
	case *BlockStatement:
		result = formatBlockStatement(n)
	case *OptionStatement:
		result = formatOptionStatement(n)
	case *ExpressionStatement:
		result = formatExpressionStatement(n)
	case *ReturnStatement:
		result = formatReturnStatement(n)
	case *VariableDeclaration:
		result = formatVariableDeclaration(n)
	case *VariableDeclarator:
		result = formatVariableDeclarator(n)
	case *CallExpression:
		result = formatCallExpression(n)
	case *PipeExpression:
		result = formatPipeExpression(n)
	case *MemberExpression:
		result = formatMemberExpression(n)
	case *IndexExpression:
		result = formatIndexExpression(n)
	case *BinaryExpression:
		result = formatBinaryExpression(n)
	case *UnaryExpression:
		result = formatUnaryExpression(n)
	case *LogicalExpression:
		result = formatLogicalExpression(n)
	case *ObjectExpression:
		result = formatObjectExpression(n)
	case *ConditionalExpression:
		result = formatConditionalExpression(n)
	case *ArrayExpression:
		result = formatArrayExpression(n)
	case *Identifier:
		result = formatIdentifier(n)
	case *PipeLiteral:
		result = formatPipeLiteral(n)
	case *StringLiteral:
		result = formatStringLiteral(n)
	case *BooleanLiteral:
		result = formatBooleanLiteral(n)
	case *FloatLiteral:
		result = formatFloatLiteral(n)
	case *IntegerLiteral:
		result = formatIntegerLiteral(n)
	case *UnsignedIntegerLiteral:
		result = formatUnsignedIntegerLiteral(n)
	case *RegexpLiteral:
		result = formatRegexpLiteral(n)
	case *DurationLiteral:
		result = formatDurationLiteral(n)
	case *DateTimeLiteral:
		result = formatDateTimeLiteral(n)
	case *ArrowFunctionExpression:
		result = formatArrowFunctionExpression(n)
	case *Property:
		result = formatProperty(n)
	default:
		// If we were able not to find the type, than this switch is wrong
		panic(fmt.Errorf("unknown type %q", n.Type()))
	}

	return result
}
