package universe

import (
	"context"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/parser"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func init() {
	runtime.RegisterPackageValue("universe", "string", stringConv)
	runtime.RegisterPackageValue("universe", "int", intConv)
	runtime.RegisterPackageValue("universe", "uint", uintConv)
	runtime.RegisterPackageValue("universe", "float", floatConv)
	runtime.RegisterPackageValue("universe", "bool", boolConv)
	runtime.RegisterPackageValue("universe", "time", timeConv)
	runtime.RegisterPackageValue("universe", "duration", durationConv)
	runtime.RegisterPackageValue("universe", "bytes", byteConv)
	runtime.RegisterPackageValue("universe", "_vectorizedFloat", vectorizedFloatConv)
}

var (
	convBoolType        = runtime.MustLookupBuiltinType("universe", "bool")
	convIntType         = runtime.MustLookupBuiltinType("universe", "int")
	convUintType        = runtime.MustLookupBuiltinType("universe", "uint")
	convFloatType       = runtime.MustLookupBuiltinType("universe", "float")
	convStringType      = runtime.MustLookupBuiltinType("universe", "string")
	convTimeType        = runtime.MustLookupBuiltinType("universe", "time")
	convDurationType    = runtime.MustLookupBuiltinType("universe", "duration")
	convBytesType       = runtime.MustLookupBuiltinType("universe", "bytes")
	convVectorFloatType = runtime.MustLookupBuiltinType("universe", "_vectorizedFloat")
)

const (
	conversionArg = "v"
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
			str = strconv.FormatInt(v.Int(), 10)
		case semantic.UInt:
			str = strconv.FormatUint(v.UInt(), 10)
		case semantic.Float:
			str = strconv.FormatFloat(v.Float(), 'f', -1, 64)
		case semantic.Bool:
			str = strconv.FormatBool(v.Bool())
		case semantic.Time:
			str = v.Time().String()
		case semantic.Duration:
			str = v.Duration().String()
		case semantic.Bytes:
			var sb strings.Builder
			var vB = v.Bytes()
			for len(vB) > 0 {
				r, size := utf8.DecodeRune(vB)
				if r == utf8.RuneError && size == 1 {
					return nil, errors.Newf(codes.Invalid, "bytes contained non utf8 bytes, cannot be converted into a string")
				}
				vB = vB[size:]

				sb.WriteRune(r)
			}
			str = sb.String()
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
			n, err := strconv.ParseInt(v.Str(), 10, 64)
			if err != nil {
				return nil, errors.Newf(codes.Invalid, "cannot convert string %q to int due to invalid syntax", v.Str())
			}
			i = n
		case semantic.Int:
			i = v.Int()
		case semantic.UInt:
			i = int64(v.UInt())
		case semantic.Float:
			i = int64(v.Float())
		case semantic.Bool:
			if v.Bool() {
				i = 1
			} else {
				i = 0
			}
		case semantic.Time:
			i = int64(v.Time())
		case semantic.Duration:
			i = int64(v.Duration().Duration())
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
			n, err := strconv.ParseUint(v.Str(), 10, 64)
			if err != nil {
				return nil, errors.Newf(codes.Invalid, "cannot convert string %q to uint due to invalid syntax", v.Str())
			}
			i = n
		case semantic.Int:
			i = uint64(v.Int())
		case semantic.UInt:
			i = v.UInt()
		case semantic.Float:
			i = uint64(v.Float())
		case semantic.Bool:
			if v.Bool() {
				i = 1
			} else {
				i = 0
			}
		case semantic.Time:
			i = uint64(v.Time())
		case semantic.Duration:
			i = uint64(v.Duration().Duration())
		default:
			return nil, errors.Newf(codes.Invalid, "cannot convert %v to uint", v.Type())
		}
		return values.NewUInt(i), nil
	},
	false,
)

func toFloatValue(v values.Value) (values.Value, error) {
	var float float64
	switch v.Type().Nature() {
	case semantic.String:
		n, err := strconv.ParseFloat(v.Str(), 64)
		if err != nil {
			return nil, errors.Newf(codes.Invalid, "cannot convert string %q to float due to invalid syntax", v.Str())
		}
		float = n
	case semantic.Int:
		float = float64(v.Int())
	case semantic.UInt:
		float = float64(v.UInt())
	case semantic.Float:
		float = v.Float()
	case semantic.Bool:
		if v.Bool() {
			float = 1
		} else {
			float = 0
		}
	default:
		return nil, errors.Newf(codes.Invalid, "cannot convert %v to float", v.Type())
	}

	return values.NewFloat(float), nil
}

var floatConv = values.NewFunction(
	"float",
	convFloatType,
	func(ctx context.Context, args values.Object) (values.Value, error) {
		v, ok := args.Get(conversionArg)
		if !ok {
			return nil, errMissingArg
		} else if v.IsNull() {
			return values.Null, nil
		}

		// When the incoming value is Dynamic, pull out the inner value.
		if v.Type().Nature() == semantic.Dynamic {
			v = v.Dynamic().Inner()
		}

		return toFloatValue(v)
	},
	false,
)

var boolConv = values.NewFunction(
	"bool",
	convBoolType,
	func(ctx context.Context, args values.Object) (values.Value, error) {
		var b bool
		v, ok := args.Get(conversionArg)
		if !ok {
			return nil, errMissingArg
		} else if v.IsNull() {
			return values.Null, nil
		}

		// When the incoming value is Dynamic, pull out the inner value.
		if v.Type().Nature() == semantic.Dynamic {
			v = v.Dynamic().Inner()
		}

		switch v.Type().Nature() {
		case semantic.String:
			switch s := v.Str(); s {
			case "true":
				b = true
			case "false":
				b = false
			default:
				return nil, errors.Newf(codes.Invalid, "cannot convert string %q to bool", s)
			}
		case semantic.Int:
			switch n := v.Int(); n {
			case 0:
				b = false
			case 1:
				b = true
			default:
				return nil, errors.Newf(codes.Invalid, "cannot convert int %d to bool, must be 0 or 1", n)
			}
		case semantic.UInt:
			switch n := v.UInt(); n {
			case 0:
				b = false
			case 1:
				b = true
			default:
				return nil, errors.Newf(codes.Invalid, "cannot convert uint %d to bool, must be 0 or 1", n)
			}
		case semantic.Float:
			switch n := v.Float(); n {
			case 0:
				b = false
			case 1:
				b = true
			default:
				return nil, errors.Newf(codes.Invalid, "cannot convert float %f to bool, must be 0 or 1", n)
			}
		case semantic.Bool:
			b = v.Bool()
		default:
			return nil, errors.Newf(codes.Invalid, "cannot convert %v to bool", v.Type())
		}
		return values.NewBool(b), nil
	},
	false,
)

var timeConv = values.NewFunction(
	"time",
	convTimeType,
	func(ctx context.Context, args values.Object) (values.Value, error) {
		var t values.Time
		v, ok := args.Get(conversionArg)
		if !ok {
			return nil, errMissingArg
		} else if v.IsNull() {
			return values.Null, nil
		}
		switch v.Type().Nature() {
		case semantic.String:
			ts, err := parser.ParseTime(v.Str())
			if err != nil {
				return nil, errors.Wrapf(err, codes.Invalid, "cannot convert string %q to time due to invalid syntax", v.Str())
			}
			t = values.Time(ts.UnixNano())
		case semantic.Int:
			t = values.Time(v.Int())
		case semantic.UInt:
			t = values.Time(v.UInt())
		case semantic.Time:
			t = v.Time()
		default:
			return nil, errors.Newf(codes.Invalid, "cannot convert %v to time", v.Type())
		}
		return values.NewTime(t), nil
	},
	false,
)

var durationConv = values.NewFunction(
	"duration",
	convDurationType,
	func(ctx context.Context, args values.Object) (values.Value, error) {
		var d values.Duration
		v, ok := args.Get(conversionArg)
		if !ok {
			return nil, errMissingArg
		} else if v.IsNull() {
			return values.Null, nil
		}

		// When the incoming value is Dynamic, pull out the inner value.
		if v.Type().Nature() == semantic.Dynamic {
			v = v.Dynamic().Inner()
		}

		switch v.Type().Nature() {
		case semantic.String:
			n, err := values.ParseDuration(v.Str())
			if err != nil {
				return nil, errors.Newf(codes.Invalid, "cannot convert string %q to duration due to invalid syntax", v.Str())
			}
			d = n
		case semantic.Int:
			d = values.ConvertDurationNsecs(time.Duration(v.Int()))
		case semantic.UInt:
			d = values.ConvertDurationNsecs(time.Duration(v.UInt()))
		case semantic.Duration:
			d = v.Duration()
		default:
			return nil, errors.Newf(codes.Invalid, "cannot convert %v to duration", v.Type())
		}
		return values.NewDuration(d), nil
	},
	false,
)

var byteConv = values.NewFunction(
	"bytes",
	convBytesType,
	func(ctx context.Context, args values.Object) (values.Value, error) {
		v, ok := args.Get(conversionArg)
		if !ok {
			return nil, errMissingArg
		}

		// When the incoming value is Dynamic, pull out the inner value.
		if v.Type().Nature() == semantic.Dynamic {
			v = v.Dynamic().Inner()
		}

		switch v.Type().Nature() {
		case semantic.String:
			return values.NewBytes([]byte(v.Str())), nil
		default:
			return nil, errors.Newf(codes.Invalid, "cannot convert %v to bytes", v.Type())
		}
	},
	false,
)

var vectorizedFloatConv = values.NewFunction(
	"_vectorizedFloat",
	convVectorFloatType,
	func(ctx context.Context, args values.Object) (values.Value, error) {
		v, ok := args.Get(conversionArg)
		if !ok {
			return nil, errMissingArg
		}

		if v.IsNull() {
			v.Retain()
			return v, nil
		}

		mem := memory.GetAllocator(ctx)

		switch v.Type().Nature() {
		case semantic.Vector:
			vec := v.Vector()

			// Delegate to row-based version when the value is constant
			if vr, ok := vec.(*values.VectorRepeatValue); ok {
				fv, err := toFloatValue(vr.Value())
				if err != nil {
					return nil, err
				}
				return values.NewVectorRepeatValue(fv), nil
			}

			arr, err := array.ToFloatConv(mem, vec.Arr())
			if err != nil {
				return nil, err
			}
			return values.NewFloatVectorValue(arr), nil
		default:
			return nil, errors.Newf(codes.Invalid, "cannot convert %v to v[float]", v.Type())
		}

	},
	false,
)
