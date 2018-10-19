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

	// Err returns an error from the inference solution if any exist.
	// There may be more than one error, each node can be inspected for specific errors.
	Err() error

	// setType updates the nodes type or type error
	setType(n Node, pt PolyType, err error)
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
			return fmt.Errorf("type error: %v != %v", k, t)
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
}

func newInferenceVisitor() *inferenceVisitor {
	return &inferenceVisitor{
		env: &Env{m: make(map[string]TS)},
		solution: &typeSolution{
			m: make(map[Node]typeAnnotation),
		},
	}
}

func (v inferenceVisitor) nest() inferenceVisitor {
	return inferenceVisitor{
		env:      v.env,
		solution: v.solution,
	}
}

func (v inferenceVisitor) Nest() NestingVisitor {
	return inferenceVisitor{
		env:      v.env.Nest(),
		solution: v.solution,
	}
}

func (v inferenceVisitor) Visit(node Node) Visitor {
	log.Printf("typeof %T", node)
	return v.nest()
}

func (v inferenceVisitor) Done(node Node) {
	t, err := v.typeof(node)
	log.Printf("typeof %T %v %v", node, t, err)
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
		*FunctionBlock,
		*FunctionParameters:
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
		return t, nil
	case *FunctionExpression:
		in := objectPolyType{}
		if n.Block.Parameters != nil {
			in.properties = make(map[string]PolyType, len(n.Block.Parameters.List))
			for _, p := range n.Block.Parameters.List {
				pt, err := v.solution.PolyTypeOf(p)
				if err != nil {
					return nil, err
				}
				in.properties[p.Key.Name] = pt
			}
		}
		var defaults objectPolyType
		if n.Defaults != nil {
			// Unify defaults
			for _, d := range n.Defaults.Properties {
				dt, err := v.solution.PolyTypeOf(d.Value)
				if err != nil {
					return nil, err
				}
				pt, ok := in.properties[d.Key.Name]
				if !ok {
					return nil, fmt.Errorf("default defined for unknown parameter %q", d.Key.Name)
				}
				if err := v.solution.Unify(dt, pt); err != nil {
					return nil, err
				}
			}
			// Collect defaults type
			t, err := v.solution.PolyTypeOf(n.Defaults)
			if err != nil {
				return nil, err
			}
			d, ok := t.(objectPolyType)
			if !ok {
				return nil, fmt.Errorf("unexpected type of defaults %T, must by an object type", t)
			}
			defaults = d
		}
		out, err := v.solution.PolyTypeOf(n.Block.Body)
		if err != nil {
			return nil, err
		}
		var pipeArgument string
		if n.Block.Parameters != nil && n.Block.Parameters.Pipe != nil {
			pipeArgument = n.Block.Parameters.Pipe.Name
		}

		t := functionPolyType{
			in:           in,
			defaults:     defaults,
			out:          out,
			pipeArgument: pipeArgument,
		}
		return t, nil
	case *FunctionParameter:
		t := v.solution.Fresh()
		ts := TS{T: t} // function parameters do not need a schema
		v.env.Set(n.Key.Name, ts)
		return t, nil
	case *CallExpression:
		ct, err := v.solution.PolyTypeOf(n.Callee)
		if err != nil {
			return nil, err
		}
		t, ok := ct.(functionPolyType)
		if !ok {
			return nil, fmt.Errorf("cannot call non function type %T", ct)
		}

		args, err := v.solution.PolyTypeOf(n.Arguments)
		if err != nil {
			return nil, err
		}
		in, ok := args.(objectPolyType)
		if !ok {
			return nil, fmt.Errorf("unexpected type of arguments %T, must be a object type", args)
		}
		//Apply defaults to arguments in type
		for k, p := range t.defaults.properties {
			if _, ok := in.properties[k]; !ok {
				in.properties[k] = p
			}
		}
		if t.pipeArgument == "" && n.Pipe != nil {
			return nil, errors.New("cannot pipe into non pipe function")
		}
		if n.Pipe != nil {
			pt, err := v.solution.PolyTypeOf(n.Pipe)
			if err != nil {
				return nil, err
			}
			in.properties[t.pipeArgument] = pt
		}

		if err := v.solution.Unify(t.in, in); err != nil {
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
			pt, err := v.solution.PolyTypeOf(p)
			if err != nil {
				return nil, err
			}
			t.properties[p.Key.Name] = pt
		}
		return t, nil
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
		// TODO(nathanielc): How can we unify against object types without having a concrete objectPolyType?
		// The below is a hack to create or mutate it if the property does not exist.
		switch t := t.(type) {
		case typeVar:
			mt := v.solution.Fresh()
			ot := NewObjectPolyType(map[string]PolyType{
				n.Property: mt,
			})
			if err := v.solution.Unify(t, ot); err != nil {
				return nil, err
			}
			return mt, nil
		case objectPolyType:
			mt, ok := t.properties[n.Property]
			if !ok {
				mt = v.solution.Fresh()
				t.properties[n.Property] = mt
			}
			return mt, nil
		default:
			return nil, fmt.Errorf("cannot get member of non object %T", t)
		}
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
	log.Println(tm)
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
	in           PolyType
	defaults     objectPolyType
	out          PolyType
	pipeArgument string
}

type PolyFunctionSignature struct {
	In           PolyType
	Defaults     PolyType
	Out          PolyType
	PipeArgument string
}

func NewFunctionPolyType(sig PolyFunctionSignature) PolyType {
	var d objectPolyType
	if sig.Defaults == nil {
		d = objectPolyType{}
	} else {
		d = sig.Defaults.(objectPolyType)
	}
	return functionPolyType{
		in:           sig.In,
		defaults:     d,
		out:          sig.Out,
		pipeArgument: sig.PipeArgument,
	}
}

func (t functionPolyType) String() string {
	return fmt.Sprintf("(%v) defaults: %v pipe: %q -> %v", t.in, t.defaults, t.pipeArgument, t.out)
}

func (t functionPolyType) freeVars() []typeVar {
	return union(t.in.freeVars(), t.out.freeVars())
}

func (t functionPolyType) instantiate(tm map[int]typeVar) PolyType {
	return functionPolyType{
		in:           t.in.instantiate(tm).(objectPolyType),
		defaults:     t.defaults.instantiate(tm).(objectPolyType),
		out:          t.out.instantiate(tm),
		pipeArgument: t.pipeArgument,
	}
}

func (t1 functionPolyType) unify(ts TypeSolution, typ PolyType) error {
	switch t2 := typ.(type) {
	case functionPolyType:
		if err := ts.Unify(t1.in, t2.in); err != nil {
			return err
		}
		if err := ts.Unify(t1.defaults, t2.defaults); err != nil {
			return err
		}
		if err := ts.Unify(t1.out, t2.out); err != nil {
			return err
		}
		if t1.pipeArgument != t2.pipeArgument {
			return errors.New("cannot unify functions with differring pipe arguments")
		}
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
	defaults, ok := t.defaults.Type()
	if !ok {
		return nil, false
	}
	out, ok := t.out.Type()
	if !ok {
		return nil, false
	}
	return NewFunctionType(FunctionSignature{
		In:           in,
		Defaults:     defaults,
		Out:          out,
		PipeArgument: t.pipeArgument,
	}), true
}

func (t1 functionPolyType) Equal(t2 PolyType) bool {
	switch t2 := t2.(type) {
	case functionPolyType:
		return t1.in.Equal(t2.in) &&
			t1.defaults.Equal(t2.defaults) &&
			t1.out.Equal(t2.out) &&
			t1.pipeArgument == t2.pipeArgument
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
		fmt.Fprintf(&builder, "%s: %v, ", k, t)
	}
	builder.WriteString("}")
	return builder.String()
}

func (t objectPolyType) freeVars() []typeVar {
	var vars []typeVar
	for _, p := range t.properties {
		vars = union(vars, p.freeVars())
	}
	return vars
}

func (t objectPolyType) instantiate(tm map[int]typeVar) PolyType {
	properties := make(map[string]PolyType, len(t.properties))
	for k, p := range t.properties {
		properties[k] = p.instantiate(tm)
	}
	return objectPolyType{
		properties: properties,
	}
}

func (t1 objectPolyType) unify(ts TypeSolution, typ PolyType) error {
	switch t2 := typ.(type) {
	case objectPolyType:
		if len(t1.properties) != len(t2.properties) {
			return fmt.Errorf("mismatched properties %v != %v", t1, t2)
		}
		for k, p1 := range t1.properties {
			p2, ok := t2.properties[k]
			if !ok {
				return fmt.Errorf("missing parameter %q", k)
			}
			err := ts.Unify(p1, p2)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("cannot unify object %v with %v", t1, typ)
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

type typeSolution struct {
	m map[Node]typeAnnotation
	// vars is a map of typeVar index to a PolyType pointer.
	// The type of the typeVar is known when the pointer points to a non-nil PolyType.
	// All pointers in the list are themselves guaranteed to be non nil.
	vars []*PolyType

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
	for idx, ptr := range s.vars {
		fmt.Fprintf(&builder, "t%d -> %v\n", idx, *ptr)
	}
	builder.WriteString("}")
	return builder.String()
}
func (s *typeSolution) FreshSolution() TypeSolution {
	ns := &typeSolution{
		m:    make(map[Node]typeAnnotation, len(s.m)),
		vars: make([]*PolyType, len(s.vars)),
	}
	tm := make(map[int]typeVar, len(s.vars))
	for i, ptr := range s.vars {
		tm[i] = typeVar{
			idx:      i,
			solution: ns,
		}
		// make fresh copies of the var pointers
		idx := s.smallestVarIndex(i)
		if idx < i {
			// Preserve existing type var mappings
			ns.vars[i] = ns.vars[idx]
		}
		ns.vars[i] = new(PolyType)
		*ns.vars[i] = *ptr
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
	ptrA := s.vars[idxA]
	// Pick the smallest index that is equal, including itself
	for idxB, ptrB := range s.vars[:idxA] {
		if ptrA == ptrB {
			return idxB
		}
	}
	return idxA
}

func (s *typeSolution) indirect(t PolyType) PolyType {
	tv, ok := t.(typeVar)
	if ok {
		t := s.vars[tv.idx]
		if *t != nil {
			return *t
		}
	}
	return t
}

func (s *typeSolution) Unify(a, b PolyType) error {
	tvA, okA := a.(typeVar)
	tvB, okB := b.(typeVar)
	log.Println("unify", a, b)

	switch {
	case !okA && !okB:
		return a.unify(s, b)
	case okA && okB:
		// tvA == tvB
		// Map all a's to b's
		s.vars[tvB.idx] = s.vars[tvA.idx]
	case okA && !okB:
		// Substitute all tvA's with b
		*s.vars[tvA.idx] = b
	case !okA && okB:
		// Substitute all tvB's with a
		*s.vars[tvB.idx] = a
	}
	return nil
}

func (s *typeSolution) Fresh() typeVar {
	idx := len(s.vars)
	s.vars = append(s.vars, new(PolyType))
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
	return *tv.solution.vars[tv.idx]
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

func (tv typeVar) freeVars() []typeVar {
	if t := tv.lookup(); t != nil {
		return t.freeVars()
	}
	return []typeVar{tv}
}

func (tv1 typeVar) instantiate(tm map[int]typeVar) PolyType {
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
