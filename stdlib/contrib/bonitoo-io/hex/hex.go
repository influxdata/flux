package hex

import (
	"context"
	goHex "encoding/hex"
	"strconv"
	"strings"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	conversionArg = "v"
	pkgName       = "contrib/bonitoo-io/hex"
)

// function is a function definition
type function func(args interpreter.Arguments) (values.Value, error)

// makeFunction constructs a values.Function from a function definition.
func makeFunction(name string, fn function) values.Function {
	mt := runtime.MustLookupBuiltinType(pkgName, name)
	return values.NewFunction(name, mt, func(ctx context.Context, args values.Object) (values.Value, error) {
		return interpreter.DoFunctionCall(fn, args)
	}, false)
}

func init() {
	runtime.RegisterPackageValue(pkgName, "string", makeFunction("string", String))
	runtime.RegisterPackageValue(pkgName, "int", makeFunction("int", Int))
	runtime.RegisterPackageValue(pkgName, "uint", makeFunction("uint", UInt))
	runtime.RegisterPackageValue(pkgName, "bytes", makeFunction("bytes", Bytes))
}

var errMissingArg = errors.Newf(codes.Invalid, "missing argument %q", conversionArg)

func String(args interpreter.Arguments) (values.Value, error) {
	var str string
	v, ok := args.Get(conversionArg)
	if !ok {
		return nil, errMissingArg
	} else if v.IsNull() {
		return values.Null, nil
	}
	switch v.Type().Nature() {
	case semantic.String:
		str = v.Str()
	case semantic.Int:
		str = strconv.FormatInt(v.Int(), 16)
	case semantic.UInt:
		str = strconv.FormatUint(v.UInt(), 16)
	case semantic.Float:
		str = strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case semantic.Bool:
		str = strconv.FormatBool(v.Bool())
	case semantic.Time:
		str = v.Time().String()
	case semantic.Duration:
		str = v.Duration().String()
	case semantic.Bytes:
		str = goHex.EncodeToString(v.Bytes())
	default:
		return nil, errors.Newf(codes.Invalid, "cannot convert %v to string", v.Type())
	}
	return values.NewString(str), nil
}

func Int(args interpreter.Arguments) (values.Value, error) {
	var i int64
	v, ok := args.Get(conversionArg)
	if !ok {
		return nil, errMissingArg
	} else if v.IsNull() {
		return values.Null, nil
	}
	switch v.Type().Nature() {
	case semantic.String:
		base := 16
		s := v.Str()
		if strings.HasPrefix(s, "0x") {
			s = s[2:]
		} else if strings.HasPrefix(s, "+0x") || strings.HasPrefix(s, "-0x") {
			base = 0
		}
		n, err := strconv.ParseInt(s, base, 64)
		if err != nil {
			return nil, errors.Newf(codes.Invalid, "cannot convert string %q to int due to invalid syntax", v.Str())
		}
		i = n
	default:
		return nil, errors.Newf(codes.Invalid, "hex cannot convert %v to int", v.Type())
	}
	return values.NewInt(i), nil
}

func UInt(args interpreter.Arguments) (values.Value, error) {
	var i uint64
	v, ok := args.Get(conversionArg)
	if !ok {
		return nil, errMissingArg
	} else if v.IsNull() {
		return values.Null, nil
	}
	switch v.Type().Nature() {
	case semantic.String:
		s := strings.TrimPrefix(v.Str(), "0x")
		n, err := strconv.ParseUint(s, 16, 64)
		if err != nil {
			return nil, errors.Newf(codes.Invalid, "cannot convert string %q to uint due to invalid syntax", v.Str())
		}
		i = n
	default:
		return nil, errors.Newf(codes.Invalid, "hex cannot convert %v to uint", v.Type())
	}
	return values.NewUInt(i), nil
}

func Bytes(args interpreter.Arguments) (values.Value, error) {
	v, ok := args.Get(conversionArg)
	if !ok {
		return nil, errMissingArg
	}
	switch v.Type().Nature() {
	case semantic.String:
		bytes, err := goHex.DecodeString(v.Str())
		if err != nil {
			return nil, errors.Newf(codes.Invalid, "cannot convert string %q to bytes due to hex decoding error: %v", v.Str(), err)
		}
		return values.NewBytes(bytes), nil
	default:
		return nil, errors.Newf(codes.Invalid, "hex cannot convert %v to bytes", v.Type())
	}
}
