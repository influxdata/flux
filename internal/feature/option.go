package feature

import (
	"context"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/edit"
	"github.com/influxdata/flux/internal/pkg/feature"
)

// InjectFromOption returns a new context with feature flags set via the debug.features Flux option
func InjectFromOption(ctx context.Context, source *ast.Package) context.Context {
	var option ast.Expression
	for _, f := range source.Files {
		init, err := edit.GetOption(source, "debug.features")
		if err == nil {
			option = init
			break
		}
	}
	// No option found, nothing to do
	if option == nil {
		return ctx
	}
	features, ok := option.(*ast.ObjectExpression)
	if !ok {
		// TODO: should this error?
		return ctx
	}
	flagger := feature.GetFlagger(ctx)
	mflagger := feature.NewMutableFlagger(flagger)
	for _, prop := range features.Properties {
		key := prop.Key.Key()
		value, ok := getValue(prop.Value)
		flag := ByKey(key)
		if ok {
			mflagger.SetFlagValue(ctx, flag, value)
		}
	}
	return Inject(ctx, mflagger)
}

func getValue(expr ast.Expression) (interface{}, bool) {
	switch expr := expr.(type) {
	// special case true false identifiers
	case *ast.Identifier:
		if expr.Name == "true" {
			return true, true
		}
		if expr.Name == "false" {
			return false, true
		}
	case *ast.IntegerFromLiteral:
		return ast.IntegerFromLiteral(expr), true
	case *ast.UnsignedIntegerLiteral:
		return ast.UnsignedIntegerFromLiteral(expr), true
	case *ast.FloatLiteral:
		return ast.FloatFromLiteral(expr), true
	case *ast.StringLiteral:
		return ast.StringFromLiteral(expr), true
	case *ast.BooleanLiteral:
		return ast.BooleanFromLiteral(expr), true
	case *ast.DateTimeLiteral:
		return ast.DateTimeFromLiteral(expr), true
	case *ast.RegexpLiteral:
		return ast.RegexpFromLiteral(expr), true
	default:
		return nil, false
	}
}
