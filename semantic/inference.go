package semantic

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/influxdata/flux/ast"
)

func Infer(n Node) {
	v := newInferenceVisitor()
	Walk(v, n)
}

type T interface {
}

type Indirecter interface {
	Indirect() T
}

func (k Kind) Unsolved() []TV {
	return nil
}
func (k Kind) Instantiate(map[int]TV) T {
	return k
}
func (k Kind) Unify(t T) error {
	switch t := t.(type) {
	case TV:
		return unifyVar(k, t.T)
	case Kind:
		if k != t {
			return fmt.Errorf("type error: %v != %v", k, t)
		}
		return nil
	default:
		return errors.New("type error")
	}
}
func (k Kind) Type() (Type, bool) {
	return k, true
}

type Env struct {
	parent *Env
	m      map[string]TS
}

func (e *Env) EnvUnsolved() []TV {
	var u []TV
	for _, ts := range e.m {
		u = union(u, ts.Unsolved())
	}
	return u
}

func (e *Env) Lookup(n string) (TS, bool) {
	ts, ok := e.m[n]
	if ok {
		return ts, true
	}
	if e.parent != nil {
		return e.parent.Lookup(n)
	}
	return TS{}, false
}

func (e *Env) Set(n string, ts TS) {
	e.m[n] = ts
}

func (e *Env) Nest() *Env {
	return &Env{
		parent: e,
		m:      make(map[string]TS),
	}
}

type Fresher int

func (f *Fresher) Fresh() TV {
	v := int(*f)
	(*f)++
	return newTV(v)
}

type TV struct {
	V int
	T **T
}

func newTV(v int) TV {
	return TV{
		V: v,
		T: new(*T),
	}
}

func (tv TV) String() string {
	if tv.T != nil && *tv.T != nil {
		return fmt.Sprintf("%v", tv.Indirect())
	}
	return fmt.Sprintf("t%d", tv.V)
}

func (tv TV) Unify(t T) error {
	err := unifyVar(t, tv.T)
	return err
}
func (tv TV) Type() (Type, bool) {
	t := tv.Indirect()
	switch t.(type) {
	case TV:
		return nil, false
	default:
		return t.Type()
	}
}

func (tv TV) Indirect() T {
	if tv.T != nil && *tv.T != nil {
		if i, ok := (**tv.T).(Indirecter); ok {
			return i.Indirect()
		}
		return (**tv.T)
	}
	return tv
}

func unifyVar(t T, r **T) error {
	if *r != nil {
		return t.Unify(**r)
	}
	if tv2, ok := t.(TV); ok && *r == *tv2.T {
		// Cyclic, no need to update
		return nil
	}
	*r = &t
	return nil
}

func (tv TV) Unsolved() []TV {
	if *tv.T != nil {
		return (**tv.T).Unsolved()
	}
	return []TV{tv}
}

func (tv1 TV) Instantiate(tm map[int]TV) T {
	if tv2, ok := tm[tv1.V]; ok {
		return tv2
	}
	return tv1
}

type TS struct {
	T    T
	List []TV
}

func (ts TS) Unsolved() []TV {
	return ts.T.Unsolved()
}

type inferenceVisitor struct {
	f    *Fresher
	env  *Env
	node Node
}

func newInferenceVisitor() *inferenceVisitor {
	return &inferenceVisitor{
		f:   new(Fresher),
		env: &Env{m: make(map[string]TS)},
	}
}

func (v *inferenceVisitor) nest() *inferenceVisitor {
	return &inferenceVisitor{
		f:   v.f,
		env: v.env,
	}
}
func (v *inferenceVisitor) nestEnv() *inferenceVisitor {
	return &inferenceVisitor{
		f:   v.f,
		env: v.env.Nest(),
	}
}

func (v *inferenceVisitor) Visit(node Node) Visitor {
	v.node = node
	log.Printf("typeof %p %T", v, v.node)
	switch node.(type) {
	case *FunctionBlock:
		return v.nestEnv()
	}
	return v.nest()
}

func (v *inferenceVisitor) Done() {
	t, err := v.typeof()
	log.Printf("typeof %p %T %v %v", v, v.node, t, err)
	v.node.setTyp(t, err)
}

func (v *inferenceVisitor) typeof() (T, error) {
	switch n := v.node.(type) {
	case *Identifier,
		*Program,
		*FunctionBlock,
		*FunctionParameters:
		return nil, nil
	case *ExpressionStatement:
		return n.Expression.typ()
	case *NativeVariableDeclaration:
		t, err := n.Init.typ()
		if err != nil {
			return nil, err
		}
		ts := v.schema(t)
		existing, ok := v.env.Lookup(n.Identifier.Name)
		if ok {
			log.Printf("unify existing %v = %v", existing.T, t)
			if err := existing.T.Unify(t); err != nil {
				return nil, err
			}
		}
		v.env.Set(n.Identifier.Name, ts)
		return t, nil
	case *FunctionExpression:
		// TODO: Type check n.Defaults
		in := objType{
			properties: make(map[string]T, len(n.Block.Parameters.List)),
		}
		for _, p := range n.Block.Parameters.List {
			pt, err := p.typ()
			if err != nil {
				return nil, err
			}
			in.properties[p.Key.Name] = pt
		}
		out, err := n.Block.Body.typ()
		if err != nil {
			return nil, err
		}

		t := funcTyp{
			in:  in,
			out: out,
		}
		return t, nil
	case *FunctionParameter:
		t := v.f.Fresh()
		ts := TS{T: t} // function parameters do not need a schema
		v.env.Set(n.Key.Name, ts)
		return t, nil
	case *CallExpression:
		ct, err := n.Callee.typ()
		if err != nil {
			return nil, err
		}
		if i, ok := ct.(Indirecter); ok {
			ct = i.Indirect()
		}
		t, ok := ct.(funcTyp)
		if !ok {
			return nil, fmt.Errorf("cannot call non function type %T", ct)
		}
		//TODO: Apply defaults to arugments here.
		//TODO: Apply pipe to arugments here.
		in, err := n.Arguments.typ()
		if err != nil {
			return nil, err
		}

		if err := t.in.Unify(in); err != nil {
			return nil, err
		}
		return t.out, nil
	case *IdentifierExpression:
		// Let-Polymorphism, each reference to an identifier
		// may have its own unique monotype.
		// Instantiate a new type for each.
		ts, ok := v.env.Lookup(n.Name)
		if !ok {
			return nil, fmt.Errorf("undefined identifier %q", n.Name)
		}
		t := v.instantiate(ts)
		return t, nil
	case *ObjectExpression:
		t := objType{
			properties: make(map[string]T, len(n.Properties)),
		}
		for _, p := range n.Properties {
			pt, err := p.typ()
			if err != nil {
				return nil, err
			}
			t.properties[p.Key.Name] = pt
		}
		return t, nil
	case *BinaryExpression:
		lt, err := n.Left.typ()
		if err != nil {
			return nil, err
		}
		rt, err := n.Right.typ()
		if err != nil {
			return nil, err
		}
		switch n.Operator {
		case
			ast.AdditionOperator,
			ast.SubtractionOperator,
			ast.MultiplicationOperator,
			ast.DivisionOperator:
			if err := lt.Unify(rt); err != nil {
				return nil, err
			}
			return lt, nil
		case
			ast.GreaterThanEqualOperator,
			ast.LessThanEqualOperator,
			ast.GreaterThanOperator,
			ast.LessThanOperator,
			ast.NotEqualOperator,
			ast.EqualOperator:
			return Bool, nil
		case
			ast.RegexpMatchOperator,
			ast.NotRegexpMatchOperator:
			if err := lt.Unify(String); err != nil {
				return nil, err
			}
			if err := rt.Unify(Regexp); err != nil {
				return nil, err
			}
			return Bool, nil
		default:
			return nil, fmt.Errorf("unsupported binary operator %v", n.Operator)
		}
	case *Property:
		return n.Value.typ()
	case *BooleanLiteral:
		return Bool, nil
	case *IntegerLiteral:
		return Int, nil
	case *StringLiteral:
		return String, nil
	default:
		return nil, fmt.Errorf("unsupported %T", n)
	}
}

func (v *inferenceVisitor) instantiate(ts TS) T {
	tm := make(map[int]TV, len(ts.List))
	for _, tv := range ts.List {
		tm[tv.V] = v.f.Fresh()
	}
	return ts.T.Instantiate(tm)
}
func (v *inferenceVisitor) schema(t T) TS {
	uv := t.Unsolved()
	ev := v.env.EnvUnsolved()
	d := diff(uv, ev)
	return TS{
		T:    t,
		List: d,
	}
}

// funcTyp represent a function, all functions transform a single input type into an output type.
type funcTyp struct {
	in  T
	out T
}

func (t funcTyp) String() string {
	return fmt.Sprintf("%v -> %v", t.in, t.out)
}

func (t funcTyp) Unsolved() []TV {
	return union(t.in.Unsolved(), t.out.Unsolved())
}

func (t funcTyp) Instantiate(tm map[int]TV) T {
	return funcTyp{
		in:  t.in.Instantiate(tm),
		out: t.out.Instantiate(tm),
	}
}

func (t1 funcTyp) Unify(typ T) error {
	switch t2 := typ.(type) {
	case TV:
		unifyVar(t1, t2.T)
	case funcTyp:
		if err := t1.in.Unify(t2.in); err != nil {
			return err
		}
		if err := t1.out.Unify(t2.out); err != nil {
			return err
		}
	default:
		return errors.New("fail")
	}
	return nil
}

func (t funcTyp) Type() (Type, bool) {
	in, ok := t.in.Type()
	if !ok {
		return nil, false
	}
	out, ok := t.out.Type()
	if !ok {
		return nil, false
	}
	return NewFunctionType(FunctionSignature{
		In:  in,
		Out: out,
	}), true
}

type objType struct {
	properties map[string]T
}

func (t objType) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	for k, t := range t.properties {
		fmt.Fprintf(&builder, "%s: %v,", k, t)
	}
	builder.WriteString("}")
	return builder.String()
}

func (t objType) Unsolved() []TV {
	var vars []TV
	for _, p := range t.properties {
		vars = union(vars, p.Unsolved())
	}
	return vars
}

func (t objType) Instantiate(tm map[int]TV) T {
	properties := make(map[string]T, len(t.properties))
	for k, p := range t.properties {
		properties[k] = p.Instantiate(tm)
	}
	return objType{
		properties: properties,
	}
}

func (t1 objType) Unify(typ T) error {
	switch t2 := typ.(type) {
	case objType:
		if len(t1.properties) != len(t2.properties) {
			return fmt.Errorf("mismatched properties")
		}
		for k, p1 := range t1.properties {
			p2, ok := t2.properties[k]
			if !ok {
				return fmt.Errorf("missing parameter %q", k)
			}
			err := p1.Unify(p2)
			if err != nil {
				return err
			}
		}
	case TV:
		unifyVar(t1, t2.T)
	default:
		return errors.New("fail")
	}
	return nil
}

func (t objType) Type() (Type, bool) {
	properties := make(map[string]Type)
	for k, p := range t.properties {
		pt, ok := p.Type()
		if !ok {
			return nil, false
		}
		properties[k] = pt
	}
	return NewObjectType(properties), true
}

func union(a, b []TV) []TV {
	u := a
	for _, v := range b {
		found := false
		for _, f := range a {
			if f == v {
				found = true
				break
			}
		}
		if !found {
			u = append(u, v)
		}
	}
	return u
}

func diff(a, b []TV) []TV {
	d := make([]TV, 0, len(a))
	for _, v := range a {
		found := false
		for _, f := range b {
			if f == v {
				found = true
				break
			}
		}
		if !found {
			d = append(d, v)
		}
	}
	return d
}

type TypeScheme interface {
	// MonoType returns the monomorphic type if such a type exists.
	MonoType() (Type, bool)

	// TypeSchemes cannot be created outside of the semantic package
	typeScheme()
}

type typeScheme struct {
	scheme TypeScheme
}

func (ts *typeScheme) TypeScheme() TypeScheme {
	if ts.scheme == nil {
		return anyTypeScheme{}
	}
	return ts.scheme
}

func (ts *typeScheme) setTypeScheme(s TypeScheme) {
	ts.scheme = s
}

//anyTypeScheme is polymorphic type scheme for any type.
type anyTypeScheme struct{}

func (ts anyTypeScheme) MonoType() (Type, bool) {
	return nil, false
}
func (ts anyTypeScheme) typeScheme() {}
