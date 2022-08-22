package function

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// Arguments provides access to the arguments of a function call.
// This struct can only be created by using one of the Register functions
// to register a function or by directly calling Invoke.
type Arguments struct {
	obj  values.Object
	used map[string]bool
}

func newArguments(obj values.Object) *Arguments {
	if obj == nil {
		return new(Arguments)
	}
	return &Arguments{
		obj:  obj,
		used: make(map[string]bool, obj.Len()),
	}
}

func (a *Arguments) Get(name string) (values.Value, bool) {
	a.used[name] = true
	v, ok := a.obj.Get(name)
	return v, ok
}

func (a *Arguments) GetRequired(name string) (values.Value, error) {
	a.used[name] = true
	v, ok := a.obj.Get(name)
	if !ok {
		return nil, errors.Newf(codes.Invalid, "missing required keyword argument %q", name)
	}
	return v, nil
}

func (a *Arguments) GetString(name string) (string, bool, error) {
	v, ok, err := a.get(name, semantic.String, false)
	if err != nil || !ok {
		return "", ok, err
	}
	return v.Str(), ok, nil
}

func (a *Arguments) GetRequiredString(name string) (string, error) {
	v, _, err := a.get(name, semantic.String, true)
	if err != nil {
		return "", err
	}
	return v.Str(), nil
}

func (a *Arguments) GetInt(name string) (int64, bool, error) {
	v, ok, err := a.get(name, semantic.Int, false)
	if err != nil || !ok {
		return 0, ok, err
	}
	return v.Int(), ok, nil
}

func (a *Arguments) GetRequiredInt(name string) (int64, error) {
	v, _, err := a.get(name, semantic.Int, true)
	if err != nil {
		return 0, err
	}
	return v.Int(), nil
}

func (a *Arguments) GetUInt(name string) (uint64, bool, error) {
	v, ok, err := a.get(name, semantic.UInt, false)
	if err != nil || !ok {
		return 0, ok, err
	}
	return v.UInt(), ok, nil
}

func (a *Arguments) GetRequiredUInt(name string) (uint64, error) {
	v, _, err := a.get(name, semantic.UInt, true)
	if err != nil {
		return 0, err
	}
	return v.UInt(), nil
}

func (a *Arguments) GetFloat(name string) (float64, bool, error) {
	v, ok, err := a.get(name, semantic.Float, false)
	if err != nil || !ok {
		return 0, ok, err
	}
	return v.Float(), ok, nil
}

func (a *Arguments) GetRequiredFloat(name string) (float64, error) {
	v, _, err := a.get(name, semantic.Float, true)
	if err != nil {
		return 0, err
	}
	return v.Float(), nil
}

func (a *Arguments) GetBool(name string) (bool, bool, error) {
	v, ok, err := a.get(name, semantic.Bool, false)
	if err != nil || !ok {
		return false, ok, err
	}
	return v.Bool(), ok, nil
}

func (a *Arguments) GetRequiredBool(name string) (bool, error) {
	v, _, err := a.get(name, semantic.Bool, true)
	if err != nil {
		return false, err
	}
	return v.Bool(), nil
}

func (a *Arguments) GetArray(name string, t semantic.Nature) (values.Array, bool, error) {
	v, ok, err := a.get(name, semantic.Array, false)
	if err != nil || !ok {
		return nil, ok, err
	}
	arr := v.Array()
	et, err := arr.Type().ElemType()
	if err != nil {
		return nil, false, err
	}
	if et.Nature() != t {
		return nil, true, errors.Newf(codes.Invalid, "keyword argument %q should be of an array of type %v, but got an array of type %v", name, t, arr.Type())
	}
	return v.Array(), ok, nil
}

func (a *Arguments) GetArrayAllowEmpty(name string, t semantic.Nature) (values.Array, bool, error) {
	v, ok, err := a.get(name, semantic.Array, false)
	if err != nil || !ok {
		return nil, ok, err
	}
	arr := v.Array()
	if arr.Len() > 0 {
		et, err := arr.Type().ElemType()
		if err != nil {
			return nil, false, err
		}
		if et.Nature() != t {
			return nil, true, errors.Newf(codes.Invalid, "keyword argument %q should be of an array of type %v, but got an array of type %v", name, t, arr.Type())
		}
	}
	return arr, ok, nil
}

func (a *Arguments) GetRequiredArray(name string, t semantic.Nature) (values.Array, error) {
	v, _, err := a.get(name, semantic.Array, true)
	if err != nil {
		return nil, err
	}
	arr := v.Array()
	et, err := arr.Type().ElemType()
	if err != nil {
		return nil, err
	}
	if et.Nature() != t {
		return nil, errors.Newf(codes.Invalid, "keyword argument %q should be of an array of type %v, but got an array of type %v", name, t, arr.Type())
	}
	return arr, nil
}

// GetRequiredArrayAllowEmpty ensures a required array (with element type) is present,
// but unlike GetRequiredArray, does not fail if the array is empty.
func (a *Arguments) GetRequiredArrayAllowEmpty(name string, t semantic.Nature) (values.Array, error) {
	v, _, err := a.get(name, semantic.Array, true)
	if err != nil {
		return nil, err
	}
	arr := v.Array()
	if arr.Array().Len() > 0 {
		et, err := arr.Type().ElemType()
		if err != nil {
			return nil, err
		}
		if et.Nature() != t {
			return nil, errors.Newf(codes.Invalid, "keyword argument %q should be of an array of type %v, but got an array of type %v", name, t, arr.Type())
		}
	}
	return arr, nil
}

func (a *Arguments) GetFunction(name string) (values.Function, bool, error) {
	v, ok, err := a.get(name, semantic.Function, false)
	if err != nil || !ok {
		return nil, ok, err
	}
	return v.Function(), ok, nil
}

func (a *Arguments) GetRequiredFunction(name string) (values.Function, error) {
	v, _, err := a.get(name, semantic.Function, true)
	if err != nil {
		return nil, err
	}
	return v.Function(), nil
}

func (a *Arguments) GetObject(name string) (values.Object, bool, error) {
	v, ok, err := a.get(name, semantic.Object, false)
	if err != nil || !ok {
		return nil, ok, err
	}
	return v.Object(), ok, nil
}

func (a *Arguments) GetRequiredObject(name string) (values.Object, error) {
	v, _, err := a.get(name, semantic.Object, true)
	if err != nil {
		return nil, err
	}
	return v.Object(), nil
}

func (a *Arguments) GetDictionary(name string) (values.Dictionary, bool, error) {
	v, ok, err := a.get(name, semantic.Dictionary, false)
	if err != nil || !ok {
		return nil, ok, err
	}
	return v.Dict(), ok, nil
}

func (a *Arguments) GetRequiredDictionary(name string) (values.Dictionary, error) {
	v, _, err := a.get(name, semantic.Dictionary, true)
	if err != nil {
		return nil, err
	}
	return v.Dict(), nil
}

func (a *Arguments) GetTime(name string) (flux.Time, bool, error) {
	v, ok := a.Get(name)
	if !ok {
		return flux.Time{}, false, nil
	}
	qt, err := flux.ToQueryTime(v)
	if err != nil {
		return flux.Time{}, ok, err
	}
	return qt, ok, nil
}

func (a *Arguments) GetRequiredTime(name string) (flux.Time, error) {
	qt, ok, err := a.GetTime(name)
	if err != nil {
		return flux.Time{}, err
	}
	if !ok {
		return flux.Time{}, errors.Newf(codes.Invalid, "missing required keyword argument %q", name)
	}
	return qt, nil
}

func (a *Arguments) GetDuration(name string) (flux.Duration, bool, error) {
	v, ok := a.Get(name)
	if !ok {
		return flux.ConvertDuration(0), false, nil
	}
	return v.Duration(), true, nil
}

func (a *Arguments) GetRequiredDuration(name string) (flux.Duration, error) {
	d, ok, err := a.GetDuration(name)
	if err != nil {
		return flux.ConvertDuration(0), err
	}
	if !ok {
		return flux.ConvertDuration(0), errors.Newf(codes.Invalid, "missing required keyword argument %q", name)
	}
	return d, nil
}

func (a *Arguments) get(name string, kind semantic.Nature, required bool) (values.Value, bool, error) {
	a.used[name] = true
	v, ok := a.obj.Get(name)
	if !ok {
		if required {
			return nil, false, errors.Newf(codes.Invalid, "missing required keyword argument %q", name)
		}
		return nil, false, nil
	}
	if v.Type().Nature() != kind {
		return nil, true, errors.Newf(codes.Invalid, "keyword argument %q should be of kind %v, but got %v", name, kind, v.Type().Nature())
	}
	return v, true, nil
}

func (a *Arguments) listUnused() []string {
	var unused []string
	if a.obj != nil {
		a.obj.Range(func(k string, v values.Value) {
			if !a.used[k] {
				unused = append(unused, k)
			}
		})
	}
	return unused
}
