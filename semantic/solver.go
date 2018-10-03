package semantic

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type SolutionMap map[Node]Type

func SolveTypes(program *Program, tenv TypeEnvironment, constraints []Constraint) (SolutionMap, error) {
	solution := make(SolutionMap)
	log.Println("tenv", tenv)
	log.Println("constraints", constraints)
	for i, a := range constraints {
		for j, b := range constraints {
			if i == j {
				continue
			}
			constraints[j] = b.Substitute(a).(Constraint)
		}
	}
	for _, c := range constraints {
		for n, tv := range tenv {
			s := tv.Substitute(c)
			typ, mono := s.MonoType()
			if mono {
				solution[n] = typ
			}
		}
	}
	log.Println("substituted", constraints)
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

func (s SolutionMap) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	for k, v := range s {
		fmt.Fprintf(&builder, "%#v = %v, ", k, v)
	}
	builder.WriteString("}")
	return builder.String()
}
