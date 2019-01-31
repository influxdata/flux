package strings

import (
	"strings"
	"testing"

	"github.com/influxdata/flux/values"
)

func TestTrim(t *testing.T) {
	testCases := []struct {
		name   string
		v      string
		cutset string
		want   string
	}{
		{
			name:   "Leading and trailing dots",
			v:      "..Koala...",
			cutset: ".",
			want:   "Koala",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			trim := generateMultiArgStringFunction("trim", strings.Trim)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "cutset": values.NewString(tc.cutset)})
			result, err := trim.Call(testCase)
			res := result.Str()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %s, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestTrimSpace(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		want string
	}{
		{
			name: "Leading and trailing spaces",
			v:    "  Giraffe  ",
			want: "Giraffe",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			trimSpace := generateStringFunction("trimSpace", strings.TrimSpace)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v)})
			result, err := trimSpace.Call(testCase)
			res := result.Str()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %s, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestTitle(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		want string
	}{
		{
			name: "lower case string",
			v:    "a giraffe",
			want: "A Giraffe",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			title := generateStringFunction("title", strings.Title)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v)})
			result, err := title.Call(testCase)
			res := result.Str()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %s, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestToUpper(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		want string
	}{
		{
			name: "lower case string",
			v:    "koala",
			want: "KOALA",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			toUpper := generateStringFunction("toUpper", strings.ToUpper)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v)})
			result, err := toUpper.Call(testCase)
			res := result.Str()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %s, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestToLower(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		want string
	}{
		{
			name: "upper case string",
			v:    "KOALA",
			want: "koala",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			toLower := generateStringFunction("toLower", strings.ToLower)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v)})
			result, err := toLower.Call(testCase)
			res := result.Str()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %s, got: %s", tc.name, tc.want, result)
			}
		})

	}
}
