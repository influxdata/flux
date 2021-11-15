package compiler

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type Func interface {
	Type() semantic.MonoType
	Eval(ctx context.Context, input []Value) (Value, error)
}

type Evaluator interface {
	Type() semantic.MonoType
	Eval(ctx context.Context, scope []Value, origin int) error
}

type compiledFn struct {
	t      semantic.MonoType
	params []valueMapper
	body   []Evaluator
	scope  Scope
	ret    int
}

// Type returns the return type of the compiled function.
func (c compiledFn) Type() semantic.MonoType {
	return c.t
}

func (c compiledFn) Eval(ctx context.Context, input []Value) (Value, error) {
	// TODO(jsternberg): I can probably remove this allocation?
	// Revisit after I get to implementing closures and function calls.
	scope := make([]Value, len(c.scope.Values()))
	copy(scope, c.scope.Values())
	for i := 0; i < len(input); i++ {
		if c.params[i] != nil {
			c.params[i].Map(input[i], scope)
		}
	}

	// Origin marks the latest jump location.
	// This gets initialized with the last executed instruction.
	// As we start at the entrypoint, we have never branched
	// As we start at the first instruction, our initial origin
	// is zero.
	origin := 0
	for i, n := 0, len(c.body); i < n; i++ {
		if err := c.body[i].Eval(ctx, scope, origin); err != nil {
			if jmp, ok := err.(jumpInterrupt); ok {
				// We received a jump interrupt signal.
				// Update the instruction pointer and the origin
				// to the new location where we jumped.
				i, origin = int(jmp)-1, i
				continue
			}
			return Value{}, err
		}
	}
	return scope[c.ret], nil
}

type evaluator struct {
	t   semantic.MonoType
	ret int
}

func (e *evaluator) Type() semantic.MonoType {
	return e.t
}

func (e *evaluator) Return() int {
	return e.ret
}

type stringPartEvaluator interface {
	Eval(ctx context.Context, scope []Value, b *strings.Builder) error
}

type stringExpressionEvaluator struct {
	evaluator
	parts []stringPartEvaluator
}

func (e *stringExpressionEvaluator) Eval(ctx context.Context, scope []Value, origin int) error {
	var b strings.Builder
	for _, p := range e.parts {
		if err := p.Eval(ctx, scope, &b); err != nil {
			return err
		}
	}
	scope[e.ret] = NewString(b.String())
	return nil
}

type textEvaluator struct {
	value string
}

func (e *textEvaluator) Eval(ctx context.Context, scope []Value, b *strings.Builder) error {
	b.WriteString(e.value)
	return nil
}

type interpolatedEvaluator struct {
	index int
}

func (e *interpolatedEvaluator) Eval(ctx context.Context, scope []Value, b *strings.Builder) error {
	o := scope[e.index]
	if o.IsNull() {
		return errors.New(codes.Invalid, "string expression evaluated to null")
	}
	v, err := stringify(o)
	if err != nil {
		return err
	}
	b.WriteString(v.Str())
	return nil
}

type recordEvaluator struct {
	evaluator
	with   int
	labels []int
}

func (e *recordEvaluator) Eval(ctx context.Context, scope []Value, origin int) error {
	// Contrary to most of the values in the scope, the record's
	// return value is pre-constructed with the default values in the scope.
	// We access it and set the attributes. This is mostly the only place where we do this.
	// This is because records are _very_ expensive to construct and used _very_ often.
	// This is in contrast to arrays which aren't constructed dynamically nearly as often.
	record := scope[e.ret]
	for i, label := range e.labels {
		if label >= 0 {
			record.Set(i, scope[label])
		}
	}

	// If there was a with clause, we copy each
	// of the values from that record into this one.
	// The type includes the with attributes after
	// the new attributes even if we executed with before
	// the attributes. This will also copy over the shadowed
	// fields that no longer "exist" anymore.
	if e.with >= 0 {
		offset := len(e.labels)
		with := scope[e.with]
		for i, n := 0, with.NumFields(); i < n; i++ {
			record.Set(i+offset, with.Get(i))
		}
	}
	return nil
}

type arrayEvaluator struct {
	evaluator
	array []int
}

func (e *arrayEvaluator) Eval(ctx context.Context, scope []Value, origin int) error {
	elements := make([]Value, len(e.array))
	for i, ev := range e.array {
		elements[i] = scope[ev]
	}
	scope[e.ret] = NewArray(e.t, elements)
	return nil
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
	// 	return NewDict(e.t), nil
	// }
	// builder := NewDictBuilder(e.t)
	// for _, item := range e.elements {
	// 	key, err := eval(ctx, item.Key, scope)
	// 	if err != nil {
	// 		return Value{}, err
	// 	}
	// 	val, err := eval(ctx, item.Val, scope)
	// 	if err != nil {
	// 		return Value{}, err
	// 	}
	// 	if err := builder.Insert(key, val); err != nil {
	// 		return Value{}, err
	// 	}
	// }
	// return builder.Dict(), nil
	return Value{}, errors.New(codes.Unimplemented)
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
	// l, err := eval(ctx, e.left, scope)
	// if err != nil {
	// 	return Value{}, err
	// }
	// r, err := eval(ctx, e.right, scope)
	// if err != nil {
	// 	return Value{}, err
	// }
	// return e.f(l, r)
	return Value{}, errors.New(codes.Unimplemented)
}

type unaryNotEvaluator struct {
	evaluator
	index int
}

func (e *unaryNotEvaluator) Eval(ctx context.Context, scope []Value, origin int) error {
	v := scope[e.index]
	if v.IsNull() {
		scope[e.ret] = v
		return nil
	}

	// TODO(jsternberg): This is presently dynamic, but it would definitely
	// be possible to use the type known ahead of time to have this be
	// non-dynamic (if possible).
	switch e.t.Nature() {
	case semantic.Int:
		scope[e.ret] = NewInt(-v.Int())
	case semantic.Float:
		scope[e.ret] = NewFloat(-v.Float())
	case semantic.Bool:
		scope[e.ret] = NewBool(!v.Bool())
	case semantic.Duration:
		scope[e.ret] = NewDuration(v.Duration().Mul(-1))
	default:
		panic(values.UnexpectedKind(e.t.Nature(), v.Nature()))
	}
	return nil
}

type unaryExistsEvaluator struct {
	evaluator
	index int
}

func (e *unaryExistsEvaluator) Eval(ctx context.Context, scope []Value, origin int) error {
	v := scope[e.index]
	scope[e.ret] = NewBool(v.IsValid())
	return nil
}

type regexpMatchEvaluator struct {
	evaluator
	left, right int
}

func (e *regexpMatchEvaluator) Eval(ctx context.Context, scope []Value, origin int) error {
	lv, rv := scope[e.left], scope[e.right]
	v := rv.Regexp().MatchString(lv.Str())
	scope[e.ret] = NewBool(v)
	return nil
}

type regexpNotMatchEvaluator struct {
	evaluator
	left, right int
}

func (e *regexpNotMatchEvaluator) Eval(ctx context.Context, scope []Value, origin int) error {
	lv, rv := scope[e.left], scope[e.right]
	v := rv.Regexp().MatchString(lv.Str())
	scope[e.ret] = NewBool(!v)
	return nil
}

type memberEvaluator struct {
	evaluator
	object   int
	property int
}

func (e *memberEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *memberEvaluator) Eval(ctx context.Context, scope []Value, origin int) error {
	o := scope[e.object]
	if o.IsNull() {
		return errors.Newf(codes.Invalid, "cannot access property of a null value; expected record")
	}
	value := o.Get(e.property)
	scope[e.ret] = value
	return nil
}

type arrayIndexEvaluator struct {
	evaluator
	array int
	index int
}

func (e *arrayIndexEvaluator) Eval(ctx context.Context, scope []Value, origin int) error {
	i := scope[e.index]
	if i.IsNull() {
		return errors.New(codes.Invalid, "cannot index into an array with null value; expected an int")
	}
	a := scope[e.array]
	ix := int(i.Int())
	if l := a.Len(); ix < 0 || ix >= l {
		return errors.Newf(codes.OutOfRange, "cannot access element %v of array of length %v", ix, l)
	}
	scope[e.ret] = a.Index(ix)
	return nil
}

type callEvaluator struct {
	t      semantic.MonoType
	callee Evaluator
	args   []Evaluator
}

func (e *callEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *callEvaluator) Eval(ctx context.Context, scope Scope) (Value, error) {
	// f, err := e.callee.Eval(ctx, scope)
	// if err != nil {
	// 	return Value{}, err
	// }
	//
	// if f.IsNull() {
	// 	return Value{}, errors.Newf(codes.Invalid, "attempt to call a null value; expected function")
	// }
	// if typ := f.Nature(); typ != semantic.Function {
	// 	return Value{}, errors.Newf(codes.Invalid, "attempt to call a value of type %s; expected function", typ)
	// }
	//
	// vargs := make([]MaybeValue, len(e.args))
	// for i, arg := range e.args {
	// 	v, err := eval(ctx, arg, scope)
	// 	if err != nil {
	// 		return Value{}, err
	// 	}
	// 	vargs[i] = SomeValue(v)
	// }
	// return f.Call(ctx, vargs)
	return Value{}, errors.New(codes.Unimplemented)
}

type functionEvaluator struct {
	t      semantic.MonoType
	fn     *semantic.FunctionExpression
	params []functionParam
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

type functionValue struct {
	t      semantic.MonoType
	fn     *semantic.FunctionExpression
	params []functionParam
	scope  values.Scope
}

// functionValue implements the interpreter.Resolver interface.
var _ interpreter.Resolver = (*functionValue)(nil)

func (f *functionValue) Resolve() (semantic.Node, error) {
	n := f.fn.Copy()
	localIdentifiers := make([]string, 0, 10)
	node, err := interpreter.ResolveIdsInFunction(f.scope, f.fn, n, &localIdentifiers)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (f functionValue) Scope() values.Scope {
	return f.scope
}

type functionParam struct {
	Key     string
	Default Evaluator
	Type    semantic.MonoType
}

func (f *functionValue) HasSideEffect() bool {
	return false
}

func (f *functionValue) Call(ctx context.Context, args values.Object) (values.Value, error) {
	// scope := nestScope(f.scope)
	// for _, p := range f.params {
	// 	a, ok := args.Get(p.Key)
	// 	if !ok && p.Default != nil {
	// 		v, err := eval(ctx, p.Default, f.scope)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		a = v
	// 	}
	// 	scope.Set(p.Key, a)
	// }
	//
	// fn, err := Compile(scope, f.fn, args.Type())
	// if err != nil {
	// 	return nil, err
	// }
	// return fn.Eval(ctx, args)
	return nil, errors.New(codes.Unimplemented)
}

func (f *functionValue) Type() semantic.MonoType { return f.t }
func (f *functionValue) IsNull() bool            { return false }
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
func (f *functionValue) Dict() values.Dictionary {
	panic(values.UnexpectedKind(semantic.Function, semantic.Dictionary))
}
func (f *functionValue) Equal(rhs values.Value) bool {
	if f.Type() != rhs.Type() {
		return false
	}
	v, ok := rhs.(*functionValue)
	return ok && (f == v)
}

type jumpInterrupt int

func (e jumpInterrupt) Error() string {
	return fmt.Sprintf("jump interrupt: %d", e)
}

type branch struct {
	test int
	t, f int
}

func (b *branch) Type() semantic.MonoType {
	return semantic.MonoType{}
}

func (b *branch) Eval(ctx context.Context, scope []Value, origin int) error {
	test := scope[b.test]
	if test.IsNull() || !test.Bool() {
		return jumpInterrupt(b.f)
	} else {
		return jumpInterrupt(b.t)
	}
}

type jump struct {
	to int
}

func (j *jump) Type() semantic.MonoType {
	return semantic.MonoType{}
}

func (j *jump) Eval(ctx context.Context, scope []Value, origin int) error {
	return jumpInterrupt(j.to)
}

// phi node is used to combine two branches back to a single value.
// It selects one of the values based on the origin branch instruction.
type phi struct {
	evaluator
	label1, index1 int
	label2, index2 int
}

func (p *phi) Eval(ctx context.Context, scope []Value, origin int) error {
	if p.label1 == origin {
		scope[p.ret] = scope[p.index1]
	} else if p.label2 == origin {
		scope[p.ret] = scope[p.index2]
	} else {
		return errors.New(codes.Internal, "invalid origin address for phi node")
	}
	return nil
}
