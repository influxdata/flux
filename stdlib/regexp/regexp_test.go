package regexp

import (
	"context"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"regexp"
	"testing"

	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func TestCompile(t *testing.T) {
	fluxFunc := SpecialFns["compile"]
	v := values.NewString("alpha32")
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(v.Str())})
	want, _ := regexp.Compile(v.Str())
	realWant := values.NewRegexp(want)
	got, err := fluxFunc.Call(context.Background(), dependenciestest.Default(), fluxArg)
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
	got, err := fluxFunc.Call(context.Background(), dependenciestest.Default(), fluxArg)
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
	got, err := fluxFunc.Call(context.Background(), dependenciestest.Default(), fluxArg)
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
	got, err := fluxFunc.Call(context.Background(), dependenciestest.Default(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if !values.NewArrayWithBacking(semantic.Int, []values.Value{values.NewInt(int64(want[0])), values.NewInt(int64(want[1]))}).Equal(got.Array()) {
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
	got, err := fluxFunc.Call(context.Background(), dependenciestest.Default(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Bool() {
		t.Errorf("expected %v, got %v", want, got.Str())
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
	got, err := fluxFunc.Call(context.Background(), dependenciestest.Default(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Str() {
		t.Errorf("input %s: expected %v, got %v", v, want, got.Str())
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
	got, err := fluxFunc.Call(context.Background(), dependenciestest.Default(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	arr := values.NewArray(semantic.String)
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
	got, err := fluxFunc.Call(context.Background(), dependenciestest.Default(), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Str() {
		t.Errorf("expected %v, got %v", want, got.Str())
	}
}
