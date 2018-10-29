package semantic

import (
	"fmt"
	"log"
	"sort"

	"github.com/influxdata/flux/ast"
)

// TypeSolution is a mapping of Nodes to their types.
type TypeSolution interface {
	// TypeOf reports the monotype of the node or an error.
	TypeOf(n Node) (Type, error)
	// TypeOf reports the polytype of the node or an error.
	PolyTypeOf(n Node) (PolyType, error)

	// AddConstraint adds a new constraint and solves again reporting any errors.
	AddConstraint(l, r PolyType) error

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

type Fresher interface {
	Fresh() typeVar
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
