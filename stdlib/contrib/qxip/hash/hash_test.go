package hash

import (
	//	"errors"
	"testing"

	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/values"
)

func Test_Sha256(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      string
		wantNull  bool
		expectErr error
	}{
		{
			name: "sha256(v:string)",
			v:    "Hello, world!",
			want: "315f5bdb76d078c43b8ac0064e4a0164612b1fce77c869345bfc94c75894edd3",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := interpreter.NewArguments(values.NewObjectWithValues(myMap))
			got, err := sha256(args)
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

func Test_Xxhash64(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      string
		wantNull  bool
		expectErr error
	}{
		{
			name: "xxhash64(v:string)",
			v:    "Hello, world!",
			want: "17691043854468224118",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := interpreter.NewArguments(values.NewObjectWithValues(myMap))
			got, err := xxhash64(args)
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

func Test_Cityhash64(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      string
		wantNull  bool
		expectErr error
	}{
		{
			name: "cityhash64(v:string)",
			v:    "Hello, world!",
			want: "2359500134450972198",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := interpreter.NewArguments(values.NewObjectWithValues(myMap))
			got, err := cityhash64(args)
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
