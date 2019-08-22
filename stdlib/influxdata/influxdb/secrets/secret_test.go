package secrets_test

import (
	"testing"

	"github.com/influxdata/flux/stdlib/influxdata/influxdb/secrets"
	"github.com/influxdata/flux/values"
)

func TestGet(t *testing.T) {
	for _, tt := range []struct {
		name string
		args map[string]values.Value
		want values.Value
		err  string
	}{
		{
			name: "normal",
			args: map[string]values.Value{
				"key": values.NewString("mykey"),
			},
			want: values.NewObjectWithValues(map[string]values.Value{
				"secretKey": values.NewString("mykey"),
			}),
		},
		{
			name: "missing argument",
			args: map[string]values.Value{},
			err:  "missing required keyword argument \"key\"",
		},
		{
			name: "wrong argument",
			args: map[string]values.Value{
				"key": values.NewInt(4),
			},
			err: "keyword argument \"key\" should be of kind string, but got int",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			args := values.NewObjectWithValues(tt.args)
			got, err := secrets.Get(args)
			if err != nil {
				if tt.err == "" {
					t.Fatalf("unexpected error: %s", err)
				} else if want, got := tt.err, err.Error(); want != got {
					t.Fatalf("unexpected error -want/+got:\n\t- %s\n\t+ %s", want, got)
				}
				return
			}

			if !tt.want.Equal(got) {
				t.Fatalf("unexpected value -want/+got:\n\t- %#v\n\t+ %#v", tt.want, got)
			}
		})
	}
}

func TestGetKeyFromValue(t *testing.T) {
	for _, tt := range []struct {
		name string
		v    values.Value
		want string
	}{
		{
			name: "valid secret key",
			v:    secrets.New("mykey"),
			want: "mykey",
		},
		{
			name: "invalid type",
			v:    values.NewString("mykey"),
		},
		{
			name: "missing key",
			v: values.NewObjectWithValues(map[string]values.Value{
				"wrongKey": values.NewString("mykey"),
			}),
		},
		{
			name: "too many properties",
			v: values.NewObjectWithValues(map[string]values.Value{
				"secretKey": values.NewString("mykey"),
				"extraKey":  values.NewString("myextrakey"),
			}),
		},
		{
			name: "invalid key type",
			v: values.NewObjectWithValues(map[string]values.Value{
				"secretKey": values.NewInt(4),
			}),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := secrets.GetKeyFromValue(tt.v)
			if !ok {
				if tt.want != "" {
					t.Fatal("expected value")
				}
				return
			}

			if want, got := tt.want, got; want != got {
				t.Fatalf("unexpected value -want/+got:\n\t- %q\n\t+ %q", tt.want, got)
			}
		})
	}
}
