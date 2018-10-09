package semantic

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type T interface {
	Unsolved() []TV
	Instantiate(tm map[int]TV) T
	Unify(t T) error
	// Type produces the monotype of this type
	Type() Type
}

func (k Kind) Unsolved() []TV {
	return nil
}
func (k Kind) Instantiate(map[int]TV) T {
	return k
}
func (k Kind) Unify(T) error {
	return nil
}
func (k Kind) Type() Type {
	return k
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
	return fmt.Sprintf("t%d", tv.V)
}

func (tv TV) Unify(t T) error {
	log.Println("TV.Unify", tv)
	err := unifyVar(t, tv.T)
	log.Println("TV.Unify", **tv.T)
	return err
}
func (tv TV) Type() Type {
	switch t := (**tv.T).(type) {
	case Type:
		return t
	case TV:
		log.Println("TV.Type rec")
		return t.Type()
	default:
		log.Printf("rec %T", **tv.T)
	}
	return nil
}

func unifyVar(t T, r **T) error {
	log.Println("unifyVar", t, r, *r)
	if *r != nil {
		return t.Unify(**r)
	}
	if tv2, ok := t.(TV); ok && *r == *tv2.T {
		// Cyclic, no need to update
		return nil
	}
	*r = &t
	log.Println("unifyVar", t, r, *r, **r)
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

func Infer(n Node) (Type, error) {
	env := &Env{m: make(map[string]TS)}
	f := newInferer()
	t, err := f.typeof(env, n)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errors.New("no type found")
	}
	log.Printf("%#v", t)
	switch t := t.(type) {
	case Type:
		return t, nil
	case TV:
		log.Printf("%#v", **t.T)
		typ := t.Type()
		if typ != nil {
			return typ, nil
		}
	}
	return nil, errors.New("fail type")
}

type inferer struct {
	f      *Fresher
	noPoly bool
}

func newInferer() *inferer {
	return &inferer{
		f: new(Fresher),
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
	case funcTyp:
		if err := t1.in.Unify(t2.in); err != nil {
			return err
		}
		if err := t1.out.Unify(t2.out); err != nil {
			return err
		}
	case TV:
		unifyVar(t1, t2.T)
	default:
		return errors.New("fail")
	}
	return nil
}

func (t funcTyp) Type() Type {
	return NewFunctionType(FunctionSignature{
		ReturnType: t.out.Type(),
	})
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

func (t objType) Type() Type {
	properties := make(map[string]Type)
	for k, p := range t.properties {
		properties[k] = p.Type()
	}
	return NewObjectType(properties)
}

func (f *inferer) typeof(env *Env, node Node) (t T, _ error) {
	log.Printf("typeof %T", node)
	defer func() {
		log.Printf("typeof %T = %v", node, t)
	}()
	switch n := node.(type) {
	case *Program:
		var t T
		var err error
		for _, s := range n.Body {
			t, err = f.typeof(env, s)
			if err != nil {
				return nil, err
			}
		}
		return t, nil
	case *ExpressionStatement:
		return f.typeof(env, n.Expression)
	case *NativeVariableDeclaration:
		t, err := f.typeof(env, n.Init)
		if err != nil {
			return nil, err
		}
		ts := f.schema(env, t)
		env.Set(n.Identifier.Name, ts)
		return t, nil
	case *FunctionExpression:
		// TODO: Type check n.Defaults
		return f.typeof(env, n.Block)
	case *FunctionBlock:
		env := env.Nest()
		in := objType{
			properties: make(map[string]T, len(n.Parameters.List)),
		}
		for _, p := range n.Parameters.List {
			pt, err := f.typeof(env, p)
			if err != nil {
				return nil, err
			}
			in.properties[p.Key.Name] = pt
		}
		f.noPoly = true
		out, err := f.typeof(env, n.Body)
		f.noPoly = false
		if err != nil {
			return nil, err
		}
		t := funcTyp{
			in:  in,
			out: out,
		}
		return t, nil
	case *FunctionParameter:
		t := f.f.Fresh()
		ts := f.schema(env, t)
		env.Set(n.Key.Name, ts)
		return t, nil
	case *CallExpression:
		ct, err := f.typeof(env, n.Callee)
		if err != nil {
			return nil, err
		}
		//TODO: Resolve through TV indirection
		t, ok := ct.(funcTyp)
		if !ok {
			return nil, fmt.Errorf("cannot call non function type %T", ct)
		}
		//TODO: Apply defaults to arugments here.
		//TODO: Apply pipe to arugments here.
		in, err := f.typeof(env, n.Arguments)
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
		ts, ok := env.Lookup(n.Name)
		if !ok {
			return nil, fmt.Errorf("undefined identifier %q", n.Name)
		}
		var t T
		if f.noPoly {
			t = ts.T
		} else {
			t = f.instantiate(ts)
		}
		log.Printf("ident: %s ts: %v t: %v", n.Name, ts, t)
		return t, nil
	case *ObjectExpression:
		t := objType{
			properties: make(map[string]T, len(n.Properties)),
		}
		for _, p := range n.Properties {
			pt, err := f.typeof(env, p)
			if err != nil {
				return nil, err
			}
			t.properties[p.Key.Name] = pt
		}
		return t, nil
	case *Property:
		return f.typeof(env, n.Value)
	case *BooleanLiteral:
		return Bool, nil
	case *IntegerLiteral:
		return Int, nil
	default:
		return nil, fmt.Errorf("unsupported %T", node)
	}
}

func (f *inferer) instantiate(ts TS) T {
	tm := make(map[int]TV, len(ts.List))
	for _, tv := range ts.List {
		tm[tv.V] = f.f.Fresh()
	}
	return ts.T.Instantiate(tm)
}
func (f *inferer) schema(env *Env, t T) TS {
	uv := t.Unsolved()
	ev := env.EnvUnsolved()
	d := diff(uv, ev)
	return TS{
		T:    t,
		List: d,
	}
}

func (f *inferer) Done() {}

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
