package semantic

import (
	"fmt"
	"sort"
	"strings"
)

// Substitution is a mapping of type variables to a poly type.
type Substitution map[Tvar]PolyType

func (s Substitution) applyToType(t PolyType) PolyType {
	tp, ok := t.apply(s)
	for ok {
		tp, ok = tp.apply(s)
	}
	return tp
}
func (s Substitution) applyToKind(k Kind) Kind {
	kind, ok := k.apply(s)
	for ok {
		kind, ok = kind.apply(s)
	}
	return kind
}
func (s Substitution) applyToTvar(tv Tvar) Tvar {
	tp, ok := s[tv]
	for ok {
		tvar, kk := tp.(Tvar)
		if !kk {
			break
		}
		tv = tvar
		tp, ok = s[tv]
	}
	return tv
}

func (s Substitution) String() string {
	var builder strings.Builder
	vars := make([]int, 0, len(s))
	for tv := range s {
		vars = append(vars, int(tv))
	}
	sort.Ints(vars)
	builder.WriteString("{")
	if len(s) > 1 {
		builder.WriteString("\n")
	}
	for i, tvi := range vars {
		tv := Tvar(tvi)
		if i != 0 {
			builder.WriteString(",\n")
		}
		fmt.Fprintf(&builder, "%v = %v", tv, s[tv])
	}
	builder.WriteString("}")
	return builder.String()
}
