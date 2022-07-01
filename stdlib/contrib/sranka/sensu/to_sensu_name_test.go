package sensu_test

import (
	"strings"
	"testing"

	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/stdlib/contrib/sranka/sensu"
	"github.com/influxdata/flux/values"
)

func TestToSensuName(t *testing.T) {
	testCases := []struct {
		input string // also name of the test
		want  string
	}{
		{
			input: "simple",
			want:  "simple",
		},
		{
			input: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.-",
			want:  "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.-",
		},
		{
			input: "no space,colon;semicolon>gt<lt",
			want:  "no_space_colon_semicolon_gt_lt",
		},
		{
			input: "",
			want:  "_",
		},
	}

	for _, tc := range testCases {
		tc := tc
		name := tc.input
		if name == "" {
			name = "<empty>"
		}
		t.Run(name, func(t *testing.T) {
			v := tc.input
			want := values.NewString(tc.want)

			args := interpreter.NewArguments(values.NewObjectWithValues(
				map[string]values.Value{
					"v": values.NewString(v),
				}),
			)

			got, err := sensu.ToSensuName(args)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !want.Equal(got) {
				t.Fatalf("unexpected value -want/+got:\n\t- %#v\n\t+ %#v", want, got)
			}
		})
	}

	t.Run("missing required argument", func(t *testing.T) {
		args := interpreter.NewArguments(values.NewObjectWithValues(
			map[string]values.Value{}),
		)

		_, err := sensu.ToSensuName(args)
		if err == nil {
			t.Fatal("error expected, but none received")
		}
		if !strings.Contains(err.Error(), "missing required") {
			t.Fatal(err)
		}
	})
}
