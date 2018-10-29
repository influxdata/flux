package inference

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/influxdata/flux/semantic"
)

type TypeExpression interface {
	FreeVars(*Ctx) TvarSet
	Type(map[Tvar]Kind) (semantic.Type, error)
}

type Type interface {
	TypeExpression
	Occurs(tv Tvar) bool
	SubstType(tv Tvar, t Type) Type
	UnifyType(map[Tvar]Kind, Type) (Substitution, error)
	// Normalize rewrites all free variables with fresh variables starting at zero.
	//Normalize() Type
}

type Kind interface {
	TypeExpression
	SubstKind(tv Tvar, t Type) Kind
	UnifyKind(map[Tvar]Kind, Kind) (Kind, Substitution, error)
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
	ks, ok := c.kindConst[tv]
	if ok {
		for _, k := range ks {
			fvs = fvs.union(k.FreeVars(c))
		}
	}
	return fvs
}
func (l Tvar) UnifyType(kinds map[Tvar]Kind, r Type) (Substitution, error) {
	switch r := r.(type) {
	case Tvar:
		if l == r {
			return nil, nil
		}
		subst := make(Substitution)
		s, err := unifyKindsByVar(kinds, l, r)
		if err != nil {
			return nil, err
		}
		subst.Merge(s)
		subst.Merge(Substitution{l: r})
		return subst, nil
	default:
		return unifyVarAndType(kinds, l, r)
	}
}

func (tv Tvar) Type(kinds map[Tvar]Kind) (semantic.Type, error) {
	k, ok := kinds[tv]
	if !ok {
		return nil, fmt.Errorf("type variable %q is not monomorphic", tv)
	}
	return k.Type(kinds)
}

type basic semantic.Kind

func (b basic) String() string {
	return semantic.Kind(b).String()
}

func (b basic) Occurs(Tvar) bool                                    { return false }
func (b basic) SubstType(Tvar, Type) Type                           { return b }
func (b basic) Type(map[Tvar]Kind) (semantic.Type, error)           { return semantic.Kind(b), nil }
func (b basic) FreeVars(*Ctx) TvarSet                               { return nil }
func (b basic) UnifyType(map[Tvar]Kind, Type) (Substitution, error) { return nil, nil }

type invalid struct{}

func (i invalid) String() string {
	return "INVALID"
}

func (i invalid) Occurs(tv Tvar) bool                                 { return false }
func (i invalid) SubstType(Tvar, Type) Type                           { return i }
func (i invalid) Type(map[Tvar]Kind) (semantic.Type, error)           { return semantic.Invalid, nil }
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
func (l list) Type(kinds map[Tvar]Kind) (semantic.Type, error) {
	t, err := l.typ.Type(kinds)
	if err != nil {
		return nil, err
	}
	return semantic.NewArrayType(t), nil
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
			subst.Merge(s)
		}
		s, err := unifyTypes(kinds, l.ret, r.ret)
		if err != nil {
			return nil, err
		}
		subst.Merge(s)
		return subst, nil
	default:
		return nil, fmt.Errorf("cannot unify list with %T", r)
	}
}
func (f function) Type(kinds map[Tvar]Kind) (semantic.Type, error) {
	ret, err := f.ret.Type(kinds)
	if err != nil {
		return nil, err
	}
	parameters := make(map[string]semantic.Type, len(f.args))
	for l, a := range f.args {
		t, err := a.Type(kinds)
		if err != nil {
			return nil, err
		}
		parameters[l] = t
	}
	in := semantic.NewObjectType(parameters)
	return semantic.NewFunctionType(semantic.FunctionSignature{
		In:  in,
		Out: ret,
	}), nil
}

type KClass struct{}

func (k KClass) FreeVars(c *Ctx) TvarSet { return nil }
func (k KClass) SubstKind(tv Tvar, t Type) Kind {
	return k
}
func (l KClass) UnifyKind(kinds map[Tvar]Kind, r Kind) (Kind, Substitution, error) {
	//TODO
	return nil, nil, nil
}
func (k KClass) Type(map[Tvar]Kind) (semantic.Type, error) {
	return nil, errors.New("KClass has no type")
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

func (l KRecord) UnifyKind(kinds map[Tvar]Kind, k Kind) (kind Kind, _ Substitution, _ error) {
	r, ok := k.(KRecord)
	if !ok {
		return nil, nil, fmt.Errorf("cannot unify record with %T", k)
	}

	// Merge fields building up a substitution
	subst := make(Substitution)
	fields := make(map[string]Type, len(l.fields)+len(r.fields))
	for f, typL := range l.fields {
		fields[f] = typL
		typR, ok := r.fields[f]
		if !ok {
			continue
		}
		s, err := unifyTypes(kinds, typL, typR)
		if err != nil {
			fields[f] = invalid{}
		}
		subst.Merge(s)
		fields[f] = subst.ApplyType(typL)
	}
	for f, typR := range r.fields {
		_, ok := l.fields[f]
		if !ok {
			fields[f] = typR
		}
	}

	// Manage label bounds
	upper := l.upper.intersect(r.upper)
	lower := l.lower.union(r.lower)

	if !upper.isSuperSet(lower) {
		return nil, nil, fmt.Errorf("unknown record accces l: %v, u: %v", lower, upper)
	}

	kr := KRecord{
		fields: fields,
		lower:  lower,
		upper:  upper,
	}
	if kr.Invalid() {
		return nil, nil, fmt.Errorf("invalid record access %v", kr)
	}
	return kr, subst, nil
}

func (k KRecord) Type(kinds map[Tvar]Kind) (semantic.Type, error) {
	properties := make(map[string]semantic.Type, len(k.upper))
	for _, l := range k.upper {
		ft, ok := k.fields[l]
		if !ok {
			return nil, fmt.Errorf("error: missing type information for %q", l)
		}
		if _, ok := ft.(invalid); !ok {
			t, err := ft.Type(kinds)
			if err != nil {
				return nil, err
			}
			properties[l] = t
		}
	}
	return semantic.NewObjectType(properties), nil
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
