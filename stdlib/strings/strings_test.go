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
			trim := generateDualArgStringFunction("trim", []string{stringArg, cutset}, strings.Trim)
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

func TestTrimPrefix(t *testing.T) {
	testCases := []struct {
		name   string
		v      string
		prefix string
		want   string
	}{
		{
			name:   "String with prefix",
			v:      "prefix_test",
			prefix: "prefix",
			want:   "_test",
		},
		{
			name:   "String without prefix",
			v:      "prefi_test",
			prefix: "prefix",
			want:   "prefi_test",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			trimPrefix := generateDualArgStringFunction("trimPrefix", []string{stringArg, prefix}, strings.TrimPrefix)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "prefix": values.NewString(tc.prefix)})
			result, err := trimPrefix.Call(testCase)
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

func TestTrimSuffix(t *testing.T) {
	testCases := []struct {
		name   string
		v      string
		suffix string
		want   string
	}{
		{
			name:   "String with suffix",
			v:      "test_suffix",
			suffix: "suffix",
			want:   "test_",
		},
		{
			name:   "String without suffix",
			v:      "test_suffi",
			suffix: "suffix",
			want:   "test_suffi",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			trimSuffix := generateDualArgStringFunction("trimSuffix", []string{stringArg, suffix}, strings.TrimSuffix)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "suffix": values.NewString(tc.suffix)})
			result, err := trimSuffix.Call(testCase)
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
			trimSpace := generateSingleArgStringFunction("trimSpace", strings.TrimSpace)
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
			title := generateSingleArgStringFunction("title", strings.Title)
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
			toUpper := generateSingleArgStringFunction("toUpper", strings.ToUpper)
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
			toLower := generateSingleArgStringFunction("toLower", strings.ToLower)
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
