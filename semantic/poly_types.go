package semantic

import (
	"fmt"
	"sort"
	"strings"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
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
	// substituteType replaces tv for t producing a new type.
	substituteType(tv Tvar, t PolyType) PolyType
	// unifyType unifies the two types given the kind constraints and produces a substitution.
	unifyType(map[Tvar]Kind, PolyType) (Substitution, error)

	// Equal reports if two types are the same.
	Equal(PolyType) bool

	// Nature reports the primitive description of the type.
	Nature() Nature
}

// Kind is a constraint in the kind domain.
type Kind interface {
	TypeExpression
	// substituteKind replaces occurences of tv with t producing a new kind.
	substituteKind(tv Tvar, t PolyType) Kind
	// unifyKind unifies the two kinds producing a new merged kind and a substitution.
	unifyKind(map[Tvar]Kind, Kind) (Kind, Substitution, error)
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
func (a Tvar) substituteType(b Tvar, t PolyType) PolyType {
	if a == b {
		return t
	}
	return a
}
func (tv Tvar) freeVars(c *Constraints) TvarSet {
	fvs := TvarSet{tv}
	if c != nil {
		for tvar, kinds := range c.kindConst {
			if tvar == tv {
				for _, k := range kinds {
					fvs = fvs.union(k.freeVars(c))
				}
				continue
			}
			for _, k := range kinds {
				if k.freeVars(nil).contains(tv) {
					fvs = fvs.union(TvarSet{tvar})
				}
			}
		}
	}
	return fvs
}
func (l Tvar) unifyType(kinds map[Tvar]Kind, r PolyType) (Substitution, error) {
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

func (tv Tvar) resolveType(kinds map[Tvar]Kind) (Type, error) {
	k, ok := kinds[tv]
	if !ok {
		return nil, errors.Newf(codes.Invalid, "type variable %q is not monomorphic", tv)
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
func (n Nature) substituteType(Tvar, PolyType) PolyType          { return n }
func (n Nature) resolveType(map[Tvar]Kind) (Type, error)         { return n, nil }
func (n Nature) MonoType() (Type, bool)                          { return n, true }
func (n Nature) resolvePolyType(map[Tvar]Kind) (PolyType, error) { return n, nil }
func (n Nature) freeVars(*Constraints) TvarSet                   { return nil }
func (n Nature) unifyType(kinds map[Tvar]Kind, t PolyType) (Substitution, error) {
	switch t := t.(type) {
	case Nature:
		if t != n {
			return nil, errors.Newf(codes.Invalid, "%v != %v", n, t)
		}
	case Tvar:
		return t.unifyType(kinds, n)
	default:
		return nil, errors.Newf(codes.Invalid, "cannot unify %v with %T", n, t)
	}
	return nil, nil
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

func (i invalid) Nature() Nature                                          { return Invalid }
func (i invalid) occurs(tv Tvar) bool                                     { return false }
func (i invalid) substituteType(Tvar, PolyType) PolyType                  { return i }
func (i invalid) resolveType(map[Tvar]Kind) (Type, error)                 { return Invalid, nil }
func (i invalid) MonoType() (Type, bool)                                  { return nil, false }
func (i invalid) resolvePolyType(map[Tvar]Kind) (PolyType, error)         { return i, nil }
func (i invalid) freeVars(*Constraints) TvarSet                           { return nil }
func (i invalid) unifyType(map[Tvar]Kind, PolyType) (Substitution, error) { return nil, nil }
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
func (a array) substituteType(tv Tvar, t PolyType) PolyType {
	return array{typ: a.typ.substituteType(tv, t)}
}
func (a array) freeVars(c *Constraints) TvarSet {
	return a.typ.freeVars(c)
}
func (a array) unifyType(kinds map[Tvar]Kind, b PolyType) (Substitution, error) {
	switch b := b.(type) {
	case array:
		return unifyTypes(kinds, a.typ, b.typ)
	case Tvar:
		return b.unifyType(kinds, a)
	default:
		return nil, errors.Newf(codes.Invalid, "cannot unify list with %T", b)
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

func (f function) substituteType(tv Tvar, typ PolyType) PolyType {
	parameters := make(map[string]PolyType, len(f.parameters))
	for k, t := range f.parameters {
		parameters[k] = t.substituteType(tv, typ)
	}
	return function{
		parameters:   parameters,
		required:     f.required.copy(),
		ret:          f.ret.substituteType(tv, typ),
		pipeArgument: f.pipeArgument,
	}
}
func (f function) freeVars(c *Constraints) TvarSet {
	fvs := f.ret.freeVars(c)
	for _, t := range f.parameters {
		fvs = fvs.union(t.freeVars(c))
	}
	return fvs
}
func (l function) unifyType(kinds map[Tvar]Kind, r PolyType) (Substitution, error) {
	switch r := r.(type) {
	case function:
		// Validate every required parameter observed in the right function
		// is observed in the left as well, excluding pipe parameters.
		for _, param := range r.required {
			if _, ok := l.parameters[param]; !ok && param != r.pipeArgument {
				// Pipe paramenters are validated below
				return nil, errors.Newf(codes.Invalid, "function does not take a parameter %q, required params %v", param, l.required)
			}
		}
		// Validate that every required parameter of the left function
		// is observed in the right function, excluding pipe parameters.
		missing := l.required.diff(r.required)
		lst := []string{}
		for _, lbl := range missing {
			if _, ok := r.parameters[lbl]; !ok && lbl != l.pipeArgument {
				// Pipe parameters are validated below
				lst = append(lst, lbl)
			}
		}
		if len(lst) > 0 {
			return nil, errors.Newf(codes.Invalid, "missing required parameter(s) %q in call to function, which requires %q",
				strings.Join(lst, ", "), l.required)
		}

		subst := make(Substitution)
		for f, tl := range l.parameters {
			tr, ok := r.parameters[f]
			if !ok {
				// Already validated missing parameters,
				// this must be the pipe parameter.
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
		if leftPipeType, ok := l.lookupPipe(l.pipeArgument); !ok {
			// If the left function does not take a pipe argument,
			// the right function must not take one either.
			if _, ok := r.lookupPipe(r.pipeArgument); ok {
				return nil, errors.New(codes.Invalid, "function does not take a pipe argument")
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
				return nil, errors.New(codes.Invalid, "function requires a pipe argument")
			}
			s, err := unifyTypes(kinds, leftPipeType, rightPipeType)
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
		return r.unifyType(kinds, l)
	default:
		return nil, errors.Newf(codes.Invalid, "cannot unify function with %T", r)
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

func (o object) substituteType(tv Tvar, typ PolyType) PolyType {
	properties := make(map[string]PolyType, len(o.krecord.properties))
	for k, t := range o.krecord.properties {
		properties[k] = t.substituteType(tv, typ)
	}
	return object{
		krecord: ObjectKind{
			properties: properties,
			lower:      o.krecord.lower.copy(),
			upper:      o.krecord.upper.copy(),
		},
	}
}
func (o object) freeVars(c *Constraints) TvarSet {
	var fvs TvarSet
	for _, t := range o.krecord.properties {
		fvs = fvs.union(t.freeVars(c))
	}
	return fvs
}

func (l object) unifyType(kinds map[Tvar]Kind, r PolyType) (Substitution, error) {
	switch r := r.(type) {
	case object:
		_, subst, err := l.krecord.unifyKind(kinds, r.krecord)
		return subst, err
	case Tvar:
		return r.unifyType(kinds, l)
	default:
		return nil, errors.Newf(codes.Invalid, "cannot unify object with %T", r)
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

type KClass struct{}

func (k KClass) freeVars(c *Constraints) TvarSet { return nil }
func (k KClass) substituteKind(tv Tvar, t PolyType) Kind {
	return k
}
func (l KClass) unifyKind(kinds map[Tvar]Kind, r Kind) (Kind, Substitution, error) {
	//TODO
	return nil, nil, nil
}
func (k KClass) resolveType(map[Tvar]Kind) (Type, error) {
	return nil, errors.New(codes.Internal, "KClass has no type")
}
func (k KClass) MonoType() (Type, bool) {
	return nil, false
}
func (k KClass) resolvePolyType(map[Tvar]Kind) (PolyType, error) {
	return nil, errors.New(codes.Internal, "KClass has no poly type")
}
func (k KClass) occurs(Tvar) bool { return false }

type ObjectKind struct {
	with       *Tvar
	properties map[string]PolyType
	lower      LabelSet
	upper      LabelSet
}

func (k ObjectKind) String() string {
	if k.with != nil {
		return fmt.Sprintf("{%v with %v %v %v}", *k.with, k.properties, k.lower, k.upper)
	}
	return fmt.Sprintf("{%v %v %v}", k.properties, k.lower, k.upper)
}

func (k ObjectKind) substituteKind(tv Tvar, t PolyType) Kind {
	properties := make(map[string]PolyType)
	for k, f := range k.properties {
		properties[k] = f.substituteType(tv, t)
	}
	var with *Tvar
	if k.with != nil {
		with = new(Tvar)
		if *k.with == tv {
			*with = tv
			v, ok := t.(Tvar)
			if ok {
				*with = v
			}
		} else {
			*with = *k.with
		}
	}
	return ObjectKind{
		with:       with,
		properties: properties,
		upper:      k.upper.copy(),
		lower:      k.lower.copy(),
	}
}
func (k ObjectKind) freeVars(c *Constraints) TvarSet {
	var fvs TvarSet
	for _, f := range k.properties {
		fvs = fvs.union(f.freeVars(c))
	}
	return fvs
}

func (l ObjectKind) unifyKind(kinds map[Tvar]Kind, k Kind) (Kind, Substitution, error) {
	r, ok := k.(ObjectKind)
	if !ok {
		return nil, nil, errors.Newf(codes.Invalid, "cannot unify record with %T", k)
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
			properties[f] = invalid{err: err}
		} else {
			subst.Merge(s)
			properties[f] = subst.ApplyType(typL)
		}
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
		return nil, nil, errors.Newf(codes.Invalid, "missing object properties %v", diff)
	}

	var with *Tvar
	switch {
	case l.with == nil && r.with == nil:
		// nothing to do
	case l.with == nil && r.with != nil:
		with = new(Tvar)
		*with = *r.with
	case l.with != nil && r.with == nil:
		with = new(Tvar)
		*with = *l.with
	case l.with != nil && r.with != nil:
		return nil, nil, errors.New(codes.Invalid, "cannot unify two object each having a with constraint")
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
			return nil, nil, errors.Wrapf(i.err, codes.Inherit, "invalid record access %q", lbl)
		}
	}
	return kr, subst, nil
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
	// A nullable type variable resolves to a NullableTvar
	if tv, ok := n.T.(Tvar); ok {
		return NullableTvar{tv}, nil
	}
	return n.T, nil
}
func (n NullableKind) substituteKind(tv Tvar, t PolyType) Kind {
	if ptv, ok := n.T.(Tvar); ok && ptv == tv {
		return NullableKind{T: t}
	}
	return n
}
func (n NullableKind) unifyKind(kinds map[Tvar]Kind, k Kind) (Kind, Substitution, error) {
	// Nullable constraint is overwritten by everything.
	return k, nil, nil
}
func (n NullableKind) occurs(tv Tvar) bool {
	return n.T.occurs(tv)
}

// NullableTvar is a type variable that might be null.
// If a type variable is constrained to be nullable (via the NullableKind),
// in order to preserve that constraint when resolving the type, we return
// a NullableTvar.
//
// TODO: This is a temporary type that will be removed once kind constraints
// can be expressed in the language of types. Unfortunately right now, external
// packages like the compiler don't have any notion of kind constraints.
type NullableTvar struct {
	Tvar
}

func (NullableTvar) Nature() Nature {
	return Nil
}

func (NullableTvar) MonoType() (Type, bool) {
	return Nil, true
}

func (t NullableTvar) Equal(p PolyType) bool {
	tv, ok := p.(NullableTvar)
	return ok && t.Tvar.Equal(tv.Tvar)
}

func (t NullableTvar) String() string {
	return fmt.Sprintf("nullable{%v}", t.Tvar)
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

func (s Scheme) Substitute(tv Tvar, t PolyType) Scheme {
	fvs := make(TvarSet, 0, len(s.Free))
	for _, ftv := range s.Free {
		if ftv != tv {
			fvs = append(fvs, ftv)
		}
	}
	return Scheme{
		T:    s.T.substituteType(tv, t),
		Free: fvs,
	}
}
