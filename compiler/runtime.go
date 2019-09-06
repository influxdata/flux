package compiler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type Func interface {
	Type() semantic.Type
	Eval(ctx context.Context, deps dependencies.Interface, input values.Object) (values.Value, error)
}

type Evaluator interface {
	Type() semantic.Type
	Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error)
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
		c.inputScope.Set(k, v)
	})
	return nil
}

func (c compiledFn) Type() semantic.Type {
	return c.fnType.FunctionSignature().Return
}

func (c compiledFn) Eval(ctx context.Context, deps dependencies.Interface, input values.Object) (values.Value, error) {
	if err := c.buildScope(input); err != nil {
		return nil, err
	}

	return eval(ctx, deps, c.root, c.inputScope)
}

type Scope interface {
	values.Scope
	Get(name string) values.Value
}

type compilerScope struct {
	values.Scope
}

func (s compilerScope) Get(name string) values.Value {
	v, ok := s.Scope.Lookup(name)
	if !ok {
		log.Println("Scope", values.FormattedScope(s.Scope))
		panic(fmt.Sprintf("attempting to access non-existant value %q", name))
	}
	return v
}

func NewScope() Scope {
	return ToScope(values.NewScope())
}
func ToScope(s values.Scope) Scope {
	return compilerScope{s}
}

func nestScope(scope Scope) Scope {
	return compilerScope{scope.Nest(nil)}
}

func eval(ctx context.Context, deps dependencies.Interface, e Evaluator, scope Scope) (values.Value, error) {
	v, err := e.Eval(ctx, deps, scope)
	if err != nil {
		return nil, err
	}

	// values.CheckKind(v.PolyType().Nature(), e.Type().Nature())
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

func (e *blockEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	var err error
	for _, b := range e.body {
		e.value, err = eval(ctx, deps, b, scope)
		if err != nil {
			return nil, err
		}
	}
	// values.CheckKind(e.value.Type().Nature(), e.Type().Nature())
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

func (e *declarationEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	v, err := eval(ctx, deps, e.init, scope)
	if err != nil {
		return nil, err
	}

	scope.Set(e.id, v)
	return v, nil
}

type stringExpressionEvaluator struct {
	t     semantic.Type
	parts []Evaluator
}

func (e *stringExpressionEvaluator) Type() semantic.Type {
	return e.t
}

func (e *stringExpressionEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	var b strings.Builder
	for _, p := range e.parts {
		v, err := p.Eval(ctx, deps, scope)
		if err != nil {
			return nil, err
		}
		b.WriteString(v.Str())
	}
	return values.NewString(b.String()), nil
}

type textEvaluator struct {
	value string
}

func (*textEvaluator) Type() semantic.Type {
	return semantic.String
}

func (e *textEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	return values.NewString(e.value), nil
}

type interpolatedEvaluator struct {
	s Evaluator
}

func (*interpolatedEvaluator) Type() semantic.Type {
	return semantic.String
}

func (e *interpolatedEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	return e.s.Eval(ctx, deps, scope)
}

type objEvaluator struct {
	t          semantic.Type
	with       *identifierEvaluator
	properties map[string]Evaluator
}

func (e *objEvaluator) Type() semantic.Type {
	return e.t
}

func (e *objEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	obj := values.NewObject()
	if e.with != nil {
		with, err := e.with.Eval(ctx, deps, scope)
		if err != nil {
			return nil, err
		}
		with.Object().Range(func(name string, v values.Value) {
			obj.Set(name, v)
		})
	}

	for k, node := range e.properties {
		v, err := eval(ctx, deps, node, scope)
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

func (e *arrayEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	arr := values.NewArray(e.t.ElementType())
	for _, ev := range e.array {
		v, err := eval(ctx, deps, ev, scope)
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

func (e *logicalEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	l, err := e.left.Eval(ctx, deps, scope)
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

	r, err := e.right.Eval(ctx, deps, scope)
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

func (e *conditionalEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	t, err := eval(ctx, deps, e.test, scope)
	if err != nil {
		return nil, err
	}

	if t.IsNull() || !t.Bool() {
		return eval(ctx, deps, e.alternate, scope)
	} else {
		return eval(ctx, deps, e.consequent, scope)
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

func (e *binaryEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	l, err := eval(ctx, deps, e.left, scope)
	if err != nil {
		return nil, err
	}
	r, err := eval(ctx, deps, e.right, scope)
	if err != nil {
		return nil, err
	}
	return e.f(l, r)
}

type unaryEvaluator struct {
	t    semantic.Type
	node Evaluator
	op   ast.OperatorKind
}

func (e *unaryEvaluator) Type() semantic.Type {
	return e.t
}

func (e *unaryEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	v, err := e.node.Eval(ctx, deps, scope)
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

func (e *integerEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	return values.NewInt(e.i), nil
}

type unsignedIntegerEvaluator struct {
	t semantic.Type
	i uint64
}

func (e *unsignedIntegerEvaluator) Type() semantic.Type {
	return e.t
}

func (e *unsignedIntegerEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	return values.NewUInt(e.i), nil
}

type stringEvaluator struct {
	t semantic.Type
	s string
}

func (e *stringEvaluator) Type() semantic.Type {
	return e.t
}

func (e *stringEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	return values.NewString(e.s), nil
}

type regexpEvaluator struct {
	t semantic.Type
	r *regexp.Regexp
}

func (e *regexpEvaluator) Type() semantic.Type {
	return e.t
}

func (e *regexpEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	return values.NewRegexp(e.r), nil
}

type booleanEvaluator struct {
	t semantic.Type
	b bool
}

func (e *booleanEvaluator) Type() semantic.Type {
	return e.t
}

func (e *booleanEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	return values.NewBool(e.b), nil
}

type floatEvaluator struct {
	t semantic.Type
	f float64
}

func (e *floatEvaluator) Type() semantic.Type {
	return e.t
}

func (e *floatEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	return values.NewFloat(e.f), nil
}

type timeEvaluator struct {
	t    semantic.Type
	time values.Time
}

func (e *timeEvaluator) Type() semantic.Type {
	return e.t
}

func (e *timeEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	return values.NewTime(e.time), nil
}

type durationEvaluator struct {
	t        semantic.Type
	duration values.Duration
}

func (e *durationEvaluator) Type() semantic.Type {
	return e.t
}

func (e *durationEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	return values.NewDuration(e.duration), nil
}

type identifierEvaluator struct {
	t    semantic.PolyType
	name string
}

func (e *identifierEvaluator) Type() semantic.Type {
	t, ok := e.t.MonoType()
	if !ok {
		return semantic.Invalid
	}
	return t
}

func (e *identifierEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	v := scope.Get(e.name)
	// Note this check pattern is slightly different than
	// the usual one:
	//
	//     CheckKind(v.Type().Nature(), e.t.Nature())
	//                  ^
	//
	// This is because variables hold polytypes that might
	// be universally quantified with one or more type
	// variables. Hence we must check the Nature of this
	// value's **polytype** as this value will return an
	// invalid monotype.
	//
	values.CheckKind(v.PolyType().Nature(), e.t.Nature())
	return v, nil
}

type memberEvaluator struct {
	t        semantic.PolyType
	object   Evaluator
	property string
}

func (e *memberEvaluator) Type() semantic.Type {
	t, ok := e.t.MonoType()
	if !ok {
		return semantic.Invalid
	}
	return t
}

func (e *memberEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	o, err := e.object.Eval(ctx, deps, scope)
	if err != nil {
		return nil, err
	}
	v, _ := o.Object().Get(e.property)
	// values.CheckKind(v.PolyType().Nature(), e.t.Nature())
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

func (e *arrayIndexEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	a, err := e.array.Eval(ctx, deps, scope)
	if err != nil {
		return nil, err
	}
	i, err := e.index.Eval(ctx, deps, scope)
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

func (e *callEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	args, err := e.args.Eval(ctx, deps, scope)
	if err != nil {
		return nil, err
	}
	f, err := e.callee.Eval(ctx, deps, scope)
	if err != nil {
		return nil, err
	}
	return f.Function().Call(ctx, deps, args.Object())
}

type functionEvaluator struct {
	t      semantic.Type
	body   Evaluator
	params []functionParam
}

func (e *functionEvaluator) Type() semantic.Type {
	return e.t
}

func (e *functionEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
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

func (f *functionValue) Call(ctx context.Context, deps dependencies.Interface, args values.Object) (values.Value, error) {
	scope := nestScope(f.scope)
	for _, p := range f.params {
		a, ok := args.Get(p.Key)
		if !ok && p.Default != nil {
			v, err := eval(ctx, deps, p.Default, f.scope)
			if err != nil {
				return nil, err
			}
			a = v
		}
		scope.Set(p.Key, a)
	}
	return eval(ctx, deps, f.body, scope)
}

func (f *functionValue) Type() semantic.Type         { return f.t }
func (f *functionValue) PolyType() semantic.PolyType { return f.t.PolyType() }
func (f *functionValue) IsNull() bool                { return false }
func (f *functionValue) Str() string {
	panic(values.UnexpectedKind(semantic.Function, semantic.String))
}
func (f *functionValue) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Function, semantic.Bytes))
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

func (noopEvaluator) Eval(ctx context.Context, deps dependencies.Interface, scope Scope) (values.Value, error) {
	return values.Null, nil
}
