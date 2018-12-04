package visit

import (
	"fmt"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func NewOptionEditor(keyMap map[string]values.Value) *OptionEditor {
	return &OptionEditor{keyMap: keyMap}
}

type OptionEditor struct {
	keyMap map[string]values.Value
}

// finds the `OptionStatement`s and returns an `optionEditor` for each one of them
func (v *OptionEditor) Visit(node ast.Node) Visitor {
	switch node.(type) {
	case *ast.OptionStatement:
		return &optionEditor{v}
	}

	return v
}

func (v *OptionEditor) Done(node ast.Node) {}

// has scope limited to one `OptionStatement`
type optionEditor struct {
	*OptionEditor
}

func (v *optionEditor) Visit(node ast.Node) Visitor {
	switch node := node.(type) {
	case *ast.Property:
		v.updateValue(node)
		// property changed, stop walking
		return nil
	}

	return v
}

func (v *optionEditor) Done(node ast.Node) {}

func (v *optionEditor) updateValue(p *ast.Property) {
	value, found := v.keyMap[p.Key.Name]
	if !found {
		return
	}

	p.Value = getLiteral(value)
}

func getLiteral(v values.Value) ast.Expression {
	var literal ast.Expression
	switch v.Type().Nature() {
	case semantic.Bool:
		literal = &ast.BooleanLiteral{Value: v.Bool()}
	case semantic.UInt:
		literal = &ast.UnsignedIntegerLiteral{Value: v.UInt()}
	case semantic.Int:
		literal = &ast.IntegerLiteral{Value: v.Int()}
	case semantic.Float:
		literal = &ast.FloatLiteral{Value: v.Float()}
	case semantic.String:
		literal = &ast.StringLiteral{Value: v.Str()}
	case semantic.Time:
		literal = &ast.DateTimeLiteral{Value: v.Time().Time()}
	case semantic.Duration:
		literal = &ast.DurationLiteral{
			Values: []ast.Duration{
				{
					Magnitude: int64(v.Duration()),
					Unit:      "ns",
				},
			},
		}
	case semantic.Regexp:
		literal = &ast.RegexpLiteral{Value: v.Regexp()}
	default:
		literal = errorLiteral(fmt.Errorf("ERROR, cannot create literal for %v", v.Type().Nature()))
	}

	return literal
}

func errorLiteral(err error) *ast.StringLiteral {
	return &ast.StringLiteral{Value: err.Error()}
}
