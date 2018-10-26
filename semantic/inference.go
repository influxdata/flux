package semantic

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/influxdata/flux/ast"
	"github.com/pkg/errors"
)

// TypeSolution is a mapping of Nodes to their types.
type TypeSolution interface {
	// TypeOf reports the monotype of the node or an error.
	TypeOf(n Node) (Type, error)
	// TypeOf reports the polytype of the node or an error.
	PolyTypeOf(n Node) (PolyType, error)

	// FreshSolution creates a new solution with fresh type variables
	FreshSolution() TypeSolution

	// Fresh creates a new type variable within the solution.
	Fresh() typeVar

	// Unify modifies the solution given that a == b.
	Unify(a, b PolyType) error

	// UnifyKind modifies the solution given that a == b.
	unifyKind(a, b K) error

	// Err returns an error from the inference solution if any exist.
	// There may be more than one error, each node can be inspected for specific errors.
	Err() error

	// setType updates the nodes type or type error
	setType(n Node, pt PolyType, err error)
}

type Constraint struct {
	a, b PolyType
}

// Infer produces a solution to type inference for a given semantic graph.
func Infer(n Node) TypeSolution {
	v := newInferenceVisitor()
	Walk(NewScopedVisitor(v), n)
	return v.solution
}

type PolyType interface {
	freeVars() []typeVar
	instantiate(tm map[int]typeVar) PolyType

	unify(ts TypeSolution, t PolyType) error

	Type() (Type, bool)

	Equal(t PolyType) bool
}

func (k Kind) freeVars() []typeVar {
	return nil
}
func (k Kind) instantiate(map[int]typeVar) PolyType {
	return k
}
func (k Kind) unify(ts TypeSolution, t PolyType) error {
	switch t := t.(type) {
	case Kind:
		if k != t {
			return fmt.Errorf("%v != %v", k, t)
		}
		return nil
	default:
		return fmt.Errorf("cannot unify primitive %v with %v", k, t)
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

func (e *Env) freeVars() []typeVar {
	var u []typeVar
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
func (e *Env) LocalLookup(n string) (TS, bool) {
	ts, ok := e.m[n]
	if ok {
		return ts, true
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

type TS struct {
	T    PolyType
	List []typeVar
}

func (ts TS) freeVars() []typeVar {
	return ts.T.freeVars()
}

type inferenceVisitor struct {
	env      *Env
	solution *typeSolution
	fe       *FunctionExpression
}

func newInferenceVisitor() inferenceVisitor {
	return inferenceVisitor{
		env: &Env{m: make(map[string]TS)},
		solution: &typeSolution{
			m: make(map[Node]typeAnnotation),
		},
	}
}

func (v inferenceVisitor) Nest() NestingVisitor {
	return inferenceVisitor{
		env:      v.env.Nest(),
		solution: v.solution,
		fe:       v.fe,
	}
}

func (v inferenceVisitor) Visit(node Node) Visitor {
	//log.Printf("typeof %T@%v", node, node.Location())
	switch n := node.(type) {
	case *FunctionExpression:
		v.fe = n
	}
	return v
}

func (v inferenceVisitor) Done(node Node) {
	t, err := v.typeof(node)
	log.Printf("typeof %T@%v %v %v", node, node.Location(), t, err)
	v.solution.setType(node, t, err)
}

func (v *inferenceVisitor) typeof(node Node) (PolyType, error) {
	if node == nil {
		panic("nil")
	}
	switch n := node.(type) {
	case *Identifier,
		*Program,
		*OptionStatement,
		*FunctionBlock:
		return nil, nil
	case *BlockStatement:
		return v.solution.PolyTypeOf(n.ReturnStatement())
	case *ReturnStatement:
		return v.solution.PolyTypeOf(n.Argument)
	case *Extern:
		return v.solution.PolyTypeOf(n.Block)
	case *ExternBlock:
		return v.solution.PolyTypeOf(n.Node)
	case *ExpressionStatement:
		return v.solution.PolyTypeOf(n.Expression)
	case *ExternalVariableDeclaration:
		t := n.ExternType
		ts := v.schema(t)
		existing, ok := v.env.Lookup(n.Identifier.Name)
		if ok {
			if err := v.solution.Unify(existing.T, t); err != nil {
				return nil, err
			}
		}
		v.env.Set(n.Identifier.Name, ts)
		return t, nil
	case *NativeVariableDeclaration:
		t, err := v.solution.PolyTypeOf(n.Init)
		if err != nil {
			return nil, err
		}
		ts := v.schema(t)
		existing, ok := v.env.LocalLookup(n.Identifier.Name)
		if ok {
			if err := v.solution.Unify(existing.T, t); err != nil {
				return nil, err
			}
		}
		v.env.Set(n.Identifier.Name, ts)
		return t, nil //TODO return nil,nil?
	case *FunctionExpression:
		in, err := v.solution.PolyTypeOf(n.Block.Parameters)
		if err != nil {
			return nil, err
		}

		//var defaults objectPolyType
		//d, err := v.solution.PolyTypeOf(n.Defaults)
		//if err != nil {
		//	return nil, err
		//}
		//if d != nil {
		//	defaults, _ = d.(objectPolyType)
		//}

		out, err := v.solution.PolyTypeOf(n.Block.Body)
		if err != nil {
			return nil, err
		}
		//var pipeArgument string
		//if n.Block.Parameters != nil && n.Block.Parameters.Pipe != nil {
		//	pipeArgument = n.Block.Parameters.Pipe.Name
		//}

		t := functionPolyType{
			in: in,
			//defaults: defaults,
			out: out,
			//pipeArgument: pipeArgument,
		}
		return t, nil
	//case *FunctionDefaults:
	//	return v.solution.PolyTypeOf(n.Object)
	case *FunctionParameters:
		properties := make(map[string]PolyType, len(n.List))
		labels := make(labelSet, len(n.List))
		for i, p := range n.List {
			pt, err := v.solution.PolyTypeOf(p)
			if err != nil {
				return nil, err
			}
			properties[p.Key.Name] = pt
			labels[i] = p.Key.Name
		}
		// Unify defaults
		if v.fe.Defaults != nil {
			for _, d := range v.fe.Defaults.Properties {
				dt, err := v.solution.PolyTypeOf(d.Value)
				if err != nil {
					return nil, err
				}
				pt, ok := properties[d.Key.Name]
				if !ok {
					return nil, fmt.Errorf("default defined for unknown parameter %q", d.Key.Name)
				}
				if err := v.solution.Unify(dt, pt); err != nil {
					return nil, err
				}
			}
		}
		ko := &objectK{
			properties: properties,
			lower:      labels,
			upper:      allLabels,
		}
		in := objectPolyType{k: ko}
		return in, nil
	case *FunctionParameter:
		t := v.solution.Fresh()
		ts := TS{T: t} // function parameters do not need a schema
		v.env.Set(n.Key.Name, ts)
		return t, nil
	case *CallExpression:
		args, err := v.solution.PolyTypeOf(n.Arguments)
		if err != nil {
			return nil, err
		}
		ct, err := v.solution.PolyTypeOf(n.Callee)
		if err != nil {
			return nil, err
		}

		out := v.solution.Fresh()
		ft := functionPolyType{
			in:  args,
			out: out,
		}

		if err := v.solution.Unify(ft, ct); err != nil {
			return nil, err
		}
		return out, nil
	case *IdentifierExpression:
		// Let-Polymorphism, each reference to an identifier
		// may have its own unique monotype.
		// Instantiate a new type for each lookup.
		ts, ok := v.env.Lookup(n.Name)
		if !ok {
			return nil, fmt.Errorf("undefined identifier %q", n.Name)
		}
		t := v.instantiate(ts)
		return t, nil
	case *ObjectExpression:
		properties := make(map[string]PolyType, len(n.Properties))
		for _, p := range n.Properties {
			pt, err := v.solution.PolyTypeOf(p)
			if err != nil {
				return nil, err
			}
			properties[p.Key.Name] = pt
		}
		return NewObjectPolyType(properties), nil
	case *ArrayExpression:
		t := arrayPolyType{
			elementType: Nil, // default to an array of nil
		}
		for i, e := range n.Elements {
			et, err := v.solution.PolyTypeOf(e)
			if err != nil {
				return nil, err
			}
			if i == 0 {
				t.elementType = et
			}
			v.solution.Unify(t.elementType, et)
		}
		return t, nil
	case *LogicalExpression:
		lt, err := v.solution.PolyTypeOf(n.Left)
		if err != nil {
			return nil, err
		}
		rt, err := v.solution.PolyTypeOf(n.Right)
		if err != nil {
			return nil, err
		}
		if err := v.solution.Unify(lt, Bool); err != nil {
			return nil, err
		}
		if err := v.solution.Unify(rt, Bool); err != nil {
			return nil, err
		}
		return Bool, err
	case *BinaryExpression:
		lt, err := v.solution.PolyTypeOf(n.Left)
		if err != nil {
			return nil, err
		}
		rt, err := v.solution.PolyTypeOf(n.Right)
		if err != nil {
			return nil, err
		}
		switch n.Operator {
		case
			ast.AdditionOperator,
			ast.SubtractionOperator,
			ast.MultiplicationOperator,
			ast.DivisionOperator:
			if err := v.solution.Unify(lt, rt); err != nil {
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
			if err := v.solution.Unify(lt, String); err != nil {
				return nil, err
			}
			if err := v.solution.Unify(rt, Regexp); err != nil {
				return nil, err
			}
			return Bool, nil
		default:
			return nil, fmt.Errorf("unsupported binary operator %v", n.Operator)
		}
	case *UnaryExpression:
		t, err := v.solution.PolyTypeOf(n.Argument)
		if err != nil {
			return nil, err
		}
		switch n.Operator {
		case ast.NotOperator:
			if err := v.solution.Unify(t, Bool); err != nil {
				return nil, err
			}
			return Bool, nil
		default:
			return t, nil
		}
	case *MemberExpression:
		t, err := v.solution.PolyTypeOf(n.Object)
		if err != nil {
			return nil, err
		}
		tv := v.solution.Fresh()
		labels := make(labelSet, 1)
		labels[0] = n.Property
		ot := objectPolyType{
			k: &objectK{
				properties: map[string]PolyType{
					n.Property: tv,
				},
				lower: labels,
				upper: allLabels,
			},
		}
		log.Println("MemberExpression", ot)
		if err := v.solution.Unify(t, ot); err != nil {
			return nil, err
		}
		return tv, nil
	case *Property:
		return v.solution.PolyTypeOf(n.Value)
	case *StringLiteral:
		return String, nil
	case *IntegerLiteral:
		return Int, nil
	case *UnsignedIntegerLiteral:
		return UInt, nil
	case *FloatLiteral:
		return Float, nil
	case *BooleanLiteral:
		return Bool, nil
	case *DateTimeLiteral:
		return Time, nil
	case *DurationLiteral:
		return Duration, nil
	case *RegexpLiteral:
		return Regexp, nil
	default:
		return nil, fmt.Errorf("unsupported node type %T", node)
	}
}

func (v *inferenceVisitor) instantiate(ts TS) PolyType {
	tm := make(map[int]typeVar, len(ts.List))
	// The type vars that are equal are resolved to the smallest var index.
	// As such we iterate over free vars from smallest to largest.
	sort.Slice(ts.List, func(i, j int) bool { return ts.List[i].idx < ts.List[j].idx })
	for _, tv := range ts.List {
		idx := v.solution.smallestVarIndex(tv.idx)
		if ntv, ok := tm[idx]; ok {
			tm[tv.idx] = ntv
		} else {
			tm[idx] = v.solution.Fresh()
		}
	}
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
	in PolyType
	//defaults PolyType
	out PolyType
	//pipeArgument string
}

type PolyFunctionSignature struct {
	In           PolyType
	Defaults     PolyType
	Out          PolyType
	PipeArgument string
}

func NewFunctionPolyType(sig PolyFunctionSignature) PolyType {
	d := sig.Defaults
	if d == nil {
		d = objectPolyType{}
	}
	return functionPolyType{
		in: sig.In,
		//defaults: d,
		out: sig.Out,
		//pipeArgument: sig.PipeArgument,
	}
}

func (t functionPolyType) String() string {
	return fmt.Sprintf("(%v) -> %v", t.in, t.out)
	//return fmt.Sprintf("(%v) %v -> %v", t.in, t.defaults, t.out)
	//return fmt.Sprintf("(%v) defaults: %v pipe: %q -> %v", t.in, t.defaults, t.pipeArgument, t.out)
}

func (t functionPolyType) freeVars() []typeVar {
	return union(t.in.freeVars(), t.out.freeVars())
}

func (t functionPolyType) instantiate(tm map[int]typeVar) (it PolyType) {
	return functionPolyType{
		in: t.in.instantiate(tm),
		//defaults: t.defaults.instantiate(tm),
		out: t.out.instantiate(tm),
		//pipeArgument: t.pipeArgument,
	}
}

func (t1 functionPolyType) unify(ts TypeSolution, typ PolyType) error {
	switch t2 := typ.(type) {
	case functionPolyType:
		if err := ts.Unify(t1.in, t2.in); err != nil {
			return err
		}
		//if err := ts.Unify(t1.defaults, t2.defaults); err != nil {
		//	return err
		//}
		if err := ts.Unify(t1.out, t2.out); err != nil {
			return err
		}
		//if t1.pipeArgument != t2.pipeArgument {
		//	return errors.New("cannot unify functions with differring pipe arguments")
		//}
	default:
		return fmt.Errorf("cannot unify function %v with %v", t1, typ)
	}
	return nil
}

func (t functionPolyType) Type() (Type, bool) {
	in, ok := t.in.Type()
	if !ok {
		return nil, false
	}
	//defaults, ok := t.defaults.Type()
	//if !ok {
	//	return nil, false
	//}
	out, ok := t.out.Type()
	if !ok {
		return nil, false
	}
	return NewFunctionType(FunctionSignature{
		In: in,
		//Defaults: defaults,
		Out: out,
		//PipeArgument: t.pipeArgument,
	}), true
}

func (t1 functionPolyType) Equal(t2 PolyType) bool {
	switch t2 := t2.(type) {
	case functionPolyType:
		return t1.in.Equal(t2.in) &&
			//t1.defaults.Equal(t2.defaults) &&
			t1.out.Equal(t2.out)
		//t1.pipeArgument == t2.pipeArgument
	default:
		return false
	}
}

type objectPolyType struct {
	k *objectK
}

func NewObjectPolyType(properties map[string]PolyType) PolyType {
	return objectPolyType{
		k: NewObjectK(properties),
	}
}

func (t objectPolyType) String() string {
	return t.k.String()
}

func (t objectPolyType) freeVars() []typeVar {
	return t.k.freeVars()
}

func (t objectPolyType) instantiate(tm map[int]typeVar) (it PolyType) {
	return objectPolyType{
		k: t.k.instantiate(tm).(*objectK),
	}
}

func (t objectPolyType) kind() K {
	return t.k
}

func (t1 objectPolyType) unify(ts TypeSolution, typ PolyType) error {
	switch t2 := typ.(type) {
	case objectPolyType:
		if err := ts.unifyKind(t1.k, t2.k); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot unify object %v with %v", t1, typ)
	}
	return nil
}

func (t objectPolyType) Type() (Type, bool) {
	properties := make(map[string]Type)
	for _, l := range t.k.lower {
		p := t.k.properties[l]
		pt, ok := p.Type()
		if !ok {
			return nil, false
		}
		properties[l] = pt
	}
	return NewObjectType(properties), true
}
func (t1 objectPolyType) Equal(t2 PolyType) bool {
	return false
	//switch t2 := t2.(type) {
	//case objectPolyType:
	//	if len(t1.properties) != len(t2.properties) {
	//		return false
	//	}
	//	for k, p1 := range t1.properties {
	//		p2, ok := t2.properties[k]
	//		if !ok || !p1.Equal(p2) {
	//			return false
	//		}
	//	}
	//	return true
	//default:
	//	return false
	//}
}

type arrayPolyType struct {
	elementType PolyType
}

func NewArrayPolyType(elementType PolyType) PolyType {
	return arrayPolyType{
		elementType: elementType,
	}
}

func (t arrayPolyType) String() string {
	return fmt.Sprintf("[%v]", t.elementType)
}

func (t arrayPolyType) freeVars() []typeVar {
	return t.elementType.freeVars()
}

func (t arrayPolyType) instantiate(tm map[int]typeVar) PolyType {
	return arrayPolyType{
		elementType: t.elementType.instantiate(tm),
	}
}

func (t1 arrayPolyType) unify(ts TypeSolution, typ PolyType) error {
	switch t2 := typ.(type) {
	case arrayPolyType:
		return ts.Unify(t1.elementType, t2.elementType)
	default:
		return fmt.Errorf("cannot unify array %v with %v", t1, typ)
	}
}

func (t arrayPolyType) Type() (Type, bool) {
	typ, mono := t.elementType.Type()
	if !mono {
		return nil, false
	}
	return NewArrayType(typ), true
}
func (t1 arrayPolyType) Equal(t2 PolyType) bool {
	switch t2 := t2.(type) {
	case arrayPolyType:
		return t1.elementType.Equal(t2.elementType)
	default:
		return false
	}
}

func union(a, b []typeVar) []typeVar {
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

func diff(a, b []typeVar) []typeVar {
	d := make([]typeVar, 0, len(a))
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

type K interface {
	unifyKind(ts TypeSolution, k K) error
	equal(K) bool
}

type Kinded interface {
	kind() K
}

type objectK struct {
	properties map[string]PolyType

	lower labelSet //union
	upper labelSet //intersection
}

func NewObjectK(properties map[string]PolyType) *objectK {
	keys := make(labelSet, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	return &objectK{
		properties: properties,
		lower:      newLabelSet(),
		upper:      keys,
	}
}
func (t *objectK) String() string {
	var builder strings.Builder
	builder.WriteString("({")
	for k, t := range t.properties {
		fmt.Fprintf(&builder, "%s: %v, ", k, t)
	}
	fmt.Fprintf(&builder, "}, %v, %v)", t.lower, t.upper)
	return builder.String()
}

func (t *objectK) freeVars() []typeVar {
	var vars []typeVar
	for _, p := range t.properties {
		vars = union(vars, p.freeVars())
	}
	return vars
}

func (a *objectK) equal(b K) bool {
	switch b := b.(type) {
	case *objectK:
		if len(a.properties) != len(b.properties) {
			return false
		}
		for k, pa := range a.properties {
			pb, ok := b.properties[k]
			if !ok {
				return false
			}
			if !pa.Equal(pb) {
				return false
			}
		}
		return a.lower.equal(b.lower) && a.upper.equal(b.upper)
	default:
		return false
	}
}

func (t *objectK) instantiate(tm map[int]typeVar) K {
	properties := make(map[string]PolyType, len(t.properties))
	for k, p := range t.properties {
		properties[k] = p.instantiate(tm)
	}
	return &objectK{
		properties: properties,
		lower:      t.lower.copy(),
		upper:      t.upper.copy(),
	}
}

func (k1 *objectK) unifyKind(ts TypeSolution, k2 K) error {
	switch k2 := k2.(type) {
	case *objectK:
		lower := k1.lower.union(k2.lower)
		upper := k1.upper.intersect(k2.upper)

		// Unify lower bound
		for _, l := range lower {
			t1 := k1.properties[l]
			t2 := k2.properties[l]
			if err := ts.Unify(t1, t2); err != nil {
				return err
			}
		}

		properties := make(map[string]PolyType, len(k1.properties)+len(k2.properties))
		// Merge properties
		for n, t1 := range k1.properties {
			properties[n] = t1
		}
		for n, t2 := range k2.properties {
			if _, ok := properties[n]; !ok {
				properties[n] = t2
			}
		}
		k := &objectK{
			properties: properties,
			lower:      lower,
			upper:      upper,
		}
		ts.(*typeSolution).updateKind(k1, k)
		ts.(*typeSolution).updateKind(k2, k)
		return nil
	default:
		return fmt.Errorf("cannot unify object kind with %T", k2)
	}
}

type typeSolution struct {
	m map[Node]typeAnnotation
	// varTypes is a map of typeVar index to a PolyType pointer.
	// The type of the typeVar is known when the pointer points to a non-nil PolyType.
	// All pointers in the list are themselves guaranteed to be non nil.
	varTypes []*PolyType

	varKinds []*K

	// err is any error encountered while computing inference.
	err error
}

type typeAnnotation struct {
	poly PolyType
	err  error
}

func (s *typeSolution) String() string {
	var builder strings.Builder
	builder.WriteString("{\n")
	for idx, ptr := range s.varTypes {
		kptr := s.varKinds[idx]
		fmt.Fprintf(&builder, "t%d -> %v,%v\n", idx, *ptr, *kptr)
		//fmt.Fprintf(&builder, "t%d -> %v\n", idx, *ptr)
	}
	builder.WriteString("}")
	return builder.String()
}
func (s *typeSolution) FreshSolution() TypeSolution {
	ns := &typeSolution{
		m:        make(map[Node]typeAnnotation, len(s.m)),
		varTypes: make([]*PolyType, len(s.varTypes)),
	}
	tm := make(map[int]typeVar, len(s.varTypes))
	for i, ptr := range s.varTypes {
		tm[i] = typeVar{
			idx:      i,
			solution: ns,
		}
		// make fresh copies of the var pointers
		idx := s.smallestVarIndex(i)
		if idx < i {
			// Preserve existing type var mappings
			ns.varTypes[i] = ns.varTypes[idx]
		}
		ns.varTypes[i] = new(PolyType)
		*ns.varTypes[i] = *ptr
	}

	for n, ta := range s.m {
		if ta.poly != nil {
			ta.poly = ta.poly.instantiate(tm)
		}
		ns.m[n] = ta
	}
	return ns
}

func (s *typeSolution) TypeOf(n Node) (Type, error) {
	ta, ok := s.m[n]
	if !ok {
		// Should this be an error?
		return nil, nil
	}
	if ta.err != nil {
		return nil, ta.err
	}
	poly := s.indirect(ta.poly)
	mono, ok := poly.Type()
	if !ok {
		return nil, errors.New("node is not monomorphic")
	}
	return mono, nil
}

func (s *typeSolution) PolyTypeOf(n Node) (PolyType, error) {
	ta := s.m[n]
	if ta.err != nil {
		return nil, ta.err
	}
	return s.indirect(ta.poly), nil
}

func (s *typeSolution) setType(n Node, poly PolyType, err error) {
	err = errors.Wrapf(err, "type error %v", n.Location())
	if s.err == nil && err != nil {
		s.err = err
	}
	s.m[n] = typeAnnotation{
		poly: poly,
		err:  err,
	}
}

// smallestVarIndex returns the smallest index of an equivalent var.
func (s *typeSolution) smallestVarIndex(idxA int) int {
	ptrA := s.varTypes[idxA]
	// Pick the smallest index that is equal, including itself
	for idxB, ptrB := range s.varTypes[:idxA] {
		if ptrA == ptrB {
			return idxB
		}
	}
	return idxA
}

func (s *typeSolution) indirect(t PolyType) (pt PolyType) {
	tv, ok := t.(typeVar)
	if ok {
		k := s.varKinds[tv.idx]
		if *k != nil {
			return objectPolyType{
				k: (*k).(*objectK),
			}
		}
		t := s.varTypes[tv.idx]
		if *t != nil {
			return *t
		}
	}
	return t
}
func (s *typeSolution) indirectK(k K) K {
	tv, ok := k.(typeVar)
	if ok {
		k := s.varKinds[tv.idx]
		if *k != nil {
			return *k
		}
	}
	return k
}

func (s *typeSolution) Unify(a, b PolyType) error {
	a = s.indirect(a)
	b = s.indirect(b)
	log.Printf("unify %v %v %v", a, b, s)
	defer func() {
		log.Println("unify done:", s)
	}()
	tvA, okA := a.(typeVar)
	tvB, okB := b.(typeVar)

	switch {
	case !okA && !okB:
		return a.unify(s, b)
	case okA && okB:
		// tvA == tvB
		// Map all a's to b's
		s.varTypes[tvA.idx] = s.varTypes[tvB.idx]
	case okA && !okB:
		// Substitute all tvA's with b
		*s.varTypes[tvA.idx] = b
		if k, ok := b.(Kinded); ok {
			*s.varKinds[tvA.idx] = k.kind()
		}
	case !okA && okB:
		// Substitute all tvB's with a
		*s.varTypes[tvB.idx] = a
		if k, ok := a.(Kinded); ok {
			*s.varKinds[tvB.idx] = k.kind()
		}
	}
	return nil
}
func (s *typeSolution) updateKind(a, b K) {
	log.Println("updateKind")
	for idx, k := range s.varKinds {
		if a.equal(*k) {
			log.Println("Updated!!")
			*s.varKinds[idx] = b
			// Is this hack reasonable?
			// I expect it breaks if we need to instantiate a record
			*s.varTypes[idx] = objectPolyType{
				k: b.(*objectK),
			}
		}
	}
}

func (s *typeSolution) unifyKind(a, b K) error {
	a = s.indirectK(a)
	b = s.indirectK(b)
	log.Printf("unifyKind %v %v", a, b)
	defer func() {
		log.Println("unifyKind done:", s)
	}()
	tvA, okA := a.(typeVar)
	tvB, okB := b.(typeVar)

	switch {
	case !okA && !okB:
		return a.unifyKind(s, b)
	case okA && okB:
		// tvA == tvB
		// Map all a's to b's
		s.varKinds[tvB.idx] = s.varKinds[tvA.idx]
	case okA && !okB:
		// Substitute all tvA's with b
		*s.varKinds[tvA.idx] = b
	case !okA && okB:
		// Substitute all tvB's with a
		*s.varKinds[tvB.idx] = a
	}
	return nil
}

func (s *typeSolution) Fresh() typeVar {
	idx := len(s.varTypes)
	s.varTypes = append(s.varTypes, new(PolyType))
	s.varKinds = append(s.varKinds, new(K))
	return typeVar{
		idx:      idx,
		solution: s,
	}
}

func (s *typeSolution) Err() error {
	return s.err
}

type typeVar struct {
	idx      int
	solution *typeSolution
}

// lookup returns the PolyType from the solution, which will be nil if it is unknown.
func (tv typeVar) lookup() PolyType {
	return *tv.solution.varTypes[tv.idx]
}

func (tv typeVar) lookupK() K {
	return *tv.solution.varKinds[tv.idx]
}

func (tv typeVar) String() string {
	if t := tv.lookup(); t != nil {
		return fmt.Sprintf("%v", t)
	}
	return fmt.Sprintf("t%d", tv.idx)
}

func (tv typeVar) unify(ts TypeSolution, t PolyType) error {
	return errors.New("unification should not reach typeVars")
}
func (tv typeVar) unifyKind(ts TypeSolution, k K) error {
	return errors.New("kind unification should not reach typeVars")
}
func (tv typeVar) Type() (Type, bool) {
	if t := tv.lookup(); t != nil {
		return t.Type()
	}
	return nil, false
}

func (tv1 typeVar) Equal(t2 PolyType) bool {
	if t1 := tv1.lookup(); t1 != nil {
		return t1.Equal(t2)
	}
	switch tv2 := t2.(type) {
	case typeVar:
		return tv1.idx == tv2.idx
	default:
		return false
	}
}
func (tv1 typeVar) equal(k2 K) bool {
	if k1 := tv1.lookupK(); k1 != nil {
		return k1.equal(k2)
	}
	switch tv2 := k2.(type) {
	case typeVar:
		return tv1.idx == tv2.idx
	default:
		return false
	}
}

func (tv typeVar) freeVars() []typeVar {
	if t := tv.lookup(); t != nil {
		return t.freeVars()
	}
	return []typeVar{tv}
}

func (tv1 typeVar) instantiate(tm map[int]typeVar) PolyType {
	if t := tv1.lookup(); t != nil {
		return t.instantiate(tm)
	}
	if tv2, ok := tm[tv1.idx]; ok {
		return tv2
	}
	return tv1
}

type Fresher interface {
	Fresh() typeVar
}

func NewFresher() Fresher {
	return new(typeSolution)
}

type labelSet []string

func newLabelSet() labelSet {
	return make(labelSet, 0, 10)
}

var allLabels = labelSet(nil)

func (s labelSet) String() string {
	if s == nil {
		return "L"
	}
	if len(s) == 0 {
		return "∅"
	}
	var builder strings.Builder
	builder.WriteString("(")
	for i, l := range s {
		if i != 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(l)
	}
	builder.WriteString(")")
	return builder.String()
}

func (s labelSet) union(o labelSet) labelSet {
	if s == nil {
		return s
	}
	union := make(labelSet, len(s), len(s)+len(o))
	copy(union, s)
LOOP:
	for _, l := range o {
		for _, lu := range union {
			if lu == l {
				continue LOOP
			}
		}
		union = append(union, l)
	}
	return union
}

func (s labelSet) intersect(o labelSet) labelSet {
	if s == nil {
		return o
	}
	if o == nil {
		return s
	}
	intersect := make(labelSet, 0, len(s))
	for _, ls := range s {
		for _, lo := range o {
			if ls == lo {
				intersect = append(intersect, ls)
				break
			}
		}
	}
	return intersect
}
func (a labelSet) equal(b labelSet) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (s labelSet) copy() labelSet {
	c := make(labelSet, len(s))
	copy(c, s)
	return c
}
