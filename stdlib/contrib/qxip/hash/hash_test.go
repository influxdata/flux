package hash

import (
	//	"errors"
	"testing"

	"github.com/InfluxCommunity/flux/interpreter"
	"github.com/InfluxCommunity/flux/values"
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

func Test_Sha1(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      string
		wantNull  bool
		expectErr error
	}{
		{
			name: "sha1(v:string)",
			v:    "Hello, world!",
			want: "943a702d06f34599aee1f8da8ef9f7296031d699",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := interpreter.NewArguments(values.NewObjectWithValues(myMap))
			got, err := sha1(args)
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

func Test_Base64(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      string
		wantNull  bool
		expectErr error
	}{
		{
			name: "b64(v:string)",
			v:    "Hello, world!",
			want: "SGVsbG8sIHdvcmxkIQ==",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := interpreter.NewArguments(values.NewObjectWithValues(myMap))
			got, err := b64(args)
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

func Test_MD5(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		want      string
		wantNull  bool
		expectErr error
	}{
		{
			name: "md5(v:string)",
			v:    "Hello, world!",
			want: "6cd3556deb0da54bca060b4c39479839",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
			}
			args := interpreter.NewArguments(values.NewObjectWithValues(myMap))
			got, err := md5(args)
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

func Test_HMac(t *testing.T) {
	testCases := []struct {
		name      string
		v         interface{}
		k         interface{}
		want      string
		wantNull  bool
		expectErr error
	}{
		{
			name: "hmac(v:string, k: key)",
			v:    "helloworld",
			k:    "123456",
			want: "75B5ueLnnGepYvh+KoevTzXCrjc=",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			myMap := map[string]values.Value{
				"v": values.New(tc.v),
				"k": values.New(tc.k),
			}
			args := interpreter.NewArguments(values.NewObjectWithValues(myMap))
			got, err := hmac(args)
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
