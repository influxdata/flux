package semantic

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type TypeExpression interface {
	FreeVars(*Constraints) TvarSet
	Type(map[Tvar]KindConstraint) (Type, error)
	MonoType() (Type, bool)
	resolvePolyType(map[Tvar]KindConstraint) (PolyType, error)
}

type PolyType interface {
	TypeExpression
	Occurs(tv Tvar) bool
	SubstType(tv Tvar, t PolyType) PolyType
	UnifyType(map[Tvar]KindConstraint, PolyType) (Substitution, error)

	Equal(PolyType) bool

	Kind() Kind
}

type KindConstraint interface {
	TypeExpression
	SubstKind(tv Tvar, t PolyType) KindConstraint
	UnifyKind(map[Tvar]KindConstraint, KindConstraint) (KindConstraint, Substitution, error)
}

// Tvar represents a type variable meaning its type could be any possible type.
type Tvar int

func (tv Tvar) Kind() Kind {
	return Invalid
}
func (tv Tvar) String() string {
	if tv == 0 {
		// tv == 0 is not considered valid,
		// we denote that by using a different
		// symbol other than t0.
		return "><"
	}
	return fmt.Sprintf("t%d", int(tv))
}

func (a Tvar) Occurs(b Tvar) bool {
	return a == b
}
func (a Tvar) SubstType(b Tvar, t PolyType) PolyType {
	if a == b {
		return t
	}
	return a
}
func (tv Tvar) FreeVars(c *Constraints) TvarSet {
	fvs := TvarSet{tv}
	if c != nil {
		ks, ok := c.kindConst[tv]
		if ok {
			for _, k := range ks {
				fvs = fvs.union(k.FreeVars(c))
			}
		}
	}
	return fvs
}
func (l Tvar) UnifyType(kinds map[Tvar]KindConstraint, r PolyType) (Substitution, error) {
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

func (tv Tvar) Type(kinds map[Tvar]KindConstraint) (Type, error) {
	k, ok := kinds[tv]
	if !ok {
		return nil, fmt.Errorf("type variable %q is not monomorphic", tv)
	}
	return k.Type(kinds)
}
func (tv Tvar) MonoType() (Type, bool) {
	return nil, false
}
func (tv Tvar) resolvePolyType(kinds map[Tvar]KindConstraint) (PolyType, error) {
	k, ok := kinds[tv]
	if !ok {
		return tv, nil
	}
	return k.resolvePolyType(kinds)
}
func (tv Tvar) Equal(t PolyType) bool {
	switch t := t.(type) {
	case Tvar:
		return tv == t
	default:
		return false
	}
}

// PolyType methods for Kind
func (k Kind) Occurs(Tvar) bool                                          { return false }
func (k Kind) SubstType(Tvar, PolyType) PolyType                         { return k }
func (k Kind) Type(map[Tvar]KindConstraint) (Type, error)                { return k, nil }
func (k Kind) MonoType() (Type, bool)                                    { return k, true }
func (k Kind) resolvePolyType(map[Tvar]KindConstraint) (PolyType, error) { return k, nil }
func (k Kind) FreeVars(*Constraints) TvarSet                             { return nil }
func (k Kind) UnifyType(kinds map[Tvar]KindConstraint, t PolyType) (Substitution, error) {
	switch t := t.(type) {
	case Kind:
		if t != k {
			return nil, fmt.Errorf("%v != %v", k, t)
		}
	case Tvar:
		return t.UnifyType(kinds, k)
	default:
		return nil, fmt.Errorf("cannot unify %v with %T", k, t)
	}
	return nil, nil
}
func (k Kind) Equal(t PolyType) bool {
	switch t := t.(type) {
	case Kind:
		return t == k
	default:
		return false
	}
}

type invalid struct{}

func (i invalid) String() string {
	return "INVALID"
}

func (i invalid) Kind() Kind                                                        { return Invalid }
func (i invalid) Occurs(tv Tvar) bool                                               { return false }
func (i invalid) SubstType(Tvar, PolyType) PolyType                                 { return i }
func (i invalid) Type(map[Tvar]KindConstraint) (Type, error)                        { return Invalid, nil }
func (i invalid) MonoType() (Type, bool)                                            { return nil, false }
func (i invalid) resolvePolyType(map[Tvar]KindConstraint) (PolyType, error)         { return i, nil }
func (i invalid) FreeVars(*Constraints) TvarSet                                     { return nil }
func (i invalid) UnifyType(map[Tvar]KindConstraint, PolyType) (Substitution, error) { return nil, nil }
func (i invalid) Equal(t PolyType) bool {
	switch t.(type) {
	case invalid:
		return true
	default:
		return false
	}
}

type list struct {
	typ PolyType
}

func NewArrayPolyType(elementType PolyType) PolyType {
	return list{typ: elementType}
}

func (l list) Kind() Kind {
	return Array
}
func (l list) String() string {
	return fmt.Sprintf("[%v]", l.typ)
}

func (l list) Occurs(tv Tvar) bool {
	return l.typ.Occurs(tv)
}
func (l list) SubstType(tv Tvar, t PolyType) PolyType {
	return list{
		typ: l.typ.SubstType(tv, t),
	}
}
func (l list) FreeVars(c *Constraints) TvarSet {
	return l.typ.FreeVars(c)
}
func (a list) UnifyType(kinds map[Tvar]KindConstraint, b PolyType) (Substitution, error) {
	switch b := b.(type) {
	case list:
		return unifyTypes(kinds, a.typ, b.typ)
	case Tvar:
		return b.UnifyType(kinds, a)
	default:
		return nil, fmt.Errorf("cannot unify list with %T", b)
	}
}
func (l list) Type(kinds map[Tvar]KindConstraint) (Type, error) {
	t, err := l.typ.Type(kinds)
	if err != nil {
		return nil, err
	}
	return NewArrayType(t), nil
}
func (l list) MonoType() (Type, bool) {
	t, ok := l.typ.MonoType()
	if !ok {
		return nil, false
	}
	return NewArrayType(t), true
}
func (l list) resolvePolyType(kinds map[Tvar]KindConstraint) (PolyType, error) {
	t, err := l.typ.resolvePolyType(kinds)
	if err != nil {
		return nil, err
	}
	return list{
		typ: t,
	}, nil
}
func (l list) Equal(t PolyType) bool {
	switch t := t.(type) {
	case list:
		return l.typ.Equal(t.typ)
	default:
		return false
	}
}

// pipeLabel is a hidden label on which all pipe arguments are passed according to type inference.
const pipeLabel = "|pipe|"

type function struct {
	parameters   map[string]PolyType
	required     LabelSet
	ret          PolyType
	pipeArgument string
}

type FunctionPolySignature struct {
	Parameters   map[string]PolyType
	Required     LabelSet
	Return       PolyType
	PipeArgument string
}

func NewFunctionPolyType(sig FunctionPolySignature) PolyType {
	return function{
		parameters:   sig.Parameters,
		required:     sig.Required.remove(sig.PipeArgument),
		ret:          sig.Return,
		pipeArgument: sig.PipeArgument,
	}
}

func (f function) Kind() Kind {
	return Function
}
func (f function) Signature() FunctionPolySignature {
	parameters := make(map[string]PolyType, len(f.parameters))
	for k, t := range f.parameters {
		parameters[k] = t
	}
	return FunctionPolySignature{
		Parameters:   parameters,
		Required:     f.required.copy(),
		Return:       f.ret,
		PipeArgument: f.pipeArgument,
	}
}

func (f function) String() string {
	var builder strings.Builder
	keys := make([]string, 0, len(f.parameters))
	for k := range f.parameters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	builder.WriteString("(")
	for i, k := range keys {
		if i != 0 {
			builder.WriteString(", ")
		}
		if f.required.contains(k) {
			builder.WriteString("^")
		}
		if f.pipeArgument == k {
			builder.WriteString("<-")
		}
		fmt.Fprintf(&builder, "%s: %v", k, f.parameters[k])
	}
	fmt.Fprintf(&builder, ") -> %v", f.ret)
	return builder.String()
}

func (f function) Occurs(tv Tvar) bool {
	for _, a := range f.parameters {
		occurs := a.Occurs(tv)
		if occurs {
			return true
		}
	}
	return f.ret.Occurs(tv)
}

func (f function) SubstType(tv Tvar, typ PolyType) PolyType {
	parameters := make(map[string]PolyType, len(f.parameters))
	for k, t := range f.parameters {
		parameters[k] = t.SubstType(tv, typ)
	}
	return function{
		parameters:   parameters,
		required:     f.required.copy(),
		ret:          f.ret.SubstType(tv, typ),
		pipeArgument: f.pipeArgument,
	}
}
func (f function) FreeVars(c *Constraints) TvarSet {
	fvs := f.ret.FreeVars(c)
	for _, t := range f.parameters {
		fvs = fvs.union(t.FreeVars(c))
	}
	return fvs
}
func (l function) UnifyType(kinds map[Tvar]KindConstraint, r PolyType) (Substitution, error) {
	switch r := r.(type) {
	case function:
		missing := l.required.diff(r.required)
		for _, lbl := range missing {
			if _, ok := r.parameters[lbl]; !ok && lbl != l.pipeArgument {
				// Pipe parameters are validated below
				return nil, fmt.Errorf("missing required parameter %q", lbl)
			}
		}
		subst := make(Substitution)
		for f, tl := range l.parameters {
			tr, ok := r.parameters[f]
			if !ok {
				// We already validated missing parameters, this must be the pipe parameter.
				continue
			}
			typl := subst.ApplyType(tl)
			typr := subst.ApplyType(tr)
			s, err := unifyTypes(kinds, typl, typr)
			if err != nil {
				return nil, err
			}
			subst.Merge(s)
		}
		// Check for valid pipe parameter
		if l.pipeArgument != "" {
			pipel, okl := l.lookupPipe(l.pipeArgument)
			if !okl {
				return nil, fmt.Errorf("left missing pipe parameter %v %v", l.pipeArgument, l)
			}
			piper, okr := r.lookupPipe(l.pipeArgument)
			if !okr {
				return nil, fmt.Errorf("right missing pipe parameter %v %v", l.pipeArgument, r)
			}
			s, err := unifyTypes(kinds, pipel, piper)
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
	case Tvar:
		return r.UnifyType(kinds, l)
	default:
		return nil, fmt.Errorf("cannot unify list with %T", r)
	}
}

func (f function) lookupPipe(label string) (PolyType, bool) {
	t, ok := f.parameters[label]
	if ok {
		return t, true
	}
	t, ok = f.parameters[pipeLabel]
	return t, ok
}

func (f function) Type(kinds map[Tvar]KindConstraint) (Type, error) {
	ret, err := f.ret.Type(kinds)
	if err != nil {
		return nil, err
	}
	parameters := make(map[string]Type, len(f.parameters))
	required := f.required.copy()
	for l, a := range f.parameters {
		if l == pipeLabel && f.pipeArgument != "" {
			l = f.pipeArgument
			required = required.remove(pipeLabel)
			required = append(required, l)
		}
		t, err := a.Type(kinds)
		if err != nil {
			return nil, err
		}
		parameters[l] = t
	}
	return NewFunctionType(FunctionSignature{
		Parameters:   parameters,
		Required:     required,
		Return:       ret,
		PipeArgument: f.pipeArgument,
	}), nil
}
func (f function) MonoType() (Type, bool) {
	ret, ok := f.ret.MonoType()
	if !ok {
		return nil, false
	}
	parameters := make(map[string]Type, len(f.parameters))
	required := f.required.copy()
	for l, a := range f.parameters {
		if l == pipeLabel && f.pipeArgument != "" {
			l = f.pipeArgument
			required = required.remove(pipeLabel)
			required = append(required, l)
		}
		t, ok := a.MonoType()
		if !ok {
			return nil, false
		}
		parameters[l] = t
	}
	return NewFunctionType(FunctionSignature{
		Parameters:   parameters,
		Required:     required,
		Return:       ret,
		PipeArgument: f.pipeArgument,
	}), true
}
func (f function) resolvePolyType(kinds map[Tvar]KindConstraint) (PolyType, error) {
	ret, err := f.ret.resolvePolyType(kinds)
	if err != nil {
		return nil, err
	}
	parameters := make(map[string]PolyType, len(f.parameters))
	required := f.required.copy()
	for l, v := range f.parameters {
		if l == pipeLabel && f.pipeArgument != "" {
			l = f.pipeArgument
			required = required.remove(pipeLabel)
			required = append(required, l)
		}
		t, err := v.resolvePolyType(kinds)
		if err != nil {
			return nil, err
		}
		parameters[l] = t
	}
	return function{
		parameters:   parameters,
		required:     required,
		ret:          ret,
		pipeArgument: f.pipeArgument,
	}, nil
}
func (f function) Equal(t PolyType) bool {
	switch t := t.(type) {
	case function:
		if len(f.parameters) != len(t.parameters) ||
			!f.required.equal(t.required) ||
			!f.ret.Equal(t.ret) ||
			f.pipeArgument != t.pipeArgument {
			return false
		}
		for k, p := range f.parameters {
			op, ok := t.parameters[k]
			if !ok || !p.Equal(op) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

type object struct {
	krecord KRecord
}

func NewEmptyObjectPolyType() PolyType {
	return NewObjectPolyType(nil, LabelSet{}, LabelSet{})
}
func NewObjectPolyType(properties map[string]PolyType, lower, upper LabelSet) PolyType {
	return object{
		krecord: KRecord{
			properties: properties,
			lower:      lower,
			upper:      upper,
		},
	}
}

func (o object) Kind() Kind {
	return Object
}
func (o object) String() string {
	return o.krecord.String()
}

func (o object) Occurs(tv Tvar) bool {
	for _, p := range o.krecord.properties {
		occurs := p.Occurs(tv)
		if occurs {
			return true
		}
	}
	return false
}

func (o object) SubstType(tv Tvar, typ PolyType) PolyType {
	properties := make(map[string]PolyType, len(o.krecord.properties))
	for k, t := range o.krecord.properties {
		properties[k] = t.SubstType(tv, typ)
	}
	return object{
		krecord: KRecord{
			properties: properties,
			lower:      o.krecord.lower.copy(),
			upper:      o.krecord.upper.copy(),
		},
	}
}
func (o object) FreeVars(c *Constraints) TvarSet {
	var fvs TvarSet
	for _, t := range o.krecord.properties {
		fvs = fvs.union(t.FreeVars(c))
	}
	return fvs
}

func (l object) UnifyType(kinds map[Tvar]KindConstraint, r PolyType) (Substitution, error) {
	switch r := r.(type) {
	case object:
		_, subst, err := l.krecord.UnifyKind(kinds, r.krecord)
		return subst, err
	case Tvar:
		return r.UnifyType(kinds, l)
	default:
		return nil, fmt.Errorf("cannot unify object with %T", r)
	}
}
func (o object) Type(kinds map[Tvar]KindConstraint) (Type, error) {
	return o.krecord.Type(kinds)
}
func (o object) MonoType() (Type, bool) {
	return o.krecord.MonoType()
}
func (o object) resolvePolyType(kinds map[Tvar]KindConstraint) (PolyType, error) {
	return o.krecord.resolvePolyType(kinds)
}
func (o object) Equal(t PolyType) bool {
	switch t := t.(type) {
	case object:
		if len(o.krecord.properties) != len(t.krecord.properties) ||
			!o.krecord.lower.equal(t.krecord.lower) ||
			!o.krecord.upper.equal(t.krecord.upper) {
			return false
		}
		for k, p := range o.krecord.properties {
			op, ok := t.krecord.properties[k]
			if !ok || !p.Equal(op) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (o object) KindConstraint() KindConstraint {
	return o.krecord
}

type KindConstrainter interface {
	KindConstraint() KindConstraint
}

type KClass struct{}

func (k KClass) FreeVars(c *Constraints) TvarSet { return nil }
func (k KClass) SubstKind(tv Tvar, t PolyType) KindConstraint {
	return k
}
func (l KClass) UnifyKind(kinds map[Tvar]KindConstraint, r KindConstraint) (KindConstraint, Substitution, error) {
	//TODO
	return nil, nil, nil
}
func (k KClass) Type(map[Tvar]KindConstraint) (Type, error) {
	return nil, errors.New("KClass has no type")
}
func (k KClass) MonoType() (Type, bool) {
	return nil, false
}
func (k KClass) resolvePolyType(map[Tvar]KindConstraint) (PolyType, error) {
	return nil, errors.New("KClass has no poly type")
}

type KRecord struct {
	properties map[string]PolyType
	lower      LabelSet
	upper      LabelSet
}

func (k KRecord) String() string {
	return fmt.Sprintf("{%v %v %v}", k.properties, k.lower, k.upper)
}

func (k KRecord) Invalid() bool {
	for _, l := range k.lower {
		t := k.properties[l]
		_, ok := t.(invalid)
		if ok {
			return true
		}
	}
	return false
}

func (k KRecord) SubstKind(tv Tvar, t PolyType) KindConstraint {
	properties := make(map[string]PolyType)
	for k, f := range k.properties {
		properties[k] = f.SubstType(tv, t)
	}
	return KRecord{
		properties: properties,
		upper:      k.upper.copy(),
		lower:      k.lower.copy(),
	}
}
func (k KRecord) FreeVars(c *Constraints) TvarSet {
	var fvs TvarSet
	for _, f := range k.properties {
		fvs = fvs.union(f.FreeVars(c))
	}
	return fvs
}

func (l KRecord) UnifyKind(kinds map[Tvar]KindConstraint, k KindConstraint) (kind KindConstraint, _ Substitution, _ error) {
	r, ok := k.(KRecord)
	if !ok {
		return nil, nil, fmt.Errorf("cannot unify record with %T", k)
	}

	// Merge properties building up a substitution
	subst := make(Substitution)
	properties := make(map[string]PolyType, len(l.properties)+len(r.properties))
	for f, typL := range l.properties {
		properties[f] = typL
		typR, ok := r.properties[f]
		if !ok {
			continue
		}
		s, err := unifyTypes(kinds, typL, typR)
		if err != nil {
			properties[f] = invalid{}
		}
		subst.Merge(s)
		properties[f] = subst.ApplyType(typL)
	}
	for f, typR := range r.properties {
		_, ok := l.properties[f]
		if !ok {
			properties[f] = typR
		}
	}

	// Manage label bounds
	upper := l.upper.intersect(r.upper)
	lower := l.lower.union(r.lower)

	if !upper.isSuperSet(lower) {
		return nil, nil, fmt.Errorf("unknown record accces l: %v, u: %v", lower, upper)
	}

	kr := KRecord{
		properties: properties,
		lower:      lower,
		upper:      upper,
	}
	if kr.Invalid() {
		return nil, nil, fmt.Errorf("invalid record access %v", kr)
	}
	return kr, subst, nil
}

func (k KRecord) Type(kinds map[Tvar]KindConstraint) (Type, error) {
	properties := make(map[string]Type, len(k.properties))
	for l, ft := range k.properties {
		if _, ok := ft.(invalid); !ok {
			t, err := ft.Type(kinds)
			if err != nil {
				return nil, err
			}
			properties[l] = t
		}
	}
	return NewObjectType(properties), nil
}
func (k KRecord) MonoType() (Type, bool) {
	properties := make(map[string]Type, len(k.properties))
	for l, ft := range k.properties {
		if _, ok := ft.(invalid); !ok {
			t, ok := ft.MonoType()
			if !ok {
				return nil, false
			}
			properties[l] = t
		}
	}
	return NewObjectType(properties), false
}
func (k KRecord) resolvePolyType(kinds map[Tvar]KindConstraint) (PolyType, error) {
	properties := make(map[string]PolyType, len(k.upper))
	for l, ft := range k.properties {
		if _, ok := ft.(invalid); !ok {
			t, err := ft.resolvePolyType(kinds)
			if err != nil {
				return nil, err
			}
			properties[l] = t
		}
	}
	return NewObjectPolyType(properties, k.lower, k.upper), nil
}

type Comparable struct{}
type Addable struct{}
type Number struct{}

type Scheme struct {
	T    PolyType
	Free TvarSet
}

func (s Scheme) Substitute(tv Tvar, t PolyType) Scheme {
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
