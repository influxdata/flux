package inference

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
)

func addKindConst(c *Ctx, t Type, k Kind) {
	switch t := t.(type) {
	case Tvar:
		c.AddKindConst(t, k)
	}
}

func generateType(c *Ctx, e Env, n semantic.Node) (Type, error) {
	switch n := n.(type) {
	case *semantic.IdentifierExpression:
		scheme, ok := e[n.Name]
		if !ok {
			return nil, fmt.Errorf("undefined identifier %q", n.Name)
		}
		return c.Inst(scheme), nil
	case *semantic.IntegerLiteral:
		return basic(semantic.Int), nil
	case *semantic.FloatLiteral:
		return basic(semantic.Float), nil
	case *semantic.StringLiteral:
		return basic(semantic.String), nil
	case *semantic.BinaryExpression:
		l, err := generateType(c, e, n.Left)
		if err != nil {
			return nil, err
		}
		r, err := generateType(c, e, n.Left)
		if err != nil {
			return nil, err
		}

		switch n.Operator {
		case ast.AdditionOperator:
			c.AddTypeConst(l, r)
			//c.addKindConst(l, Addable{})
			//c.addKindConst(r, Addable{})
			return l, nil
		}
		// TODO add all cases so this isn't needed
		return l, nil
	case *semantic.FunctionExpression:
		args := make(map[string]Type, len(n.Block.Parameters.List))
		required := make([]string, 0, len(args))
		for _, arg := range n.Block.Parameters.List {
			tv := c.Fresh()
			e[arg.Key.Name] = Scheme{T: tv}
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
		ret, err := generateType(c, e, n.Block.Body)
		if err != nil {
			return nil, err
		}
		return function{
			args:     args,
			required: required,
			ret:      ret,
		}, nil
	case *semantic.CallExpression:
		typ, err := generateType(c, e, n.Callee)
		if err != nil {
			return nil, err
		}
		args := make(map[string]Type, len(n.Arguments.Properties))
		required := make([]string, 0, len(args))
		for _, arg := range n.Arguments.Properties {
			t, err := generateType(c, e, arg.Value)
			if err != nil {
				return nil, err
			}
			args[arg.Key.Name] = t
			required = append(required, arg.Key.Name)
		}
		ft := function{
			args:     args,
			required: required,
			ret:      c.Fresh(),
		}
		c.AddTypeConst(typ, ft)
		return ft.ret, nil
	case *semantic.ObjectExpression:
		fields := make(map[string]Type, len(n.Properties))
		upper := make([]string, 0, len(fields))
		for _, field := range n.Properties {
			t, err := generateType(c, e, field.Value)
			if err != nil {
				return nil, err
			}
			fields[field.Key.Name] = t
			upper = append(upper, field.Key.Name)
		}
		tv := c.Fresh()
		c.AddKindConst(tv, KRecord{
			fields: fields,
			lower:  newLabelSet(),
			upper:  upper,
		})
		return tv, nil
	case *semantic.MemberExpression:
		ptv := c.Fresh()
		t, err := generateType(c, e, n.Object)
		if err != nil {
			return nil, err
		}
		tv, ok := t.(Tvar)
		if !ok {
			return nil, errors.New("member object must be a type variable")
		}
		c.AddKindConst(tv, KRecord{
			fields: map[string]Type{n.Property: ptv},
			lower:  LabelSet{n.Property},
			upper:  allLabels,
		})
		return ptv, nil
	default:
		return nil, fmt.Errorf("unsupported %T", n)
	}
}

func unifyTypes(kinds map[Tvar]Kind, l, r Type) (Substitution, error) {
	log.Println("unify", l, r)
	return l.UnifyType(kinds, r)
}

func unifyKinds(kinds map[Tvar]Kind, tv Tvar, l, r Kind) (Substitution, error) {
	return l.UnifyKind(kinds, tv, r)
}

func unifyVarAndType(kinds map[Tvar]Kind, tv Tvar, t Type) (Substitution, error) {
	if t.Occurs(tv) {
		return nil, errors.New("cycle found")
	}
	unifyKindsByType(kinds, tv, t)
	return Substitution{tv: t}, nil
}

func unifyKindsByVar(kinds map[Tvar]Kind, l, r Tvar) (Substitution, error) {
	kl, okl := kinds[l]
	kr, okr := kinds[r]
	switch {
	case okl && okr:
		subst, err := unifyKinds(kinds, r, kl, kr)
		if err != nil {
			return nil, err
		}
		//log.Println("deleting", l)
		//delete(kinds, l)
		return subst, nil
	case okl && !okr:
		kinds[r] = kl
		//log.Println("deleting", l)
		//delete(kinds, l)
	}
	return nil, nil
}

func unifyKindsByType(kinds map[Tvar]Kind, tv Tvar, t Type) (Substitution, error) {
	k, ok := kinds[tv]
	if !ok {
		return nil, nil
	}
	switch k.(type) {
	case KRecord:
		_, ok := t.(Tvar)
		if !ok {
			return nil, errors.New("invalid type for kind")
		}
	}
	return nil, nil
}

func solve(program *semantic.Program) (sol *solution, err error) {
	sol = &solution{
		nodes: make(map[semantic.Node]Type),
	}
	c := NewCtx()
	env := make(Env)
	defer func() {
		//log.Println("Ctx", c)
		//log.Println("Env", env)
		log.Println("Err", err)
		//log.Println("Solution", sol)
	}()
	for _, s := range program.Body {
		switch s := s.(type) {
		case *semantic.NativeVariableDeclaration:
			t, err := generateType(c, env, s.Init)
			if err != nil {
				return nil, err
			}
			scheme := Scheme{
				T:    t,
				Free: t.FreeVars(c),
			}
			env[s.Identifier.Name] = scheme
			sol.nodes[s] = t
		case *semantic.ExpressionStatement:
			t, err := generateType(c, env, s.Expression)
			if err != nil {
				return nil, err
			}
			sol.nodes[s] = t
		}
	}
	log.Println("Generated Constraints", c)

	kinds := make(map[Tvar]Kind, len(c.kindConst))
	subst := make(Substitution)
	defer func() {
		log.Println("kinds", kindsMap(kinds))
		log.Println("subst", subst)
	}()
	for tv, kind := range c.kindConst {
		kinds[tv] = kind
	}

	for _, tc := range c.typeConst {
		l := subst.ApplyType(tc.l)
		r := subst.ApplyType(tc.r)
		s, err := unifyTypes(kinds, l, r)
		if err != nil {
			return nil, err
		}
		subst = subst.Merge(s)
	}

	skinds := make(map[Tvar]Kind, len(kinds))
	for tv, k := range kinds {
		k = subst.ApplyKind(k)
		tv = subst.ApplyTvar(tv)
		skinds[tv] = k
	}
	kinds = skinds
	env = subst.ApplyEnv(env)
	return sol, nil
}

func Infer(program *semantic.Program) (Solution, error) {
	return solve(program)
}

type Solution interface {
	TypeOf(n semantic.Node) (Type, error)
}

type solution struct {
	kinds map[Tvar]Kind
	nodes map[semantic.Node]Type
}

func (s *solution) TypeOf(n semantic.Node) (Type, error) {
	return s.nodes[n], nil
}

type kindsMap map[Tvar]Kind

func (kinds kindsMap) String() string {
	var builder strings.Builder
	vars := make([]int, 0, len(kinds))
	for tv := range kinds {
		vars = append(vars, int(tv))
	}
	sort.Ints(vars)
	builder.WriteString("{\n")
	for i, tvi := range vars {
		tv := Tvar(tvi)
		if i != 0 {
			builder.WriteString(",\n")
		}
		fmt.Fprintf(&builder, "%v = %v", tv, kinds[tv])
	}
	builder.WriteString("}")
	return builder.String()
}
