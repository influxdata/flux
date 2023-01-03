package hash

import (
	"context"
	_sha256 "crypto/sha256"
	"encoding/hex"
	"github.com/cespare/xxhash/v2"
	"strconv"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	conversionArg = "v"
	pkgName       = "contrib/qxip/hash"
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
	runtime.RegisterPackageValue(pkgName, "sha256", makeFunction("sha256", sha256))
	runtime.RegisterPackageValue(pkgName, "xxhash64", makeFunction("xxhash64", xxhash64))
	runtime.RegisterPackageValue(pkgName, "cityhash64", makeFunction("cityhash64", cityhash64))
}

var errMissingArg = errors.Newf(codes.Invalid, "missing argument %q", conversionArg)

func sha256(args interpreter.Arguments) (values.Value, error) {
	v, ok := args.Get(conversionArg)
	if !ok {
		return nil, errMissingArg
	} else if v.IsNull() {
		return values.Null, nil
	}
	switch v.Type().Nature() {
	case semantic.String:
		s := v.Str()
		hash := _sha256.Sum256([]byte(s))
		str := hex.EncodeToString(hash[:])
		return values.NewString(str), nil
	default:
		return nil, errors.Newf(codes.Invalid, "hash cannot convert %v to sha256", v.Type())
	}
}

func xxhash64(args interpreter.Arguments) (values.Value, error) {
	v, ok := args.Get(conversionArg)
	if !ok {
		return nil, errMissingArg
	} else if v.IsNull() {
		return values.Null, nil
	}
	switch v.Type().Nature() {
	case semantic.String:
		s := v.Str()
		hash := xxhash.Sum64([]byte(s))
		str := strconv.FormatUint(hash, 10)
		return values.NewString(str), nil
	default:
		return nil, errors.Newf(codes.Invalid, "hash cannot convert %v to sha256", v.Type())
	}
}

func cityhash64(args interpreter.Arguments) (values.Value, error) {
	v, ok := args.Get(conversionArg)
	if !ok {
		return nil, errMissingArg
	} else if v.IsNull() {
		return values.Null, nil
	}
	switch v.Type().Nature() {
	case semantic.String:
		s := v.Str()
		hash := CityHash64([]byte(s), uint32(len(s)))
		str := strconv.FormatUint(hash, 10)
		return values.NewString(str), nil
	default:
		return nil, errors.Newf(codes.Invalid, "hash cannot convert %v to sha256", v.Type())
	}
}
