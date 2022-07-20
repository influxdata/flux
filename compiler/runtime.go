package compiler

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

type Func interface {
	Type() semantic.MonoType
	Eval(ctx context.Context, input values.Object) (values.Value, error)
}

type Evaluator interface {
	Type() semantic.MonoType
	Eval(ctx context.Context, scope Scope) (values.Value, error)
}

type compiledFn struct {
	root        Evaluator
	parentScope Scope
}

// Type returns the return type of the compiled function.
func (c compiledFn) Type() semantic.MonoType {
	return c.root.Type()
}

func (c compiledFn) Eval(ctx context.Context, input values.Object) (values.Value, error) {
	inputScope := nestScope(c.parentScope)
	input.Range(func(k string, v values.Value) {
		inputScope.Set(k, v)
		v.Retain()
	})

	defer releaseScope(inputScope)
	return eval(ctx, c.root, inputScope)
}

func releaseScope(s Scope) {
	s.LocalRange(func(k string, v values.Value) {
		v.Release()
	})
}

type Scope interface {
	values.Scope
	Get(name string) values.Value
}

type runtimeScope struct {
	values.Scope
}

func (s runtimeScope) Get(name string) values.Value {
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
	if s == nil {
		return nil
	}
	return runtimeScope{s}
}

func nestScope(scope Scope) Scope {
	return runtimeScope{scope.Nest(nil)}
}

func eval(ctx context.Context, e Evaluator, scope Scope) (values.Value, error) {
	v, err := e.Eval(ctx, scope)
	if err != nil {
		return nil, err
	}
	return v, nil
}

type blockEvaluator struct {
	t    semantic.MonoType
	body []Evaluator
}

func (e *blockEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *blockEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	for _, b := range e.body[:len(e.body)-1] {
		value, err := eval(ctx, b, scope)
		if err != nil {
			return nil, err
		}
		value.Release()
	}
	return eval(ctx, e.body[len(e.body)-1], scope)
}

type returnEvaluator struct {
	Evaluator
}

type declarationEvaluator struct {
	t    semantic.MonoType
	id   string
	init Evaluator
}

func (e *declarationEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *declarationEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	v, err := eval(ctx, e.init, scope)
	if err != nil {
		return nil, err
	}

	scope.Set(e.id, v)
	v.Retain()

	return v, nil
}

type stringExpressionEvaluator struct {
	parts []Evaluator
}

func (e *stringExpressionEvaluator) Type() semantic.MonoType {
	return semantic.BasicString
}

func (e *stringExpressionEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	var b strings.Builder
	for _, p := range e.parts {
		v, err := p.Eval(ctx, scope)
		if err != nil {
			return nil, err
		}

		if v.IsNull() {
			return nil, errors.New(codes.Invalid, "string expression evaluated to null")
		}
		b.WriteString(v.Str())
	}
	return values.NewString(b.String()), nil
}

type textEvaluator struct {
	value string
}

func (*textEvaluator) Type() semantic.MonoType {
	return semantic.BasicString
}

func (e *textEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	return values.NewString(e.value), nil
}

type interpolatedEvaluator struct {
	s Evaluator
}

func (*interpolatedEvaluator) Type() semantic.MonoType {
	return semantic.BasicString
}

func (e *interpolatedEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	o, err := e.s.Eval(ctx, scope)
	if err != nil {
		return nil, err
	}
	defer o.Release()
	return values.Stringify(o)
}

type objEvaluator struct {
	t          semantic.MonoType
	with       *identifierEvaluator
	properties map[string]Evaluator
}

func (e *objEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *objEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	return values.BuildObject(func(set values.ObjectSetter) error {
		if e.with != nil {
			with, err := e.with.Eval(ctx, scope)
			if err != nil {
				return err
			}
			if with.IsNull() {
				return errors.New(codes.Invalid, `null value on left hand side of "with" in record literal`)
			}
			if typ := with.Type().Nature(); typ != semantic.Object {
				return errors.Newf(codes.Invalid, `value on left hand side of "with" in record literal has type %s; expected record`, typ)
			}
			with.Object().Range(func(name string, v values.Value) {
				set(name, v)
			})
		}

		for k, node := range e.properties {
			v, err := eval(ctx, node, scope)
			if err != nil {
				return err
			}
			set(k, v)
		}
		return nil
	})
}

type arrayEvaluator struct {
	t     semantic.MonoType
	array []Evaluator
}

func (e *arrayEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *arrayEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	arr := values.NewArray(e.t)
	for _, ev := range e.array {
		v, err := eval(ctx, ev, scope)
		if err != nil {
			return nil, err
		}
		arr.Append(v)
	}
	return arr, nil
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

func (e *dictEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	if len(e.elements) == 0 {
		return values.NewEmptyDict(e.t), nil
	}
	builder := values.NewDictBuilder(e.t)
	for _, item := range e.elements {
		key, err := eval(ctx, item.Key, scope)
		if err != nil {
			return nil, err
		}
		val, err := eval(ctx, item.Val, scope)
		if err != nil {
			return nil, err
		}
		if err := builder.Insert(key, val); err != nil {
			return nil, err
		}
	}
	return builder.Dict(), nil
}

type logicalEvaluator struct {
	operator    ast.LogicalOperatorKind
	left, right Evaluator
}

func (e *logicalEvaluator) Type() semantic.MonoType {
	return semantic.BasicBool
}

func (e *logicalEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	l, err := e.left.Eval(ctx, scope)
	if err != nil {
		return nil, err
	}
	defer l.Release()
	if typ := l.Type().Nature(); !l.IsNull() && typ != semantic.Bool {
		return nil, errors.Newf(codes.Invalid, "cannot use operand of type %s with logical %s; expected boolean", typ, e.operator)
	}

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
		panic(errors.Newf(codes.Internal, "unknown logical operator %v", e.operator))
	}

	r, err := e.right.Eval(ctx, scope)
	if err != nil {
		return nil, err
	}
	if typ := r.Type().Nature(); !r.IsNull() && typ != semantic.Bool {
		return nil, errors.Newf(codes.Invalid, "cannot use operand of type %s with logical %s; expected boolean", typ, e.operator)
	}

	return r, nil
}

type logicalVectorEvaluator struct {
	operator    ast.LogicalOperatorKind
	left, right Evaluator
}

func (e *logicalVectorEvaluator) Type() semantic.MonoType {
	return semantic.BasicBool
}

func (e *logicalVectorEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	l, err := e.left.Eval(ctx, scope)
	if err != nil {
		return nil, err
	}
	defer l.Release()

	if typ := l.Type().Nature(); typ != semantic.Vector {
		return nil, errors.Newf(codes.Invalid, "cannot use vectorized %s operation on non-vector %s", e.operator, typ)
	} else if elemType := l.Vector().ElementType().Nature(); elemType != semantic.Bool {
		return nil, errors.Newf(codes.Invalid, "cannot use operand of type %s with logical %s; expected boolean", elemType, e.operator)
	}
	lv := l.Vector().Arr().(*array.Boolean)

	r, err := e.right.Eval(ctx, scope)
	if err != nil {
		return nil, err
	}

	if typ := r.Type().Nature(); typ != semantic.Vector {
		return nil, errors.Newf(codes.Invalid, "cannot use vectorized %s operation on non-vector %s", e.operator, typ)
	} else if elemType := r.Vector().ElementType().Nature(); elemType != semantic.Bool {
		return nil, errors.Newf(codes.Invalid, "cannot use operand of type %s with logical %s; expected boolean", elemType, e.operator)
	}
	rv := r.Vector().Arr().(*array.Boolean)

	mem := memory.GetAllocator(ctx)

	switch e.operator {
	case ast.AndOperator:
		res, err := array.And(lv, rv, mem)
		if err != nil {
			return nil, err
		}
		return values.NewVectorValue(res, semantic.BasicBool), nil
	case ast.OrOperator:
		res, err := array.Or(lv, rv, mem)
		if err != nil {
			return nil, err
		}
		return values.NewVectorValue(res, semantic.BasicBool), nil
	default:
		panic(errors.Newf(codes.Internal, "unknown logical operator %v", e.operator))
	}
}

type conditionalEvaluator struct {
	test       Evaluator
	consequent Evaluator
	alternate  Evaluator
}

func (e *conditionalEvaluator) Type() semantic.MonoType {
	return e.alternate.Type()
}

func (e *conditionalEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	t, err := eval(ctx, e.test, scope)
	if err != nil {
		return nil, err
	}
	defer t.Release()
	if typ := t.Type().Nature(); !t.IsNull() && typ != semantic.Bool {
		return nil, errors.Newf(codes.Invalid, "cannot use test of type %s in conditional expression; expected boolean", typ)
	}

	if t.IsNull() || !t.Bool() {
		return eval(ctx, e.alternate, scope)
	} else {
		return eval(ctx, e.consequent, scope)
	}
}

type conditionalVectorEvaluator struct {
	test       Evaluator
	consequent Evaluator
	alternate  Evaluator
}

func (e *conditionalVectorEvaluator) Type() semantic.MonoType {
	return e.alternate.Type()
}

func (e *conditionalVectorEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	t, err := eval(ctx, e.test, scope)
	if err != nil {
		return nil, err
	}
	defer t.Release()

	typ := t.Type()
	// Will err if typ is not Vector or Array, but that's fine.
	// We're testing first to make sure this is a Vector.
	etyp, _ := typ.ElemType()

	if !t.IsNull() && !(typ.Nature() == semantic.Vector && etyp.Nature() == semantic.Bool) {
		return nil, errors.Newf(codes.Invalid, "cannot use test of type %s in vectorized conditional expression; expected vector of boolean", typ)
	}

	// FIXME: vectorized impl here

	mem := memory.GetAllocator(ctx)

	tv := t.Vector()
	// If `t` is vec repeat, skip the varied checks and immediately select the
	// branch to return.
	// FIXME: currently there's no way to vectorize bool literals.
	//   This branch will never run until boolean literals work and/or we have
	//	 const folding working with equality operators.
	//   <https://github.com/influxdata/flux/issues/4997>
	//   <https://github.com/influxdata/flux/issues/4608>
	if vr, ok := tv.(*values.VectorRepeatValue); ok {
		if vr.Value().Bool() {
			return eval(ctx, e.consequent, scope)
		} else {
			return eval(ctx, e.alternate, scope)
		}
	}

	tva := tv.Arr().(*array.Boolean)
	n := tva.Len()

	// Scan to see if we have all one outcome for the test.
	var initialOutcome interface{}
	varied := false
	for i := 0; i < n; i++ {
		if tva.IsValid(i) {
			if initialOutcome == nil {
				initialOutcome = tva.Value(i)
			}
			if initialOutcome != tva.Value(i) {
				varied = true
				break
			}
		}
	}

	if !varied {
		if initialOutcome.(bool) {
			return eval(ctx, e.consequent, scope)
		} else {
			return eval(ctx, e.alternate, scope)
		}
	}

	c, err := eval(ctx, e.consequent, scope)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	a, err := eval(ctx, e.alternate, scope)
	if err != nil {
		return nil, err
	}
	defer a.Release()

	return values.VectorConditional(tv, c.Vector(), a.Vector(), mem)
}

type binaryEvaluator struct {
	t           semantic.MonoType
	left, right Evaluator
	f           values.BinaryFunction
}

func (e *binaryEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *binaryEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	l, err := eval(ctx, e.left, scope)
	if err != nil {
		return nil, err
	}
	defer l.Release()
	r, err := eval(ctx, e.right, scope)
	if err != nil {
		return nil, err
	}
	defer r.Release()

	return e.f(l, r)
}

type binaryVectorEvaluator struct {
	t           semantic.MonoType
	left, right Evaluator
	f           values.BinaryVectorFunction
}

func (e *binaryVectorEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *binaryVectorEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	l, err := eval(ctx, e.left, scope)
	if err != nil {
		return nil, err
	}
	defer l.Release()
	r, err := eval(ctx, e.right, scope)
	if err != nil {
		return nil, err
	}
	defer r.Release()
	mem := memory.GetAllocator(ctx)

	if mem == nil {
		return nil, errors.Newf(codes.Invalid, "missing allocator, cannot use vectorized operators")
	}

	return e.f(l, r, mem)
}

type constVectorEvaluator struct {
	t semantic.MonoType
	v Evaluator
}

func (e *constVectorEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *constVectorEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	v, err := eval(ctx, e.v, scope)
	if err != nil {
		return nil, err
	}
	v.Retain()
	return values.NewVectorRepeatValue(v), nil
}

type unaryEvaluator struct {
	t    semantic.MonoType
	node Evaluator
	op   ast.OperatorKind
}

func (e *unaryEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *unaryEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	v, err := e.node.Eval(ctx, scope)
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
			return nil, errors.Newf(codes.Internal, "unknown unary operator: %s", e.op)
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
			return values.NewDuration(v.Duration().Mul(-1)), nil
		default:
			panic(values.UnexpectedKind(e.t.Nature(), v.Type().Nature()))
		}
	}(v)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type integerEvaluator struct {
	i int64
}

func (e *integerEvaluator) Type() semantic.MonoType {
	return semantic.BasicInt
}

func (e *integerEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	return values.NewInt(e.i), nil
}

type unsignedIntegerEvaluator struct {
	i uint64
}

func (e *unsignedIntegerEvaluator) Type() semantic.MonoType {
	return semantic.BasicUint
}

func (e *unsignedIntegerEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	return values.NewUInt(e.i), nil
}

type stringEvaluator struct {
	s string
}

func (e *stringEvaluator) Type() semantic.MonoType {
	return semantic.BasicString
}

func (e *stringEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	return values.NewString(e.s), nil
}

type regexpEvaluator struct {
	r *regexp.Regexp
}

func (e *regexpEvaluator) Type() semantic.MonoType {
	return semantic.BasicRegexp
}

func (e *regexpEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	return values.NewRegexp(e.r), nil
}

type booleanEvaluator struct {
	b bool
}

func (e *booleanEvaluator) Type() semantic.MonoType {
	return semantic.BasicBool
}

func (e *booleanEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	return values.NewBool(e.b), nil
}

type floatEvaluator struct {
	f float64
}

func (e *floatEvaluator) Type() semantic.MonoType {
	return semantic.BasicFloat
}

func (e *floatEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	return values.NewFloat(e.f), nil
}

type timeEvaluator struct {
	time values.Time
}

func (e *timeEvaluator) Type() semantic.MonoType {
	return semantic.BasicTime
}

func (e *timeEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	return values.NewTime(e.time), nil
}

type durationEvaluator struct {
	duration values.Duration
}

func (e *durationEvaluator) Type() semantic.MonoType {
	return semantic.BasicDuration
}

func (e *durationEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	return values.NewDuration(e.duration), nil
}

type identifierEvaluator struct {
	t    semantic.MonoType
	name string
}

func (e *identifierEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *identifierEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	v := scope.Get(e.name)
	v.Retain()
	return v, nil
}

type memberEvaluator struct {
	t        semantic.MonoType
	object   Evaluator
	property string
	nullable bool
}

func (e *memberEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *memberEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	o, err := e.object.Eval(ctx, scope)
	if err != nil {
		return nil, err
	}
	defer o.Release()
	if o.IsNull() {
		return nil, errors.Newf(codes.Invalid, "cannot access property of a null value; expected record")
	}
	if typ := o.Type().Nature(); typ != semantic.Object {
		return nil, errors.Newf(codes.Invalid, "cannot access property of a value with type %s; expected record", typ)
	}

	v, ok := o.Object().Get(e.property)
	if !ok && !e.nullable {
		return nil, errors.Newf(codes.Invalid, "member %q with type %s is not in the record", e.property, e.t.Nature())
	}
	v.Retain()
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

func (e *arrayIndexEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	a, err := e.array.Eval(ctx, scope)
	if err != nil {
		return nil, err
	}
	defer a.Release()
	if a.IsNull() {
		return nil, errors.New(codes.Invalid, "cannot index into a null value; expected an array")
	}
	if typ := a.Type().Nature(); typ != semantic.Array {
		return nil, errors.Newf(codes.Invalid, "cannot index into a value of type %s; expected an array", typ)
	}
	i, err := e.index.Eval(ctx, scope)
	if err != nil {
		return nil, err
	}
	defer i.Release()
	if i.IsNull() {
		return nil, errors.New(codes.Invalid, "cannot index into an array with null value; expected an int")
	}
	if typ := i.Type().Nature(); typ != semantic.Int {
		return nil, errors.Newf(codes.Invalid, "cannot index into an array with value of type %s; expected an int", typ)
	}
	ix := int(i.Int())
	l := a.Array().Len()
	if ix < 0 || ix >= l {
		return nil, errors.Newf(codes.OutOfRange, "cannot access element %v of array of length %v", ix, l)
	}
	v := a.Array().Get(ix)
	v.Retain()
	return v, nil
}

type callEvaluator struct {
	t      semantic.MonoType
	callee Evaluator
	args   Evaluator
}

func (e *callEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *callEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	args, err := e.args.Eval(ctx, scope)
	if err != nil {
		return nil, err
	}
	defer args.Release()
	f, err := e.callee.Eval(ctx, scope)
	if err != nil {
		return nil, err
	}
	defer f.Release()
	if f.IsNull() {
		return nil, errors.Newf(codes.Invalid, "attempt to call a null value; expected function")
	}
	if typ := f.Type().Nature(); typ != semantic.Function {
		return nil, errors.Newf(codes.Invalid, "attempt to call a value of type %s; expected function", typ)
	}

	return f.Function().Call(ctx, args.Object())
}

type functionEvaluator struct {
	t      semantic.MonoType
	fn     *semantic.FunctionExpression
	params []functionParam
}

func (e *functionEvaluator) Type() semantic.MonoType {
	return e.t
}

func (e *functionEvaluator) Eval(ctx context.Context, scope Scope) (values.Value, error) {
	return &functionValue{
		t:      e.t,
		fn:     e.fn,
		params: e.params,
		scope:  scope,
	}, nil
}

type functionValue struct {
	t      semantic.MonoType
	fn     *semantic.FunctionExpression
	params []functionParam
	scope  Scope
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
	scope := nestScope(f.scope)
	for _, p := range f.params {
		a, ok := args.Get(p.Key)
		if !ok && p.Default != nil {
			v, err := eval(ctx, p.Default, f.scope)
			if err != nil {
				return nil, err
			}
			a = v
		} else {
			a.Retain()
		}
		scope.Set(p.Key, a)
	}
	defer releaseScope(scope)

	fn, err := Compile(scope, f.fn, args.Type())
	if err != nil {
		return nil, err
	}
	return fn.Eval(ctx, args)
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
func (f *functionValue) Vector() values.Vector {
	panic(values.UnexpectedKind(semantic.Function, semantic.Vector))
}
func (f *functionValue) Equal(rhs values.Value) bool {
	if f.Type() != rhs.Type() {
		return false
	}
	v, ok := rhs.(*functionValue)
	return ok && (f == v)
}

func (f *functionValue) Retain()  {}
func (f *functionValue) Release() {}
