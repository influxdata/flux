package hash

import (
	"context"
	_hmac "crypto/hmac"
	_md5 "crypto/md5"
	_sha1 "crypto/sha1"
	_sha256 "crypto/sha256"
	_b64 "encoding/base64"
	"encoding/hex"
	"github.com/cespare/xxhash/v2"
	"strconv"

	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/internal/errors"
	"github.com/InfluxCommunity/flux/interpreter"
	"github.com/InfluxCommunity/flux/runtime"
	"github.com/InfluxCommunity/flux/semantic"
	"github.com/InfluxCommunity/flux/values"
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
	runtime.RegisterPackageValue(pkgName, "sha1", makeFunction("sha1", sha1))
	runtime.RegisterPackageValue(pkgName, "xxhash64", makeFunction("xxhash64", xxhash64))
	runtime.RegisterPackageValue(pkgName, "cityhash64", makeFunction("cityhash64", cityhash64))
	runtime.RegisterPackageValue(pkgName, "md5", makeFunction("md5", md5))
	runtime.RegisterPackageValue(pkgName, "b64", makeFunction("b64", b64))
	runtime.RegisterPackageValue(pkgName, "hmac", makeFunction("hmac", hmac))
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

func sha1(args interpreter.Arguments) (values.Value, error) {
	v, ok := args.Get(conversionArg)
	if !ok {
		return nil, errMissingArg
	} else if v.IsNull() {
		return values.Null, nil
	}
	switch v.Type().Nature() {
	case semantic.String:
		s := v.Str()
		hash := _sha1.Sum([]byte(s))
		str := hex.EncodeToString(hash[:])
		return values.NewString(str), nil
	default:
		return nil, errors.Newf(codes.Invalid, "hash cannot convert %v to sha1", v.Type())
	}
}

func md5(args interpreter.Arguments) (values.Value, error) {
	v, ok := args.Get(conversionArg)
	if !ok {
		return nil, errMissingArg
	} else if v.IsNull() {
		return values.Null, nil
	}
	switch v.Type().Nature() {
	case semantic.String:
		s := v.Str()
		hash := _md5.Sum([]byte(s))
		str := hex.EncodeToString(hash[:])
		return values.NewString(str), nil
	default:
		return nil, errors.Newf(codes.Invalid, "hash cannot convert %v to md5", v.Type())
	}
}

func b64(args interpreter.Arguments) (values.Value, error) {
	v, ok := args.Get(conversionArg)
	if !ok {
		return nil, errMissingArg
	} else if v.IsNull() {
		return values.Null, nil
	}
	switch v.Type().Nature() {
	case semantic.String:
		s := v.Str()
		str := _b64.URLEncoding.EncodeToString([]byte(s))
		return values.NewString(str), nil
	default:
		return nil, errors.Newf(codes.Invalid, "hash cannot convert %v to b64", v.Type())
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

func hmac(args interpreter.Arguments) (values.Value, error) {
	v, ok := args.Get(conversionArg)
	if !ok {
		return nil, errMissingArg
	} else if v.IsNull() {
		return values.Null, nil
	}
	k, kok := args.Get("k")
	if !kok {
		return nil, errMissingArg
	} else if k.IsNull() {
		return values.Null, nil
	}
	switch v.Type().Nature() {
	case semantic.String:
		key_for_sign := []byte(k.Str())
		h := _hmac.New(_sha1.New, key_for_sign)
		h.Write([]byte(v.Str()))
		str := _b64.StdEncoding.EncodeToString(h.Sum(nil))
		return values.NewString(str), nil
	default:
		return nil, errors.Newf(codes.Invalid, "hash cannot convert %v to hmac", v.Type())
	}
}
