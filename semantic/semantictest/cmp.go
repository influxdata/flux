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
	cmp.Transformer("MonoType", func(mt semantic.MonoType) string {
		return mt.String()
	}),
	cmp.Transformer("PolyType", func(pt semantic.PolyType) string {
		return pt.String()
	}),
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

	cmpopts.IgnoreFields(semantic.ArrayExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.Package{}, "Loc"),
	cmpopts.IgnoreFields(semantic.File{}, "Loc"),
	cmpopts.IgnoreFields(semantic.PackageClause{}, "Loc"),
	cmpopts.IgnoreFields(semantic.ImportDeclaration{}, "Loc"),
	cmpopts.IgnoreFields(semantic.Block{}, "Loc"),
	cmpopts.IgnoreFields(semantic.OptionStatement{}, "Loc"),
	cmpopts.IgnoreFields(semantic.BuiltinStatement{}, "Loc"),
	cmpopts.IgnoreFields(semantic.TestStatement{}, "Loc"),
	cmpopts.IgnoreFields(semantic.ExpressionStatement{}, "Loc"),
	cmpopts.IgnoreFields(semantic.ReturnStatement{}, "Loc"),
	cmpopts.IgnoreFields(semantic.NativeVariableAssignment{}, "Loc"),
	cmpopts.IgnoreFields(semantic.MemberAssignment{}, "Loc"),
	cmpopts.IgnoreFields(semantic.ArrayExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.FunctionExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.FunctionBlock{}, "Loc"),
	cmpopts.IgnoreFields(semantic.FunctionParameters{}, "Loc"),
	cmpopts.IgnoreFields(semantic.FunctionParameter{}, "Loc"),
	cmpopts.IgnoreFields(semantic.BinaryExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.CallExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.ConditionalExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.LogicalExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.MemberExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.IndexExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.ObjectExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.UnaryExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.Property{}, "Loc"),
	cmpopts.IgnoreFields(semantic.IdentifierExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.Identifier{}, "Loc"),
	cmpopts.IgnoreFields(semantic.BooleanLiteral{}, "Loc"),
	cmpopts.IgnoreFields(semantic.DateTimeLiteral{}, "Loc"),
	cmpopts.IgnoreFields(semantic.DurationLiteral{}, "Loc"),
	cmpopts.IgnoreFields(semantic.IntegerLiteral{}, "Loc"),
	cmpopts.IgnoreFields(semantic.FloatLiteral{}, "Loc"),
	cmpopts.IgnoreFields(semantic.RegexpLiteral{}, "Loc"),
	cmpopts.IgnoreFields(semantic.StringLiteral{}, "Loc"),
	cmpopts.IgnoreFields(semantic.UnsignedIntegerLiteral{}, "Loc"),
	cmpopts.IgnoreFields(semantic.StringExpression{}, "Loc"),
	cmpopts.IgnoreFields(semantic.TextPart{}, "Loc"),
	cmpopts.IgnoreFields(semantic.InterpolatedPart{}, "Loc"),
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
	case semantic.Function:
		// Just use the function type when comparing functions
		return map[string]interface{}{
			"type": v.Type().String(),
		}
	default:
		panic(fmt.Errorf("unexpected value type %v", v.Type()))
	}
}
