package semantic

import (
	"fmt"
	"log"
	"strings"

	"github.com/influxdata/flux/ast"
	"github.com/pkg/errors"
)

func GenerateConstraints(node Node, annotator Annotator) *Constraints {
	cg := ConstraintGenerator{
		cs: &Constraints{
			f:           annotator.f,
			annotations: annotator.annotations,
			kindConst:   make(map[Tvar][]KindConstraint),
		},
		env: NewEnv(),
	}
	log.Println("Pre", cg.cs)
	Walk(NewScopedVisitor(cg), node)
	log.Println("GenerateConstraints", cg.cs)
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
	if a.Type != nil {
		v.cs.annotations[node] = a
		if !a.Var.Equal(a.Type) {
			v.cs.AddTypeConst(a.Var, a.Type, node.Location())
		}

	}
	a.Err = errors.Wrapf(a.Err, "type error %v", node.Location())
	log.Printf("typeof %T@%v %v %v %v", node, node.Location(), a.Var, a.Type, a.Err)
}

func (v ConstraintGenerator) lookup(n Node) (PolyType, error) {
	a, ok := v.cs.annotations[n]
	if !ok {
		return nil, fmt.Errorf("no annotation found for %T@%v", n, n.Location())
	}
	if a.Type == nil {
		return nil, fmt.Errorf("no type annotation found for %T@%v", n, n.Location())
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
	case *ExternalVariableDeclaration:
		t := n.ExternType
		existing, ok := v.env.LocalLookup(n.Identifier.Name)
		if ok {
			v.cs.AddTypeConst(t, existing.T, n.Location())
		}
		scheme := v.scheme(t)
		v.env.Set(n.Identifier.Name, scheme)
		return nil, nil

	case *NativeVariableDeclaration:
		t, err := v.lookup(n.Init)
		if err != nil {
			return nil, err
		}
		existing, ok := v.env.LocalLookup(n.Identifier.Name)
		if ok {
			v.cs.AddTypeConst(t, existing.T, n.Location())
		}
		scheme := v.scheme(t)
		v.env.Set(n.Identifier.Name, scheme)
		return nil, nil
	case *IdentifierExpression:
		scheme, ok := v.env.Lookup(n.Name)
		if !ok {
			return nil, fmt.Errorf("undefined identifier %q", n.Name)
		}
		t := v.cs.Instantiate(scheme, n.Location())
		return t, nil
	case *ReturnStatement:
		return v.lookup(n.Argument)
	case *BlockStatement:
		return v.lookup(n.ReturnStatement())
	case *BinaryExpression:
		l, err := v.lookup(n.Left)
		if err != nil {
			return nil, err
		}
		r, err := v.lookup(n.Right)
		if err != nil {
			return nil, err
		}

		switch n.Operator {
		case
			ast.AdditionOperator,
			ast.SubtractionOperator,
			ast.MultiplicationOperator,
			ast.DivisionOperator:
			v.cs.AddTypeConst(l, r, n.Location())
			return l, nil
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
			v.cs.AddTypeConst(l, String, n.Left.Location())
			v.cs.AddTypeConst(r, Regexp, n.Right.Location())
			return Bool, nil
		default:
			return nil, fmt.Errorf("unsupported binary operator %v", n.Operator)
		}
	case *FunctionExpression:
		parameters := make(map[string]PolyType, len(n.Block.Parameters.List))
		required := make([]string, 0, len(parameters))
		for _, param := range n.Block.Parameters.List {
			t, err := v.lookup(param)
			if err != nil {
				return nil, err
			}
			parameters[param.Key.Name] = t
			hasDefault := false
			if n.Defaults != nil {
				for _, p := range n.Defaults.Properties {
					if p.Key.Name == param.Key.Name {
						hasDefault = true
						dt, err := v.lookup(p)
						if err != nil {
							return nil, err
						}
						v.cs.AddTypeConst(t, dt, p.Location())
						break
					}
				}
			}
			if !hasDefault {
				required = append(required, param.Key.Name)
			}

		}
		ret, err := v.lookup(n.Block)
		if err != nil {
			return nil, err
		}
		return function{
			parameters: parameters,
			required:   required,
			ret:        ret,
		}, nil
	case *FunctionParameter:
		v.env.Set(n.Key.Name, Scheme{T: nodeVar})
		return nodeVar, nil
	case *FunctionBlock:
		return v.lookup(n.Body)
	case *CallExpression:
		typ, err := v.lookup(n.Callee)
		if err != nil {
			return nil, err
		}
		parameters := make(map[string]PolyType, len(n.Arguments.Properties))
		required := make([]string, 0, len(parameters))
		for _, arg := range n.Arguments.Properties {
			t, err := v.lookup(arg.Value)
			if err != nil {
				return nil, err
			}
			parameters[arg.Key.Name] = t
			required = append(required, arg.Key.Name)
		}
		ft := function{
			parameters: parameters,
			required:   required,
			ret:        v.cs.f.Fresh(),
		}
		v.cs.AddTypeConst(typ, ft, n.Location())
		return ft.ret, nil
	case *ObjectExpression:
		properties := make(map[string]PolyType, len(n.Properties))
		upper := make([]string, 0, len(properties))
		for _, field := range n.Properties {
			t, err := v.lookup(field.Value)
			if err != nil {
				return nil, err
			}
			properties[field.Key.Name] = t
			upper = append(upper, field.Key.Name)
		}
		v.cs.AddKindConst(nodeVar, KRecord{
			properties: properties,
			lower:      EmptyLabelSet(),
			upper:      upper,
		})
		return nodeVar, nil
	case *Property:
		return v.lookup(n.Value)
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
			properties: map[string]PolyType{n.Property: ptv},
			lower:      LabelSet{n.Property},
			upper:      AllLabels,
		})
		return ptv, nil
	case *ArrayExpression:
		at := list{typ: NewObjectPolyType(nil, EmptyLabelSet(), AllLabels)}
		if len(n.Elements) > 0 {
			et, err := v.lookup(n.Elements[0])
			if err != nil {
				return nil, err
			}
			at.typ = et
			for _, el := range n.Elements[1:] {
				elt, err := v.lookup(n.Elements[0])
				if err != nil {
					return nil, err
				}
				v.cs.AddTypeConst(et, elt, el.Location())
			}
		}
		return at, nil
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

	// Explictly list nodes that do not produce constraints
	case *Program,
		*Extern,
		*ExternBlock,
		*Identifier,
		*FunctionParameters,
		*ExpressionStatement:
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported %T", n)
	}
}

type Constraints struct {
	f           *fresher
	annotations map[Node]annotation

	typeConst []TypeConstraint
	kindConst map[Tvar][]KindConstraint
}

type TypeConstraint struct {
	l, r PolyType
	loc  ast.SourceLocation
}

func (tc TypeConstraint) String() string {
	return fmt.Sprintf("%v = %v @ %v", tc.l, tc.r, tc.loc)
}

func (c *Constraints) AddTypeConst(l, r PolyType, loc ast.SourceLocation) {
	c.typeConst = append(c.typeConst, TypeConstraint{
		l:   l,
		r:   r,
		loc: loc,
	})
}

func (c *Constraints) AddKindConst(tv Tvar, k KindConstraint) {
	c.kindConst[tv] = append(c.kindConst[tv], k)
}

func (c *Constraints) Instantiate(s Scheme, loc ast.SourceLocation) (t PolyType) {
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
		fvs := tc.l.FreeVars(c)
		// Only add new constraints that constrain the left hand free vars
		if fvs.hasIntersect(s.Free) {
			l := subst.ApplyType(tc.l)
			r := subst.ApplyType(tc.r)
			c.AddTypeConst(l, r, loc)
		}
	}

	return subst.ApplyType(s.T)
}

func (c *Constraints) String() string {
	var builder strings.Builder
	builder.WriteString("{\nannotations:\n")
	for n, ann := range c.annotations {
		fmt.Fprintf(&builder, "%T@%v = %v,\n", n, n.Location(), ann.Var)
	}
	builder.WriteString("types:\n")
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
