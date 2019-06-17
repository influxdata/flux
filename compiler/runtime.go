package compiler

import (
	"fmt"
	"regexp"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

type Func interface {
	Type() semantic.Type
	Eval(input values.Object) (values.Value, error)
}

type Evaluator interface {
	Type() semantic.Type
	Eval(scope Scope) (values.Value, error)
}

type compiledFn struct {
	root       Evaluator
	fnType     semantic.Type
	inputScope Scope
}

func (c compiledFn) validate(input values.Object) error {
	sig := c.fnType.FunctionSignature()
	properties := input.Type().Properties()
	if len(properties) != len(sig.Parameters) {
		return errors.New("mismatched parameters and properties")
	}
	for k, v := range sig.Parameters {
		if !values.AssignableTo(properties[k], v) {
			return fmt.Errorf("parameter %q has the wrong type, expected %v got %v", k, v, properties[k])
		}
	}
	return nil
}

func (c compiledFn) buildScope(input values.Object) error {
	if err := c.validate(input); err != nil {
		return err
	}
	input.Range(func(k string, v values.Value) {
		c.inputScope[k] = v
	})
	return nil
}

func (c compiledFn) Type() semantic.Type {
	return c.fnType.FunctionSignature().Return
}

func (c compiledFn) Eval(input values.Object) (values.Value, error) {
	if err := c.buildScope(input); err != nil {
		return nil, err
	}

	return eval(c.root, c.inputScope)
}

type Scope map[string]values.Value

func (s Scope) Type(name string) semantic.Type {
	return s[name].Type()
}
func (s Scope) Set(name string, v values.Value) {
	s[name] = v
}
func (s Scope) Get(name string) values.Value {
	v := s[name]
	if v == nil {
		panic("attempting to access non-existant value")
	}
	return v
}

func (s Scope) Copy() Scope {
	n := make(Scope, len(s))
	for k, v := range s {
		n[k] = v
	}
	return n
}

func eval(e Evaluator, scope Scope) (values.Value, error) {
	v, err := e.Eval(scope)
	if err != nil {
		return nil, err
	}
	values.CheckKind(e.Type().Nature(), v.Type().Nature())
	return v, nil
}

type blockEvaluator struct {
	t     semantic.Type
	body  []Evaluator
	value values.Value
}

func (e *blockEvaluator) Type() semantic.Type {
	return e.t
}

func (e *blockEvaluator) Eval(scope Scope) (values.Value, error) {
	var err error
	for _, b := range e.body {
		e.value, err = eval(b, scope)
		if err != nil {
			return nil, err
		}
	}
	values.CheckKind(e.value.Type().Nature(), e.Type().Nature())
	return e.value, nil
}

type returnEvaluator struct {
	Evaluator
}

type declarationEvaluator struct {
	t    semantic.Type
	id   string
	init Evaluator
}

func (e *declarationEvaluator) Type() semantic.Type {
	return e.t
}

func (e *declarationEvaluator) Eval(scope Scope) (values.Value, error) {
	v, err := eval(e.init, scope)
	if err != nil {
		return nil, err
	}

	scope.Set(e.id, v)
	return v, nil
}

type objEvaluator struct {
	t          semantic.Type
	with       *identifierEvaluator
	properties map[string]Evaluator
}

func (e *objEvaluator) Type() semantic.Type {
	return e.t
}

func (e *objEvaluator) Eval(scope Scope) (values.Value, error) {
	obj := values.NewObject()
	if e.with != nil {
		with, err := e.with.Eval(scope)
		if err != nil {
			return nil, err
		}
		with.Object().Range(func(name string, v values.Value) {
			obj.Set(name, v)
		})
	}

	for k, node := range e.properties {
		v, err := eval(node, scope)
		if err != nil {
			return nil, err
		}
		obj.Set(k, v)
	}

	return obj, nil
}

type arrayEvaluator struct {
	t     semantic.Type
	array []Evaluator
}

func (e *arrayEvaluator) Type() semantic.Type {
	return e.t
}

func (e *arrayEvaluator) Eval(scope Scope) (values.Value, error) {
	arr := values.NewArray(e.t.ElementType())
	for _, ev := range e.array {
		v, err := eval(ev, scope)
		if err != nil {
			return nil, err
		}
		arr.Append(v)
	}
	return arr, nil
}

type logicalEvaluator struct {
	t           semantic.Type
	operator    ast.LogicalOperatorKind
	left, right Evaluator
}

func (e *logicalEvaluator) Type() semantic.Type {
	return e.t
}

func (e *logicalEvaluator) Eval(scope Scope) (values.Value, error) {
	l, err := e.left.Eval(scope)
	if err != nil {
		return nil, err
	}
	values.CheckKind(l.Type().Nature(), e.t.Nature())

	switch e.operator {
	case ast.AndOperator:
		if l.IsNull() || !l.Bool() {
			return values.NewBool(false), nil
		}
	case ast.OrOperator:
		if !l.IsNull() && l.Bool() {
			return values.NewBool(true), nil
		}
	default:
		panic(fmt.Errorf("unknown logical operator %v", e.operator))
	}

	r, err := e.right.Eval(scope)
	if err != nil {
		return nil, err
	}
	return r, nil
}

type conditionalEvaluator struct {
	t          semantic.Type
	test       Evaluator
	consequent Evaluator
	alternate  Evaluator
}

func (e *conditionalEvaluator) Type() semantic.Type {
	return e.t
}

func (e *conditionalEvaluator) Eval(scope Scope) (values.Value, error) {
	t, err := eval(e.test, scope)
	if err != nil {
		return nil, err
	}

	if t.Bool() {
		return eval(e.consequent, scope)
	} else {
		return eval(e.alternate, scope)
	}
}

type binaryEvaluator struct {
	t           semantic.Type
	left, right Evaluator
	f           values.BinaryFunction
}

func (e *binaryEvaluator) Type() semantic.Type {
	return e.t
}

func (e *binaryEvaluator) Eval(scope Scope) (values.Value, error) {
	l, err := eval(e.left, scope)
	if err != nil {
		return nil, err
	}
	r, err := eval(e.right, scope)
	if err != nil {
		return nil, err
	}
	return e.f(l, r), nil
}

type unaryEvaluator struct {
	t    semantic.Type
	node Evaluator
	op   ast.OperatorKind
}

func (e *unaryEvaluator) Type() semantic.Type {
	return e.t
}

func (e *unaryEvaluator) Eval(scope Scope) (values.Value, error) {
	v, err := e.node.Eval(scope)
	if err != nil {
		return nil, err
	}

	ret, err := func(v values.Value) (values.Value, error) {
		if e.op == ast.ExistsOperator {
			return values.NewBool(!v.IsNull()), nil
		}

		// If the value is null, return it immediately.
		if v.IsNull() {
			return v, nil
		}

		switch e.op {
		case ast.AdditionOperator:
			// Do nothing.
			return v, nil
		case ast.SubtractionOperator, ast.NotOperator:
			// Fallthrough to below.
		default:
			return nil, fmt.Errorf("unknown unary operator: %s", e.op)
		}

		// The subtraction operator falls through to here.
		switch v.Type().Nature() {
		case semantic.Int:
			return values.NewInt(-v.Int()), nil
		case semantic.Float:
			return values.NewFloat(-v.Float()), nil
		case semantic.Bool:
			return values.NewBool(!v.Bool()), nil
		case semantic.Duration:
			return values.NewDuration(-v.Duration()), nil
		default:
			panic(values.UnexpectedKind(e.t.Nature(), v.Type().Nature()))
		}
	}(v)
	if err != nil {
		return nil, err
	}
	values.CheckKind(ret.Type().Nature(), e.t.Nature())
	return ret, nil
}

type integerEvaluator struct {
	t semantic.Type
	i int64
}

func (e *integerEvaluator) Type() semantic.Type {
	return e.t
}

func (e *integerEvaluator) Eval(scope Scope) (values.Value, error) {
	return values.NewInt(e.i), nil
}

type stringEvaluator struct {
	t semantic.Type
	s string
}

func (e *stringEvaluator) Type() semantic.Type {
	return e.t
}

func (e *stringEvaluator) Eval(scope Scope) (values.Value, error) {
	return values.NewString(e.s), nil
}

type regexpEvaluator struct {
	t semantic.Type
	r *regexp.Regexp
}

func (e *regexpEvaluator) Type() semantic.Type {
	return e.t
}

func (e *regexpEvaluator) Eval(scope Scope) (values.Value, error) {
	return values.NewRegexp(e.r), nil
}

type booleanEvaluator struct {
	t semantic.Type
	b bool
}

func (e *booleanEvaluator) Type() semantic.Type {
	return e.t
}

func (e *booleanEvaluator) Eval(scope Scope) (values.Value, error) {
	return values.NewBool(e.b), nil
}

type floatEvaluator struct {
	t semantic.Type
	f float64
}

func (e *floatEvaluator) Type() semantic.Type {
	return e.t
}

func (e *floatEvaluator) Eval(scope Scope) (values.Value, error) {
	return values.NewFloat(e.f), nil
}

type timeEvaluator struct {
	t    semantic.Type
	time values.Time
}

func (e *timeEvaluator) Type() semantic.Type {
	return e.t
}

func (e *timeEvaluator) Eval(scope Scope) (values.Value, error) {
	return values.NewTime(e.time), nil
}

type durationEvaluator struct {
	t        semantic.Type
	duration values.Duration
}

func (e *durationEvaluator) Type() semantic.Type {
	return e.t
}

func (e *durationEvaluator) Eval(scope Scope) (values.Value, error) {
	return values.NewDuration(e.duration), nil
}

type identifierEvaluator struct {
	t    semantic.Type
	name string
}

func (e *identifierEvaluator) Type() semantic.Type {
	return e.t
}

func (e *identifierEvaluator) Eval(scope Scope) (values.Value, error) {
	v := scope.Get(e.name)
	values.CheckKind(v.Type().Nature(), e.t.Nature())
	return v, nil
}

type valueEvaluator struct {
	value values.Value
}

func (e *valueEvaluator) Type() semantic.Type {
	return e.value.Type()
}

func (e *valueEvaluator) Eval(scope Scope) (values.Value, error) {
	return e.value, nil
}

type memberEvaluator struct {
	t        semantic.Type
	object   Evaluator
	property string
}

func (e *memberEvaluator) Type() semantic.Type {
	return e.t
}

func (e *memberEvaluator) Eval(scope Scope) (values.Value, error) {
	o, err := e.object.Eval(scope)
	if err != nil {
		return nil, err
	}
	v, _ := o.Object().Get(e.property)
	values.CheckKind(v.Type().Nature(), e.t.Nature())
	return v, nil
}

type arrayIndexEvaluator struct {
	t     semantic.Type
	array Evaluator
	index Evaluator
}

func (e *arrayIndexEvaluator) Type() semantic.Type {
	return e.t
}

func (e *arrayIndexEvaluator) Eval(scope Scope) (values.Value, error) {
	a, err := e.array.Eval(scope)
	if err != nil {
		return nil, err
	}
	i, err := e.index.Eval(scope)
	if err != nil {
		return nil, err
	}
	return a.Array().Get(int(i.Int())), nil
}

type callEvaluator struct {
	t      semantic.Type
	callee Evaluator
	args   Evaluator
}

func (e *callEvaluator) Type() semantic.Type {
	return e.t
}

func (e *callEvaluator) Eval(scope Scope) (values.Value, error) {
	args, err := e.args.Eval(scope)
	if err != nil {
		return nil, err
	}
	f, err := e.callee.Eval(scope)
	if err != nil {
		return nil, err
	}
	return f.Function().Call(args.Object())
}

type functionEvaluator struct {
	t      semantic.Type
	body   Evaluator
	params []functionParam
}

func (e *functionEvaluator) Type() semantic.Type {
	return e.t
}

func (e *functionEvaluator) Eval(scope Scope) (values.Value, error) {
	return &functionValue{
		t:      e.t,
		body:   e.body,
		params: e.params,
		scope:  scope,
	}, nil
}

type functionValue struct {
	t      semantic.Type
	body   Evaluator
	params []functionParam
	scope  Scope
}

type functionParam struct {
	Key     string
	Default Evaluator
	Type    semantic.Type
}

func (f *functionValue) HasSideEffect() bool {
	return false
}

func (f *functionValue) Call(args values.Object) (values.Value, error) {
	scope := f.scope.Copy()
	for _, p := range f.params {
		a, ok := args.Get(p.Key)
		if !ok && p.Default != nil {
			v, err := eval(p.Default, f.scope)
			if err != nil {
				return nil, err
			}
			a = v
		}
		scope.Set(p.Key, a)
	}
	return eval(f.body, scope)
}

func (f *functionValue) Type() semantic.Type         { return f.t }
func (f *functionValue) PolyType() semantic.PolyType { return f.t.PolyType() }
func (f *functionValue) IsNull() bool                { return false }
func (f *functionValue) Str() string {
	panic(values.UnexpectedKind(semantic.Function, semantic.String))
}
func (f *functionValue) Int() int64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.Int))
}
func (f *functionValue) UInt() uint64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.UInt))
}
func (f *functionValue) Float() float64 {
	panic(values.UnexpectedKind(semantic.Function, semantic.Float))
}
func (f *functionValue) Bool() bool {
	panic(values.UnexpectedKind(semantic.Function, semantic.Bool))
}
func (f *functionValue) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Function, semantic.Time))
}
func (f *functionValue) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Function, semantic.Duration))
}
func (f *functionValue) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Function, semantic.Regexp))
}
func (f *functionValue) Array() values.Array {
	panic(values.UnexpectedKind(semantic.Function, semantic.Array))
}
func (f *functionValue) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Function, semantic.Object))
}
func (f *functionValue) Function() values.Function {
	return f
}
func (f *functionValue) Equal(rhs values.Value) bool {
	if f.Type() != rhs.Type() {
		return false
	}
	v, ok := rhs.(*functionValue)
	return ok && (f == v)
}

type noopEvaluator struct{}

func (noopEvaluator) Type() semantic.Type {
	return semantic.Nil
}

func (noopEvaluator) Eval(scope Scope) (values.Value, error) {
	return values.Null, nil
}
