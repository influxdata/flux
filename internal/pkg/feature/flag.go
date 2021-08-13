package feature

import (
	"context"
	"fmt"
)

// Flag represents a generic feature flag with a key and a default.
type Flag interface {
	// Key returns the programmatic backend identifier for the flag.
	Key() string
	// Default returns the type-agnostic zero value for the flag.
	// Type-specific flag implementations may expose a typed default
	// (e.g. BoolFlag includes a boolean Default field).
	Default() interface{}
}

// MakeFlag constructs a Flag. The concrete implementation is inferred from the provided default.
func MakeFlag(name, key, owner string, defaultValue interface{}) Flag {
	b := MakeBase(name, key, owner, defaultValue)
	switch v := defaultValue.(type) {
	case bool:
		return BoolFlag{b, v}
	case float64:
		return FloatFlag{b, v}
	case int32:
		return IntFlag{b, v}
	case int:
		return IntFlag{b, int32(v)}
	case string:
		return StringFlag{b, v}
	default:
		return StringFlag{b, fmt.Sprintf("%v", v)}
	}
}

// Base implements the base of the Flag type.
type Base struct {
	// name of the flag.
	name string
	// key is the programmatic backend identifier for the flag.
	key string
	// defaultValue for the flag.
	defaultValue interface{}
	// owner is an individual or team responsible for the flag.
	owner string
}

var _ Flag = Base{}

// MakeBase constructs a flag.
func MakeBase(name, key, owner string, defaultValue interface{}) Base {
	return Base{
		name:         name,
		key:          key,
		owner:        owner,
		defaultValue: defaultValue,
	}
}

// Key returns the programmatic backend identifier for the flag.
func (f Base) Key() string {
	return f.key
}

// Default returns the type-agnostic zero value for the flag.
func (f Base) Default() interface{} {
	return f.defaultValue
}

func (f Base) value(ctx context.Context) interface{} {
	flagger := GetFlagger(ctx)
	return flagger.FlagValue(ctx, f)
}

// StringFlag implements Flag for string values.
type StringFlag struct {
	Base
	defaultString string
}

var _ Flag = StringFlag{}

// MakeStringFlag returns a string flag with the given Base and default.
func MakeStringFlag(name, key, owner string, defaultValue string) StringFlag {
	b := MakeBase(name, key, owner, defaultValue)
	return StringFlag{b, defaultValue}
}

// String value of the flag on the request context.
func (f StringFlag) String(ctx context.Context) string {
	s, ok := f.value(ctx).(string)
	if !ok {
		return f.defaultString
	}
	return s
}

// FloatFlag implements Flag for float values.
type FloatFlag struct {
	Base
	defaultFloat float64
}

var _ Flag = FloatFlag{}

// MakeFloatFlag returns a string flag with the given Base and default.
func MakeFloatFlag(name, key, owner string, defaultValue float64) FloatFlag {
	b := MakeBase(name, key, owner, defaultValue)
	return FloatFlag{b, defaultValue}
}

// Float value of the flag on the request context.
func (f FloatFlag) Float(ctx context.Context) float64 {
	v, ok := f.value(ctx).(float64)
	if !ok {
		return f.defaultFloat
	}
	return v
}

// IntFlag implements Flag for integer values.
type IntFlag struct {
	Base
	defaultInt int32
}

var _ Flag = IntFlag{}

// MakeIntFlag returns a string flag with the given Base and default.
func MakeIntFlag(name, key, owner string, defaultValue int32) IntFlag {
	b := MakeBase(name, key, owner, defaultValue)
	return IntFlag{b, defaultValue}
}

// Int value of the flag on the request context.
func (f IntFlag) Int(ctx context.Context) int32 {
	i, ok := f.value(ctx).(int32)
	if !ok {
		return f.defaultInt
	}
	return i
}

// BoolFlag implements Flag for boolean values.
type BoolFlag struct {
	Base
	defaultBool bool
}

var _ Flag = BoolFlag{}

// MakeBoolFlag returns a string flag with the given Base and default.
func MakeBoolFlag(name, key, owner string, defaultValue bool) BoolFlag {
	b := MakeBase(name, key, owner, defaultValue)
	return BoolFlag{b, defaultValue}
}

// Enabled indicates whether flag is true or false on the request context.
func (f BoolFlag) Enabled(ctx context.Context) bool {
	i, ok := f.value(ctx).(bool)
	if !ok {
		return f.defaultBool
	}
	return i
}
