package math

import (
	"context"
	"fmt"
	"math"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var SpecialFns map[string]values.Function

func generateMathFunctionX(name string, mathFn func(float64) float64) values.Function {
	return values.NewFunction(
		name,
		semantic.MustLookupBuiltinType("math", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			v, ok := args.Get("x")
			if !ok {
				return nil, errors.New(codes.Invalid, "missing argument x")
			}

			if v.Type().Nature() == semantic.Float {
				return values.NewFloat(mathFn(v.Float())), nil
			}

			return nil, fmt.Errorf("cannot convert argument of type %v to float", v.Type().Nature())
		}, false,
	)
}

func generateMathFunctionXY(name string, mathFn func(float64, float64) float64, argNames ...string) values.Function {
	if argNames == nil {
		argNames = []string{"x", "y"}
	}
	return values.NewFunction(
		name,
		semantic.MustLookupBuiltinType("math", name),
		func(ctx context.Context, args values.Object) (values.Value, error) {
			v1, ok := args.Get(argNames[0])
			if !ok {
				return nil, fmt.Errorf("missing argument %s", argNames[0])
			}
			v2, ok := args.Get(argNames[1])
			if !ok {
				return nil, fmt.Errorf("missing argument %s", argNames[1])
			}

			if v1.Type().Nature() == semantic.Float {
				if v2.Type().Nature() == semantic.Float {
					return values.NewFloat(mathFn(v1.Float(), v2.Float())), nil
				} else {
					return nil, fmt.Errorf("cannot convert argument %s of type %v to float", argNames[1], v2.Type().Nature())
				}
			}
			return nil, fmt.Errorf("cannot convert argument %s of type %v to float", argNames[0], v1.Type().Nature())
		}, false,
	)
}

func init() {
	// constants
	runtime.RegisterPackageValue("math", "pi", values.NewFloat(math.Pi))
	runtime.RegisterPackageValue("math", "e", values.NewFloat(math.E))
	runtime.RegisterPackageValue("math", "phi", values.NewFloat(math.Phi))
	runtime.RegisterPackageValue("math", "sqrt2", values.NewFloat(math.Sqrt2))
	runtime.RegisterPackageValue("math", "sqrte", values.NewFloat(math.SqrtE))
	runtime.RegisterPackageValue("math", "sqrtpi", values.NewFloat(math.SqrtPi))
	runtime.RegisterPackageValue("math", "sqrtphi", values.NewFloat(math.SqrtPhi))
	runtime.RegisterPackageValue("math", "log2e", values.NewFloat(math.Log2E))
	runtime.RegisterPackageValue("math", "ln2", values.NewFloat(math.Ln2))
	runtime.RegisterPackageValue("math", "ln10", values.NewFloat(math.Ln10))
	runtime.RegisterPackageValue("math", "log10e", values.NewFloat(math.Log10E))

	runtime.RegisterPackageValue("math", "maxfloat", values.NewFloat(math.MaxFloat64))
	runtime.RegisterPackageValue("math", "smallestNonzeroFloat", values.NewFloat(math.SmallestNonzeroFloat64))
	runtime.RegisterPackageValue("math", "maxint", values.NewInt(math.MaxInt64))
	runtime.RegisterPackageValue("math", "minint", values.NewFloat(math.MinInt64))
	runtime.RegisterPackageValue("math", "maxuint", values.NewUInt(math.MaxUint64))

	runtime.RegisterPackageValue("math", "abs", generateMathFunctionX("abs", math.Abs))
	runtime.RegisterPackageValue("math", "acos", generateMathFunctionX("acos", math.Acos))
	runtime.RegisterPackageValue("math", "acosh", generateMathFunctionX("acosh", math.Acosh))
	runtime.RegisterPackageValue("math", "asin", generateMathFunctionX("asin", math.Asin))
	runtime.RegisterPackageValue("math", "asinh", generateMathFunctionX("asinh", math.Asinh))
	runtime.RegisterPackageValue("math", "atan", generateMathFunctionX("atan", math.Atan))
	runtime.RegisterPackageValue("math", "atan2", generateMathFunctionXY("atan2", math.Atan2))
	runtime.RegisterPackageValue("math", "atanh", generateMathFunctionX("atanh", math.Atanh))
	runtime.RegisterPackageValue("math", "cbrt", generateMathFunctionX("cbrt", math.Cbrt))
	runtime.RegisterPackageValue("math", "ceil", generateMathFunctionX("ceil", math.Ceil))
	runtime.RegisterPackageValue("math", "copysign", generateMathFunctionXY("copysign", math.Copysign))
	runtime.RegisterPackageValue("math", "cos", generateMathFunctionX("cos", math.Cos))
	runtime.RegisterPackageValue("math", "cosh", generateMathFunctionX("cosh", math.Cosh))
	runtime.RegisterPackageValue("math", "dim", generateMathFunctionXY("dim", math.Dim))
	runtime.RegisterPackageValue("math", "erf", generateMathFunctionX("erf", math.Erf))
	runtime.RegisterPackageValue("math", "erfc", generateMathFunctionX("erfc", math.Erfc))
	runtime.RegisterPackageValue("math", "erfcinv", generateMathFunctionX("erfcinv", math.Erfcinv))
	runtime.RegisterPackageValue("math", "erfinv", generateMathFunctionX("erfinv", math.Erfinv))
	runtime.RegisterPackageValue("math", "exp", generateMathFunctionX("exp", math.Exp))
	runtime.RegisterPackageValue("math", "exp2", generateMathFunctionX("exp2", math.Exp2))
	runtime.RegisterPackageValue("math", "expm1", generateMathFunctionX("expm1", math.Expm1))
	runtime.RegisterPackageValue("math", "floor", generateMathFunctionX("floor", math.Floor))
	runtime.RegisterPackageValue("math", "gamma", generateMathFunctionX("gamma", math.Gamma))
	runtime.RegisterPackageValue("math", "hypot", generateMathFunctionXY("hypot", math.Hypot, "p", "q"))
	runtime.RegisterPackageValue("math", "j0", generateMathFunctionX("j0", math.J0))
	runtime.RegisterPackageValue("math", "j1", generateMathFunctionX("j1", math.J1))
	runtime.RegisterPackageValue("math", "log", generateMathFunctionX("log", math.Log))
	runtime.RegisterPackageValue("math", "log10", generateMathFunctionX("log10", math.Log10))
	runtime.RegisterPackageValue("math", "log1p", generateMathFunctionX("log1p", math.Log1p))
	runtime.RegisterPackageValue("math", "log2", generateMathFunctionX("log2", math.Log2))
	runtime.RegisterPackageValue("math", "logb", generateMathFunctionX("logb", math.Logb))
	// TODO: change to max and min when we eliminate namespace collisions
	runtime.RegisterPackageValue("math", "mMax", generateMathFunctionXY("mMax", math.Max))
	runtime.RegisterPackageValue("math", "mMin", generateMathFunctionXY("mMin", math.Min))
	runtime.RegisterPackageValue("math", "mod", generateMathFunctionXY("mod", math.Mod))
	runtime.RegisterPackageValue("math", "nextafter", generateMathFunctionXY("nextafter", math.Nextafter))
	runtime.RegisterPackageValue("math", "pow", generateMathFunctionXY("pow", math.Pow))
	runtime.RegisterPackageValue("math", "remainder", generateMathFunctionXY("remainder", math.Remainder))
	runtime.RegisterPackageValue("math", "round", generateMathFunctionX("round", math.Round))
	runtime.RegisterPackageValue("math", "roundtoeven", generateMathFunctionX("roundtoeven", math.RoundToEven))
	runtime.RegisterPackageValue("math", "sin", generateMathFunctionX("sin", math.Sin))
	runtime.RegisterPackageValue("math", "sinh", generateMathFunctionX("sinh", math.Sinh))
	runtime.RegisterPackageValue("math", "sqrt", generateMathFunctionX("sqrt", math.Sqrt))
	runtime.RegisterPackageValue("math", "tan", generateMathFunctionX("tan", math.Tan))
	runtime.RegisterPackageValue("math", "tanh", generateMathFunctionX("tanh", math.Tanh))
	runtime.RegisterPackageValue("math", "trunc", generateMathFunctionX("trunc", math.Trunc))
	runtime.RegisterPackageValue("math", "y0", generateMathFunctionX("y0", math.Y0))
	runtime.RegisterPackageValue("math", "y1", generateMathFunctionX("y1", math.Y1))

	SpecialFns = map[string]values.Function{
		// float --> uint
		"float64bits": values.NewFunction(
			"float64bits",
			semantic.MustLookupBuiltinType("math", "float64bits"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("f")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument f")
				}

				if v1.Type().Nature() == semantic.Float {
					return values.NewUInt(math.Float64bits(v1.Float())), nil
				}
				return nil, fmt.Errorf("cannot convert argument f of type %v to float", v1.Type().Nature())
			}, false,
		),
		"float64frombits": values.NewFunction(
			"float64frombits",
			semantic.MustLookupBuiltinType("math", "float64frombits"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("b")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument b")
				}

				if v1.Type().Nature() == semantic.UInt {
					return values.NewFloat(math.Float64frombits(v1.UInt())), nil
				}
				return nil, fmt.Errorf("cannot convert argument b of type %v to uint", v1.Type().Nature())
			}, false,
		),
		// float --> int
		"ilogb": values.NewFunction(
			"ilogb",
			semantic.MustLookupBuiltinType("math", "ilogb"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("x")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument x")
				}

				if v1.Type().Nature() == semantic.Float {
					return values.NewInt(int64(math.Ilogb(v1.Float()))), nil
				}
				return nil, fmt.Errorf("cannot convert argument x of type %v to float", v1.Type().Nature())
			}, false,
		),
		// float --> {frac: float, exp: int}
		"frexp": values.NewFunction(
			"frexp",
			semantic.MustLookupBuiltinType("math", "frexp"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("f")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument f")
				}

				if v1.Type().Nature() == semantic.Float {
					frac, exp := math.Frexp(v1.Float())
					return values.NewObjectWithValues(map[string]values.Value{"frac": values.NewFloat(frac), "exp": values.NewInt(int64(exp))}), nil
				}
				return nil, fmt.Errorf("cannot convert argument f of type %v to float", v1.Type().Nature())
			}, false,
		),
		// float --> {lgamma: float, sign: int}
		"lgamma": values.NewFunction(
			"lgamma",
			semantic.MustLookupBuiltinType("math", "lgamma"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("x")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument x")
				}

				if v1.Type().Nature() == semantic.Float {
					lgamma, sign := math.Lgamma(v1.Float())
					return values.NewObjectWithValues(map[string]values.Value{"lgamma": values.NewFloat(lgamma), "sign": values.NewInt(int64(sign))}), nil
				}
				return nil, fmt.Errorf("cannot convert argument x of type %v to float", v1.Type().Nature())
			}, false,
		),
		// float --> {int: float, frac: float}
		"modf": values.NewFunction(
			"modf",
			semantic.MustLookupBuiltinType("math", "modf"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("f")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument f")
				}

				if v1.Type().Nature() == semantic.Float {
					intres, frac := math.Modf(v1.Float())
					return values.NewObjectWithValues(map[string]values.Value{"int": values.NewFloat(intres), "frac": values.NewFloat(frac)}), nil
				}
				return nil, fmt.Errorf("cannot convert argument f of type %v to float", v1.Type().Nature())
			}, false,
		),
		// float --> {sin: float, cos: float}
		"sincos": values.NewFunction(
			"sincos",
			semantic.MustLookupBuiltinType("math", "sincos"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("x")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument x")
				}

				if v1.Type().Nature() == semantic.Float {
					sin, cos := math.Sin(v1.Float()), math.Cos(v1.Float())
					return values.NewObjectWithValues(map[string]values.Value{"sin": values.NewFloat(sin), "cos": values.NewFloat(cos)}), nil
				}
				return nil, fmt.Errorf("cannot convert argument x of type %v to float", v1.Type().Nature())
			}, false,
		),
		// float, int --> bool
		"isInf": values.NewFunction(
			"isInf",
			semantic.MustLookupBuiltinType("math", "isInf"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("f")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument f")
				}
				v2, ok := args.Get("sign")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument sign")
				}

				if v1.Type().Nature() == semantic.Float {
					if v2.Type().Nature() == semantic.Int {
						return values.NewBool(math.IsInf(v1.Float(), int(v2.Int()))), nil
					} else {
						return nil, fmt.Errorf("cannot convert argument sign of type %v to int", v2.Type().Nature())
					}
				}
				return nil, fmt.Errorf("cannot convert argument f of type %v to float", v1.Type().Nature())
			}, false,
		),
		// float --> bool
		"isNaN": values.NewFunction(
			"isNaN",
			semantic.MustLookupBuiltinType("math", "isNaN"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("f")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument f")
				}

				if v1.Type().Nature() == semantic.Float {
					return values.NewBool(math.IsNaN(v1.Float())), nil
				}
				return nil, fmt.Errorf("cannot convert argument f of type %v to float", v1.Type().Nature())
			}, false,
		),
		// float --> bool
		"signbit": values.NewFunction(
			"signbit",
			semantic.MustLookupBuiltinType("math", "signbit"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("x")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument x")
				}

				if v1.Type().Nature() == semantic.Float {
					return values.NewBool(math.Signbit(v1.Float())), nil
				}
				return nil, fmt.Errorf("cannot convert argument x of type %v to float", v1.Type().Nature())
			}, false,
		),
		// () --> float
		"NaN": values.NewFunction(
			"NaN",
			semantic.MustLookupBuiltinType("math", "NaN"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				return values.NewFloat(math.NaN()), nil
			}, false,
		),
		// (int) --> float
		"mInf": values.NewFunction(
			"inf",
			semantic.MustLookupBuiltinType("math", "mInf"),
			func(ctx context.Context, args values.Object) (values.Value, error) {

				v1, ok := args.Get("sign")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument sign")
				}

				if v1.Type().Nature() == semantic.Int {
					return values.NewFloat(math.Inf(int(v1.Int()))), nil
				}
				return nil, fmt.Errorf("cannot convert argument sign of type %v to int", v1.Type().Nature())
			}, false,
		),
		// (int, float) --> float
		"jn": values.NewFunction(
			"jn",
			semantic.MustLookupBuiltinType("math", "jn"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("n")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument n")
				}
				v2, ok := args.Get("x")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument x")
				}

				if v1.Type().Nature() == semantic.Int {
					if v2.Type().Nature() == semantic.Float {
						return values.NewFloat(math.Jn(int(v1.Int()), v2.Float())), nil
					} else {
						return nil, fmt.Errorf("cannot convert argument x of type %v to float", v2.Type().Nature())
					}
				}
				return nil, fmt.Errorf("cannot convert argument n of type %v to int", v1.Type().Nature())
			}, false,
		),
		// (int, float) --> float
		"yn": values.NewFunction(
			"yn",
			semantic.MustLookupBuiltinType("math", "yn"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("n")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument n")
				}
				v2, ok := args.Get("x")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument x")
				}

				if v1.Type().Nature() == semantic.Int {
					if v2.Type().Nature() == semantic.Float {
						return values.NewFloat(math.Yn(int(v1.Int()), v2.Float())), nil
					} else {
						return nil, fmt.Errorf("cannot convert argument x of type %v to float", v2.Type().Nature())
					}
				}
				return nil, fmt.Errorf("cannot convert argument n of type %v to int", v1.Type().Nature())
			}, false,
		),
		// (float, int) --> float
		"ldexp": values.NewFunction(
			"ldexp",
			semantic.MustLookupBuiltinType("math", "ldexp"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("frac")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument frac")
				}
				v2, ok := args.Get("exp")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument exp")
				}

				if v1.Type().Nature() == semantic.Float {
					if v2.Type().Nature() == semantic.Int {
						return values.NewFloat(math.Ldexp(v1.Float(), int(v2.Int()))), nil
					} else {
						return nil, fmt.Errorf("cannot convert argument exp of type %v to int", v2.Type().Nature())
					}
				}
				return nil, fmt.Errorf("cannot convert argument frac of type %v to float", v1.Type().Nature())
			}, false,
		),
		// int --> float
		"pow10": values.NewFunction(
			"pow10",
			semantic.MustLookupBuiltinType("math", "pow10"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("n")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument frac")
				}

				if v1.Type().Nature() == semantic.Int {
					return values.NewFloat(math.Pow10(int(v1.Int()))), nil
				}
				return nil, fmt.Errorf("cannot convert argument n of type %v to int", v1.Type().Nature())
			}, false,
		),
	}

	// special case args and/or return types not worth generalizing
	runtime.RegisterPackageValue("math", "float64bits", SpecialFns["float64bits"])
	runtime.RegisterPackageValue("math", "float64frombits", SpecialFns["float64frombits"])
	runtime.RegisterPackageValue("math", "ilogb", SpecialFns["ilogb"])
	runtime.RegisterPackageValue("math", "frexp", SpecialFns["frexp"])
	runtime.RegisterPackageValue("math", "lgamma", SpecialFns["lgamma"])
	runtime.RegisterPackageValue("math", "modf", SpecialFns["modf"])
	runtime.RegisterPackageValue("math", "sincos", SpecialFns["sincos"])
	runtime.RegisterPackageValue("math", "isInf", SpecialFns["isInf"])
	runtime.RegisterPackageValue("math", "isNaN", SpecialFns["isNaN"])
	runtime.RegisterPackageValue("math", "signbit", SpecialFns["signbit"])
	runtime.RegisterPackageValue("math", "NaN", SpecialFns["NaN"])
	runtime.RegisterPackageValue("math", "mInf", SpecialFns["mInf"])
	runtime.RegisterPackageValue("math", "jn", SpecialFns["jn"])
	runtime.RegisterPackageValue("math", "yn", SpecialFns["yn"])
	runtime.RegisterPackageValue("math", "ldexp", SpecialFns["ldexp"])
	runtime.RegisterPackageValue("math", "pow10", SpecialFns["pow10"])
}
