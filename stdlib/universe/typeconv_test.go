package universe

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/apache/arrow/go/v7/arrow/array"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

func TestTypeconv_String(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      string
		wantNull  bool
		expectErr error
	}{
		{
			name: "string(v:1)",
			v:    int64(541),
			want: "541",
		},
		{
			name: "string(v:2)",
			v:    uint64(501),
			want: "501",
		},
		{
			name: "string(v:3)",
			v:    float64(653.28),
			want: "653.28",
		},
		{
			name: "string(v:4)",
			v:    bool(true),
			want: "true",
		},
		{
			name: "string(v:5)",
			v:    bool(false),
			want: "false",
		},
		{
			name: "string(v:6)",
			v:    values.Time(1136239445999999999),
			want: "2006-01-02T22:04:05.999999999Z",
		},
		{
			name: "string(v:7)",
			v:    values.ConvertDurationNsecs(184000000000),
			want: "3m4s",
		},
		{
			name: "string(v:8)",
			v:    []byte{120},
			want: "x",
		},
		{
			name: "string(v:9)",
			v:    []byte{194, 167},
			want: "§",
		},
		{
			name: "string(v:10)",
			v:    int64(-541),
			want: "-541",
		},
		{
			name:     "string(v:nil)",
			v:        nil,
			wantNull: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := values.NewObjectWithValues(myMap)
			c := stringConv
			ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
			defer deps.Finish()
			got, err := c.Call(ctx, args)
			if err != nil {
				if tc.expectErr == nil {
					t.Errorf("unexpected error - want: <nil>, got: %s", err.Error())
				} else if want, got := tc.expectErr.Error(), err.Error(); got != want {
					t.Errorf("unexpected error - want: %s, got: %s", want, got)
				}
				return
			}
			if !tc.wantNull {
				want := values.NewString(tc.want)
				if !got.Equal(want) {
					t.Errorf("Wanted: %s, got: %v", want, got)
				}
			} else {
				if !got.IsNull() {
					t.Errorf("Wanted: %v, got: %v", values.Null, got)
				}
			}
		})
	}
}

func TestTypeconv_Int(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      int64
		wantNull  bool
		expectErr error
	}{
		{
			name: "int64(v:1)",
			v:    "4615",
			want: int64(4615),
		},
		{
			name: "int64(v:2)",
			v:    uint64(123),
			want: int64(123),
		},
		{
			name: "int64(v:3)",
			v:    float64(-728),
			want: int64(-728),
		},
		{
			name: "int64(v:4)",
			v:    true,
			want: int64(1),
		},
		{
			name: "int64(v:4)",
			v:    false,
			want: int64(0),
		},
		{
			name: "int64(v:5)",
			v:    values.Time(1136239445999999999),
			want: int64(1136239445999999999),
		},
		{
			name: "int64(v:6)",
			v:    values.ConvertDurationNsecs(123456789),
			want: int64(123456789),
		},
		{
			name:      "int64(error)",
			v:         "notanumber",
			want:      0,
			expectErr: errors.New("cannot convert string \"notanumber\" to int due to invalid syntax"),
		},
		{
			name:     "int64(v:nil)",
			v:        nil,
			wantNull: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := values.NewObjectWithValues(myMap)
			c := intConv
			ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
			defer deps.Finish()
			got, err := c.Call(ctx, args)
			if err != nil {
				if tc.expectErr == nil {
					t.Errorf("unexpected error - want: <nil>, got: %s", err.Error())
				} else if want, got := tc.expectErr.Error(), err.Error(); got != want {
					t.Errorf("unexpected error - want: %s, got: %s", want, got)
				}
				return
			}
			if !tc.wantNull {
				want := values.NewInt(tc.want)
				if !got.Equal(want) {
					t.Errorf("Wanted: %s, got: %v", want, got)
				}
			} else {
				if !got.IsNull() {
					t.Errorf("Wanted: %v, got: %v", values.Null, got)
				}
			}
		})
	}
}

func TestTypeconv_UInt(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      uint64
		wantNull  bool
		expectErr error
	}{
		{
			name: "uint64(v:1)",
			v:    "4615",
			want: uint64(4615),
		},
		{
			name: "uint64(v:2)",
			v:    int64(123),
			want: uint64(123),
		},
		{
			name: "uint64(v:3)",
			v:    float64(728),
			want: uint64(728),
		},
		{
			name: "uint64(v:4)",
			v:    true,
			want: uint64(1),
		},
		{
			name: "uint64(v:4)",
			v:    false,
			want: uint64(0),
		},
		{
			name: "uint64(v:5)",
			v:    values.Time(1136239445999999999),
			want: uint64(1136239445999999999),
		},
		{
			name: "uint64(v:6)",
			v:    values.ConvertDurationNsecs(123456789),
			want: uint64(123456789),
		},
		{
			name:      "uint64(error)",
			v:         "NaN",
			want:      0,
			expectErr: errors.New("cannot convert string \"NaN\" to uint due to invalid syntax"),
		},
		{
			name:     "uint64(v:nil)",
			v:        nil,
			wantNull: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := values.NewObjectWithValues(myMap)
			c := uintConv
			ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
			defer deps.Finish()
			got, err := c.Call(ctx, args)
			if err != nil {
				if tc.expectErr == nil {
					t.Errorf("unexpected error - want: <nil>, got: %s", err.Error())
				} else if want, got := tc.expectErr.Error(), err.Error(); got != want {
					t.Errorf("unexpected error - want: %s, got: %s", want, got)
				}
				return
			}
			if !tc.wantNull {
				want := values.NewUInt(tc.want)
				if !got.Equal(want) {
					t.Errorf("Wanted: %s, got: %v", want, got)
				}
			} else {
				if !got.IsNull() {
					t.Errorf("Wanted: %v, got: %v", values.Null, got)
				}
			}
		})
	}
}

func TestTypeconv_Bool(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      bool
		wantNull  bool
		expectErr error
	}{
		{
			name: "bool(v:1)",
			v:    int64(1),
			want: true,
		},
		{
			name: "bool(v:1)",
			v:    int64(0),
			want: false,
		},
		{
			name: "bool(v:2)",
			v:    "true",
			want: true,
		},
		{
			name: "bool(v:2)",
			v:    "false",
			want: false,
		},
		{
			name: "bool(v:3)",
			v:    uint64(1),
			want: true,
		},
		{
			name: "bool(v:3)",
			v:    uint64(0),
			want: false,
		},
		{
			name: "bool(v:4)",
			v:    float64(1),
			want: true,
		},
		{
			name: "bool(v:4)",
			v:    float64(0),
			want: false,
		},
		{
			name:      "bool(error)",
			v:         "asdf",
			want:      false,
			expectErr: errors.New("cannot convert string \"asdf\" to bool"),
		},
		{
			name:     "bool(v:nil)",
			v:        nil,
			wantNull: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := values.NewObjectWithValues(myMap)
			c := boolConv
			ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
			defer deps.Finish()
			got, err := c.Call(ctx, args)
			if err != nil {
				if tc.expectErr == nil {
					t.Errorf("unexpected error - want: <nil>, got: %s", err.Error())
				} else if want, got := tc.expectErr.Error(), err.Error(); got != want {
					t.Errorf("unexpected error - want: %s, got: %s", want, got)
				}
				return
			}
			if !tc.wantNull {
				want := values.NewBool(tc.want)
				if !got.Equal(want) {
					t.Errorf("Wanted: %s, got: %v", want, got)
				}
			} else {
				if !got.IsNull() {
					t.Errorf("Wanted: %v, got: %v", values.Null, got)
				}
			}
		})
	}
}

func TestTypeconv_VectorsToVectorizedFloat(t *testing.T) {
	alloc := memory.NewResourceAllocator(nil)

	testCases := []struct {
		name       string
		vSlice     []interface{}
		vRepeat    interface{}
		want       []interface{}
		wantNull   bool
		wantRepeat float64
		expectErr  error
	}{
		{
			name: "vectoredFloat to vectoredFloat",
			vSlice: []interface{}{
				float64(4615.123),
				float64(0.1),
				nil,
			},
			want: []interface{}{
				float64(4615.123),
				float64(0.1),
				nil,
			},
		},
		{
			name: "vectoredInt to vectoredFloat",
			vSlice: []interface{}{
				int64(123),
				int64(0),
				nil,
			},
			want: []interface{}{
				float64(123.0),
				float64(0.0),
				nil,
			},
		},
		{
			name: "vectoredString to vectoredFloat",
			vSlice: []interface{}{
				"123.456",
				"0.0",
				"NaN",
			},
			want: []interface{}{
				float64(123.456),
				float64(0.0),
				float64(math.NaN()),
			},
		},
		{
			name: "vectoredUint to vectoredFloat",
			vSlice: []interface{}{
				uint64(123),
				uint64(0),
				nil,
			},
			want: []interface{}{
				float64(123),
				float64(0),
				nil,
			},
		},
		{
			name: "vectoredBool to vectoredFloat",
			vSlice: []interface{}{
				true,
				false,
				nil,
			},
			want: []interface{}{
				float64(1),
				float64(0),
				nil,
			},
		},
		{
			name:       "RepeatVector(Int) to vectoredFloat",
			vRepeat:    int64(1234),
			wantRepeat: float64(1234.0),
		},
		{
			name:       "RepeatVector(Bool) to vectoredFloat",
			vRepeat:    true,
			wantRepeat: float64(1.0),
		},
		{
			name:       "RepeatVector(Uint) to vectoredFloat",
			vRepeat:    uint64(1234),
			wantRepeat: float64(1234.0),
		},
		{
			name:       "RepeatVector(String) to vectoredFloat",
			vRepeat:    "1234.567",
			wantRepeat: float64(1234.567),
		},
		{
			name:       "RepeatVector(Float) to vectoredFloat",
			vRepeat:    float64(1234.567),
			wantRepeat: float64(1234.567),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v values.Value

			if tc.vRepeat == nil {
				v = values.NewVectorFromElements(alloc, tc.vSlice...)
			} else {
				v = values.NewVectorRepeatValue(values.New(tc.vRepeat))
			}

			myMap := map[string]values.Value{
				"v": v,
			}

			args := values.NewObjectWithValues(myMap)
			c := vectorizedFloatConv
			ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
			defer deps.Finish()
			gotVal, err := c.Call(memory.WithAllocator(ctx, alloc), args)
			if err != nil {
				if tc.expectErr == nil {
					t.Errorf("unexpected error - want: <nil>, got: %s", err.Error())
				} else if want, got := tc.expectErr.Error(), err.Error(); got != want {
					t.Errorf("unexpected error - want: %s, got: %s", want, got)
				}
				return
			}

			got := gotVal.Vector()
			if tc.vRepeat != nil {
				if !got.IsRepeat() {
					t.Error("Unexpected error: wanted float vector repeat")
				}

				gotRepeatVal := got.(*values.VectorRepeatValue).Value()
				wantValue := values.New(tc.wantRepeat)

				if !gotRepeatVal.Equal(wantValue) {
					t.Errorf("Wanted: vector repeat: %v, got: %v", wantValue, gotRepeatVal)
				}
			} else if !tc.wantNull {
				got := got.Arr().(*array.Float64)
				want := values.NewVectorFromElements(alloc, tc.want...).Arr().(*array.Float64)

				if got.Len() != want.Len() {
					t.Errorf("Unexpected error want: count(%v), got: count(%v)", want.Len(), got.Len())
				}

				for i := 0; i < want.Len(); i++ {
					if want.Value(i) != got.Value(i) && !(math.IsNaN(want.Value(i)) && math.IsNaN(got.Value(i))) {
						t.Errorf("Wanted v[%v] => %v, got: %v", i, want.Value(i), got.Value(i))
					}
				}
			} else {
				if !got.IsNull() {
					t.Errorf("Wanted: %v, got: %v", values.Null, got)
				}
			}
		})
	}
}

func TestTypeconv_Float(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      float64
		wantNull  bool
		expectErr error
	}{
		{
			name: "float64(v:1)",
			v:    "4615.123",
			want: float64(4615.123),
		},
		{
			name: "float64(v:2)",
			v:    uint64(123),
			want: float64(123),
		},
		{
			name: "float64(v:3)",
			v:    float64(728),
			want: float64(728),
		},
		{
			name: "float64(v:4)",
			v:    true,
			want: float64(1),
		},
		{
			name: "float64(v:5)",
			v:    false,
			want: float64(0),
		},
		{
			name: "float64(v:6)",
			v:    int64(-753),
			want: float64(-753),
		},
		{
			name: "float64(v:7)",
			v:    "+Inf",
			want: float64(math.Inf(+1)),
		},
		{
			name: "float64(v:8)",
			v:    "-Inf",
			want: float64(math.Inf(-1)),
		},
		{
			name:      "float64(v:8)",
			v:         "NaN",
			want:      float64(math.NaN()),
			expectErr: errors.New("Test failed, got: NaN, want: NaN"),
		},
		{
			name:      "float(error)",
			v:         "ThisIsNotANumber",
			want:      float64(0),
			expectErr: errors.New("cannot convert string \"ThisIsNotANumber\" to float due to invalid syntax"),
		},
		{
			name:     "float(v:nil)",
			v:        nil,
			wantNull: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := values.NewObjectWithValues(myMap)
			c := floatConv
			ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
			defer deps.Finish()
			got, err := c.Call(ctx, args)
			if err != nil {
				if tc.expectErr == nil {
					t.Errorf("unexpected error - want: <nil>, got: %s", err.Error())
				} else if want, got := tc.expectErr.Error(), err.Error(); got != want {
					t.Errorf("unexpected error - want: %s, got: %s", want, got)
				}
				return
			}
			if !tc.wantNull {
				want := values.NewFloat(tc.want)
				if !got.Equal(want) {
					// NaN == NaN evaluates to false, so need a special check
					if math.IsNaN(tc.want) && math.IsNaN(got.Float()) {
						return
					}
					t.Errorf("Wanted: %s, got: %v", want, got)
				}
			} else {
				if !got.IsNull() {
					t.Errorf("Wanted: %v, got: %v", values.Null, got)
				}
			}
		})
	}
}

func TestTypeconv_Time(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      values.Time
		wantNull  bool
		expectErr error
	}{
		{
			name: "time(v:1)",
			v:    int64(1136239445),
			want: values.Time(1136239445),
		},
		{
			name: "time(v:2)",
			v:    uint64(1136239445),
			want: values.Time(1136239445),
		},
		{
			name: "time(v:3)",
			v:    values.Time(1136239445),
			want: values.Time(1136239445),
		},
		{
			name: "time(v:4)",
			v:    "2006-01-02T22:04:05.999999999Z",
			want: values.Time(1136239445999999999),
		},
		{
			name:      "time(error)",
			v:         "NotATime",
			want:      values.Time(0),
			expectErr: errors.New("cannot convert string \"NotATime\" to time due to invalid syntax: parsing time \"NotATime\" as \"2006-01-02T15:04:05.999999999Z07:00\": cannot parse \"NotATime\" as \"2006\""),
		},
		{
			name:     "time(v:nil)",
			v:        nil,
			wantNull: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := values.NewObjectWithValues(myMap)
			c := timeConv
			ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
			defer deps.Finish()
			got, err := c.Call(ctx, args)
			if err != nil {
				if tc.expectErr == nil {
					t.Errorf("unexpected error - want: <nil>, got: %s", err.Error())
				} else if want, got := tc.expectErr.Error(), err.Error(); got != want {
					t.Errorf("unexpected error - want: %s, got: %s", want, got)
				}
				return
			}
			if !tc.wantNull {
				want := values.NewTime(tc.want)
				if !got.Equal(want) {
					t.Errorf("Wanted: %s, got: %v", want, got)
				}
			} else {
				if !got.IsNull() {
					t.Errorf("Wanted: %v, got: %v", values.Null, got)
				}
			}
		})
	}
}

func TestTypeconv_Duration(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      values.Duration
		wantNull  bool
		expectErr error
	}{
		{
			name: "duration(v:1)",
			v:    int64(123456789),
			want: values.ConvertDurationNsecs(123456789),
		},
		{
			name: "duration(v:2)",
			v:    uint64(123456789),
			want: values.ConvertDurationNsecs(123456789),
		},
		{
			name: "duration(v:3)",
			v:    "4s2ns",
			want: values.ConvertDurationNsecs(4000000002),
		},
		{
			name: "duration(v:4s2ns)",
			v:    values.ConvertDurationNsecs(4000000002),
			want: values.ConvertDurationNsecs(4000000002),
		},
		{
			name:      "duration(error)",
			v:         "not_a_duration",
			want:      values.ConvertDurationNsecs(0),
			expectErr: errors.New("cannot convert string \"not_a_duration\" to duration due to invalid syntax"),
		},
		{
			name:     "duration(v:nil)",
			v:        nil,
			wantNull: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := values.NewObjectWithValues(myMap)
			c := durationConv
			ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
			defer deps.Finish()
			got, err := c.Call(ctx, args)
			if err != nil {
				if tc.expectErr == nil {
					t.Errorf("unexpected error - want: <nil>, got: %s", err.Error())
				} else if want, got := tc.expectErr.Error(), err.Error(); got != want {
					t.Errorf("unexpected error - want: %s, got: %s", want, got)
				}
				return
			}
			if !tc.wantNull {
				want := values.NewDuration(tc.want)
				if !got.Equal(want) {
					t.Errorf("Wanted: %s, got: %v", want, got)
				}
			} else {
				if !got.IsNull() {
					t.Errorf("Wanted: %v, got: %v", values.Null, got)
				}
			}
		})
	}
}
