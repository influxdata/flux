package ast

import (
	"fmt"
	"reflect"
	"regexp"
	"time"
)

func NewOptionEditor(keyMap map[string]interface{}) *OptionEditor {
	return &OptionEditor{keyMap: keyMap}
}

type OptionEditor struct {
	keyMap map[string]interface{}
}

// finds the `OptionStatement`s and returns an `optionEditor` for each one of them
func (v *OptionEditor) Visit(node Node) Visitor {
	switch node.(type) {
	case *OptionStatement:
		return &optionEditor{v}
	}

	return v
}

func (v *OptionEditor) Done(node Node) {}

// has scope limited to one `OptionStatement`
type optionEditor struct {
	*OptionEditor
}

func (v *optionEditor) Visit(node Node) Visitor {
	switch node := node.(type) {
	case *Property:
		v.updateValue(node)
		// property changed, stop walking
		return nil
	}

	return v
}

func (v *optionEditor) Done(node Node) {}

func (v *optionEditor) updateValue(p *Property) {
	value, found := v.keyMap[p.Key.Name]
	if !found {
		return
	}

	var literal Expression
	switch val := value.(type) {
	case string:
		literal = &StringLiteral{Value: val}
	case bool:
		literal = &BooleanLiteral{Value: val}
	case float64:
		literal = &FloatLiteral{Value: val}
	case int64:
		literal = &IntegerLiteral{Value: val}
	case int:
		literal = &IntegerLiteral{Value: int64(val)}
	case uint64:
		literal = &UnsignedIntegerLiteral{Value: val}
	case regexp.Regexp:
		literal = &RegexpLiteral{Value: &val}
	case *regexp.Regexp:
		literal = &RegexpLiteral{Value: val}
	case Duration:
		// TODO should we allow to specify more than one duration?
		literal = &DurationLiteral{Values: []Duration{val}}
	case time.Time:
		literal = &DateTimeLiteral{Value: val}
	default:
		errString := fmt.Sprintf("ERROR, cannot create literal for %v", reflect.TypeOf(val))
		literal = &StringLiteral{Value: errString}
	}

	p.Value = literal
}
