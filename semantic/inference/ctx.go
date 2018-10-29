package inference

import (
	"fmt"
	"strings"
)

type Ctx struct {
	typeConst []TypeConstraint
	kindConst map[Tvar][]Kind
	f         fresher
}

type TypeConstraint struct {
	l, r Type
}

func (tc TypeConstraint) String() string {
	return fmt.Sprintf("%v = %v", tc.l, tc.r)
}

func (tc TypeConstraint) FreeVars(c *Ctx) TvarSet {
	return tc.l.FreeVars(c).union(tc.r.FreeVars(c))
}

func NewCtx() *Ctx {
	return &Ctx{
		kindConst: make(map[Tvar][]Kind),
	}
}

func (c *Ctx) AddTypeConst(l, r Type) {
	c.typeConst = append(c.typeConst, TypeConstraint{
		l: l,
		r: r,
	})
}

func (c *Ctx) AddKindConst(tv Tvar, k Kind) {
	c.kindConst[tv] = append(c.kindConst[tv], k)
}

func (c *Ctx) Inst(s Scheme) (t Type) {
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

func (c *Ctx) String() string {
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
