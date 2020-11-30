package dict

import (
	"context"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const pkgpath = "dict"

// FromList will convert a list of values into a Dictionary.
func FromList(args interpreter.Arguments) (values.Value, error) {
	pairs, err := args.GetRequiredArray("pairs", semantic.Object)
	if err != nil {
		return nil, err
	}

	// Retrieve the array element type to determine the
	// types for the Dictionary.
	elemType, err := pairs.Type().ElemType()
	if err != nil {
		return nil, err
	}

	// Retrieve the properties in a sorted way so that the
	// properties are in a defined order. Since the symbol
	// key will always be before value in order, then key
	// will be the first element and value the second.
	props, err := elemType.SortedProperties()
	if err != nil {
		return nil, err
	}

	// Exactly two properties.
	if len(props) != 2 {
		return nil, errors.New(codes.Internal, "expected exactly 2 properties for the array element type")
	}

	// First property should be "key".
	if props[0].Name() != "key" {
		return nil, errors.New(codes.Internal, "missing key property from array element")
	}
	keyType, err := props[0].TypeOf()
	if err != nil {
		return nil, err
	}

	// Second property should be "value".
	if props[1].Name() != "value" {
		return nil, errors.New(codes.Internal, "missing value property from array element")
	}
	valueType, err := props[1].TypeOf()
	if err != nil {
		return nil, err
	}

	// Construct the dictionary from each element in the array.
	dictType := semantic.NewDictType(keyType, valueType)
	builder := values.NewDictBuilder(dictType)

	// Track any errors that happen when building the dictionary.
	pairs.Array().Range(func(i int, v values.Value) {
		if err != nil {
			return
		}

		o := v.Object()
		key, _ := o.Get("key")
		value, _ := o.Get("value")
		err = builder.Insert(key, value)
	})
	if err != nil {
		return nil, err
	}
	return builder.Dict(), nil
}

// Get will retrieve a value from a Dictionary.
func Get(args interpreter.Arguments) (values.Value, error) {
	from, err := args.GetRequiredDictionary("dict")
	if err != nil {
		return nil, err
	}

	key, err := args.GetRequired("key")
	if err != nil {
		return nil, err
	}

	def, err := args.GetRequired("default")
	if err != nil {
		return nil, err
	}
	return from.Get(key, def), nil
}

// Insert will insert a value into a Dictionary and
// return the new Dictionary. It will not modify
// the original Dictionary.
func Insert(args interpreter.Arguments) (values.Value, error) {
	dict, err := args.GetRequiredDictionary("dict")
	if err != nil {
		return nil, err
	}

	key, err := args.GetRequired("key")
	if err != nil {
		return nil, err
	}

	value, err := args.GetRequired("value")
	if err != nil {
		return nil, err
	}
	return dict.Insert(key, value)
}

// Remove will remove a value from a Dictionary and
// return the new Dictionary. It will not modify
// the original Dictionary.
func Remove(args interpreter.Arguments) (values.Value, error) {
	dict, err := args.GetRequiredDictionary("dict")
	if err != nil {
		return nil, err
	}

	key, err := args.GetRequired("key")
	if err != nil {
		return nil, err
	}
	return dict.Remove(key), nil
}

// function is a function definition.
type function func(args interpreter.Arguments) (values.Value, error)

// makeFunction will construct a values.Function from a function definition.
func makeFunction(name string, fn function) values.Function {
	mt := runtime.MustLookupBuiltinType(pkgpath, name)
	return values.NewFunction(name, mt, func(ctx context.Context, args values.Object) (values.Value, error) {
		return interpreter.DoFunctionCall(fn, args)
	}, false)
}

// registerFunction will create a function from the function definition
// and register it in this package with the given name.
func registerFunction(name string, fn function) {
	runtime.RegisterPackageValue(pkgpath, name, makeFunction(name, fn))
}

func init() {
	registerFunction("fromList", FromList)
	registerFunction("get", Get)
	registerFunction("insert", Insert)
	registerFunction("remove", Remove)
}
