package semantic

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/influxdata/flux/ast"
	"github.com/pkg/errors"
)

func SolveConstraints(cs *Constraints) (TypeSolution, error) {
	s := &Solution{cs: cs}
	err := s.solve()
	if err != nil {
		return nil, err
	}
	return s, nil
}

type Solution struct {
	cs    *Constraints
	kinds kindsMap
}

func (sol *Solution) solve() error {
	// Create substituion
	subst := make(Substitution)
	// Create map of unified kind constraints
	kinds := make(map[Tvar]KindConstraint, len(sol.cs.kindConst))

	// Initialize unified kinds with first kind constraint
	for tv, ks := range sol.cs.kindConst {
		kinds[tv] = ks[0]
	}

	// Unify all kind constraints
	for tvl, ks := range sol.cs.kindConst {
		for _, k := range ks {
			tvr := subst.ApplyTvar(tvl)
			kind := kinds[tvr]
			s, err := unifyKinds(kinds, tvl, tvr, kind, k)
			if err != nil {
				return err
			}
			subst.Merge(s)
		}
	}

	// Unify all type constraints
	for _, tc := range sol.cs.typeConst {
		l := subst.ApplyType(tc.l)
		r := subst.ApplyType(tc.r)
		s, err := unifyTypes(kinds, l, r)
		if err != nil {
			return errors.Wrapf(err, "type error %v", tc.loc)
		}
		subst.Merge(s)
	}

	// Apply substituion to kind constraints
	sol.kinds = make(map[Tvar]KindConstraint, len(kinds))
	for tv, k := range kinds {
		k = subst.ApplyKind(k)
		tv = subst.ApplyTvar(tv)
		sol.kinds[tv] = k
	}
	for n, ann := range sol.cs.annotations {
		if ann.Type != nil {
			ann.Type = subst.ApplyType(ann.Type)
			sol.cs.annotations[n] = ann
		}
	}
	log.Println("subst", subst)
	return nil
}

func (s *Solution) TypeOf(n Node) (Type, error) {
	a, ok := s.cs.annotations[n]
	if !ok {
		return nil, nil
	}
	if a.Err != nil {
		return nil, a.Err
	}
	return a.Type.Type(s.kinds)
}

func (s *Solution) PolyTypeOf(n Node) (PolyType, error) {
	a, ok := s.cs.annotations[n]
	if !ok {
		return nil, nil
	}
	if a.Err != nil {
		return nil, a.Err
	}
	if a.Type == nil {
		return nil, fmt.Errorf("node %T@%v has no poly type", n, n.Location())
	}
	return a.Type.polyType(s.kinds)
}

func (s *Solution) AddConstraint(l, r PolyType) error {
	s.kinds = nil
	s.cs.AddTypeConst(l, r, ast.SourceLocation{})
	return s.solve()
}

func unifyTypes(kinds map[Tvar]KindConstraint, l, r PolyType) (s Substitution, _ error) {
	log.Println("unifyTypes", l, r)
	return l.UnifyType(kinds, r)
}

func unifyKinds(kinds map[Tvar]KindConstraint, tvl, tvr Tvar, l, r KindConstraint) (Substitution, error) {
	k, s, err := l.UnifyKind(kinds, r)
	if err != nil {
		return nil, err
	}
	log.Printf("unifyKinds %v = %v == %v = %v ==> %v :: %v", tvl, l, tvr, r, k, s)
	kinds[tvr] = k
	if tvl != tvr {
		log.Println("unifyKinds.deleting", tvl)
		delete(s, tvl)
	}
	return s, nil
}

func unifyVarAndType(kinds map[Tvar]KindConstraint, tv Tvar, t PolyType) (Substitution, error) {
	if t.Occurs(tv) {
		return nil, fmt.Errorf("type var %v occurs in %v creating a cycle", tv, t)
	}
	unifyKindsByType(kinds, tv, t)
	return Substitution{tv: t}, nil
}

func unifyKindsByVar(kinds map[Tvar]KindConstraint, l, r Tvar) (Substitution, error) {
	kl, okl := kinds[l]
	kr, okr := kinds[r]
	switch {
	case okl && okr:
		return unifyKinds(kinds, l, r, kl, kr)
	case okl && !okr:
		kinds[r] = kl
		log.Println("unifyKindsByVar.deleting", l)
		delete(kinds, l)
	}
	return nil, nil
}

func unifyKindsByType(kinds map[Tvar]KindConstraint, tv Tvar, t PolyType) (Substitution, error) {
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

type kindsMap map[Tvar]KindConstraint

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
