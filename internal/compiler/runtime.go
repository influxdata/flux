package compiler

import (
	"context"
	"regexp"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
)

type Func interface {
	Type() semantic.MonoType
	Eval(ctx context.Context, input Value) (Value, error)
}

type Evaluator interface {
	Type() semantic.MonoType
	Eval(ctx context.Context, scope Scope) (Value, error)
}

type compiledFn struct {
	root       Evaluator
	inputScope Scope
}

func (c compiledFn) buildScope(input Value) error {
	input.Range(func(i int, _ string, value Value) {
		c.inputScope.Set(i, value)
	})
	return nil
}

// Type returns the return type of the compiled function.
func (c compiledFn) Type() semantic.MonoType {
	return c.root.Type()
}

func (c compiledFn) Eval(ctx context.Context, input Value) (Value, error) {
	if err := c.buildScope(input); err != nil {
		return Value{}, err
	}

	return eval(ctx, c.root, c.inputScope)
}

type Scope struct {
	data *scope
}

func NewScope() Scope {
	return Scope{
		data: &scope{},
	}
}

type scope struct {
	names  []string
	values []Value
}

func (s Scope) Declare(name string) int {
	s.data.names = append(s.data.names, name)
	s.data.values = append(s.data.values, Value{})
	return len(s.data.names) - 1
}

func (s Scope) Set(i int, v Value) {
	s.data.values[i] = v
}

func (s Scope) Get(i int) Value {
	return s.data.values[i]
}

func (s Scope) Index(name string) int {
	for i := len(s.data.names) - 1; i >= 0; i-- {
		if s.data.names[i] == name {
			return i
		}
	}
	return -1
}

// func ToScope(s values.Scope) Scope {
// 	if s == nil {
// 		return nil
// 	}
// 	return runtimeScope{s}
// }
//
// func nestScope(scope Scope) Scope {
// 	return runtimeScope{scope.Nest(nil)}
// }

func eval(ctx context.Context, e Evaluator, scope Scope) (Value, error) {
	v, err := e.Eval(ctx, scope)
	if err != nil {
		return Value{}, err
	}
	return v, nil
}

type blockEvaluator struct {
	t     semantic.MonoType
	body  []Evaluator
	value Value
}

func (e *blockEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *blockEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	var err error
	for _, b := range e.body {
		e.value, err = eval(ctx, b, scope)
		if err != nil {
			return Value{}, err
		}
	}
	return e.value, nil
}

type returnEvaluator struct {
	Evaluator
}

type declarationEvaluator struct {
	t    semantic.MonoType
	id   int
	init Evaluator
}

func (e *declarationEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *declarationEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	v, err := eval(ctx, e.init, scope)
	if err != nil {
		return Value{}, err
	}

	scope.Set(e.id, v)
	return v, nil
}

type stringExpressionEvaluator struct {
	parts []Evaluator
}

func (e *stringExpressionEvaluator) Type() semantic.MonoType {
	return semantic.BasicString
}

func (e *stringExpressionEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	// var b strings.Builder
	// for _, p := range e.parts {
	// 	v, err := p.Eval(ctx, scope)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	// if v.IsNull() {
	// 	// 	return nil, errors.New(codes.Invalid, "string expression evaluated to null")
	// 	// }
	// 	b.WriteString(v.Str())
	// }
	// return values.NewString(b.String()), nil
	return Value{}, errors.New(codes.Unimplemented)
}

type textEvaluator struct {
	value string
}

func (*textEvaluator) Type() semantic.MonoType {
	return semantic.BasicString
}

func (e *textEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	// return values.NewString(e.value), nil
	return Value{}, errors.New(codes.Unimplemented)
}

type interpolatedEvaluator struct {
	s Evaluator
}

func (*interpolatedEvaluator) Type() semantic.MonoType {
	return semantic.BasicString
}

func (e *interpolatedEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	// o, err := e.s.Eval(ctx, scope)
	// if err != nil {
	// 	return nil, err
	// }
	// v, err := values.Stringify(o)
	// return v, err
	return Value{}, errors.New(codes.Unimplemented)
}

type objEvaluator struct {
	t          semantic.MonoType
	with       *identifierEvaluator
	properties map[string]Evaluator
}

func (e *objEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *objEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	return Value{}, errors.New(codes.Unimplemented)
	// return values.BuildObject(func(set values.ObjectSetter) error {
	// 	if e.with != nil {
	// 		with, err := e.with.Eval(ctx, scope)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		with.Object().Range(func(name string, v values.Value) {
	// 			set(name, v)
	// 		})
	// 	}
	//
	// 	for k, node := range e.properties {
	// 		v, err := eval(ctx, node, scope)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		set(k, v)
	// 	}
	// 	return nil
	// })
}

type arrayEvaluator struct {
	t     semantic.MonoType
	array []Evaluator
}

func (e *arrayEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *arrayEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	// arr := values.NewArray(e.t)
	// for _, ev := range e.array {
	// 	v, err := eval(ctx, ev, scope)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	arr.Append(v)
	// }
	// return arr, nil
	return Value{}, errors.New(codes.Unimplemented)
}

type dictEvaluator struct {
	elements []struct {
		Key Evaluator
		Val Evaluator
	}
	t semantic.MonoType
}

func (e *dictEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *dictEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	// if len(e.elements) == 0 {
	// 	return values.NewEmptyDict(e.t), nil
	// }
	// builder := values.NewDictBuilder(e.t)
	// for _, item := range e.elements {
	// 	key, err := eval(ctx, item.Key, scope)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	val, err := eval(ctx, item.Val, scope)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if err := builder.Insert(key, val); err != nil {
	// 		return nil, err
	// 	}
	// }
	// return builder.Dict(), nil
	return Value{}, errors.New(codes.Unimplemented)
}

type logicalEvaluator struct {
	operator    ast.LogicalOperatorKind
	left, right Evaluator
}

func (e *logicalEvaluator) Type() semantic.MonoType {
	return semantic.BasicBool
}

func (e *logicalEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	l, err := e.left.Eval(ctx, scope)
	if err != nil {
		return Value{}, err
	}

	switch e.operator {
	case ast.AndOperator:
		if /*l.IsNull() ||*/ !l.Bool() {
			return NewBool(false), nil
		}
	case ast.OrOperator:
		if /*!l.IsNull() &&*/ l.Bool() {
			return NewBool(true), nil
		}
	default:
		panic(errors.Newf(codes.Internal, "unknown logical operator %v", e.operator))
	}

	r, err := e.right.Eval(ctx, scope)
	if err != nil {
		return Value{}, err
	}
	return r, nil
}

type conditionalEvaluator struct {
	test       Evaluator
	consequent Evaluator
	alternate  Evaluator
}

func (e *conditionalEvaluator) Type() semantic.MonoType {
	return e.alternate.Type()
}

func (e *conditionalEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	t, err := eval(ctx, e.test, scope)
	if err != nil {
		return Value{}, err
	}

	if /*t.IsNull() ||*/ !t.Bool() {
		return eval(ctx, e.alternate, scope)
	} else {
		return eval(ctx, e.consequent, scope)
	}
}

type binaryEvaluator struct {
	t           semantic.MonoType
	left, right Evaluator
	f           BinaryFunction
}

func (e *binaryEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *binaryEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	l, err := e.left.Eval(ctx, scope)
	if err != nil {
		return Value{}, err
	}
	r, err := e.right.Eval(ctx, scope)
	if err != nil {
		return Value{}, err
	}
	return e.f(l, r)
}

type unaryEvaluator struct {
	t    semantic.MonoType
	node Evaluator
	op   ast.OperatorKind
}

func (e *unaryEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *unaryEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	v, err := e.node.Eval(ctx, scope)
	if err != nil {
		return Value{}, err
	}

	ret, err := func(v Value) (Value, error) {
		if e.op == ast.ExistsOperator {
			return Value{}, errors.New(codes.Unimplemented)
			// return NewBool(!v.IsNull()), nil
		}

		// If the value is null, return it immediately.
		// if v.IsNull() {
		// 	return v, nil
		// }

		switch e.op {
		case ast.AdditionOperator:
			// Do nothing.
			return v, nil
		case ast.SubtractionOperator, ast.NotOperator:
			// Fallthrough to below.
		default:
			return Value{}, errors.Newf(codes.Internal, "unknown unary operator: %s", e.op)
		}

		// The subtraction operator falls through to here.
		switch v.Nature() {
		case semantic.Int:
			return NewInt(-v.Int()), nil
		case semantic.Float:
			return NewFloat(-v.Float()), nil
		case semantic.Bool:
			return NewBool(!v.Bool()), nil
		case semantic.Duration:
			return Value{}, errors.New(codes.Unimplemented)
			// return values.NewDuration(v.Duration().Mul(-1)), nil
		default:
			panic(UnexpectedKind(e.t.Nature(), v.Nature()))
		}
	}(v)
	if err != nil {
		return Value{}, err
	}
	return ret, nil
}

type integerEvaluator struct {
	i int64
}

func (e *integerEvaluator) Type() semantic.MonoType {
	return semantic.BasicInt
}

func (e *integerEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	return NewInt(e.i), nil
}

type unsignedIntegerEvaluator struct {
	i uint64
}

func (e *unsignedIntegerEvaluator) Type() semantic.MonoType {
	return semantic.BasicUint
}

func (e *unsignedIntegerEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	return Value{}, errors.New(codes.Unimplemented)
	// return NewUint(e.i), nil
}

type stringEvaluator struct {
	s string
}

func (e *stringEvaluator) Type() semantic.MonoType {
	return semantic.BasicString
}

func (e *stringEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	return NewString(e.s), nil
}

type regexpEvaluator struct {
	r *regexp.Regexp
}

func (e *regexpEvaluator) Type() semantic.MonoType {
	return semantic.BasicRegexp
}

func (e *regexpEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	// return values.NewRegexp(e.r), nil
	return Value{}, errors.New(codes.Unimplemented)
}

type booleanEvaluator struct {
	b bool
}

func (e *booleanEvaluator) Type() semantic.MonoType {
	return semantic.BasicBool
}

func (e *booleanEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	return NewBool(e.b), nil
}

type floatEvaluator struct {
	f float64
}

func (e *floatEvaluator) Type() semantic.MonoType {
	return semantic.BasicFloat
}

func (e *floatEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	return NewFloat(e.f), nil
}

func (e *floatEvaluator) EvalFloat(ctx context.Context, scope Scope) (float64, error) {
	return e.f, nil
}

// type timeEvaluator struct {
// 	time values.Time
// }
//
// func (e *timeEvaluator) Type() semantic.MonoType {
// 	return semantic.BasicTime
// }
//
// func (e *timeEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
// 	// return values.NewTime(e.time), nil
// 	return Value{}, errors.New(codes.Unimplemented)
// }
//
// type durationEvaluator struct {
// 	duration values.Duration
// }
//
// func (e *durationEvaluator) Type() semantic.MonoType {
// 	return semantic.BasicDuration
// }
//
// func (e *durationEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
// 	// return values.NewDuration(e.duration), nil
// 	return Value{}, errors.New(codes.Unimplemented)
// }

type identifierEvaluator struct {
	t    semantic.MonoType
	name int
}

func (e *identifierEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *identifierEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	v := scope.Get(e.name)
	return v, nil
}

func (e *identifierEvaluator) EvalFloat(ctx context.Context, scope Scope) (float64, error) {
	v, err := e.Eval(ctx, scope)
	if err != nil {
		return 0, err
	}
	return v.Float(), nil
}

type memberEvaluator struct {
	t        semantic.MonoType
	object   Evaluator
	property int
	nullable bool
}

func (e *memberEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *memberEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	o, err := e.object.Eval(ctx, scope)
	if err != nil {
		return Value{}, err
	}
	v, ok := o.Get(e.property)
	if !ok && !e.nullable {
		return Value{}, errors.Newf(codes.Invalid, "member %q with type %s is not in the record", e.property, e.t.Nature())
	}
	return v, nil
}

type arrayIndexEvaluator struct {
	t     semantic.MonoType
	array Evaluator
	index Evaluator
}

func (e *arrayIndexEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *arrayIndexEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	// a, err := e.array.Eval(ctx, scope)
	// if err != nil {
	// 	return nil, err
	// }
	// i, err := e.index.Eval(ctx, scope)
	// if err != nil {
	// 	return nil, err
	// }
	// ix := int(i.Int())
	// l := a.Array().Len()
	// if ix < 0 || ix >= l {
	// 	return nil, errors.Newf(codes.OutOfRange, "cannot access element %v of array of length %v", ix, l)
	// }
	// return a.Array().Get(ix), nil
	return Value{}, errors.New(codes.Unimplemented)
}

type callEvaluator struct {
	t      semantic.MonoType
	callee Evaluator
	args   Evaluator
}

func (e *callEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *callEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	// args, err := e.args.Eval(ctx, scope)
	// if err != nil {
	// 	return nil, err
	// }
	// f, err := e.callee.Eval(ctx, scope)
	// if err != nil {
	// 	return nil, err
	// }
	// return f.Function().Call(ctx, args.Object())
	return Value{}, errors.New(codes.Unimplemented)
}

type functionEvaluator struct {
	t  semantic.MonoType
	fn *semantic.FunctionExpression
	// params []functionParam
}

func (e *functionEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *functionEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	// return &functionValue{
	// 	t:      e.t,
	// 	fn:     e.fn,
	// 	params: e.params,
	// 	scope:  scope,
	// }, nil
	return Value{}, errors.New(codes.Unimplemented)
}

// type functionValue struct {
// 	t      semantic.MonoType
// 	fn     *semantic.FunctionExpression
// 	params []functionParam
// 	scope  Scope
// }

// functionValue implements the interpreter.Resolver interface.
// var _ interpreter.Resolver = (*functionValue)(nil)
//
// func (f *functionValue) Resolve() (semantic.Node, error) {
// 	n := f.fn.Copy()
// 	localIdentifiers := make([]string, 0, 10)
// 	node, err := interpreter.ResolveIdsInFunction(f.scope, f.fn, n, &localIdentifiers)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return node, nil
// }
//
// func (f functionValue) Scope() values.Scope {
// 	return f.scope
// }
//
// type functionParam struct {
// 	Key     string
// 	Default Evaluator
// 	Type    semantic.MonoType
// }
//
// func (f *functionValue) HasSideEffect() bool {
// 	return false
// }
//
// func (f *functionValue) Call(ctx context.Context, args values.Object) (values.Value, error) {
// 	scope := nestScope(f.scope)
// 	for _, p := range f.params {
// 		a, ok := args.Get(p.Key)
// 		if !ok && p.Default != nil {
// 			v, err := eval(ctx, p.Default, f.scope)
// 			if err != nil {
// 				return nil, err
// 			}
// 			a = v
// 		}
// 		scope.Set(p.Key, a)
// 	}
//
// 	fn, err := Compile(scope, f.fn, args.Type())
// 	if err != nil {
// 		return nil, err
// 	}
// 	return fn.Eval(ctx, args)
// }
//
// func (f *functionValue) Type() semantic.MonoType { return f.t }
// func (f *functionValue) IsNull() bool            { return false }
// func (f *functionValue) Str() string {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.String))
// }
// func (f *functionValue) Bytes() []byte {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.Bytes))
// }
// func (f *functionValue) Int() int64 {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.Int))
// }
// func (f *functionValue) UInt() uint64 {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.UInt))
// }
// func (f *functionValue) Float() float64 {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.Float))
// }
// func (f *functionValue) Bool() bool {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.Bool))
// }
// func (f *functionValue) Time() values.Time {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.Time))
// }
// func (f *functionValue) Duration() values.Duration {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.Duration))
// }
// func (f *functionValue) Regexp() *regexp.Regexp {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.Regexp))
// }
// func (f *functionValue) Array() values.Array {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.Array))
// }
// func (f *functionValue) Object() values.Object {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.Object))
// }
// func (f *functionValue) Function() values.Function {
// 	return f
// }
// func (f *functionValue) Dict() values.Dictionary {
// 	panic(values.UnexpectedKind(semantic.Function, semantic.Dictionary))
// }
// func (f *functionValue) Equal(rhs values.Value) bool {
// 	if f.Type() != rhs.Type() {
// 		return false
// 	}
// 	v, ok := rhs.(*functionValue)
// 	return ok && (f == v)
// }
