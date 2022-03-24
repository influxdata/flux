package regexp

import (
	"context"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func TestCompile(t *testing.T) {
	fluxFunc := SpecialFns["compile"]
	v := values.NewString("alpha32")
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(v.Str())})
	want, _ := regexp.Compile(v.Str())
	realWant := values.NewRegexp(want)
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	got, err := fluxFunc.Call(ctx, fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if realWant.Regexp().String() != got.Regexp().String() {
		t.Errorf("input %s: expected %v, got %v", v, want, got.Regexp())
	}
}

func TestQuoteMeta(t *testing.T) {
	fluxFunc := SpecialFns["quoteMeta"]
	v := values.NewString("Escaping symbols like: .+*?()|[]{}^$")
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(v.Str())})
	want := regexp.QuoteMeta(v.Str())
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	got, err := fluxFunc.Call(ctx, fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Str() {
		t.Errorf("input %s: expected %v, got %v", v, want, got.Str())
	}
}

func TestFindString(t *testing.T) {
	fluxFunc := SpecialFns["findString"]
	re := regexp.MustCompile(`foo.?`)
	r := values.NewRegexp(re)
	v := values.NewString("seafood fool")
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"r": values.NewRegexp(r.Regexp()), "v": values.NewString(v.Str())})
	want := r.Regexp().FindString(v.Str())
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	got, err := fluxFunc.Call(ctx, fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Str() {
		t.Errorf("expected %v, got %v", want, got.Str())
	}
}

func TestFindStringIndex(t *testing.T) {
	fluxFunc := SpecialFns["findStringIndex"]
	re := regexp.MustCompile(`ab?`)
	r := values.NewRegexp(re)
	v := values.NewString("tablett")
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"r": values.NewRegexp(r.Regexp()), "v": values.NewString(v.Str())})
	want := r.Regexp().FindStringIndex(v.Str())
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	got, err := fluxFunc.Call(ctx, fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if !values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicInt), []values.Value{values.NewInt(int64(want[0])), values.NewInt(int64(want[1]))}).Equal(got.Array()) {
		t.Errorf("expected %v, got %v", want, got.Array())
	}
}

func TestMatchRegexpString(t *testing.T) {
	fluxFunc := SpecialFns["matchRegexpString"]
	re := regexp.MustCompile(`(gopher){2}`)
	r := values.NewRegexp(re)
	v := values.NewString("gophergophergopher")
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"r": values.NewRegexp(r.Regexp()), "v": values.NewString(v.Str())})
	want := r.Regexp().MatchString(v.Str())
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	got, err := fluxFunc.Call(ctx, fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Bool() {
		t.Errorf("expected %v, got %v", want, got.Str())
	}
}

func TestMatchRegexpStringNullV(t *testing.T) {
	fluxFunc := SpecialFns["matchRegexpString"]
	re := regexp.MustCompile(`(gopher){2}`)
	r := values.NewRegexp(re)
	vStrNullV := values.NewNull(semantic.BasicString)
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"r": values.NewRegexp(r.Regexp()), "v": vStrNullV})
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	_, err := fluxFunc.Call(ctx, fluxArg)
	wantErr := errors.New(codes.Invalid, "cannot execute function containing argument r of type regexp value (gopher){2} and argument v of type string value <nil>")
	if !cmp.Equal(wantErr, err) {
		t.Errorf("input %s: expected %v, got %v", vStrNullV, wantErr, err)
	}
}

func TestReplaceAllString(t *testing.T) {
	fluxFunc := SpecialFns["replaceAllString"]
	re := regexp.MustCompile(`a(x*)b`)
	r := values.NewRegexp(re)
	v := values.NewString("-ab-axxb-")
	tStr := values.NewString("T")
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"r": values.NewRegexp(r.Regexp()), "v": values.NewString(v.Str()), "t": values.NewString(tStr.Str())})
	want := re.ReplaceAllString(v.Str(), tStr.Str())
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	got, err := fluxFunc.Call(ctx, fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Str() {
		t.Errorf("input %s: expected %v, got %v", v, want, got.Str())
	}
}

func TestReplaceAllStringNullT(t *testing.T) {
	fluxFunc := SpecialFns["replaceAllString"]
	re := regexp.MustCompile(`a(x*)b`)
	r := values.NewRegexp(re)
	v := values.NewString("-ab-axxb-")
	tStrNullV := values.NewNull(semantic.BasicString)
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"r": values.NewRegexp(r.Regexp()), "v": values.NewString(v.Str()), "t": tStrNullV})
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	_, err := fluxFunc.Call(ctx, fluxArg)
	wantErr := errors.New(codes.Invalid, "cannot execute function containing argument r of type regexp value a(x*)b, argument v of type string value -ab-axxb-, and argument t of type string value <nil>")
	if !cmp.Equal(wantErr, err) {
		t.Errorf("input %s: expected %v, got %v", tStrNullV, wantErr, err)
	}
}

func TestSplitRegexp(t *testing.T) {
	fluxFunc := SpecialFns["splitRegexp"]
	re := regexp.MustCompile("a*")
	r := values.NewRegexp(re)
	v := values.NewString("abaabaccadaaae")
	i := values.NewInt(5)
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"r": values.NewRegexp(r.Regexp()), "v": values.NewString(v.Str()), "i": values.NewInt(i.Int())})
	want := r.Regexp().Split(v.Str(), int(i.Int()))
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	got, err := fluxFunc.Call(ctx, fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	arr := values.NewArray(semantic.NewArrayType(semantic.BasicString))
	for _, z := range want {
		arr.Append(values.NewString(z))
	}
	if !arr.Equal(got.Array()) {
		t.Errorf("expected %v, got %v", want, got.Array())
	}
}

func TestGetString(t *testing.T) {
	fluxFunc := SpecialFns["getString"]
	re := regexp.MustCompile("a*")
	r := values.NewRegexp(re)
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"r": values.NewRegexp(r.Regexp())})
	want := re.String()
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	got, err := fluxFunc.Call(ctx, fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Str() {
		t.Errorf("expected %v, got %v", want, got.Str())
	}
}

func TestGetStringNullR(t *testing.T) {
	fluxFunc := SpecialFns["getString"]
	regexpNullV := values.NewNull(semantic.BasicRegexp)
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"r": regexpNullV})
	wantErr := errors.New(codes.Invalid, "cannot execute function containing argument r of type regexp value <nil>")
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()
	_, err := fluxFunc.Call(ctx, fluxArg)
	if !cmp.Equal(wantErr, err) {
		t.Errorf("input %s: expected %v, got %v", regexpNullV, wantErr, err)
	}
}
