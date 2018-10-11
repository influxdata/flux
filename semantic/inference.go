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

type PolyType interface {
	freeVars() []*TV
	instantiate(tm map[int]*TV) PolyType

	Unify(t PolyType) error

	Type() (Type, bool)

	Equal(t PolyType) bool
}

type Indirecter interface {
	Indirect() PolyType
}

func (k Kind) freeVars() []*TV {
	return nil
}
func (k Kind) instantiate(map[int]*TV) PolyType {
	return k
}
func (k Kind) Unify(t PolyType) error {
	switch t := t.(type) {
	case *TV:
		return t.Unify(k)
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
func (k Kind) PolyType() PolyType {
	return k
}
func (k Kind) Equal(t PolyType) bool {
	switch t := t.(type) {
	case Kind:
		return k == t
	default:
		return false
	}
}

type Env struct {
	parent *Env
	m      map[string]TS
}

func (e *Env) freeVars() []*TV {
	var u []*TV
	for _, ts := range e.m {
		u = union(u, ts.freeVars())
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

func (f *Fresher) Fresh() *TV {
	v := int(*f)
	(*f)++
	return newTV(v)
}

// TV is a type variable.
// By using a pointer to a poly type,
// type variables are updated to point to their solved type.
// This update occurs during unification.
type TV struct {
	V int
	// T is a reference to a poly type.
	// If *T is not nil, then the type of the variable is known.
	// T itself is guaranteed to be non nil.
	T *PolyType
}

func newTV(v int) *TV {
	return &TV{
		V: v,
		T: new(PolyType),
	}
}

func (tv *TV) String() string {
	if *tv.T != nil {
		return fmt.Sprintf("%v", tv.Indirect())
	}
	return fmt.Sprintf("t%d", tv.V)
}

// Unify ensures the types are equal, updating the type variable necessary.
func (tv1 *TV) Unify(t PolyType) error {
	if *tv1.T != nil {
		return (*tv1.T).Unify(t)
	}
	switch tv2 := t.(type) {
	case *TV:
		log.Println(tv1.V, " ==> ", tv2.V)
		// Rename the variables to be the same.
		// Make both type variables point to the same PolyType.
		// This way if one or the other is further unified both will be.
		*tv1 = *tv2
	default:
		log.Println(tv1.V, " ==> ", t)
		// Update the referenced poly type since it is now known.
		*tv1.T = t
	}
	return nil
}
func (tv *TV) Type() (Type, bool) {
	t := tv.Indirect()
	switch t.(type) {
	case *TV:
		return nil, false
	default:
		return t.Type()
	}
}
func (tv *TV) Equal(t PolyType) bool {
	if *tv.T != nil {
		return (*tv.T).Equal(t)
	}
	switch t := t.(type) {
	case *TV:
		return tv.V == t.V
	default:
		return false
	}
}

func (tv *TV) Indirect() PolyType {
	if *tv.T != nil {
		return *tv.T
	}
	return tv
}

func (tv *TV) freeVars() []*TV {
	if *tv.T != nil {
		return (*tv.T).freeVars()
	}
	return []*TV{tv}
}

func (tv1 *TV) instantiate(tm map[int]*TV) PolyType {
	if tv2, ok := tm[tv1.V]; ok {
		return tv2
	}
	return tv1
}

type TS struct {
	T    PolyType
	List []*TV
}

func (ts TS) freeVars() []*TV {
	return ts.T.freeVars()
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
	node.annotateType(nil, nil)
	v.node = node
	log.Printf("typeof %p %T", v, v.node)
	switch node.(type) {
	case *ExternBlock,
		*BlockStatement,
		*FunctionBlock:
		return v.nestEnv()
	}
	return v.nest()
}

func (v *inferenceVisitor) Done() {
	t, err := v.typeof()
	log.Printf("typeof %p %T %v %v", v, v.node, t, err)
	v.node.annotateType(t, err)
}

func (v *inferenceVisitor) typeof() (PolyType, error) {
	switch n := v.node.(type) {
	case *Identifier,
		*Program,
		*FunctionBlock,
		*FunctionParameters:
		return nil, nil
	case *BlockStatement:
		return n.ReturnStatement().PolyType()
	case *ReturnStatement:
		return n.Argument.PolyType()
	case *Extern:
		return n.Block.PolyType()
	case *ExternBlock:
		return n.Node.PolyType()
	case *ExpressionStatement:
		return n.Expression.PolyType()
	case *ExternalVariableDeclaration:
		t := n.ExternType
		ts := v.schema(t)
		existing, ok := v.env.Lookup(n.Identifier.Name)
		if ok {
			if err := existing.T.Unify(t); err != nil {
				return nil, err
			}
		}
		v.env.Set(n.Identifier.Name, ts)
		return t, nil
	case *NativeVariableDeclaration:
		t, err := n.Init.PolyType()
		if err != nil {
			return nil, err
		}
		ts := v.schema(t)
		existing, ok := v.env.Lookup(n.Identifier.Name)
		if ok {
			if err := existing.T.Unify(t); err != nil {
				return nil, err
			}
		}
		v.env.Set(n.Identifier.Name, ts)
		return t, nil
	case *FunctionExpression:
		// TODO: Type check n.Defaults
		in := objectPolyType{
			properties: make(map[string]PolyType, len(n.Block.Parameters.List)),
		}
		for _, p := range n.Block.Parameters.List {
			pt, err := p.PolyType()
			if err != nil {
				return nil, err
			}
			in.properties[p.Key.Name] = pt
		}
		out, err := n.Block.Body.PolyType()
		if err != nil {
			return nil, err
		}

		t := functionPolyType{
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
		ct, err := n.Callee.PolyType()
		if err != nil {
			return nil, err
		}
		if i, ok := ct.(Indirecter); ok {
			ct = i.Indirect()
		}
		t, ok := ct.(functionPolyType)
		if !ok {
			return nil, fmt.Errorf("cannot call non function type %T", ct)
		}
		//TODO: Apply defaults to arugments here.
		//TODO: Apply pipe to arugments here.
		in, err := n.Arguments.PolyType()
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
		// instantiate a new type for each.
		ts, ok := v.env.Lookup(n.Name)
		if !ok {
			return nil, fmt.Errorf("undefined identifier %q", n.Name)
		}
		t := v.instantiate(ts)
		return t, nil
	case *ObjectExpression:
		t := objectPolyType{
			properties: make(map[string]PolyType, len(n.Properties)),
		}
		for _, p := range n.Properties {
			pt, err := p.PolyType()
			if err != nil {
				return nil, err
			}
			t.properties[p.Key.Name] = pt
		}
		return t, nil
	case *BinaryExpression:
		lt, err := n.Left.PolyType()
		if err != nil {
			return nil, err
		}
		rt, err := n.Right.PolyType()
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
		return n.Value.PolyType()
	case *BooleanLiteral:
		return Bool, nil
	case *IntegerLiteral:
		return Int, nil
	case *FloatLiteral:
		return Float, nil
	case *StringLiteral:
		return String, nil
	default:
		return nil, fmt.Errorf("unsupported %T", n)
	}
}

func (v *inferenceVisitor) instantiate(ts TS) PolyType {
	tm := make(map[int]*TV, len(ts.List))
	for _, tv := range ts.List {
		tm[tv.V] = v.f.Fresh()
	}
	log.Println("tm", tm)
	return ts.T.instantiate(tm)
}
func (v *inferenceVisitor) schema(t PolyType) TS {
	uv := t.freeVars()
	ev := v.env.freeVars()
	d := diff(uv, ev)
	return TS{
		T:    t,
		List: d,
	}
}

// functionPolyType represent a function, all functions transform a single input type into an output type.
type functionPolyType struct {
	in  PolyType
	out PolyType
}

func NewFunctionPolyType(in, out PolyType) PolyType {
	return functionPolyType{
		in:  in,
		out: out,
	}
}

func (t functionPolyType) String() string {
	return fmt.Sprintf("(%v) -> %v", t.in, t.out)
}

func (t functionPolyType) freeVars() []*TV {
	return union(t.in.freeVars(), t.out.freeVars())
}

func (t functionPolyType) instantiate(tm map[int]*TV) PolyType {
	return functionPolyType{
		in:  t.in.instantiate(tm),
		out: t.out.instantiate(tm),
	}
}

func (t1 functionPolyType) Unify(typ PolyType) error {
	switch t2 := typ.(type) {
	case *TV:
		return t2.Unify(t1)
	case functionPolyType:
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

func (t functionPolyType) Type() (Type, bool) {
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

func (t1 functionPolyType) Equal(t2 PolyType) bool {
	switch t2 := t2.(type) {
	case functionPolyType:
		return t1.in.Equal(t2.in) && t1.out.Equal(t2.out)
	default:
		return false
	}
}

type objectPolyType struct {
	properties map[string]PolyType
}

func NewObjectPolyType(properties map[string]PolyType) PolyType {
	return objectPolyType{
		properties: properties,
	}
}

func (t objectPolyType) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	for k, t := range t.properties {
		fmt.Fprintf(&builder, "%s: %v,", k, t)
	}
	builder.WriteString("}")
	return builder.String()
}

func (t objectPolyType) freeVars() []*TV {
	var vars []*TV
	for _, p := range t.properties {
		vars = union(vars, p.freeVars())
	}
	return vars
}

func (t objectPolyType) instantiate(tm map[int]*TV) PolyType {
	properties := make(map[string]PolyType, len(t.properties))
	for k, p := range t.properties {
		properties[k] = p.instantiate(tm)
	}
	return objectPolyType{
		properties: properties,
	}
}

func (t1 objectPolyType) Unify(typ PolyType) error {
	switch t2 := typ.(type) {
	case objectPolyType:
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
	case *TV:
		return t2.Unify(t1)
	default:
		return errors.New("fail")
	}
	return nil
}

func (t objectPolyType) Type() (Type, bool) {
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
func (t1 objectPolyType) Equal(t2 PolyType) bool {
	switch t2 := t2.(type) {
	case objectPolyType:
		if len(t1.properties) != len(t2.properties) {
			return false
		}
		for k, p1 := range t1.properties {
			p2, ok := t2.properties[k]
			if !ok || !p1.Equal(p2) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func union(a, b []*TV) []*TV {
	u := a
	for _, v := range b {
		found := false
		for _, f := range a {
			if f.Equal(v) {
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

func diff(a, b []*TV) []*TV {
	d := make([]*TV, 0, len(a))
	for _, v := range a {
		found := false
		for _, f := range b {
			if f.Equal(v) {
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
