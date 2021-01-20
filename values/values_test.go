package values_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func TestNew(t *testing.T) {
	for _, tt := range []struct {
		v    interface{}
		want values.Value
	}{
		{v: "abc", want: values.NewString("abc")},
		{v: int64(4), want: values.NewInt(4)},
		{v: uint64(4), want: values.NewUInt(4)},
		{v: float64(6.0), want: values.NewFloat(6.0)},
		{v: true, want: values.NewBool(true)},
		{v: values.Time(1000), want: values.NewTime(values.Time(1000))},
		{v: values.ConvertDurationNsecs(1), want: values.NewDuration(values.ConvertDurationNsecs(1))},
		{v: regexp.MustCompile(`.+`), want: values.NewRegexp(regexp.MustCompile(`.+`))},
	} {
		t.Run(fmt.Sprint(tt.want.Type()), func(t *testing.T) {
			if want, got := tt.want, values.New(tt.v); !want.Equal(got) {
				t.Fatalf("unexpected value -want/+got\n\t- %s\n\t+ %s", want, got)
			}
		})
	}
}

func TestNewNull(t *testing.T) {
	v := values.NewNull(semantic.BasicString)
	if want, got := true, v.IsNull(); want != got {
		t.Fatalf("unexpected value -want/+got\n\t- %v\n\t+ %v", want, got)
	}
}

func TestUnexpectedKind(t *testing.T) {
	err := values.UnexpectedKind(semantic.String, semantic.Float)
	if want, got := "unexpected kind: got \"string\" expected \"float\", trace:", err.Error(); !strings.HasPrefix(got, want) {
		t.Fatalf("unexpected error -want prefix/+got\n\t- %q\n\t+ %q", want, got)
	}

	// Ensure that it can be read as a *flux.Error.
	var ferr *flux.Error
	if !errors.As(err, &ferr) {
		t.Fatal("could not read unexpected kind error as a flux error")
	}

	if want, got := codes.Internal, ferr.Code; want != got {
		t.Fatalf("unexpected code -want/+got:\n\t- %s\n\t+ %s", want, got)
	}
}

// result stores results from the benchmark at the package level.
// Assigning to a global value prevents the optimizer from removing
// the assignment as it cannot determine whether the value will
// be read or not.
var result struct {
	Str      string
	Bytes    []byte
	Int      int64
	Uint     uint64
	Float    float64
	Bool     bool
	Time     values.Time
	Duration values.Duration
}

func BenchmarkValue_Str(b *testing.B) {
	v := values.NewString("abc")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result.Str = v.Str()
	}
}

func BenchmarkValue_Bytes(b *testing.B) {
	v := values.NewBytes([]byte("abc"))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result.Bytes = v.Bytes()
	}
}

func BenchmarkValue_Int(b *testing.B) {
	// Numbers 0-255 do not cause an allocation because
	// of an optimization in go 1.15. We want to avoid this
	// optimization to avoid skewing the benchmark.
	v := values.NewInt(int64(rand.Intn(1000) + 256))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result.Int = v.Int()
	}
}

func BenchmarkValue_UInt(b *testing.B) {
	v := values.NewUInt(rand.Uint64())
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result.Uint = v.UInt()
	}
}

func BenchmarkValue_Float(b *testing.B) {
	v := values.NewFloat(rand.Float64())
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result.Float = v.Float()
	}
}

func BenchmarkValue_Bool(b *testing.B) {
	v := values.NewBool(true)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result.Bool = v.Bool()
	}
}
