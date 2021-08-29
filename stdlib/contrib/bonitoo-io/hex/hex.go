package hex

import (
	"context"
	goHex "encoding/hex"
	"strconv"
	"strings"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	conversionArg = "v"
	pkgName       = "contrib/bonitoo-io/hex"
)

func init() {
	runtime.RegisterPackageValue(pkgName, "string", stringConv)
	runtime.RegisterPackageValue(pkgName, "int", intConv)
	runtime.RegisterPackageValue(pkgName, "uint", uintConv)
	runtime.RegisterPackageValue(pkgName, "bytes", bytesConv)
}

var (
	convIntType    = runtime.MustLookupBuiltinType(pkgName, "int")
	convUintType   = runtime.MustLookupBuiltinType(pkgName, "uint")
	convStringType = runtime.MustLookupBuiltinType(pkgName, "string")
	convBytesType  = runtime.MustLookupBuiltinType(pkgName, "bytes")
)

var errMissingArg = errors.Newf(codes.Invalid, "missing argument %q", conversionArg)

var stringConv = values.NewFunction(
	"string",
	convStringType,
	func(ctx context.Context, args values.Object) (values.Value, error) {
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
	},
	false,
)

var intConv = values.NewFunction(
	"int",
	convIntType,
	func(ctx context.Context, args values.Object) (values.Value, error) {
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
			return nil, errors.Newf(codes.Invalid, "cannot convert %v to int", v.Type())
		}
		return values.NewInt(i), nil
	},
	false,
)

var uintConv = values.NewFunction(
	"uint",
	convUintType,
	func(ctx context.Context, args values.Object) (values.Value, error) {
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
			return nil, errors.Newf(codes.Invalid, "cannot convert %v to uint", v.Type())
		}
		return values.NewUInt(i), nil
	},
	false,
)

var bytesConv = values.NewFunction(
	"bytes",
	convBytesType,
	func(ctx context.Context, args values.Object) (values.Value, error) {
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
			return nil, errors.Newf(codes.Invalid, "cannot convert %v to bytes", v.Type())
		}
	},
	false,
)
