package inference

import (
	"fmt"
	"sort"
	"strings"
)

type Substitution map[Tvar]Type

func (s Substitution) ApplyType(t Type) Type {
	for tv, typ := range s {
		t = t.SubstType(tv, typ)
	}
	return t
}
func (s Substitution) ApplyScheme(ts Scheme) Scheme {
	for tv, typ := range s {
		ts.T = ts.T.SubstType(tv, typ)
	}
	return ts
}

func (s Substitution) ApplyKind(k Kind) Kind {
	for tv, typ := range s {
		k = k.SubstKind(tv, typ)
	}
	return k
}

func (s Substitution) ApplyEnv(env Env) Env {
	for tv, typ := range s {
		for n, ts := range env {
			env[n] = ts.Substitute(tv, typ)
		}
	}
	return env
}

func (s Substitution) ApplyTvar(tv Tvar) Tvar {
	switch t := s[tv].(type) {
	case Tvar:
		return t
	default:
		return tv
	}
}

// Merge r into l.
func (l Substitution) Merge(r Substitution) {
	// Apply right to all of l
	for tvL, tL := range l {
		l[tvL] = r.ApplyType(tL)
	}
	// Add missing key from r to l
	for tvR, tR := range r {
		if _, ok := l[tvR]; !ok {
			l[tvR] = tR
		}
	}
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
