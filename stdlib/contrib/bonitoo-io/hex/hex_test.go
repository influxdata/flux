package hex

import (
	"errors"
	"testing"

	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/values"
)

func Test_String(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      string
		wantNull  bool
		expectErr error
	}{
		{
			name: "string(v:int64)",
			v:    int64(541),
			want: "21d",
		},
		{
			name: "string(v:uint64)",
			v:    uint64(501),
			want: "1f5",
		},
		{
			name: "string(v:float64)",
			v:    float64(653.28),
			want: "653.28",
		},
		{
			name: "string(v:bool/true)",
			v:    bool(true),
			want: "true",
		},
		{
			name: "string(v:bool/false)",
			v:    bool(false),
			want: "false",
		},
		{
			name: "string(v:time)",
			v:    values.Time(1136239445999999999),
			want: "2006-01-02T22:04:05.999999999Z",
		},
		{
			name: "string(v:duration)",
			v:    values.ConvertDurationNsecs(184000000000),
			want: "3m4s",
		},
		{
			name: "string(v:byte[1])",
			v:    []byte{120},
			want: "78",
		},
		{
			name: "string(v:byte[2])",
			v:    []byte{194, 1},
			want: "c201",
		},
		{
			name: "string(v:int64-negative)",
			v:    int64(-541),
			want: "-21d",
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
			args := interpreter.NewArguments(values.NewObjectWithValues(myMap))
			got, err := String(args)
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

func Test_Int(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      int64
		wantNull  bool
		expectErr error
	}{
		{
			name: "int64(v:hexString)",
			v:    "21d",
			want: int64(541),
		},
		{
			name: "int64(v:hexString/negative)",
			v:    "-21d",
			want: int64(-541),
		},
		{
			name: "int64(v:hexString/positive)",
			v:    "+21d",
			want: int64(541),
		},
		{
			name: "int64(v:0xHexString)",
			v:    "0x21d",
			want: int64(541),
		},
		{
			name: "int64(v:0xHexString/negative)",
			v:    "-0x21d",
			want: int64(-541),
		},
		{
			name: "int64(v:0xHexString/positive)",
			v:    "+0x21d",
			want: int64(541),
		},
		{
			name:      "int64(v:uint64)",
			v:         uint64(123),
			expectErr: errors.New("hex cannot convert uint to int"),
		},
		{
			name:      "int64(v:float64)",
			v:         float64(-728),
			expectErr: errors.New("hex cannot convert float to int"),
		},
		{
			name:      "int64(v:boolean/true)",
			v:         true,
			expectErr: errors.New("hex cannot convert bool to int"),
		},
		{
			name:      "int64(v:boolean/false)",
			v:         false,
			expectErr: errors.New("hex cannot convert bool to int"),
		},
		{
			name:      "int64(v:time)",
			v:         values.Time(1136239445999999999),
			expectErr: errors.New("hex cannot convert time to int"),
		},
		{
			name:      "int64(v:duration)",
			v:         values.ConvertDurationNsecs(123456789),
			expectErr: errors.New("hex cannot convert duration to int"),
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
			args := interpreter.NewArguments(values.NewObjectWithValues(myMap))
			got, err := Int(args)
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

func Test_UInt(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      uint64
		wantNull  bool
		expectErr error
	}{
		{
			name: "uint64(v:hexString)",
			v:    "21d",
			want: uint64(541),
		},
		{
			name: "uint64(v:0xhexString)",
			v:    "0x21d",
			want: uint64(541),
		},
		{
			name:      "uint64(v:int64)",
			v:         int64(123),
			expectErr: errors.New("hex cannot convert int to uint"),
		},
		{
			name:      "uint64(v:float64)",
			v:         float64(728),
			expectErr: errors.New("hex cannot convert float to uint"),
		},
		{
			name:      "uint64(v:boolean/true)",
			v:         true,
			expectErr: errors.New("hex cannot convert bool to uint"),
		},
		{
			name:      "uint64(v:boolean/false)",
			v:         false,
			expectErr: errors.New("hex cannot convert bool to uint"),
		},
		{
			name:      "uint64(v:time)",
			v:         values.Time(1136239445999999999),
			expectErr: errors.New("hex cannot convert time to uint"),
		},
		{
			name:      "uint64(v:duration)",
			v:         values.ConvertDurationNsecs(123456789),
			expectErr: errors.New("hex cannot convert duration to uint"),
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
			args := interpreter.NewArguments(values.NewObjectWithValues(myMap))
			got, err := UInt(args)
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

func Test_Bytes(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      []byte
		wantNull  bool
		expectErr error
	}{
		{
			name: "bytes(v:hexString)",
			v:    "6869",
			want: []byte("hi"),
		},
		{
			name:      "bytes(v:invalidHexString)",
			v:         "6",
			expectErr: errors.New("cannot convert string \"6\" to bytes due to hex decoding error: encoding/hex: odd length hex string"),
		},
		{
			name:      "bytes(v:invalidType)",
			v:         true,
			expectErr: errors.New("hex cannot convert bool to bytes"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := interpreter.NewArguments(values.NewObjectWithValues(myMap))
			got, err := Bytes(args)
			if err != nil {
				if tc.expectErr == nil {
					t.Errorf("unexpected error - want: <nil>, got: %s", err.Error())
				} else if want, got := tc.expectErr.Error(), err.Error(); got != want {
					t.Errorf("unexpected error - want: %s, got: %s", want, got)
				}
				return
			}
			if !tc.wantNull {
				want := values.NewBytes(tc.want)
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
