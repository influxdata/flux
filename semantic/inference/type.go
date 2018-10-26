package inference

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/influxdata/flux/semantic"
)

type Basic interface {
}

type Type interface {
	Occurs(tv Tvar) bool
	SubstType(tv Tvar, t Type) Type
	FreeVars(*Ctx) TvarSet
	UnifyType(map[Tvar]Kind, Type) (Substitution, error)
}

type Kind interface {
	Invalid() bool
	SubstKind(tv Tvar, t Type) Kind
	FreeVars(*Ctx) TvarSet
	UnifyKind(map[Tvar]Kind, Tvar, Kind) (Substitution, error)
	Merge(c *Ctx, k Kind) Kind
}

type TypeClass interface {
}

type Tvar int

func (tv Tvar) String() string {
	return fmt.Sprintf("t%d", int(tv))
}

func (a Tvar) Occurs(b Tvar) bool {
	return a == b
}
func (a Tvar) SubstType(b Tvar, t Type) Type {
	if a == b {
		return t
	}
	return a
}
func (tv Tvar) FreeVars(c *Ctx) TvarSet {
	fvs := TvarSet{tv}
	k, ok := c.kindConst[tv]
	if ok {
		fvs = fvs.union(k.FreeVars(c))
	}
	return fvs
}
func (l Tvar) UnifyType(kinds map[Tvar]Kind, r Type) (Substitution, error) {
	switch r := r.(type) {
	case Tvar:
		if l == r {
			return nil, nil
		}
		subst, err := unifyKindsByVar(kinds, l, r)
		if err != nil {
			return nil, err
		}
		s := subst.Merge(Substitution{l: r})
		return s, nil
	default:
		return unifyVarAndType(kinds, l, r)
	}
}

type basic semantic.Kind

func (b basic) String() string {
	return semantic.Kind(b).String()
}

func (b basic) Occurs(Tvar) bool                                    { return false }
func (b basic) SubstType(Tvar, Type) Type                           { return b }
func (b basic) FreeVars(*Ctx) TvarSet                               { return nil }
func (b basic) UnifyType(map[Tvar]Kind, Type) (Substitution, error) { return nil, nil }

type invalid struct{}

func (i invalid) String() string {
	return "INVALID"
}

func (i invalid) Occurs(tv Tvar) bool                                 { return false }
func (i invalid) SubstType(Tvar, Type) Type                           { return i }
func (i invalid) FreeVars(*Ctx) TvarSet                               { return nil }
func (i invalid) UnifyType(map[Tvar]Kind, Type) (Substitution, error) { return nil, nil }

type list struct {
	typ Type
}

func (l list) String() string {
	return fmt.Sprintf("[%v]", l.typ)
}

func (l list) Occurs(tv Tvar) bool {
	return l.typ.Occurs(tv)
}
func (l list) SubstType(tv Tvar, t Type) Type {
	return list{
		typ: l.typ.SubstType(tv, t),
	}
}
func (l list) FreeVars(c *Ctx) TvarSet {
	return l.typ.FreeVars(c)
}
func (a list) UnifyType(kinds map[Tvar]Kind, b Type) (Substitution, error) {
	switch b := b.(type) {
	case list:
		return unifyTypes(kinds, a.typ, b.typ)
	default:
		return nil, fmt.Errorf("cannot unify list with %T", b)
	}
}

type function struct {
	args     map[string]Type
	required LabelSet
	ret      Type
}

func (f function) String() string {
	var builder strings.Builder
	keys := make([]string, 0, len(f.args))
	for k := range f.args {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	builder.WriteString("(")
	for i, k := range keys {
		if i != 0 {
			builder.WriteString(", ")
		}
		if f.required.contains(k) {
			builder.WriteString("*")
		}
		fmt.Fprintf(&builder, "%s: %v", k, f.args[k])
	}
	fmt.Fprintf(&builder, ") -> %v", f.ret)
	return builder.String()
}

func (f function) Occurs(tv Tvar) bool {
	for _, a := range f.args {
		occurs := a.Occurs(tv)
		if occurs {
			return true
		}
	}
	return f.ret.Occurs(tv)
}

func (f function) SubstType(tv Tvar, typ Type) Type {
	args := make(map[string]Type, len(f.args))
	for k, t := range f.args {
		args[k] = t.SubstType(tv, typ)
	}
	return function{
		args:     args,
		required: f.required.copy(),
		ret:      f.ret.SubstType(tv, typ),
	}
}
func (f function) FreeVars(c *Ctx) TvarSet {
	fvs := f.ret.FreeVars(c)
	for _, t := range f.args {
		fvs = fvs.union(t.FreeVars(c))
	}
	return fvs
}
func (l function) UnifyType(kinds map[Tvar]Kind, r Type) (Substitution, error) {
	switch r := r.(type) {
	case function:
		if !l.required.isSubSet(r.required) {
			return nil, fmt.Errorf("missing required parameters l: %v r: %v", l.required, r.required)
		}
		subst := make(Substitution)
		for f, tl := range l.args {
			tr, ok := r.args[f]
			if !ok && l.required.contains(f) {
				return nil, errors.New("missing")
			}
			typl := subst.ApplyType(tl)
			typr := subst.ApplyType(tr)
			s, err := unifyTypes(kinds, typl, typr)
			if err != nil {
				return nil, err
			}
			subst = subst.Merge(s)
		}
		s, err := unifyTypes(kinds, l.ret, r.ret)
		if err != nil {
			return nil, err
		}
		subst = subst.Merge(s)
		return subst, nil
	default:
		return nil, fmt.Errorf("cannot unify list with %T", r)
	}
}

type KClass struct{}

func (k KClass) Invalid() bool           { return false }
func (k KClass) FreeVars(c *Ctx) TvarSet { return nil }
func (k KClass) SubstKind(tv Tvar, t Type) Kind {
	return k
}
func (a KClass) UnifyKind(kinds map[Tvar]Kind, tv Tvar, b Kind) (Substitution, error) {
	//TODO
	return nil, nil
}
func (a KClass) Merge(c *Ctx, b Kind) Kind {
	//TODO
	return a
}

type KRecord struct {
	fields map[string]Type
	lower  LabelSet
	upper  LabelSet
}

func (k KRecord) String() string {
	return fmt.Sprintf("{%v %v %v}", k.fields, k.lower, k.upper)
}

func (k KRecord) Invalid() bool {
	for _, l := range k.lower {
		t := k.fields[l]
		_, ok := t.(invalid)
		if ok {
			return true
		}
	}
	return false
}

func (k KRecord) SubstKind(tv Tvar, t Type) Kind {
	fields := make(map[string]Type)
	for k, f := range k.fields {
		fields[k] = f.SubstType(tv, t)
	}
	return KRecord{
		fields: fields,
		upper:  k.upper.copy(),
		lower:  k.lower.copy(),
	}
}
func (k KRecord) FreeVars(c *Ctx) TvarSet {
	var fvs TvarSet
	for _, f := range k.fields {
		fvs = fvs.union(f.FreeVars(c))
	}
	return fvs
}

func (a KRecord) UnifyKind(kinds map[Tvar]Kind, tv Tvar, k Kind) (Substitution, error) {
	b, ok := k.(KRecord)
	if !ok {
		return nil, fmt.Errorf("cannot unify record with %T", k)
	}

	// Merge fields building up a substitution
	subst := make(Substitution)
	fields := make(map[string]Type, len(a.fields)+len(b.fields))
	for f, typA := range a.fields {
		typB, ok := b.fields[f]
		if !ok {
			fields[f] = typA
			continue
		}
		s, err := typA.UnifyType(kinds, typB)
		if err != nil {
			fields[f] = invalid{}
		}
		subst = subst.Merge(s)
	}
	for f, typB := range b.fields {
		_, ok := a.fields[f]
		if !ok {
			fields[f] = typB
		}
	}

	// Manage label bounds
	upper := a.upper.intersect(b.upper)
	lower := a.lower.union(b.lower)

	if !upper.isSuperSet(lower) {
		return nil, fmt.Errorf("unknown record accces l: %v, u: %v", lower, upper)
	}

	kr := KRecord{
		fields: fields,
		lower:  lower,
		upper:  upper,
	}
	if kr.Invalid() {
		return nil, fmt.Errorf("invalid record access %v", kr)
	}
	kinds[tv] = kr
	return subst, nil
}

func (a KRecord) Merge(c *Ctx, k Kind) Kind {
	b, ok := k.(KRecord)
	if !ok {
		//return nil, fmt.Errorf("cannot merge record with %T", k)
		panic("boo")
	}

	// Merge fields building up a substitution
	fields := make(map[string]Type, len(a.fields)+len(b.fields))
	for f, typA := range a.fields {
		fields[f] = typA
	}
	for f, typB := range b.fields {
		_, ok := fields[f]
		if !ok {
			fields[f] = typB
		}
	}

	// Manage label bounds
	upper := a.upper.intersect(b.upper)
	lower := a.lower.union(b.lower)

	for _, l := range lower {
		// Passing nil types here?
		// Means fail, how to fail here?
		c.AddTypeConst(a.fields[l], b.fields[l])
	}
	return KRecord{
		fields: fields,
		lower:  lower,
		upper:  upper,
	}
}

type Comparable struct{}
type Addable struct{}
type Number struct{}

type Scheme struct {
	T    Type
	Free TvarSet
}

func (s Scheme) Substitute(tv Tvar, t Type) Scheme {
	fvs := make(TvarSet, 0, len(s.Free))
	for _, ftv := range s.Free {
		if ftv != tv {
			fvs = append(fvs, ftv)
		}
	}
	return Scheme{
		T:    s.T.SubstType(tv, t),
		Free: fvs,
	}
}
