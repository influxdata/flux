package function

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// Arguments defines the interface for reading an Argument.
type Arguments interface {
	// Get will retrieve the Value associated with
	// the name. If the argument was not provided,
	// false will be returned.
	Get(name string) (values.Value, bool)
}

// ReadArgs will read Arguments into a spec.
// The spec must be a pointer to a struct value.
//
// Each field is set from a value with the corresponding name.
// The default name for a field is to convert the exported
// field name to lowerCamelCase.
//
// The basic types of int, uint, float, and string are converted
// directly from the corresponding value type. All of the different
// sizes for the numeric types are supported.
//
// Array types are converted into slice types. Go array types
// are not supported.
//
// Object types are converted into maps or structs. If a map,
// the key must be a string and the value can be anything normally
// supported. If a struct type, the same process is used as
// reading a spec.
//
// If a type implements the Argument interface, the ReadArg
// function will be used instead of the normal assignments.
//
// For all other types, it will attempt to assign the
// values.Value type. If this isn't possible, an error will
// be returned.
//
// To customize the behavior, the `flux` struct tag can be used.
//
// 		type CustomSpec struct {
//			Tables       *function.TableObject `flux:"tables,required"`
//			Column       string                `flux:",required"`
//			DefaultValue values.Value          `flux:"default"`
//			Ignored      string                `flux:"-"`
//		}
//
// An alternate name can be used with the `flux` struct tag.
// The `-` string can be used to tell the argument reader to
// ignore that value. Non-exported values are always ignored.
//
// A comma can be used to separate options or no name can be given
// and a comma can be used to use the default name with options.
// The `required` option marks a field as required.
func ReadArgs(spec interface{}, args Arguments, a *flux.Administration) error {
	if a == nil {
		a = &flux.Administration{}
	}

	v := reflect.ValueOf(spec)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New(codes.Internal, "spec must be a pointer to a struct")
	}

	elem := v.Elem()
	elemType := elem.Type()
	for i, n := 0, elemType.NumField(); i < n; i++ {
		ftype := elemType.Field(i)
		name, required, ok := getArgInfo(ftype)
		if !ok {
			continue
		}

		fval := elem.Field(i)
		if err := readArg(fval, args, name, required, a); err != nil {
			return err
		}
	}
	return nil
}

func getArgInfo(field reflect.StructField) (name string, required, ok bool) {
	name = field.Tag.Get("flux")
	if name == "-" {
		return
	}

	options := strings.Split(name, ",")
	for _, opt := range options[1:] {
		if opt == "required" {
			required = true
		}
	}
	name = options[0]

	if name == "" {
		argN := []rune(field.Name)
		if len(argN) > 0 {
			argN[0] = unicode.ToLower(argN[0])
		}
		name = string(argN)
	}
	ok = true
	return
}

func readArg(fv reflect.Value, args Arguments, name string, required bool, a *flux.Administration) error {
	arg, ok := args.Get(name)
	if !ok && required {
		return errors.Newf(codes.Invalid, "missing required keyword argument %q", name)
	} else if !ok {
		return nil
	}

	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			v := reflect.New(fv.Type().Elem())
			fv.Set(v)
		}
		fv = fv.Elem()
	}
	return readArgValue(fv, arg, name, a)
}

func readArgValue(fv reflect.Value, arg values.Value, name string, a *flux.Administration) error {
	// If the field is implements the Argument interface,
	// always use that.
	if fv.CanAddr() {
		iv := fv.Addr().Interface()
		if argV, ok := iv.(Argument); ok {
			return argV.ReadArg(name, arg, a)
		}
	}

	switch fv.Kind() {
	case reflect.String:
		return readStrArg(fv, arg, name)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return readIntArg(fv, arg, name)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return readUintArg(fv, arg, name)
	case reflect.Float64, reflect.Float32:
		return readFloatArg(fv, arg, name)
	case reflect.Slice:
		return readArrayArg(fv, arg, name, a)
	case reflect.Map:
		return readObjectArg(fv, arg, name, a)
	case reflect.Struct:
		return readStructArg(fv, arg, name, a)
	default:
		argV := reflect.ValueOf(arg)
		if argV.Type().AssignableTo(fv.Type()) {
			fv.Set(argV)
			return nil
		}
		return errors.Newf(codes.Internal, "invalid argument type: %T", fv.Interface())
	}
}

func readStrArg(fv reflect.Value, arg values.Value, name string) error {
	if arg.Type().Nature() != semantic.String {
		return errors.Newf(codes.Invalid, "keyword argument %q should be of kind %v, but got %v", name, semantic.String, arg.Type().Nature())
	}
	fv.SetString(arg.Str())
	return nil
}

func readIntArg(fv reflect.Value, arg values.Value, name string) error {
	if arg.Type().Nature() != semantic.Int {
		return errors.Newf(codes.Invalid, "keyword argument %q should be of kind %v, but got %v", name, semantic.Int, arg.Type().Nature())
	}
	// TODO(jsternberg): Check for integer size.
	fv.SetInt(arg.Int())
	return nil
}

func readUintArg(fv reflect.Value, arg values.Value, name string) error {
	if arg.Type().Nature() != semantic.UInt {
		return errors.Newf(codes.Invalid, "keyword argument %q should be of kind %v, but got %v", name, semantic.UInt, arg.Type().Nature())
	}
	// TODO(jsternberg): Check for float size.
	fv.SetUint(arg.UInt())
	return nil
}

func readFloatArg(fv reflect.Value, arg values.Value, name string) error {
	if arg.Type().Nature() != semantic.Float {
		return errors.Newf(codes.Invalid, "keyword argument %q should be of kind %v, but got %v", name, semantic.Float, arg.Type().Nature())
	}
	// TODO(jsternberg): Check for float size.
	fv.SetFloat(arg.Float())
	return nil
}

func readArrayArg(fv reflect.Value, arg values.Value, name string, a *flux.Administration) error {
	if arg.Type().Nature() != semantic.Array {
		return errors.Newf(codes.Invalid, "keyword argument %q should be of kind %v, but got %v", name, semantic.Array, arg.Type().Nature())
	}
	arr := arg.Array()

	arrv := reflect.MakeSlice(fv.Type(), arr.Len(), arr.Len())
	for i, sz := 0, arr.Len(); i < sz; i++ {
		name := fmt.Sprintf("%s[%d]", name, i)
		if err := readArgValue(arrv.Index(i), arr.Get(i), name, a); err != nil {
			return err
		}
	}
	fv.Set(arrv)
	return nil
}

func readObjectArg(fv reflect.Value, arg values.Value, name string, a *flux.Administration) (err error) {
	if fv.Type().Key().Kind() != reflect.String {
		return errors.Newf(codes.Internal, "only string keys are supported for map types")
	} else if arg.Type().Nature() != semantic.Object {
		return errors.Newf(codes.Invalid, "keyword argument %q should be of kind %v, but got %v", name, semantic.Object, arg.Type().Nature())
	}
	obj := arg.Object()

	objv := reflect.MakeMapWithSize(fv.Type(), obj.Len())
	obj.Range(func(key string, v values.Value) {
		if err != nil {
			return
		}
		name := fmt.Sprintf("%s[%s]", name, key)
		keyV := reflect.ValueOf(key)

		var elem reflect.Value
		if elemT := fv.Type().Elem(); elemT.Kind() == reflect.Ptr {
			elem = reflect.New(elemT.Elem())
		} else {
			elem = reflect.New(elemT).Elem()
		}

		if err = readArgValue(elem, v, name, a); err != nil {
			return
		}
		objv.SetMapIndex(keyV, elem)
	})
	fv.Set(objv)
	return err
}

func readStructArg(fv reflect.Value, arg values.Value, name string, a *flux.Administration) error {
	if arg.Type().Nature() != semantic.Object {
		return errors.Newf(codes.Invalid, "keyword argument %q should be of kind %v, but got %v", name, semantic.Object, arg.Type().Nature())
	}
	args := interpreter.NewArguments(arg.Object())
	return ReadArgs(fv.Addr().Interface(), args, a)
}

// Argument defines the interface for reading an argument
// into a value.
type Argument interface {
	ReadArg(name string, arg values.Value, a *flux.Administration) error
}

// TableObject is a flux.TableObject that implements the Argument interface.
type TableObject struct {
	*flux.TableObject
}

func (t *TableObject) ReadArg(name string, arg values.Value, a *flux.Administration) error {
	o, ok := arg.(*flux.TableObject)
	if !ok {
		return errors.Newf(codes.Invalid, "argument is not a table object: got %T", arg)
	}
	t.TableObject = o
	a.AddParent(o)
	return nil
}
