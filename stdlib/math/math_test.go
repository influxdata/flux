package math

import (
	"context"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"math"
	"math/rand"
	"testing"

	"github.com/influxdata/flux/values"
)

func TestMathFunctionsX(t *testing.T) {
	testCases := []struct {
		name   string
		mathFn func(float64) float64
	}{
		{"abs", math.Abs},
		{"abs", math.Abs},
		{"acos", math.Acos},
		{"acosh", math.Acosh},
		{"asin", math.Asin},
		{"asinh", math.Asinh},
		{"atan", math.Atan},
		{"atanh", math.Atanh},
		{"cbrt", math.Cbrt},
		{"ceil", math.Ceil},
		{"cos", math.Cos},
		{"cosh", math.Cosh},
		{"erf", math.Erf},
		{"erfc", math.Erfc},
		{"erfcinv", math.Erfcinv},
		{"erfinv", math.Erfinv},
		{"exp", math.Exp},
		{"exp2", math.Exp2},
		{"expm1", math.Expm1},
		{"floor", math.Floor},
		{"gamma", math.Gamma},
		{"j0", math.J0},
		{"j1", math.J1},
		{"log", math.Log},
		{"log10", math.Log10},
		{"log1p", math.Log1p},
		{"log2", math.Log2},
		{"logb", math.Logb},
		{"round", math.Round},
		{"roundtoeven", math.RoundToEven},
		{"sin", math.Sin},
		{"sinh", math.Sinh},
		{"sqrt", math.Sqrt},
		{"tan", math.Tan},
		{"tanh", math.Tanh},
		{"trunc", math.Trunc},
		{"y0", math.Y0},
		{"y1", math.Y1},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxFn := generateMathFunctionX(tc.name, tc.mathFn)

			x := rand.Float64()
			got := tc.mathFn(x)

			fluxArg := values.NewObjectWithValues(map[string]values.Value{"x": values.NewFloat(x)})
			result, err := fluxFn.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
			if err != nil {
				t.Fatal(err)
			}
			want := result.Float()

			if want != got && !(math.IsNaN(want) && math.IsNaN(got)) {
				t.Errorf("math.%s function result input %f: expected %f, got %f", tc.name, x, want, got)
			}
		})
	}
}

func TestMathFunctionsXY(t *testing.T) {
	testCases := []struct {
		name   string
		mathFn func(float64, float64) float64
		xname  string
		yname  string
	}{
		{"hypot", math.Hypot, "p", "q"},
		{"max", math.Max, "x", "y"},
		{"min", math.Min, "x", "y"},
		{"mod", math.Mod, "x", "y"},
		{"nextafter", math.Nextafter, "x", "y"},
		{"pow", math.Pow, "x", "y"},
		{"remainder", math.Remainder, "x", "y"},
		{"dim", math.Dim, "x", "y"},
		{"copysign", math.Copysign, "x", "y"},
		{"atan2", math.Atan2, "x", "y"},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxFn := generateMathFunctionXY(tc.name, tc.mathFn, tc.xname, tc.yname)

			x := rand.Float64()
			y := rand.Float64()
			got := tc.mathFn(x, y)

			fluxArg := values.NewObjectWithValues(map[string]values.Value{tc.xname: values.NewFloat(x), tc.yname: values.NewFloat(y)})
			result, err := fluxFn.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
			if err != nil {
				t.Fatal(err)
			}
			want := result.Float()

			if floatsNotEqual(want, got) {
				t.Errorf("math.%s function result input %f: expected %f, got %f", tc.name, x, want, got)
			}
		})
	}
}

func TestFloat64Bits(t *testing.T) {
	fluxFunc := SpecialFns["float64bits"]
	f := rand.Float64()
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"f": values.NewFloat(f)})
	want := math.Float64bits(f)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.UInt() {
		t.Errorf("input %f: expected %d, got %d", f, want, got.UInt())
	}

}

func TestFloat64FromBits(t *testing.T) {
	fluxFunc := SpecialFns["float64frombits"]
	b := rand.Uint64()
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"b": values.NewUInt(b)})
	want := math.Float64frombits(b)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if floatsNotEqual(want, got.Float()) {
		t.Errorf("input %d: expected %f, got %f", b, want, got.Float())
	}

}

func TestIlogb(t *testing.T) {
	fluxFunc := SpecialFns["ilogb"]
	x := rand.Float64()
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"x": values.NewFloat(x)})
	want := math.Ilogb(x)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != int(got.Int()) {
		t.Errorf("input %f: expected %d, got %f", x, want, got)
	}

}

func TestFrexp(t *testing.T) {
	fluxFunc := SpecialFns["frexp"]
	f := rand.Float64()
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"f": values.NewFloat(f)})
	wantfrac, wantexp := math.Frexp(f)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}

	gotFrac, ok := got.Object().Get("frac")
	if !ok {
		t.Error("frexp missing result for frac")
	}
	gotExp, ok := got.Object().Get("exp")
	if !ok {
		t.Error("frexp missing result for exp")
	}
	if floatsNotEqual(wantfrac, gotFrac.Float()) || int(gotExp.Int()) != wantexp {
		t.Errorf("input %f: expected (%f,%d), got (%f, %d)", f, wantfrac, wantexp, gotFrac, gotExp)
	}

}

func TestLGamma(t *testing.T) {
	fluxFunc := SpecialFns["lgamma"]
	x := rand.Float64()
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"x": values.NewFloat(x)})
	wantLGamma, wantSign := math.Lgamma(x)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}

	gotLGamma, ok := got.Object().Get("lgamma")
	if !ok {
		t.Error("lgamma missing result for lgamma")
	}
	gotSign, ok := got.Object().Get("sign")
	if !ok {
		t.Error("lgamma missing result for sign")
	}
	if floatsNotEqual(wantLGamma, gotLGamma.Float()) || int(gotSign.Int()) != wantSign {
		t.Errorf("input %f: expected (%f,%d), got (%f, %d)", x, wantLGamma, wantSign, gotLGamma, gotSign)
	}

}

func TestModf(t *testing.T) {
	fluxFunc := SpecialFns["modf"]
	f := rand.Float64()
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"f": values.NewFloat(f)})
	wantInt, wantFrac := math.Modf(f)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}

	gotInt, ok := got.Object().Get("int")
	if !ok {
		t.Error("lgamma missing result for lgamma")
	}
	gotFrac, ok := got.Object().Get("frac")
	if !ok {
		t.Error("lgamma missing result for sign")
	}
	if floatsNotEqual(wantInt, gotInt.Float()) || floatsNotEqual(wantFrac, gotFrac.Float()) {
		t.Errorf("input %f: expected (%f,%f), got (%f, %f)", f, wantInt, wantFrac, gotInt, gotFrac)
	}

}

func TestSinCos(t *testing.T) {
	fluxFunc := SpecialFns["sincos"]
	x := rand.Float64()
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"x": values.NewFloat(x)})
	wantSin, wantCos := math.Sincos(x)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}

	gotSin, ok := got.Object().Get("sin")
	if !ok {
		t.Error("sincos missing result for sin")
	}
	gotCos, ok := got.Object().Get("cos")
	if !ok {
		t.Error("cos missing result for cos")
	}
	if floatsNotEqual(wantSin, gotSin.Float()) || floatsNotEqual(wantCos, gotCos.Float()) {
		t.Errorf("input %f: expected (%f,%f), got (%f, %f)", x, wantSin, wantCos, gotSin, gotCos)
	}

}

func TestIsInf(t *testing.T) {
	fluxFunc := SpecialFns["isInf"]
	f := rand.Float64()
	sign := rand.Int()
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"f": values.NewFloat(f), "sign": values.NewInt(int64(sign))})
	want := math.IsInf(f, sign)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Bool() {
		t.Errorf("input %f,%d: expected %t, got %t", f, sign, want, got.Bool())
	}

}

func TestIsNaN(t *testing.T) {
	fluxFunc := SpecialFns["isNaN"]
	f := rand.Float64()
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"f": values.NewFloat(f)})
	want := math.IsNaN(f)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Bool() {
		t.Errorf("input %f: expected %t, got %t", f, want, got.Bool())
	}

}

func TestSignBit(t *testing.T) {
	fluxFunc := SpecialFns["signbit"]
	x := rand.Float64()
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"x": values.NewFloat(x)})
	want := math.Signbit(x)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Bool() {
		t.Errorf("input %f: expected %t, got %t", x, want, got.Bool())
	}

}

func TestNaN(t *testing.T) {
	fluxFunc := SpecialFns["NaN"]

	fluxArg := values.NewObjectWithValues(map[string]values.Value{})
	want := math.NaN()
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if !(math.IsNaN(want) && math.IsNaN(got.Float())) {
		t.Errorf("expected %f, got %f", want, got.Float())
	}

}

func TestInf(t *testing.T) {
	fluxFunc := SpecialFns["mInf"]

	sign := rand.Intn(5000)
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"sign": values.NewInt(int64(sign))})
	want := math.Inf(sign)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if !(math.IsInf(want, sign) && math.IsInf(got.Float(), sign)) {
		t.Errorf("input %d: expected %f, got %f", sign, want, got.Float())
	}

}

func TestJn(t *testing.T) {
	fluxFunc := SpecialFns["jn"]
	x := rand.Float64()
	n := rand.Intn(5000)
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"x": values.NewFloat(x), "n": values.NewInt(int64(n))})
	want := math.Jn(n, x)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if floatsNotEqual(want, got.Float()) {
		t.Errorf("input %f, %d: expected %f, got %f", x, n, want, got)
	}

}

func TestYn(t *testing.T) {
	fluxFunc := SpecialFns["yn"]
	x := rand.Float64()
	n := rand.Intn(5000)
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"x": values.NewFloat(x), "n": values.NewInt(int64(n))})
	want := math.Yn(n, x)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if floatsNotEqual(want, got.Float()) {
		t.Errorf("result input %f, %d: expected %f, got %f", x, n, want, got.Float())
	}

}

func TestLdexp(t *testing.T) {
	fluxFunc := SpecialFns["ldexp"]
	frac := rand.Float64()
	exp := rand.Intn(5000)
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"frac": values.NewFloat(frac), "exp": values.NewInt(int64(exp))})
	want := math.Ldexp(frac, exp)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if floatsNotEqual(want, got.Float()) {
		t.Errorf("input %f, %d: expected %f, got %f", frac, exp, want, got)
	}

}

func TestPow10(t *testing.T) {
	fluxFunc := SpecialFns["pow10"]
	n := rand.Intn(5000)
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"n": values.NewInt(int64(n))})
	want := math.Pow10(n)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.NewTestDependenciesInterface(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if floatsNotEqual(want, got.Float()) {
		t.Errorf("math.pow10 function result input %d: expected %f, got %f", n, want, got)
	}

}

func floatsNotEqual(want, got float64) bool {
	return want != got && !(math.IsNaN(want) && math.IsNaN(got))
}
