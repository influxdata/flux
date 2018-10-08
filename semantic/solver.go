package semantic

import (
	"fmt"
	"strings"
)

func SolveTypes(tenv TypeEnvironment, constraints ConstraintSet) (SolutionMap, error) {
	solution := make(SolutionMap, len(tenv))
	//log.Println("tenv", tenv)
	//log.Println("constraints", constraints)

	substitution := make(Substitution, len(tenv))
	for _, c := range constraints {
		// Apply substitution to the current equation
		e := Equation{
			left:  c.left,
			right: c.right,
		}
		//log.Println("applying", e)
		for _, tv := range e.Vars() {
			s := substitution[tv]
			if s != nil {
				c := substitution.Constraint(tv)
				e = e.Apply(c)
			}
		}

		cs, err := e.Constraints()
		if err != nil {
			return nil, err
		}
		//log.Println("equation cs", cs)
		for _, c := range cs {
			// Substitute the constraint into the substitution
			for tv, s := range substitution {
				if s != nil {
					substitution[tv] = s.Substitute(c)
				}
			}
			// Add the constraint to the substitution
			substitution[c.left] = c.right
		}
	}

	// TODO(nathanielc): Add occurence check

	//log.Println("substitution", substitution)
	for n, tv := range tenv {
		s := substitution[tv]
		typ, mono := s.MonoType()
		if mono {
			solution[n] = typ
		}
		if e, ok := n.(Expression); ok {
			e.setTypeScheme(s)
		}
	}
	return solution, nil
}

type SolutionMap map[Node]Type

func (s SolutionMap) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	for k, v := range s {
		fmt.Fprintf(&builder, "%#v = %v, ", k, v)
	}
	builder.WriteString("}")
	return builder.String()
}

// Equation represents two Substitutables that are equated
type Equation struct {
	left  Substitutable
	right Substitutable
}

// Apply substitues the constraint to both sides of the equation
func (e Equation) Apply(c Constraint) Equation {
	e.left = e.left.Substitute(c)
	e.right = e.right.Substitute(c)
	return e
}

// Vars reports the type vars referenced within the equation
func (e Equation) Vars() []TypeVar {
	return append(e.left.Vars(), e.right.Vars()...)
}

// Constraints returns a list of constraints produced by this equation.
// In most cases the equation is a direct constraint, when one side is a type var.
// It is possible for both sides to be compound substitutable expressions in which case
// there may be up to two constraints, one for each side.
func (e Equation) Constraints() ([]Constraint, error) {
	lv, lok := e.left.(TypeVar)
	rv, rok := e.right.(TypeVar)
	if lok && rok {
		return []Constraint{{
			left:  lv,
			right: rv,
		}}, nil
	} else if lok {
		return []Constraint{{
			left:  lv,
			right: e.right,
		}}, nil
	} else if rok {
		return []Constraint{{
			left:  rv,
			right: e.left,
		}}, nil
	} else {
		// Neither hand is a single TypeVar,
		// each hand could become its own Constraint
		var cons []Constraint
		if c, ok := e.left.(Constraint); ok {
			cons = append(cons, c)
		}
		if c, ok := e.right.(Constraint); ok {
			cons = append(cons, c)
		}
		if len(cons) == 0 {
			lt, lok := e.left.(Type)
			rt, rok := e.right.(Type)
			if lok && rok {
				if lt != rt {
					return nil, fmt.Errorf("type error: %v != %v", lt, rt)
				}
			}
		}
		return cons, nil
	}
}

func (e Equation) String() string {
	return fmt.Sprintf("%v = %v", e.left, e.right)
}

// Substitution is a mapping of TypeVar (which are ints) to Substitutable.
// A Constraint can be created from the Substitution given a TypeVar.
type Substitution []Substitutable

func (s Substitution) Constraint(tv TypeVar) Constraint {
	return Constraint{
		left:  tv,
		right: s[tv],
	}
}

func (s Substitution) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	if len(s) > 1 {
		builder.WriteString("\n")
	}
	for i, sub := range s {
		if i != 0 {
			builder.WriteString(",\n")
		}
		fmt.Fprintf(&builder, "%v", Constraint{
			left:  TypeVar(i),
			right: sub,
		})
	}
	if len(s) > 1 {
		builder.WriteString("\n")
	}
	builder.WriteString("}")
	return builder.String()
}
