package regexp

import (
	"github.com/influxdata/flux/values"
	"regexp"
	"testing"
)

func TestCompileRegexp(t *testing.T) {
	fluxFunc := SpecialFns["compile"]
	v := values.NewString("alpha32")
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(v.Str())})
	want, _ := regexp.Compile(v.Str())
	realWant := values.NewRegexp(want)
	got, err := fluxFunc.Call(fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if realWant.Regexp().String() != got.Regexp().String() {
		t.Errorf("input %s: expected %v, got %v", v, want, got.Regexp())
	}
}
