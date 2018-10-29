package semantic

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/influxdata/flux/ast"
)

func GenerateConstraints(node Node, annotator *Annotator) *Constraints {
	cg := ConstraintGenerator{
		cs: &Constraints{
			f:           annotator.f,
			annotations: annotator.annotations,
		},
	}
	Walk(NewScopedVisitor(cg), node)
	return cg.cs
}

type ConstraintGenerator struct {
	cs  *Constraints
	env *Env
}

func (v ConstraintGenerator) Nest() NestingVisitor {
	return ConstraintGenerator{
		cs:  v.cs,
		env: v.env.Nest(),
	}
}

func (v ConstraintGenerator) Visit(node Node) Visitor {
	return v
}

func (v ConstraintGenerator) Done(node Node) {
	a := v.cs.annotations[node]
	a.Type, a.Err = v.typeof(node)
	v.cs.annotations[node] = a
	v.cs.AddTypeConst(a.Var, a.Type)
	log.Printf("typeof %T@%v %v %v %v", node, node.Location(), a.Var, a.Type, a.Err)
}

func (v ConstraintGenerator) lookup(n Node) (PolyType, error) {
	a, ok := v.cs.annotations[n]
	if !ok {
		return nil, fmt.Errorf("no annotation found for %v", n)
	}
	return a.Type, a.Err
}
func (v ConstraintGenerator) scheme(t PolyType) Scheme {
	ftv := t.FreeVars(v.cs).diff(v.env.FreeVars())
	return Scheme{
		T:    t,
		Free: ftv,
	}
}

func (v ConstraintGenerator) typeof(n Node) (PolyType, error) {
	nodeVar := v.cs.annotations[n].Var
	switch n := n.(type) {
	case *NativeVariableDeclaration:
		t, err := v.lookup(n.Init)
		if err != nil {
			return nil, err
		}
		scheme := v.scheme(t)
		v.env.Set(n.Identifier.Name, scheme)
		return nil, nil
	case *IdentifierExpression:
		scheme, ok := v.env.Lookup(n.Name)
		if !ok {
			return nil, fmt.Errorf("undefined identifier %q", n.Name)
		}
		return v.cs.Inst(scheme), nil
	case *IntegerLiteral:
		return basic(Int), nil
	case *FloatLiteral:
		return basic(Float), nil
	case *StringLiteral:
		return basic(String), nil
	case *BinaryExpression:
		l, err := v.lookup(n.Left)
		if err != nil {
			return nil, err
		}
		r, err := v.lookup(n.Left)
		if err != nil {
			return nil, err
		}

		switch n.Operator {
		case ast.AdditionOperator:
			v.cs.AddTypeConst(l, r)
			//c.addKindConst(l, Addable{})
			//c.addKindConst(r, Addable{})
			return l, nil
		}
		// TODO add all cases so this isn't needed
		return l, nil
	case *FunctionExpression:
		args := make(map[string]PolyType, len(n.Block.Parameters.List))
		required := make([]string, 0, len(args))
		for _, arg := range n.Block.Parameters.List {
			tv := v.cs.f.Fresh()
			v.env.Set(arg.Key.Name, Scheme{T: tv})
			args[arg.Key.Name] = tv
			hasDefault := false
			for _, p := range n.Defaults.Properties {
				if p.Key.Name == arg.Key.Name {
					hasDefault = true
					break
				}
			}
			if !hasDefault {
				required = append(required, arg.Key.Name)
			}
		}
		ret, err := v.lookup(n.Block)
		if err != nil {
			return nil, err
		}
		return function{
			args:     args,
			required: required,
			ret:      ret,
		}, nil
	case *CallExpression:
		typ, err := v.lookup(n.Callee)
		if err != nil {
			return nil, err
		}
		args := make(map[string]PolyType, len(n.Arguments.Properties))
		required := make([]string, 0, len(args))
		for _, arg := range n.Arguments.Properties {
			t, err := v.lookup(arg.Value)
			if err != nil {
				return nil, err
			}
			args[arg.Key.Name] = t
			required = append(required, arg.Key.Name)
		}
		ft := function{
			args:     args,
			required: required,
			ret:      v.cs.f.Fresh(),
		}
		v.cs.AddTypeConst(typ, ft)
		return ft.ret, nil
	case *ObjectExpression:
		fields := make(map[string]PolyType, len(n.Properties))
		upper := make([]string, 0, len(fields))
		for _, field := range n.Properties {
			t, err := v.lookup(field.Value)
			if err != nil {
				return nil, err
			}
			fields[field.Key.Name] = t
			upper = append(upper, field.Key.Name)
		}
		v.cs.AddKindConst(nodeVar, KRecord{
			fields: fields,
			lower:  newLabelSet(),
			upper:  upper,
		})
		return nodeVar, nil
	case *MemberExpression:
		ptv := v.cs.f.Fresh()
		t, err := v.lookup(n.Object)
		if err != nil {
			return nil, err
		}
		tv, ok := t.(Tvar)
		if !ok {
			return nil, errors.New("member object must be a type variable")
		}
		v.cs.AddKindConst(tv, KRecord{
			fields: map[string]PolyType{n.Property: ptv},
			lower:  LabelSet{n.Property},
			upper:  allLabels,
		})
		return ptv, nil
	default:
		return nil, fmt.Errorf("unsupported %T", n)
	}
}

type Constraints struct {
	f           fresher
	annotations map[Node]annotation

	typeConst []TypeConstraint
	kindConst map[Tvar][]KindConstraint
}

type TypeConstraint struct {
	l, r PolyType
}

func (tc TypeConstraint) String() string {
	return fmt.Sprintf("%v = %v", tc.l, tc.r)
}

func (tc TypeConstraint) FreeVars(c *Constraints) TvarSet {
	return tc.l.FreeVars(c).union(tc.r.FreeVars(c))
}

func NewConstraints() *Constraints {
	return &Constraints{
		kindConst: make(map[Tvar][]KindConstraint),
	}
}

func (c *Constraints) AddTypeConst(l, r PolyType) {
	c.typeConst = append(c.typeConst, TypeConstraint{
		l: l,
		r: r,
	})
}

func (c *Constraints) AddKindConst(tv Tvar, k KindConstraint) {
	c.kindConst[tv] = append(c.kindConst[tv], k)
}

func (c *Constraints) Inst(s Scheme) (t PolyType) {
	if len(s.Free) == 0 {
		return s.T
	}
	// Create a substituion for the new type variables
	subst := make(Substitution, len(s.Free))
	for _, tv := range s.Free {
		fresh := c.f.Fresh()
		subst[tv] = fresh
	}

	// Add any new kind constraints
	for _, tv := range s.Free {
		ks, ok := c.kindConst[tv]
		if ok {
			ntv := subst.ApplyTvar(tv)
			for _, k := range ks {
				nk := subst.ApplyKind(k)
				c.AddKindConst(ntv, nk)
			}
		}
	}

	// Add any new type constraints
	for _, tc := range c.typeConst {
		fvs := tc.FreeVars(c)
		// Only add new constraints that will change
		if fvs.hasIntersect(s.Free) {
			l := subst.ApplyType(tc.l)
			r := subst.ApplyType(tc.r)
			c.AddTypeConst(l, r)
		}
	}

	return subst.ApplyType(s.T)
}

func (c *Constraints) String() string {
	var builder strings.Builder
	builder.WriteString("{\ntypes:\n")
	for _, tc := range c.typeConst {
		fmt.Fprintf(&builder, "%v,\n", tc)
	}
	builder.WriteString("kinds:\n")
	for tv, ks := range c.kindConst {
		fmt.Fprintf(&builder, "%v = %v,\n", tv, ks)
	}
	builder.WriteString("}")
	return builder.String()
}
