package semantic

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// TypeExpression represents an expression describing a type.
type TypeExpression interface {
	// MonoType produces a monotype of the expression.
	MonoType() (Type, bool)
	// freeVars reports the free unbound type variables in the expression.
	freeVars(*Constraints) TvarSet
	// resolveType produces a monotype of the type expression given the kind constraints.
	resolveType(map[Tvar]Kind) (Type, error)
	// resolvePolyType applies the kind constraints producing a new self describing poly type.
	resolvePolyType(map[Tvar]Kind) (PolyType, error)
}

// PolyType represents a polymorphic type, meaning that the type may have multiple free type variables.
type PolyType interface {
	TypeExpression

	// occurs reports whether tv is a free variable in the type.
	occurs(tv Tvar) bool

	// apply applies a substitution to a type
	// apply returns false if there were no substitutions made
	apply(sub Substitution) (PolyType, bool)

	// unifyType unifies two parametric types by producing a substitution that
	// when applied to both types, produces the same parametric type.
	//
	// If the input substitution is already a unifier for both types it returns
	// that same substitution. Otherwise the input substitution is not a unifier
	// for the two types, and it returns an expanded one that is.
	unifyType(map[Tvar]Kind, Substitution, PolyType) (Substitution, error)

	// Equal reports if two types are the same.
	Equal(PolyType) bool

	// Nature reports the primitive description of the type.
	Nature() Nature
}

// Kind is a constraint in the kind domain.
type Kind interface {
	TypeExpression

	// apply applies a substitution to a kind
	// apply returns false if there were no substitutions made
	apply(sub Substitution) (Kind, bool)

	// unifyKind unifies two kinds producing a new merged kind and a new updated
	// substitution.
	//
	// If the input substitution is already a unifier for both kinds it returns
	// that same substitution. Otherwise the input substitution is not a unifier
	// for the two kinds, and it returns an expanded one that is.
	unifyKind(map[Tvar]Kind, Substitution, Kind) (Kind, Substitution, error)

	// occurs reports whether tv occurs in this kind.
	occurs(tv Tvar) bool
}

// Tvar represents a type variable meaning its type could be any possible type.
type Tvar int

func (tv Tvar) Nature() Nature {
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

func (a Tvar) occurs(b Tvar) bool {
	return a == b
}
func (a Tvar) apply(sub Substitution) (PolyType, bool) {
	if tp, ok := sub[a]; ok {
		return tp, true
	}
	return a, false
}
func (tv Tvar) freeVars(c *Constraints) TvarSet {
	fvs := TvarSet{tv}
	if c != nil {
		ks, ok := c.kindConst[tv]
		if ok {
			for _, k := range ks {
				fvs = fvs.union(k.freeVars(c))
			}
		}
	}
	return fvs
}
func (l Tvar) unifyType(kinds map[Tvar]Kind, s Substitution, r PolyType) (Substitution, error) {
	// Use the substitution to reduce the left and right hand sides.
	// This ensures we never add a transitive cycle in the substitution.
	tl := s.applyToType(l)
	tr := s.applyToType(r)
	switch tl := tl.(type) {
	case Tvar:
		tv, ok := tr.(Tvar)
		if !ok {
			return unifyVarAndType(kinds, s, tl, tr)
		}
		if tl == tv {
			return s, nil
		}
		sub, err := unifyKindsByVar(kinds, s, tl, tv)
		if err != nil {
			return nil, err
		}
		return unifyVarAndType(kinds, sub, tl, tv)
	default:
		if tv, ok := tr.(Tvar); ok {
			return unifyVarAndType(kinds, s, tv, tl)
		}
		return tl.unifyType(kinds, s, tr)
	}
}

func (tv Tvar) resolveType(kinds map[Tvar]Kind) (Type, error) {
	k, ok := kinds[tv]
	if !ok {
		return nil, fmt.Errorf("type variable %q is not monomorphic", tv)
	}
	return k.resolveType(kinds)
}
func (tv Tvar) MonoType() (Type, bool) {
	return nil, false
}
func (tv Tvar) resolvePolyType(kinds map[Tvar]Kind) (PolyType, error) {
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

// PolyType methods for Nature
func (n Nature) occurs(Tvar) bool                                { return false }
func (n Nature) apply(sub Substitution) (PolyType, bool)         { return n, false }
func (n Nature) resolveType(map[Tvar]Kind) (Type, error)         { return n, nil }
func (n Nature) MonoType() (Type, bool)                          { return n, true }
func (n Nature) resolvePolyType(map[Tvar]Kind) (PolyType, error) { return n, nil }
func (n Nature) freeVars(*Constraints) TvarSet                   { return nil }
func (n Nature) unifyType(kinds map[Tvar]Kind, s Substitution, t PolyType) (Substitution, error) {
	switch t := t.(type) {
	case Nature:
		if t != n {
			return nil, fmt.Errorf("%v != %v", n, t)
		}
	case Tvar:
		return t.unifyType(kinds, s, n)
	default:
		return nil, fmt.Errorf("cannot unify %v with %T", n, t)
	}
	return s, nil
}
func (n Nature) Equal(t PolyType) bool {
	switch t := t.(type) {
	case Nature:
		return t == n
	default:
		return false
	}
}

type invalid struct {
	err error
}

func (i invalid) String() string {
	return "INVALID"
}

func (i invalid) Nature() Nature                                  { return Invalid }
func (i invalid) occurs(tv Tvar) bool                             { return false }
func (i invalid) apply(sub Substitution) (PolyType, bool)         { return i, false }
func (i invalid) resolveType(map[Tvar]Kind) (Type, error)         { return Invalid, nil }
func (i invalid) MonoType() (Type, bool)                          { return nil, false }
func (i invalid) resolvePolyType(map[Tvar]Kind) (PolyType, error) { return i, nil }
func (i invalid) freeVars(*Constraints) TvarSet                   { return nil }
func (i invalid) unifyType(map[Tvar]Kind, Substitution, PolyType) (Substitution, error) {
	return nil, nil
}
func (i invalid) Equal(t PolyType) bool {
	switch t.(type) {
	case invalid:
		return true
	default:
		return false
	}
}

type array struct {
	typ PolyType
}

func NewArrayPolyType(elementType PolyType) PolyType {
	return array{typ: elementType}
}

func (a array) Nature() Nature {
	return Array
}
func (a array) String() string {
	return fmt.Sprintf("[%v]", a.typ)
}

func (a array) occurs(tv Tvar) bool {
	return a.typ.occurs(tv)
}
func (a array) apply(sub Substitution) (PolyType, bool) {
	tp, ok := a.typ.apply(sub)
	return array{typ: tp}, ok
}
func (a array) freeVars(c *Constraints) TvarSet {
	return a.typ.freeVars(c)
}
func (a array) unifyType(kinds map[Tvar]Kind, s Substitution, b PolyType) (Substitution, error) {
	switch b := b.(type) {
	case array:
		return unifyTypes(kinds, s, a.typ, b.typ)
	case Tvar:
		return b.unifyType(kinds, s, a)
	default:
		return nil, fmt.Errorf("cannot unify list with %T", b)
	}
}
func (a array) resolveType(kinds map[Tvar]Kind) (Type, error) {
	t, err := a.typ.resolveType(kinds)
	if err != nil {
		return nil, err
	}
	return NewArrayType(t), nil
}
func (a array) MonoType() (Type, bool) {
	t, ok := a.typ.MonoType()
	if !ok {
		return nil, false
	}
	return NewArrayType(t), true
}
func (a array) resolvePolyType(kinds map[Tvar]Kind) (PolyType, error) {
	t, err := a.typ.resolvePolyType(kinds)
	if err != nil {
		return nil, err
	}
	return array{typ: t}, nil
}
func (a array) Equal(t PolyType) bool {
	if arr, ok := t.(array); ok {
		return a.typ.Equal(arr.typ)
	}
	return false
}

type ArrayKind struct {
	elementType PolyType
}

func (k ArrayKind) String() string {
	return fmt.Sprintf("ArrayKind: [%v]", k.elementType)
}
func (k ArrayKind) apply(sub Substitution) (Kind, bool) {
	tp, ok := k.elementType.apply(sub)
	return ArrayKind{elementType: tp}, ok
}
func (k ArrayKind) freeVars(c *Constraints) TvarSet {
	return k.elementType.freeVars(c)
}
func (k ArrayKind) unifyKind(kinds map[Tvar]Kind, s Substitution, r Kind) (Kind, Substitution, error) {
	if r, ok := r.(ArrayKind); ok {
		sub, err := unifyTypes(kinds, s, k.elementType, r.elementType)
		if err != nil {
			return nil, nil, err
		}
		return k, sub, nil
	}
	return nil, nil, fmt.Errorf("cannot unify array with %T", k)
}

func (k ArrayKind) resolveType(kinds map[Tvar]Kind) (Type, error) {
	typ, err := k.elementType.resolveType(kinds)
	if err != nil {
		return nil, err
	}
	return NewArrayType(typ), nil
}
func (k ArrayKind) MonoType() (Type, bool) {
	m, ok := k.elementType.MonoType()
	if !ok {
		return nil, false
	}
	return NewArrayType(m), true
}
func (k ArrayKind) resolvePolyType(kinds map[Tvar]Kind) (PolyType, error) {
	typ, err := k.elementType.resolvePolyType(kinds)
	if err != nil {
		return nil, err
	}
	return NewArrayPolyType(typ), nil
}
func (k ArrayKind) occurs(tv Tvar) bool {
	return k.elementType.occurs(tv)
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

func (f function) Nature() Nature {
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

func (f function) occurs(tv Tvar) bool {
	for _, a := range f.parameters {
		occurs := a.occurs(tv)
		if occurs {
			return true
		}
	}
	return f.ret.occurs(tv)
}

func (f function) apply(sub Substitution) (PolyType, bool) {
	parameters := make(map[string]PolyType, len(f.parameters))
	var mod bool
	for k, t := range f.parameters {
		tp, ok := t.apply(sub)
		parameters[k] = tp
		mod = mod || ok
	}
	ret, ok := f.ret.apply(sub)
	mod = mod || ok
	return function{
		parameters:   parameters,
		required:     f.required.copy(),
		ret:          ret,
		pipeArgument: f.pipeArgument,
	}, mod
}
func (f function) freeVars(c *Constraints) TvarSet {
	fvs := f.ret.freeVars(c)
	for _, t := range f.parameters {
		fvs = fvs.union(t.freeVars(c))
	}
	return fvs
}
func (l function) unifyType(kinds map[Tvar]Kind, s Substitution, r PolyType) (Substitution, error) {
	switch r := r.(type) {
	case function:
		// Validate every required parameter observed in the right function
		// is observed in the left as well, excluding pipe parameters.
		for _, param := range r.required {
			if _, ok := l.parameters[param]; !ok && param != r.pipeArgument {
				// Pipe paramenters are validated below
				return nil, fmt.Errorf("function does not take a parameter %q", param)
			}
		}
		// Validate that every required parameter of the left function
		// is observed in the right function, excluding pipe parameters.
		missing := l.required.diff(r.required)
		for _, lbl := range missing {
			if _, ok := r.parameters[lbl]; !ok && lbl != l.pipeArgument {
				// Pipe parameters are validated below
				return nil, fmt.Errorf("missing required parameter %q", lbl)
			}
		}
		for f, tl := range l.parameters {
			tr, ok := r.parameters[f]
			if !ok {
				// Already validated missing parameters,
				// this must be the pipe parameter.
				continue
			}
			sub, err := unifyTypes(kinds, s, tl, tr)
			if err != nil {
				return nil, err
			}
			s = sub
		}
		if leftPipeType, ok := l.lookupPipe(l.pipeArgument); !ok {
			// If the left function does not take a pipe argument,
			// the right function must not take one either.
			if _, ok := r.lookupPipe(r.pipeArgument); ok {
				return nil, fmt.Errorf("function does not take a pipe argument")
			}
		} else {
			var pipeArgument string
			if l.pipeArgument != "" {
				pipeArgument = l.pipeArgument
			} else {
				pipeArgument = r.pipeArgument
			}
			// If the left function takes a pipe argument, the
			// the right must as well, and the types must unify.
			rightPipeType, ok := r.lookupPipe(pipeArgument)
			if !ok {
				return nil, fmt.Errorf("function requires a pipe argument")
			}
			sub, err := unifyTypes(kinds, s, leftPipeType, rightPipeType)
			if err != nil {
				return nil, err
			}
			s = sub
		}
		sub, err := unifyTypes(kinds, s, l.ret, r.ret)
		if err != nil {
			return nil, err
		}
		return sub, nil
	case Tvar:
		return r.unifyType(kinds, s, l)
	default:
		return nil, fmt.Errorf("cannot unify function with %T", r)
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

func (f function) resolveType(kinds map[Tvar]Kind) (Type, error) {
	ret, err := f.ret.resolveType(kinds)
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
		t, err := a.resolveType(kinds)
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
func (f function) resolvePolyType(kinds map[Tvar]Kind) (PolyType, error) {
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
	krecord ObjectKind
}

func NewEmptyObjectPolyType() PolyType {
	return NewObjectPolyType(nil, LabelSet{}, LabelSet{})
}

// NewObjectPolyType creates a PolyType representing an object.
// A map of properties and their types may be provided.
// Lower is a set of labels that must exist on the object,
// and upper is a set of labels that may exist on the object.
// Upper must be a superset of lower.
// The map must contain an entry for all lables in the lower set.
// Use AllLabels() to represent the infinite set of all possible labels.
func NewObjectPolyType(properties map[string]PolyType, lower, upper LabelSet) PolyType {
	return object{
		krecord: ObjectKind{
			properties: properties,
			lower:      lower,
			upper:      upper,
		},
	}
}

func (o object) Nature() Nature {
	return Object
}
func (o object) String() string {
	return o.krecord.String()
}

func (o object) occurs(tv Tvar) bool {
	return o.krecord.occurs(tv)
}

func (o object) apply(sub Substitution) (PolyType, bool) {
	properties := make(map[string]PolyType, len(o.krecord.properties))
	var mod bool
	for k, t := range o.krecord.properties {
		tp, ok := t.apply(sub)
		properties[k] = tp
		mod = mod || ok
	}
	return object{
		krecord: ObjectKind{
			properties: properties,
			lower:      o.krecord.lower.copy(),
			upper:      o.krecord.upper.copy(),
		},
	}, mod
}
func (o object) freeVars(c *Constraints) TvarSet {
	var fvs TvarSet
	for _, t := range o.krecord.properties {
		fvs = fvs.union(t.freeVars(c))
	}
	return fvs
}

func (l object) unifyType(kinds map[Tvar]Kind, s Substitution, r PolyType) (Substitution, error) {
	switch r := r.(type) {
	case object:
		_, sub, err := l.krecord.unifyKind(kinds, s, r.krecord)
		return sub, err
	case Tvar:
		return r.unifyType(kinds, s, l)
	default:
		return nil, fmt.Errorf("cannot unify object with %T", r)
	}
}
func (o object) resolveType(kinds map[Tvar]Kind) (Type, error) {
	return o.krecord.resolveType(kinds)
}
func (o object) MonoType() (Type, bool) {
	return o.krecord.MonoType()
}
func (o object) resolvePolyType(kinds map[Tvar]Kind) (PolyType, error) {
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

func (o object) KindConstraint() Kind {
	return o.krecord
}

type KindConstrainter interface {
	KindConstraint() Kind
}

type ObjectKind struct {
	with       PolyType
	properties map[string]PolyType
	lower      LabelSet
	upper      LabelSet
}

func (k ObjectKind) String() string {
	if k.with != nil {
		return fmt.Sprintf("{%v with %v %v %v}", k.with, k.properties, k.lower, k.upper)
	}
	return fmt.Sprintf("{%v %v %v}", k.properties, k.lower, k.upper)
}

func (k ObjectKind) apply(sub Substitution) (Kind, bool) {
	properties := make(map[string]PolyType)
	var mod bool
	for k, f := range k.properties {
		tp, ok := f.apply(sub)
		properties[k] = tp
		mod = mod || ok
	}
	var with PolyType
	if k.with != nil {
		tp, ok := k.with.apply(sub)
		with = tp
		mod = mod || ok
	}
	return ObjectKind{
		with:       with,
		properties: properties,
		upper:      k.upper.copy(),
		lower:      k.lower.copy(),
	}, mod
}
func (k ObjectKind) freeVars(c *Constraints) TvarSet {
	var fvs TvarSet
	for _, f := range k.properties {
		fvs = fvs.union(f.freeVars(c))
	}
	return fvs
}

func (l ObjectKind) unifyKind(kinds map[Tvar]Kind, s Substitution, k Kind) (Kind, Substitution, error) {
	r, ok := k.(ObjectKind)
	if !ok {
		return nil, nil, fmt.Errorf("cannot unify record with %T", k)
	}

	properties := make(map[string]PolyType, len(l.properties)+len(r.properties))
	for f, typL := range l.properties {
		properties[f] = typL
		typR, ok := r.properties[f]
		if !ok {
			continue
		}
		sub, err := unifyTypes(kinds, s, typL, typR)
		if err != nil {
			properties[f] = invalid{err: err}
		} else {
			properties[f] = sub.applyToType(typL)
		}
		s = sub
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

	// Ensure that all of the values that are missing are allowed to be missing.
	diff := lower.diff(upper)
	for _, lbl := range diff {
		if ptv, ok := properties[lbl].(Tvar); ok {
			// If this tvar is nullable, then it is allowed
			// to be missing.
			kind := kinds[ptv]
			if _, ok := kind.(NullableKind); ok {
				continue
			}
		}
		return nil, nil, fmt.Errorf("missing object properties %v", diff)
	}

	var with PolyType
	switch {
	case l.with == nil && r.with == nil:
		// nothing to do
	case l.with == nil && r.with != nil:
		with = r.with
	case l.with != nil && r.with == nil:
		with = l.with
	case l.with != nil && r.with != nil:
		return nil, nil, errors.New("cannot unify two object each having a with constraint")
	}

	kr := ObjectKind{
		with:       with,
		properties: properties,
		lower:      lower,
		upper:      upper,
	}
	// Check for invalid records in the properties.
	for lbl, t := range kr.properties {
		i, ok := t.(invalid)
		if ok {
			return nil, nil, errors.Wrapf(i.err, "invalid record access %q", lbl)
		}
	}
	return kr, s, nil
}

func (k ObjectKind) resolveType(kinds map[Tvar]Kind) (Type, error) {
	properties := make(map[string]Type, len(k.properties))
	for l, ft := range k.properties {
		if _, ok := ft.(invalid); !ok {
			t, err := ft.resolveType(kinds)
			if err != nil {
				return nil, err
			}
			properties[l] = t
		}
	}

	return NewObjectType(properties), nil
}
func (k ObjectKind) MonoType() (Type, bool) {
	if k.with != nil {
		return nil, false
	}
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
	return NewObjectType(properties), true
}
func (k ObjectKind) resolvePolyType(kinds map[Tvar]Kind) (PolyType, error) {
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
func (k ObjectKind) occurs(tv Tvar) bool {
	for _, p := range k.properties {
		occurs := p.occurs(tv)
		if occurs {
			return true
		}
	}
	return false
}

// NullableKind indicates that it is possible for this
// variable to be the null type if no other type is
// more appropriate.
type NullableKind struct {
	T PolyType
}

func (n NullableKind) MonoType() (Type, bool) {
	return n.T.MonoType()
}
func (NullableKind) freeVars(*Constraints) TvarSet { return nil }
func (n NullableKind) resolveType(kinds map[Tvar]Kind) (Type, error) {
	return Nil, nil
}
func (n NullableKind) resolvePolyType(kinds map[Tvar]Kind) (PolyType, error) {
	return n.T, nil
}
func (n NullableKind) apply(sub Substitution) (Kind, bool) {
	tv, ok := n.T.(Tvar)
	if !ok {
		return n, false
	}
	tp, ok := sub[tv]
	if !ok {
		return n, false
	}
	return NullableKind{T: tp}, true
}
func (n NullableKind) unifyKind(kinds map[Tvar]Kind, s Substitution, k Kind) (Kind, Substitution, error) {
	// Nullable constraint is overwritten by everything.
	return k, s, nil
}
func (n NullableKind) occurs(tv Tvar) bool {
	return n.T.occurs(tv)
}

type Comparable struct{}
type Addable struct{}
type Number struct{}

type Scheme struct {
	T    PolyType
	Free TvarSet
}

// freeVars returns the free vars unioned with the free vars in T.
func (s Scheme) freeVars(c *Constraints) TvarSet {
	return s.Free.union(s.T.freeVars(c))
}
