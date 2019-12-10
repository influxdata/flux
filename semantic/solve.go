package semantic

import (
	"fmt"
	"sort"
	"strings"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// SolveConstraints solves the type inference problem defined by the constraints.
func SolveConstraints(cs *Constraints) (TypeSolution, error) {
	s := &Solution{cs: cs}
	err := s.solve()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Solution implement TypeSolution and solves the unification problem.
type Solution struct {
	cs    *Constraints
	kinds kindsMap
}

func (s *Solution) Fresh() Tvar {
	return s.cs.f.Fresh()
}

func (s *Solution) FreshSolution() TypeSolution {
	return &Solution{
		cs: s.cs.Copy(),
	}
}

// solve uses Robinson flavor unification to solve the constraints.
// Robison unification is the idea that given a constraint that two types are equal, those types are unified.
//
// Unifying two types means to do one of the following:
//  1. Given two primitive types assert the types are the same or report an error.
//  2. Given a type variable and another type, record that the type variable now has the given type.
//  3. Recurse into children types of compound types, for example unify the return types of functions.
//
// The unification process has two domains over which it operates.
// The type domain and the kind domain.
// Unifying types occurs as explained above.
// Unifying kinds is the same process except in the kind domain.
// The domains are NOT independent, unifying two types may require that two kinds be unified.
// Similarly unifying two kinds may require that two types be unified.
//
// These two separate domains allow for structural polymorphism among other things.
// Specifically the structure of objects is constrained in the kind domain not the type domain.
// See "Simple Type Inference for Structural Polymorphism" Jacques Garrigue https://caml.inria.fr/pub/papers/garrigue-structural_poly-fool02.pdf for details on this approach.
func (sol *Solution) solve() error {
	// Create map of unified kind constraints
	kinds := make(map[Tvar]Kind, len(sol.cs.kindConst))

	// Unify all kind constraints
	subst, err := sol.unifyKinds(kinds)
	if err != nil {
		return err
	}

	// Unify all type constraints
	for _, tc := range sol.cs.typeConst {
		l := subst.ApplyType(tc.l)
		r := subst.ApplyType(tc.r)
		s, err := unifyTypes(kinds, l, r)
		if err != nil {
			return errors.Wrapf(err, codes.Invalid, "type error %v", tc.loc)
		}
		subst.Merge(s)
	}

	// Apply substitution to kind constraints
	sol.kinds = make(map[Tvar]Kind, len(kinds))
	for tv, k := range kinds {
		k = subst.ApplyKind(k)
		tv = subst.ApplyTvar(tv)
		sol.kinds[tv] = k
	}

	// Apply substitution to the type annotations
	for n, ann := range sol.cs.annotations {
		if ann.Type != nil {
			ann.Type = subst.ApplyType(ann.Type)
			sol.cs.annotations[n] = ann
		}
	}
	//log.Println("subst", subst)
	//log.Println("kinds", sol.kinds)
	return nil
}

// unifyKinds will unify the kinds from the kind constraints
// into a single kind per type variable.
//
// The kinds will be unified so that kinds that depend on other
// type variables will be unified after the kinds for the dependents
// are unified.
func (sol *Solution) unifyKinds(kinds map[Tvar]Kind) (Substitution, error) {
	// Create substitution.
	subst := make(Substitution)

	// Iterate through each of the kind constraints
	// and set the ones that only have one constraint
	// since they do not have to be unified.
	// Mark down which constraints we have not visited.
	unvisited := make(map[Tvar]bool, len(sol.cs.kindConst))
	for tv, ks := range sol.cs.kindConst {
		if len(ks) == 1 {
			kinds[tv] = ks[0]
		} else {
			unvisited[tv] = true
		}
	}

	// Continuously iterate through the kind constraints
	// and unify any kinds where all dependencies have
	// already been unified.
	for len(unvisited) > 0 {
		// Track if we have visited at least one kind
		// to avoid a recursive type.
		once := false
		for tvl := range unvisited {
			// We may want to visit this constraint.
			// Check if any other unvisited tvar
			// occurs in this type.
			ks := sol.cs.kindConst[tvl]
			if canVisit := func() bool {
				for tvr := range unvisited {
					if tvl == tvr {
						continue
					}

					for _, k := range ks {
						if k.occurs(tvr) {
							return false
						}
					}
				}
				return true
			}(); !canVisit {
				continue
			}
			once = true
			delete(unvisited, tvl)

			// We can visit this constraint so let's do that.
			kinds[tvl] = ks[0]
			for _, k := range ks[1:] {
				tvr := subst.ApplyTvar(tvl)
				kind := kinds[tvr]
				s, err := unifyKinds(kinds, tvl, tvr, kind, k)
				if err != nil {
					return nil, err
				}
				subst.Merge(s)
			}
		}

		if !once {
			remaining := make([]Tvar, 0, len(unvisited))
			for tv := range unvisited {
				remaining = append(remaining, tv)
			}
			return nil, errors.Newf(codes.Internal, "unable to resolve tvars for all kinds because of a cycle: %v", remaining)
		}
	}
	return subst, nil
}

func (s *Solution) TypeOf(n Node) (Type, error) {
	a, ok := s.cs.annotations[n]
	if !ok {
		return nil, nil
	}
	if a.Err != nil {
		return nil, a.Err
	}
	return a.Type.resolveType(s.kinds)
}

func (s *Solution) PolyTypeOf(n Node) (PolyType, error) {
	a, ok := s.cs.annotations[n]
	if !ok {
		return nil, errors.Newf(codes.Internal, "no type annotation for node %T@%v", n, n.Location())
	}
	if a.Err != nil {
		return nil, a.Err
	}
	if a.Type == nil {
		return nil, errors.Newf(codes.Internal, "node %T@%v has no poly type", n, n.Location())
	}
	return a.Type.resolvePolyType(s.kinds)
}

func (s *Solution) AddConstraint(l, r PolyType) error {
	if l == nil || r == nil {
		return errors.New(codes.Invalid, "cannot add type constraint on nil types")
	}
	s.kinds = nil
	s.cs.AddTypeConst(l, r, ast.SourceLocation{})
	return s.solve()
}

func unifyTypes(kinds map[Tvar]Kind, l, r PolyType) (s Substitution, _ error) {
	//log.Printf("unifyTypes %v == %v", l, r)
	return l.unifyType(kinds, r)
}

func unifyKinds(kinds map[Tvar]Kind, tvl, tvr Tvar, l, r Kind) (Substitution, error) {
	k, s, err := l.unifyKind(kinds, r)
	if err != nil {
		return nil, err
	}
	//log.Printf("unifyKinds %v = %v == %v = %v ==> %v :: %v", tvl, l, tvr, r, k, s)
	kinds[tvr] = k
	if tvl != tvr {
		// The substitution now knows that tvl = tvr
		// No need to keep the kind constraints around for tvl
		delete(kinds, tvl)
	}
	return s, nil
}

func unifyVarAndType(kinds map[Tvar]Kind, tv Tvar, t PolyType) (Substitution, error) {
	if t.occurs(tv) {
		return nil, errors.Newf(codes.Internal, "type var %v occurs in %v creating a cycle", tv, t)
	}
	return Substitution{tv: t}, nil
}

func unifyKindsByVar(kinds map[Tvar]Kind, l, r Tvar) (Substitution, error) {
	kl, okl := kinds[l]
	kr, okr := kinds[r]
	switch {
	case okl && okr:
		return unifyKinds(kinds, l, r, kl, kr)
	case okl && !okr:
		kinds[r] = kl
		delete(kinds, l)
	}
	return nil, nil
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

// SolutionMap represents a mapping of nodes to their poly types.
type SolutionMap map[Node]PolyType

// CreateSolutionMap constructs a new solution map from the nodes and type solution.
// Any type errors in the type solution are ignored.
func CreateSolutionMap(node Node, sol TypeSolution) SolutionMap {
	solMap := make(SolutionMap)
	Walk(CreateVisitor(func(node Node) {
		t, _ := sol.PolyTypeOf(node)
		if t != nil {
			solMap[node] = t
		}

	}), node)
	return solMap
}

func (s SolutionMap) String() string {
	var builder strings.Builder
	builder.WriteString("{\n")
	nodes := make([]Node, 0, len(s))
	for n := range s {
		nodes = append(nodes, n)
	}
	SortNodes(nodes)
	for _, n := range nodes {
		t := s[n]
		fmt.Fprintf(&builder, "%T@%v: %v\n", n, n.Location(), t)
	}
	builder.WriteString("}")
	return builder.String()
}

// SortNodes sorts a list of nodes by their source locations.
func SortNodes(nodes []Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Location().Less(nodes[j].Location())
	})
}
