package semantic

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type SolutionMap map[Node]Type

func SolveTypes(program *Program, tenv map[Node]TypeVar, constraints []Constraint) (SolutionMap, error) {
	solution := make(SolutionMap)
	log.Println(constraints)
	changed := true
	for changed {
		changed = false
		for i, a := range constraints {
			for j, b := range constraints {
				if i == j {
					continue
				}
				c, chg := substitute(b, a)
				constraints[j] = c
				changed = changed || chg
			}
		}
	}
	changed = true
	for changed {
		changed = false
		for _, c := range constraints {
			typ, mono := c.right.MonoType()
			if mono {
				for n, tv := range tenv {
					if tv == c.left {
						if _, ok := solution[n]; !ok {
							solution[n] = typ
							changed = true
						}
					}
				}
			}
		}
		log.Println(constraints, tenv, solution)
	}
	if len(solution) != len(tenv) {
		for n := range tenv {
			if _, ok := solution[n]; !ok {
				solution[n] = Invalid
			}
		}
		return solution, errors.New("incomplete type solution")
	}
	return solution, nil
}

func substitute(a, b Constraint) (Constraint, bool) {
	if r, ok := a.right.(TypeVar); ok && r == b.left {
		c := a
		c.right = b.right
		log.Println(a, " <-> ", b, " >> ", c)
		return c, true
	}
	return a, false
}

func (s SolutionMap) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	for k, v := range s {
		builder.WriteString(fmt.Sprintf("%T", k))
		builder.WriteString(" = ")
		builder.WriteString(fmt.Sprintf("%v", v))
		builder.WriteString(", ")
	}
	builder.WriteString("}")
	return builder.String()
}
