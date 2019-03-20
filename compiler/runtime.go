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
	EvalString(input values.Object) (string, error)
	EvalInt(input values.Object) (int64, error)
	EvalUInt(input values.Object) (uint64, error)
	EvalFloat(input values.Object) (float64, error)
	EvalBool(input values.Object) (bool, error)
	EvalTime(input values.Object) (values.Time, error)
	EvalDuration(input values.Object) (values.Duration, error)
	EvalRegexp(input values.Object) (*regexp.Regexp, error)
	EvalArray(input values.Object) (values.Array, error)
	EvalObject(input values.Object) (values.Object, error)
	EvalFunction(input values.Object) (values.Function, error)
}

type Evaluator interface {
	Type() semantic.Type
	EvalString(scope Scope) (string, error)
	EvalInt(scope Scope) (int64, error)
	EvalUInt(scope Scope) (uint64, error)
	EvalFloat(scope Scope) (float64, error)
	EvalBool(scope Scope) (bool, error)
	EvalTime(scope Scope) (values.Time, error)
	EvalDuration(scope Scope) (values.Duration, error)
	EvalRegexp(scope Scope) (*regexp.Regexp, error)
	EvalArray(scope Scope) (values.Array, error)
	EvalObject(scope Scope) (values.Object, error)
	EvalFunction(scope Scope) (values.Function, error)
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
		if properties[k] != v {
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

func (c compiledFn) EvalString(input values.Object) (string, error) {
	if err := c.buildScope(input); err != nil {
		return "", err
	}
	return c.root.EvalString(c.inputScope)
}
func (c compiledFn) EvalBool(input values.Object) (bool, error) {
	if err := c.buildScope(input); err != nil {
		return false, err
	}
	return c.root.EvalBool(c.inputScope)
}
func (c compiledFn) EvalInt(input values.Object) (int64, error) {
	if err := c.buildScope(input); err != nil {
		return 0, err
	}
	return c.root.EvalInt(c.inputScope)
}
func (c compiledFn) EvalUInt(input values.Object) (uint64, error) {
	if err := c.buildScope(input); err != nil {
		return 0, err
	}
	return c.root.EvalUInt(c.inputScope)
}
func (c compiledFn) EvalFloat(input values.Object) (float64, error) {
	if err := c.buildScope(input); err != nil {
		return 0, err
	}
	return c.root.EvalFloat(c.inputScope)
}
func (c compiledFn) EvalTime(input values.Object) (values.Time, error) {
	if err := c.buildScope(input); err != nil {
		return 0, err
	}
	return c.root.EvalTime(c.inputScope)
}
func (c compiledFn) EvalDuration(input values.Object) (values.Duration, error) {
	if err := c.buildScope(input); err != nil {
		return 0, err
	}
	return c.root.EvalDuration(c.inputScope)
}
func (c compiledFn) EvalRegexp(input values.Object) (*regexp.Regexp, error) {
	if err := c.buildScope(input); err != nil {
		return nil, err
	}
	return c.root.EvalRegexp(c.inputScope)
}
func (c compiledFn) EvalArray(input values.Object) (values.Array, error) {
	if err := c.buildScope(input); err != nil {
		return nil, err
	}
	return c.root.EvalArray(c.inputScope)
}
func (c compiledFn) EvalObject(input values.Object) (values.Object, error) {
	if err := c.buildScope(input); err != nil {
		return nil, err
	}
	return c.root.EvalObject(c.inputScope)
}
func (c compiledFn) EvalFunction(input values.Object) (values.Function, error) {
	if err := c.buildScope(input); err != nil {
		return nil, err
	}
	return c.root.EvalFunction(c.inputScope)
}

type Scope map[string]values.Value

func (s Scope) Type(name string) semantic.Type {
	return s[name].Type()
}
func (s Scope) Set(name string, v values.Value) {
	s[name] = v
}

func (s Scope) GetString(name string) string {
	return s[name].Str()
}
func (s Scope) GetInt(name string) int64 {
	return s[name].Int()
}
func (s Scope) GetUInt(name string) uint64 {
	return s[name].UInt()
}
func (s Scope) GetFloat(name string) float64 {
	return s[name].Float()
}
func (s Scope) GetBool(name string) bool {
	return s[name].Bool()
}
func (s Scope) GetTime(name string) values.Time {
	return s[name].Time()
}
func (s Scope) GetDuration(name string) values.Duration {
	return s[name].Duration()
}
func (s Scope) GetRegexp(name string) *regexp.Regexp {
	return s[name].Regexp()
}
func (s Scope) GetArray(name string) values.Array {
	return s[name].Array()
}
func (s Scope) GetObject(name string) values.Object {
	return s[name].Object()
}
func (s Scope) GetFunction(name string) values.Function {
	return s[name].Function()
}

func (s Scope) Copy() Scope {
	n := make(Scope, len(s))
	for k, v := range s {
		n[k] = v
	}
	return n
}

func eval(e Evaluator, scope Scope) (values.Value, error) {
	var v values.Value
	var err error
	switch e.Type().Nature() {
	case semantic.String:
		var v0 string
		v0, err = e.EvalString(scope)
		if err == nil {
			v = values.NewString(v0)
		}
	case semantic.Int:
		var v0 int64
		v0, err = e.EvalInt(scope)
		if err == nil {
			v = values.NewInt(v0)
		}
	case semantic.UInt:
		var v0 uint64
		v0, err = e.EvalUInt(scope)
		if err == nil {
			v = values.NewUInt(v0)
		}
	case semantic.Float:
		var v0 float64
		v0, err = e.EvalFloat(scope)
		if err == nil {
			v = values.NewFloat(v0)
		}
	case semantic.Bool:
		var v0 bool
		v0, err = e.EvalBool(scope)
		if err == nil {
			v = values.NewBool(v0)
		}
	case semantic.Time:
		var v0 values.Time
		v0, err = e.EvalTime(scope)
		if err == nil {
			v = values.NewTime(v0)
		}
	case semantic.Duration:
		var v0 values.Duration
		v0, err = e.EvalDuration(scope)
		if err == nil {
			v = values.NewDuration(v0)
		}
	case semantic.Regexp:
		var v0 *regexp.Regexp
		v0, err = e.EvalRegexp(scope)
		if err == nil {
			v = values.NewRegexp(v0)
		}
	case semantic.Array:
		v, err = e.EvalArray(scope)
	case semantic.Object:
		v, err = e.EvalObject(scope)
	case semantic.Function:
		v, err = e.EvalFunction(scope)
	case semantic.Nil:
		return nil, nil
	default:
		err = fmt.Errorf("eval: unknown type: %v", e.Type())
	}

	return v, err
}

type blockEvaluator struct {
	t     semantic.Type
	body  []Evaluator
	value values.Value
}

func (e *blockEvaluator) Type() semantic.Type {
	return e.t
}

func (e *blockEvaluator) eval(scope Scope) error {
	var err error
	for _, b := range e.body {
		e.value, err = eval(b, scope)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *blockEvaluator) EvalString(scope Scope) (string, error) {
	values.CheckKind(e.t.Nature(), semantic.String)
	err := e.eval(scope)
	if err != nil {
		return "", err
	}
	return e.value.Str(), nil
}
func (e *blockEvaluator) EvalInt(scope Scope) (int64, error) {
	values.CheckKind(e.t.Nature(), semantic.Int)
	err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return e.value.Int(), nil
}
func (e *blockEvaluator) EvalUInt(scope Scope) (uint64, error) {
	values.CheckKind(e.t.Nature(), semantic.UInt)
	err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return e.value.UInt(), nil
}
func (e *blockEvaluator) EvalFloat(scope Scope) (float64, error) {
	values.CheckKind(e.t.Nature(), semantic.Float)
	err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return e.value.Float(), nil
}
func (e *blockEvaluator) EvalBool(scope Scope) (bool, error) {
	values.CheckKind(e.t.Nature(), semantic.Bool)
	err := e.eval(scope)
	if err != nil {
		return false, err
	}
	return e.value.Bool(), nil
}
func (e *blockEvaluator) EvalTime(scope Scope) (values.Time, error) {
	values.CheckKind(e.t.Nature(), semantic.Time)
	err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return e.value.Time(), nil
}
func (e *blockEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	values.CheckKind(e.t.Nature(), semantic.Duration)
	err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return e.value.Duration(), nil
}
func (e *blockEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	values.CheckKind(e.t.Nature(), semantic.Regexp)
	err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return e.value.Regexp(), nil
}
func (e *blockEvaluator) EvalArray(scope Scope) (values.Array, error) {
	values.CheckKind(e.t.Nature(), semantic.Object)
	err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return e.value.Array(), nil
}
func (e *blockEvaluator) EvalObject(scope Scope) (values.Object, error) {
	values.CheckKind(e.t.Nature(), semantic.Object)
	err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return e.value.Object(), nil
}
func (e *blockEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	values.CheckKind(e.t.Nature(), semantic.Object)
	err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return e.value.Function(), nil
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

func (e *declarationEvaluator) eval(scope Scope) error {
	v, err := eval(e.init, scope)
	if err != nil {
		return err
	}

	scope.Set(e.id, v)
	return nil
}

func (e *declarationEvaluator) EvalString(scope Scope) (string, error) {
	err := e.eval(scope)
	if err != nil {
		return "", err
	}
	return scope.GetString(e.id), nil
}
func (e *declarationEvaluator) EvalInt(scope Scope) (int64, error) {
	err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return scope.GetInt(e.id), nil
}
func (e *declarationEvaluator) EvalUInt(scope Scope) (uint64, error) {
	err := e.eval(scope)
	if err != nil {
		return 0, err
	}

	return scope.GetUInt(e.id), nil
}
func (e *declarationEvaluator) EvalFloat(scope Scope) (float64, error) {
	err := e.eval(scope)
	if err != nil {
		return 0.0, err
	}

	return scope.GetFloat(e.id), nil
}
func (e *declarationEvaluator) EvalBool(scope Scope) (bool, error) {
	err := e.eval(scope)
	if err != nil {
		return false, err
	}
	return scope.GetBool(e.id), nil
}
func (e *declarationEvaluator) EvalTime(scope Scope) (values.Time, error) {
	err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return scope.GetTime(e.id), nil
}
func (e *declarationEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	err := e.eval(scope)
	if err != nil {
		return 0, err
	}

	return scope.GetDuration(e.id), nil
}
func (e *declarationEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return scope.GetRegexp(e.id), nil
}
func (e *declarationEvaluator) EvalArray(scope Scope) (values.Array, error) {
	err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return scope.GetArray(e.id), nil
}
func (e *declarationEvaluator) EvalObject(scope Scope) (values.Object, error) {
	err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return scope.GetObject(e.id), nil
}
func (e *declarationEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return scope.GetFunction(e.id), nil
}

type objEvaluator struct {
	t          semantic.Type
	properties map[string]Evaluator
}

func (e *objEvaluator) Type() semantic.Type {
	return e.t
}

func (e *objEvaluator) EvalString(scope Scope) (string, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.String))
}
func (e *objEvaluator) EvalInt(scope Scope) (int64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Int))
}
func (e *objEvaluator) EvalUInt(scope Scope) (uint64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.UInt))
}
func (e *objEvaluator) EvalFloat(scope Scope) (float64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Float))
}
func (e *objEvaluator) EvalBool(scope Scope) (bool, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Bool))
}
func (e *objEvaluator) EvalTime(scope Scope) (values.Time, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Time))
}
func (e *objEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Duration))
}
func (e *objEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *objEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *objEvaluator) EvalObject(scope Scope) (values.Object, error) {
	obj := values.NewObject()
	for k, node := range e.properties {
		v, err := eval(node, scope)
		if err != nil {
			return nil, err
		}
		obj.Set(k, v)
	}
	return obj, nil
}
func (e *objEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type arrayEvaluator struct {
	t     semantic.Type
	array []Evaluator
}

func (e *arrayEvaluator) Type() semantic.Type {
	return e.t
}

func (e *arrayEvaluator) EvalString(scope Scope) (string, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.String))
}
func (e *arrayEvaluator) EvalInt(scope Scope) (int64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Int))
}
func (e *arrayEvaluator) EvalUInt(scope Scope) (uint64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.UInt))
}
func (e *arrayEvaluator) EvalFloat(scope Scope) (float64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Float))
}
func (e *arrayEvaluator) EvalBool(scope Scope) (bool, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Bool))
}
func (e *arrayEvaluator) EvalTime(scope Scope) (values.Time, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Time))
}
func (e *arrayEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Duration))
}
func (e *arrayEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *arrayEvaluator) EvalArray(scope Scope) (values.Array, error) {
	arr := values.NewArray(e.t)
	for _, ev := range e.array {
		v, err := eval(ev, scope)
		if err != nil {
			return nil, err
		}
		arr.Append(v)
	}
	return arr, nil
}
func (e *arrayEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *arrayEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type logicalEvaluator struct {
	t           semantic.Type
	operator    ast.LogicalOperatorKind
	left, right Evaluator
}

func (e *logicalEvaluator) Type() semantic.Type {
	return e.t
}

func (e *logicalEvaluator) EvalString(scope Scope) (string, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.String))
}
func (e *logicalEvaluator) EvalInt(scope Scope) (int64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Int))
}
func (e *logicalEvaluator) EvalUInt(scope Scope) (uint64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.UInt))
}
func (e *logicalEvaluator) EvalFloat(scope Scope) (float64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Float))
}
func (e *logicalEvaluator) EvalBool(scope Scope) (bool, error) {
	l, err := e.left.EvalBool(scope)
	if err != nil {
		return false, err
	}

	switch e.operator {
	case ast.AndOperator:
		if !l {
			return false, nil
		}
	case ast.OrOperator:
		if l {
			return true, nil
		}
	default:
		panic(fmt.Errorf("unknown logical operator %v", e.operator))
	}

	r, err := e.right.EvalBool(scope)
	if err != nil {
		return false, err
	}
	return r, nil
}
func (e *logicalEvaluator) EvalTime(scope Scope) (values.Time, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Time))
}
func (e *logicalEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Duration))
}
func (e *logicalEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *logicalEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *logicalEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *logicalEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type binaryEvaluator struct {
	t           semantic.Type
	left, right Evaluator
	f           values.BinaryFunction
}

func (e *binaryEvaluator) Type() semantic.Type {
	return e.t
}

func (e *binaryEvaluator) eval(scope Scope) (values.Value, values.Value, error) {
	l, err := eval(e.left, scope)
	if err != nil {
		return nil, nil, err
	}
	r, err := eval(e.right, scope)
	if err != nil {
		return nil, nil, err
	}
	return l, r, nil
}

func (e *binaryEvaluator) EvalString(scope Scope) (string, error) {
	l, r, err := e.eval(scope)
	if err != nil {
		return "", err
	}
	return e.f(l, r).Str(), nil
}
func (e *binaryEvaluator) EvalInt(scope Scope) (int64, error) {
	l, r, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return e.f(l, r).Int(), nil
}
func (e *binaryEvaluator) EvalUInt(scope Scope) (uint64, error) {
	l, r, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return e.f(l, r).UInt(), nil
}
func (e *binaryEvaluator) EvalFloat(scope Scope) (float64, error) {
	l, r, err := e.eval(scope)
	if err != nil {
		return 0.0, err
	}
	return e.f(l, r).Float(), nil
}
func (e *binaryEvaluator) EvalBool(scope Scope) (bool, error) {
	l, r, err := e.eval(scope)
	if err != nil {
		return false, err
	}
	return e.f(l, r).Bool(), nil
}
func (e *binaryEvaluator) EvalTime(scope Scope) (values.Time, error) {
	l, r, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return e.f(l, r).Time(), nil
}
func (e *binaryEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	l, r, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return e.f(l, r).Duration(), nil
}
func (e *binaryEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *binaryEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *binaryEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *binaryEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type unaryEvaluator struct {
	t    semantic.Type
	node Evaluator
}

func (e *unaryEvaluator) Type() semantic.Type {
	return e.t
}

func (e *unaryEvaluator) EvalString(scope Scope) (string, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.String))
}
func (e *unaryEvaluator) EvalInt(scope Scope) (int64, error) {
	v, err := e.node.EvalInt(scope)
	if err != nil {
		return 0, err
	}
	// There is only one integer unary operator
	return -v, nil
}
func (e *unaryEvaluator) EvalUInt(scope Scope) (uint64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.UInt))
}
func (e *unaryEvaluator) EvalFloat(scope Scope) (float64, error) {
	v, err := e.node.EvalFloat(scope)
	if err != nil {
		return 0, err
	}
	// There is only one float unary operator
	return -v, nil
}
func (e *unaryEvaluator) EvalBool(scope Scope) (bool, error) {
	v, err := e.node.EvalBool(scope)
	if err != nil {
		return false, err
	}
	// There is only one bool unary operator
	return !v, nil
}
func (e *unaryEvaluator) EvalTime(scope Scope) (values.Time, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Time))
}
func (e *unaryEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	v, err := e.node.EvalDuration(scope)
	if err != nil {
		return 0, err
	}
	// There is only one duration unary operator
	return -v, nil
}
func (e *unaryEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *unaryEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *unaryEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *unaryEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type integerEvaluator struct {
	t semantic.Type
	i int64
}

func (e *integerEvaluator) Type() semantic.Type {
	return e.t
}

func (e *integerEvaluator) EvalString(scope Scope) (string, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.String))
}
func (e *integerEvaluator) EvalInt(scope Scope) (int64, error) {
	return e.i, nil
}
func (e *integerEvaluator) EvalUInt(scope Scope) (uint64, error) {
	return uint64(e.i), nil
}
func (e *integerEvaluator) EvalFloat(scope Scope) (float64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Float))
}
func (e *integerEvaluator) EvalBool(scope Scope) (bool, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Bool))
}
func (e *integerEvaluator) EvalTime(scope Scope) (values.Time, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Time))
}
func (e *integerEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Duration))
}
func (e *integerEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *integerEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *integerEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *integerEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type stringEvaluator struct {
	t semantic.Type
	s string
}

func (e *stringEvaluator) Type() semantic.Type {
	return e.t
}

func (e *stringEvaluator) EvalString(scope Scope) (string, error) {
	return e.s, nil
}
func (e *stringEvaluator) EvalInt(scope Scope) (int64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Int))
}
func (e *stringEvaluator) EvalUInt(scope Scope) (uint64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.UInt))
}
func (e *stringEvaluator) EvalFloat(scope Scope) (float64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Float))
}
func (e *stringEvaluator) EvalBool(scope Scope) (bool, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Bool))
}
func (e *stringEvaluator) EvalTime(scope Scope) (values.Time, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Time))
}
func (e *stringEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Duration))
}
func (e *stringEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *stringEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *stringEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *stringEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type regexpEvaluator struct {
	t semantic.Type
	r *regexp.Regexp
}

func (e *regexpEvaluator) Type() semantic.Type {
	return e.t
}

func (e *regexpEvaluator) EvalString(scope Scope) (string, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.String))
}
func (e *regexpEvaluator) EvalInt(scope Scope) (int64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Int))
}
func (e *regexpEvaluator) EvalUInt(scope Scope) (uint64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.UInt))
}
func (e *regexpEvaluator) EvalFloat(scope Scope) (float64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Float))
}
func (e *regexpEvaluator) EvalBool(scope Scope) (bool, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Bool))
}
func (e *regexpEvaluator) EvalTime(scope Scope) (values.Time, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Time))
}
func (e *regexpEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Duration))
}
func (e *regexpEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	return e.r, nil
}
func (e *regexpEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *regexpEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *regexpEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type booleanEvaluator struct {
	t semantic.Type
	b bool
}

func (e *booleanEvaluator) Type() semantic.Type {
	return e.t
}

func (e *booleanEvaluator) EvalString(scope Scope) (string, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.String))
}
func (e *booleanEvaluator) EvalInt(scope Scope) (int64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Int))
}
func (e *booleanEvaluator) EvalUInt(scope Scope) (uint64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.UInt))
}
func (e *booleanEvaluator) EvalFloat(scope Scope) (float64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Float))
}
func (e *booleanEvaluator) EvalBool(scope Scope) (bool, error) {
	return e.b, nil
}
func (e *booleanEvaluator) EvalTime(scope Scope) (values.Time, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Time))
}
func (e *booleanEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Duration))
}
func (e *booleanEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *booleanEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *booleanEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *booleanEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type floatEvaluator struct {
	t semantic.Type
	f float64
}

func (e *floatEvaluator) Type() semantic.Type {
	return e.t
}

func (e *floatEvaluator) EvalString(scope Scope) (string, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.String))
}
func (e *floatEvaluator) EvalInt(scope Scope) (int64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Int))
}
func (e *floatEvaluator) EvalUInt(scope Scope) (uint64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.UInt))
}
func (e *floatEvaluator) EvalFloat(scope Scope) (float64, error) {
	return e.f, nil
}
func (e *floatEvaluator) EvalBool(scope Scope) (bool, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Bool))
}
func (e *floatEvaluator) EvalTime(scope Scope) (values.Time, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Time))
}
func (e *floatEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Duration))
}
func (e *floatEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *floatEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *floatEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *floatEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type timeEvaluator struct {
	t    semantic.Type
	time values.Time
}

func (e *timeEvaluator) Type() semantic.Type {
	return e.t
}

func (e *timeEvaluator) EvalString(scope Scope) (string, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.String))
}
func (e *timeEvaluator) EvalInt(scope Scope) (int64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Int))
}
func (e *timeEvaluator) EvalUInt(scope Scope) (uint64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.UInt))
}
func (e *timeEvaluator) EvalFloat(scope Scope) (float64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Float))
}
func (e *timeEvaluator) EvalBool(scope Scope) (bool, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Bool))
}
func (e *timeEvaluator) EvalTime(scope Scope) (values.Time, error) {
	return e.time, nil
}
func (e *timeEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Duration))
}
func (e *timeEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *timeEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *timeEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *timeEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type durationEvaluator struct {
	t        semantic.Type
	duration values.Duration
}

func (e *durationEvaluator) Type() semantic.Type {
	return e.t
}

func (e *durationEvaluator) EvalString(scope Scope) (string, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.String))
}
func (e *durationEvaluator) EvalInt(scope Scope) (int64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Int))
}
func (e *durationEvaluator) EvalUInt(scope Scope) (uint64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.UInt))
}
func (e *durationEvaluator) EvalFloat(scope Scope) (float64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Float))
}
func (e *durationEvaluator) EvalBool(scope Scope) (bool, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Bool))
}
func (e *durationEvaluator) EvalTime(scope Scope) (values.Time, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Time))
}
func (e *durationEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	return e.duration, nil
}
func (e *durationEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *durationEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *durationEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *durationEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Function))
}

type identifierEvaluator struct {
	t    semantic.Type
	name string
}

func (e *identifierEvaluator) Type() semantic.Type {
	return e.t
}

func (e *identifierEvaluator) EvalString(scope Scope) (string, error) {
	return scope.GetString(e.name), nil
}
func (e *identifierEvaluator) EvalInt(scope Scope) (int64, error) {
	return scope.GetInt(e.name), nil
}
func (e *identifierEvaluator) EvalUInt(scope Scope) (uint64, error) {
	return scope.GetUInt(e.name), nil
}
func (e *identifierEvaluator) EvalFloat(scope Scope) (float64, error) {
	return scope.GetFloat(e.name), nil
}
func (e *identifierEvaluator) EvalBool(scope Scope) (bool, error) {
	return scope.GetBool(e.name), nil
}
func (e *identifierEvaluator) EvalTime(scope Scope) (values.Time, error) {
	return scope.GetTime(e.name), nil
}
func (e *identifierEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	return scope.GetDuration(e.name), nil
}
func (e *identifierEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	return scope.GetRegexp(e.name), nil
}
func (e *identifierEvaluator) EvalArray(scope Scope) (values.Array, error) {
	return scope.GetArray(e.name), nil
}
func (e *identifierEvaluator) EvalObject(scope Scope) (values.Object, error) {
	return scope.GetObject(e.name), nil
}
func (e *identifierEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	return scope.GetFunction(e.name), nil
}

type valueEvaluator struct {
	value values.Value
}

func (e *valueEvaluator) Type() semantic.Type {
	return e.value.Type()
}

func (e *valueEvaluator) EvalString(scope Scope) (string, error) {
	return e.value.Str(), nil
}
func (e *valueEvaluator) EvalInt(scope Scope) (int64, error) {
	return e.value.Int(), nil
}
func (e *valueEvaluator) EvalUInt(scope Scope) (uint64, error) {
	return e.value.UInt(), nil
}
func (e *valueEvaluator) EvalFloat(scope Scope) (float64, error) {
	return e.value.Float(), nil
}
func (e *valueEvaluator) EvalBool(scope Scope) (bool, error) {
	return e.value.Bool(), nil
}
func (e *valueEvaluator) EvalTime(scope Scope) (values.Time, error) {
	return e.value.Time(), nil
}
func (e *valueEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	return e.value.Duration(), nil
}
func (e *valueEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	return e.value.Regexp(), nil
}
func (e *valueEvaluator) EvalArray(scope Scope) (values.Array, error) {
	return e.value.Array(), nil
}
func (e *valueEvaluator) EvalObject(scope Scope) (values.Object, error) {
	return e.value.Object(), nil
}
func (e *valueEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	return e.value.Function(), nil
}

type memberEvaluator struct {
	t        semantic.Type
	object   Evaluator
	property string
}

func (e *memberEvaluator) Type() semantic.Type {
	return e.t
}

func (e *memberEvaluator) EvalString(scope Scope) (string, error) {
	o, err := e.object.EvalObject(scope)
	if err != nil {
		return "", err
	}
	v, _ := o.Get(e.property)
	return v.Str(), nil
}
func (e *memberEvaluator) EvalInt(scope Scope) (int64, error) {
	o, err := e.object.EvalObject(scope)
	if err != nil {
		return 0, err
	}
	v, _ := o.Get(e.property)
	return v.Int(), nil
}
func (e *memberEvaluator) EvalUInt(scope Scope) (uint64, error) {
	o, err := e.object.EvalObject(scope)
	if err != nil {
		return 0, err
	}
	v, _ := o.Get(e.property)
	return v.UInt(), nil
}
func (e *memberEvaluator) EvalFloat(scope Scope) (float64, error) {
	o, err := e.object.EvalObject(scope)
	if err != nil {
		return 0.0, err
	}
	v, _ := o.Get(e.property)
	return v.Float(), nil
}
func (e *memberEvaluator) EvalBool(scope Scope) (bool, error) {
	o, err := e.object.EvalObject(scope)
	if err != nil {
		return false, err
	}
	v, _ := o.Get(e.property)
	return v.Bool(), nil
}
func (e *memberEvaluator) EvalTime(scope Scope) (values.Time, error) {
	o, err := e.object.EvalObject(scope)
	if err != nil {
		return 0, err
	}
	v, _ := o.Get(e.property)
	return v.Time(), nil
}
func (e *memberEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	o, err := e.object.EvalObject(scope)
	if err != nil {
		return 0, err
	}
	v, _ := o.Get(e.property)
	return v.Duration(), nil
}
func (e *memberEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	o, err := e.object.EvalObject(scope)
	if err != nil {
		return nil, err
	}
	v, _ := o.Get(e.property)
	return v.Regexp(), nil
}
func (e *memberEvaluator) EvalArray(scope Scope) (values.Array, error) {
	o, err := e.object.EvalObject(scope)
	if err != nil {
		return nil, nil
	}
	v, _ := o.Get(e.property)
	return v.Array(), nil
}
func (e *memberEvaluator) EvalObject(scope Scope) (values.Object, error) {
	o, err := e.object.EvalObject(scope)
	if err != nil {
		return nil, nil
	}
	v, _ := o.Get(e.property)
	return v.Object(), nil
}
func (e *memberEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	o, err := e.object.EvalObject(scope)
	if err != nil {
		return nil, err
	}
	v, _ := o.Get(e.property)
	return v.Function(), nil
}

type arrayIndexEvaluator struct {
	t     semantic.Type
	array Evaluator
	index Evaluator
}

func (e *arrayIndexEvaluator) Type() semantic.Type {
	return e.t
}

func (e *arrayIndexEvaluator) eval(scope Scope) (values.Value, error) {
	a, err := e.array.EvalArray(scope)
	if err != nil {
		return nil, err
	}
	i, err := e.index.EvalInt(scope)
	if err != nil {
		return nil, err
	}
	return a.Get(int(i)), nil
}

func (e *arrayIndexEvaluator) EvalString(scope Scope) (string, error) {
	v, err := e.eval(scope)
	if err != nil {
		return "", err
	}
	return v.Str(), nil
}
func (e *arrayIndexEvaluator) EvalInt(scope Scope) (int64, error) {
	v, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return v.Int(), nil
}
func (e *arrayIndexEvaluator) EvalUInt(scope Scope) (uint64, error) {
	v, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return v.UInt(), nil
}
func (e *arrayIndexEvaluator) EvalFloat(scope Scope) (float64, error) {
	v, err := e.eval(scope)
	if err != nil {
		return 0.0, err
	}
	return v.Float(), nil
}
func (e *arrayIndexEvaluator) EvalBool(scope Scope) (bool, error) {
	v, err := e.eval(scope)
	if err != nil {
		return false, err
	}
	return v.Bool(), nil
}
func (e *arrayIndexEvaluator) EvalTime(scope Scope) (values.Time, error) {
	v, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return v.Time(), nil
}
func (e *arrayIndexEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	v, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return v.Duration(), nil
}
func (e *arrayIndexEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	v, err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return v.Regexp(), nil
}
func (e *arrayIndexEvaluator) EvalArray(scope Scope) (values.Array, error) {
	v, err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return v.Array(), nil
}
func (e *arrayIndexEvaluator) EvalObject(scope Scope) (values.Object, error) {
	v, err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return v.Object(), nil
}
func (e *arrayIndexEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	v, err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return v.Function(), nil
}

type callEvaluator struct {
	t      semantic.Type
	callee Evaluator
	args   Evaluator
}

func (e *callEvaluator) Type() semantic.Type {
	return e.t
}

func (e *callEvaluator) eval(scope Scope) (values.Value, error) {
	args, err := e.args.EvalObject(scope)
	if err != nil {
		return nil, err
	}
	f, err := e.callee.EvalFunction(scope)
	if err != nil {
		return nil, err
	}
	return f.Call(args)
}

func (e *callEvaluator) EvalString(scope Scope) (string, error) {
	v, err := e.eval(scope)
	if err != nil {
		return "", err
	}
	return v.Str(), nil
}
func (e *callEvaluator) EvalInt(scope Scope) (int64, error) {
	v, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return v.Int(), nil
}
func (e *callEvaluator) EvalUInt(scope Scope) (uint64, error) {
	v, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return v.UInt(), nil
}
func (e *callEvaluator) EvalFloat(scope Scope) (float64, error) {
	v, err := e.eval(scope)
	if err != nil {
		return 0.0, err
	}
	return v.Float(), nil
}
func (e *callEvaluator) EvalBool(scope Scope) (bool, error) {
	v, err := e.eval(scope)
	if err != nil {
		return false, err
	}
	return v.Bool(), nil
}
func (e *callEvaluator) EvalTime(scope Scope) (values.Time, error) {
	v, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return v.Time(), nil
}
func (e *callEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	v, err := e.eval(scope)
	if err != nil {
		return 0, err
	}
	return v.Duration(), nil
}
func (e *callEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	v, err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return v.Regexp(), nil
}
func (e *callEvaluator) EvalArray(scope Scope) (values.Array, error) {
	v, err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return v.Array(), nil
}
func (e *callEvaluator) EvalObject(scope Scope) (values.Object, error) {
	v, err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return v.Object(), nil
}
func (e *callEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	v, err := e.eval(scope)
	if err != nil {
		return nil, err
	}
	return v.Function(), nil
}

type functionEvaluator struct {
	t      semantic.Type
	body   Evaluator
	params []functionParam
}

func (e *functionEvaluator) Type() semantic.Type {
	return e.t
}

func (e *functionEvaluator) EvalString(scope Scope) (string, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.String))
}
func (e *functionEvaluator) EvalInt(scope Scope) (int64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Int))
}
func (e *functionEvaluator) EvalUInt(scope Scope) (uint64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.UInt))
}
func (e *functionEvaluator) EvalFloat(scope Scope) (float64, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Float))
}
func (e *functionEvaluator) EvalBool(scope Scope) (bool, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Bool))
}
func (e *functionEvaluator) EvalTime(scope Scope) (values.Time, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Time))
}
func (e *functionEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Duration))
}
func (e *functionEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Regexp))
}
func (e *functionEvaluator) EvalArray(scope Scope) (values.Array, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Array))
}
func (e *functionEvaluator) EvalObject(scope Scope) (values.Object, error) {
	panic(values.UnexpectedKind(e.t.Nature(), semantic.Object))
}
func (e *functionEvaluator) EvalFunction(scope Scope) (values.Function, error) {
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

func (f *functionValue) Type() semantic.Type {
	return f.t
}
func (f *functionValue) PolyType() semantic.PolyType {
	return f.t.PolyType()
}

func (f *functionValue) IsNull() bool {
	return false
}
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

type noopEvaluator struct {
}

func (noopEvaluator) Type() semantic.Type {
	return semantic.Nil
}

func (noopEvaluator) EvalString(scope Scope) (string, error) {
	return "", nil
}

func (noopEvaluator) EvalInt(scope Scope) (int64, error) {
	return 0, nil
}

func (noopEvaluator) EvalUInt(scope Scope) (uint64, error) {
	return 0, nil
}

func (noopEvaluator) EvalFloat(scope Scope) (float64, error) {
	return 0.0, nil
}

func (noopEvaluator) EvalBool(scope Scope) (bool, error) {
	return false, nil
}

func (noopEvaluator) EvalTime(scope Scope) (values.Time, error) {
	return 0, nil
}

func (noopEvaluator) EvalDuration(scope Scope) (values.Duration, error) {
	return 0, nil
}

func (noopEvaluator) EvalRegexp(scope Scope) (*regexp.Regexp, error) {
	return nil, nil
}

func (noopEvaluator) EvalArray(scope Scope) (values.Array, error) {
	return nil, nil
}

func (noopEvaluator) EvalObject(scope Scope) (values.Object, error) {
	return nil, nil
}

func (noopEvaluator) EvalFunction(scope Scope) (values.Function, error) {
	return nil, nil
}
