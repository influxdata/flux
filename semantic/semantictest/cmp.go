package semantictest

import (
	"fmt"
	"regexp"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var CmpOptions = []cmp.Option{
	cmp.Comparer(func(x, y *regexp.Regexp) bool { return x.String() == y.String() }),
	cmp.Transformer("Value", TransformValue),

	cmpopts.IgnoreUnexported(semantic.ArrayExpression{}),
	cmpopts.IgnoreUnexported(semantic.Package{}),
	cmpopts.IgnoreUnexported(semantic.File{}),
	cmpopts.IgnoreUnexported(semantic.PackageClause{}),
	cmpopts.IgnoreUnexported(semantic.ImportDeclaration{}),
	cmpopts.IgnoreUnexported(semantic.Block{}),
	cmpopts.IgnoreUnexported(semantic.OptionStatement{}),
	cmpopts.IgnoreUnexported(semantic.BuiltinStatement{}),
	cmpopts.IgnoreUnexported(semantic.TestStatement{}),
	cmpopts.IgnoreUnexported(semantic.ExpressionStatement{}),
	cmpopts.IgnoreUnexported(semantic.ReturnStatement{}),
	cmpopts.IgnoreUnexported(semantic.NativeVariableAssignment{}),
	cmpopts.IgnoreUnexported(semantic.MemberAssignment{}),
	cmpopts.IgnoreUnexported(semantic.Extern{}),
	cmpopts.IgnoreUnexported(semantic.ExternalVariableAssignment{}),
	cmpopts.IgnoreUnexported(semantic.ArrayExpression{}),
	cmpopts.IgnoreUnexported(semantic.FunctionExpression{}),
	cmpopts.IgnoreUnexported(semantic.FunctionBlock{}),
	cmpopts.IgnoreUnexported(semantic.FunctionParameters{}),
	cmpopts.IgnoreUnexported(semantic.FunctionParameter{}),
	cmpopts.IgnoreUnexported(semantic.BinaryExpression{}),
	cmpopts.IgnoreUnexported(semantic.CallExpression{}),
	cmpopts.IgnoreUnexported(semantic.ConditionalExpression{}),
	cmpopts.IgnoreUnexported(semantic.LogicalExpression{}),
	cmpopts.IgnoreUnexported(semantic.MemberExpression{}),
	cmpopts.IgnoreUnexported(semantic.IndexExpression{}),
	cmpopts.IgnoreUnexported(semantic.ObjectExpression{}),
	cmpopts.IgnoreUnexported(semantic.UnaryExpression{}),
	cmpopts.IgnoreUnexported(semantic.Property{}),
	cmpopts.IgnoreUnexported(semantic.IdentifierExpression{}),
	cmpopts.IgnoreUnexported(semantic.Identifier{}),
	cmpopts.IgnoreUnexported(semantic.BooleanLiteral{}),
	cmpopts.IgnoreUnexported(semantic.DateTimeLiteral{}),
	cmpopts.IgnoreUnexported(semantic.DurationLiteral{}),
	cmpopts.IgnoreUnexported(semantic.IntegerLiteral{}),
	cmpopts.IgnoreUnexported(semantic.FloatLiteral{}),
	cmpopts.IgnoreUnexported(semantic.RegexpLiteral{}),
	cmpopts.IgnoreUnexported(semantic.StringLiteral{}),
	cmpopts.IgnoreUnexported(semantic.UnsignedIntegerLiteral{}),
	cmpopts.IgnoreUnexported(semantic.StringExpression{}),
	cmpopts.IgnoreUnexported(semantic.TextPart{}),
	cmpopts.IgnoreUnexported(semantic.InterpolatedPart{}),
}

func TransformValue(v values.Value) map[string]interface{} {
	if v.IsNull() {
		return map[string]interface{}{
			"type":  v.Type(),
			"value": nil,
		}
	}

	switch v.Type().Nature() {
	case semantic.Int:
		return map[string]interface{}{
			"type":  semantic.Int.String(),
			"value": v.Int(),
		}
	case semantic.UInt:
		return map[string]interface{}{
			"type":  semantic.UInt.String(),
			"value": v.UInt(),
		}
	case semantic.Float:
		return map[string]interface{}{
			"type":  semantic.Float.String(),
			"value": v.Float(),
		}
	case semantic.String:
		return map[string]interface{}{
			"type":  semantic.String.String(),
			"value": v.Str(),
		}
	case semantic.Bool:
		return map[string]interface{}{
			"type":  semantic.Bool.String(),
			"value": v.Bool(),
		}
	case semantic.Time:
		return map[string]interface{}{
			"type":  semantic.Time.String(),
			"value": v.Time(),
		}
	case semantic.Duration:
		return map[string]interface{}{
			"type":  semantic.Duration.String(),
			"value": v.Duration(),
		}
	case semantic.Regexp:
		return map[string]interface{}{
			"type":  semantic.Regexp.String(),
			"value": v.Regexp(),
		}
	case semantic.Array:
		elements := make([]map[string]interface{}, v.Array().Len())
		for i := range elements {
			elements[i] = TransformValue(v.Array().Get(i))
		}
		return map[string]interface{}{
			"type":     semantic.Array.String(),
			"elements": elements,
		}
	case semantic.Object:
		elements := make(map[string]interface{})
		v.Object().Range(func(name string, v values.Value) {
			elements[name] = TransformValue(v)
		})
		return map[string]interface{}{
			"type":     semantic.Object.String(),
			"elements": elements,
		}
	default:
		panic(fmt.Errorf("unexpected value type %v", v.Type()))
	}
}
