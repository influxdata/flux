package semantic

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type SolutionMap map[Node]Type

func SolveTypes(program *Program, tenv map[Node]TypeVar, constraints []Substitutable) (SolutionMap, error) {
	solution := make(SolutionMap)
	log.Println("constraints", constraints)
	countFreeVars := 2
	for countFreeVars > 0 {
		countFreeVars--
		for i, a := range constraints {
			c, ok := a.(Constraint)
			if !ok {
				continue
			}
			for j, b := range constraints {
				if i == j {
					continue
				}
				constraints[j] = b.Substitute(c)
			}
		}
	}
	for _, c := range constraints {
		c, ok := c.(Constraint)
		if !ok {
			continue
		}
		for n, tv := range tenv {
			s := tv.Substitute(c)
			typ, mono := s.MonoType()
			if mono {
				solution[n] = typ
			}
		}
	}
	log.Println(constraints, tenv, solution)
	// Validate we got a complete solution
	if len(solution) != len(tenv) {
		// Populate the missing solutions with Invalid
		for n := range tenv {
			if _, ok := solution[n]; !ok {
				solution[n] = Invalid
			}
		}
		return solution, errors.New("incomplete type solution")
	}
	return solution, nil
}

//func substitute(a, b Constraint) Constraint {
//	if r, ok := a.right.(TypeVar); ok && r == b.left {
//		c := a
//		c.right = b.right
//		log.Println(a, " <-> ", b, " >> ", c)
//		return c
//	}
//	return a
//}

func (s SolutionMap) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	for k, v := range s {
		fmt.Fprintf(&builder, "%#v", k)
		builder.WriteString(" = ")
		fmt.Fprintf(&builder, "%v", v)
		builder.WriteString(", ")
	}
	builder.WriteString("}")
	return builder.String()
}
