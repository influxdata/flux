package semantic

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

// Type is the representation of a Flux type.
// Type is a monomorphic, meaning that it represents a single type and is not polymorphic.
// See TypeScheme for polymorphic types.
//
// Type values are comparable and as such can be used as map keys and directly comparison using the == operator.
// Two types are equal if they represent identical types.
//
// Do NOT embed this type into other interfaces or structs as that will invalidate the comparison properties of the interface.
type Type interface {
	// Kind returns the specific kind of this type.
	Kind() Kind

	// PropertyType returns the type of a given property.
	// It panics if the type's Kind is not Object
	PropertyType(name string) Type

	// Properties returns a map of all property types.
	// It panics if the type's Kind is not Object
	Properties() map[string]Type

	// ElementType return the type of elements in the array.
	// It panics if the type's Kind is not Array.
	ElementType() Type

	// InType reports the input type of the function
	// It panics if the type's Kind is not Function.
	InType() Type

	// OutType reports the output type of the function
	// It panics if the type's Kind is not Function.
	OutType() Type

	// PipeArgument reports the name of the argument that can be pipe into.
	// It panics if the type's Kind is not Function.
	PipeArgument() string

	PolyType() PolyType

	// Types cannot be created outside of the semantic package
	// This is needed so that we can cache type definitions.
	typ()
}

type Kind int

const (
	Invalid Kind = iota
	Nil
	String
	Int
	UInt
	Float
	Bool
	Time
	Duration
	Regexp
	Array
	Object
	Function
)

var kindNames = []string{
	Invalid:  "invalid",
	Nil:      "nil",
	String:   "string",
	Int:      "int",
	UInt:     "uint",
	Float:    "float",
	Bool:     "bool",
	Time:     "time",
	Duration: "duration",
	Regexp:   "regexp",
	Array:    "array",
	Object:   "object",
	Function: "function",
}

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "kind" + strconv.Itoa(int(k))
}
func (k Kind) MonoType() (Type, bool) {
	switch k {
	case
		String,
		Int,
		UInt,
		Float,
		Bool,
		Time,
		Duration,
		Regexp:
		return k, true
	default:
		return nil, false
	}
}
func (k Kind) typeScheme() {}

func (k Kind) Kind() Kind {
	return k
}
func (k Kind) PropertyType(name string) Type {
	panic(fmt.Errorf("cannot get type of property %q, from kind %q", name, k))
}
func (k Kind) Properties() map[string]Type {
	panic(fmt.Errorf("cannot get properties from kind %s", k))
}
func (k Kind) ElementType() Type {
	panic(fmt.Errorf("cannot get element type from kind %s", k))
}
func (k Kind) Params() map[string]Type {
	panic(fmt.Errorf("cannot get parameters from kind %s", k))
}
func (k Kind) PipeArgument() string {
	panic(fmt.Errorf("cannot get pipe argument name from kind %s", k))
}
func (k Kind) InType() Type {
	panic(fmt.Errorf("cannot get in type from kind %s", k))
}
func (k Kind) OutType() Type {
	panic(fmt.Errorf("cannot get out type from kind %s", k))
}
func (k Kind) typ() {}

type arrayType struct {
	elementType Type
}

func (t *arrayType) String() string {
	return fmt.Sprintf("[%v]", t.elementType)
}

func (t *arrayType) Kind() Kind {
	return Array
}
func (t *arrayType) PropertyType(name string) Type {
	panic(fmt.Errorf("cannot get property type of kind %s", t.Kind()))
}
func (t *arrayType) Properties() map[string]Type {
	panic(fmt.Errorf("cannot get properties type of kind %s", t.Kind()))
}
func (t *arrayType) ElementType() Type {
	return t.elementType
}
func (t *arrayType) PipeArgument() string {
	panic(fmt.Errorf("cannot get pipe argument name from kind %s", t.Kind()))
}
func (t *arrayType) InType() Type {
	panic(fmt.Errorf("cannot get in type of kind %s", t.Kind()))
}
func (t *arrayType) OutType() Type {
	panic(fmt.Errorf("cannot get out type of kind %s", t.Kind()))
}

func (t *arrayType) PolyType() PolyType {
	panic("not implemented")
}

func (t *arrayType) typ() {}

// arrayTypeCache caches *arrayType values.
//
// Since arrayTypes only have a single field elementType we can key
// all arrayTypes by their elementType.
var arrayTypeCache struct {
	sync.Mutex // Guards stores (but not loads) on m.

	// m is a map[Type]*arrayType keyed by the elementType of the array.
	// Elements in m are append-only and thus safe for concurrent reading.
	m sync.Map
}

// TODO(nathanielc): Make empty array types polymorphic over element type?
var EmptyArrayType = NewArrayType(Nil)

func NewArrayType(elementType Type) Type {
	// Lookup arrayType in cache by elementType
	if t, ok := arrayTypeCache.m.Load(elementType); ok {
		return t.(*arrayType)
	}

	// Type not found in cache, lock and retry.
	arrayTypeCache.Lock()
	defer arrayTypeCache.Unlock()

	// First read again while holding the lock.
	if t, ok := arrayTypeCache.m.Load(elementType); ok {
		return t.(*arrayType)
	}

	// Still no cache entry, add it.
	at := &arrayType{
		elementType: elementType,
	}
	arrayTypeCache.m.Store(elementType, at)

	return at
}

type objectType struct {
	properties map[string]Type
}

func (t *objectType) String() string {
	var buf bytes.Buffer
	buf.Write([]byte("{"))
	for k, prop := range t.properties {
		fmt.Fprintf(&buf, "%s: %v,", k, prop)
	}
	buf.WriteRune('}')

	return buf.String()
}

func (t *objectType) Kind() Kind {
	return Object
}
func (t *objectType) PropertyType(name string) Type {
	typ, ok := t.properties[name]
	if ok {
		return typ
	}
	return Invalid
}
func (t *objectType) Properties() map[string]Type {
	return t.properties
}
func (t *objectType) ElementType() Type {
	panic(fmt.Errorf("cannot get element type of kind %s", t.Kind()))
}
func (t *objectType) PipeArgument() string {
	panic(fmt.Errorf("cannot get pipe argument name from kind %s", t.Kind()))
}
func (t *objectType) InType() Type {
	panic(fmt.Errorf("cannot get in type of kind %s", t.Kind()))
}
func (t *objectType) OutType() Type {
	panic(fmt.Errorf("cannot get out type of kind %s", t.Kind()))
}
func (t *objectType) PolyType() PolyType {
	properties := make(map[string]PolyType)
	for k, p := range t.properties {
		properties[k] = p.PolyType()
	}
	return NewObjectPolyType(properties)
}

func (t *objectType) typ() {}

func (t *objectType) equal(o *objectType) bool {
	if t == o {
		return true
	}

	if len(t.properties) != len(o.properties) {
		return false
	}

	for k, vtyp := range t.properties {
		ovtyp, ok := o.properties[k]
		if !ok {
			return false
		}
		if ovtyp != vtyp {
			return false
		}
	}
	return true
}

// objectTypeCache caches all *objectTypes.
//
// Since objectTypes are identified by their properties,
// a hash is computed of the property names and kinds to reduce the search space.
var objectTypeCache struct {
	sync.Mutex // Guards stores (but not loads) on m.

	// m is a map[uint32][]*objectType keyed by the hash calculated of the object's properties' name and kind.
	// Elements in m are append-only and thus safe for concurrent reading.
	m sync.Map
}

var EmptyObject = NewObjectType(nil)

func NewObjectType(propertyTypes map[string]Type) Type {
	propertyNames := make([]string, 0, len(propertyTypes))
	for name := range propertyTypes {
		propertyNames = append(propertyNames, name)
	}
	sort.Strings(propertyNames)

	sum := fnv.New32a()
	for _, p := range propertyNames {
		t := propertyTypes[p]

		// track hash of property names and kinds
		sum.Write([]byte(p))
		binary.Write(sum, binary.LittleEndian, t.Kind())
	}

	// Create new object type
	ot := &objectType{
		properties: propertyTypes,
	}

	// Simple linear search after hash lookup
	h := sum.Sum32()
	if ts, ok := objectTypeCache.m.Load(h); ok {
		for _, t := range ts.([]*objectType) {
			if t.equal(ot) {
				return t
			}
		}
	}

	// Type not found in cache, lock and retry.
	objectTypeCache.Lock()
	defer objectTypeCache.Unlock()

	// First read again while holding the lock.
	var types []*objectType
	if ts, ok := objectTypeCache.m.Load(h); ok {
		types = ts.([]*objectType)
		for _, t := range types {
			if t.equal(ot) {
				return t
			}
		}
	}

	// Make copy of properties since we can't trust that the source will not be modified
	properties := make(map[string]Type)
	for k, v := range ot.properties {
		properties[k] = v
	}
	ot.properties = properties

	// Still no cache entry, add it.
	objectTypeCache.m.Store(h, append(types, ot))

	return ot
}

type functionType struct {
	in           Type
	out          Type
	pipeArgument string
}

func (t *functionType) String() string {
	return fmt.Sprintf("(%v) -> %v", t.in, t.out)
}

func (t *functionType) Kind() Kind {
	return Function
}
func (t *functionType) PropertyType(name string) Type {
	panic(fmt.Errorf("cannot get property type of kind %s", t.Kind()))
}
func (t *functionType) Properties() map[string]Type {
	panic(fmt.Errorf("cannot get properties type of kind %s", t.Kind()))
}
func (t *functionType) ElementType() Type {
	panic(fmt.Errorf("cannot get element type of kind %s", t.Kind()))
}
func (t *functionType) InType() Type {
	return t.in
}
func (t *functionType) OutType() Type {
	return t.out
}
func (t *functionType) PipeArgument() string {
	return t.pipeArgument
}

func (t *functionType) PolyType() PolyType {
	return NewFunctionPolyType(t.in.PolyType(), nil, t.out.PolyType())
}

func (t *functionType) typ() {}

func (t *functionType) equal(o *functionType) bool {
	return t == o || *t == *o
}

// functionTypeCache caches all *functionTypes.
var functionTypeCache = struct {
	sync.Mutex // Guards access to cache map
	cache      map[functionType]*functionType
}{
	cache: make(map[functionType]*functionType),
}

type FunctionSignature struct {
	In           Type // Must always be an object type
	Defaults     Type
	Out          Type
	PipeArgument string
}

func NewFunctionType(sig FunctionSignature) Type {
	// Create new object type
	ft := &functionType{
		in:           sig.In,
		out:          sig.Out,
		pipeArgument: sig.PipeArgument,
	}

	functionTypeCache.Lock()
	defer functionTypeCache.Unlock()

	t, ok := functionTypeCache.cache[*ft]
	if ok {
		return t
	}

	// Still no cache entry, add it.
	functionTypeCache.cache[*ft] = ft
	return ft
}
