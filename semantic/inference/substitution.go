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

func (a Substitution) Merge(b Substitution) (m Substitution) {
	m = make(Substitution, len(a)+len(b))
	// Apply B to all of A
	for tvA, tA := range a {
		m[tvA] = b.ApplyType(tA)
	}
	// Add any missing from B
	for tvB, tB := range b {
		_, ok := m[tvB]
		if !ok {
			m[tvB] = tB
		}
	}
	return m
}

func (s Substitution) String() string {
	var builder strings.Builder
	vars := make([]int, 0, len(s))
	for tv := range s {
		vars = append(vars, int(tv))
	}
	sort.Ints(vars)
	builder.WriteString("{\n")
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
