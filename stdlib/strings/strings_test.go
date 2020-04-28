package strings

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode"

	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/semantic"
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
			trim := generateDualArgStringFunction("trim", []string{stringArgV, cutset}, strings.Trim)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "cutset": values.NewString(tc.cutset)})
			result, err := trim.Call(dependenciestest.Default().Inject(context.Background()), testCase)
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
			trimPrefix := generateDualArgStringFunction("trimPrefix", []string{stringArgV, prefix}, strings.TrimPrefix)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "prefix": values.NewString(tc.prefix)})
			result, err := trimPrefix.Call(dependenciestest.Default().Inject(context.Background()), testCase)
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
			trimSuffix := generateDualArgStringFunction("trimSuffix", []string{stringArgV, suffix}, strings.TrimSuffix)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "suffix": values.NewString(tc.suffix)})
			result, err := trimSuffix.Call(dependenciestest.Default().Inject(context.Background()), testCase)
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
			result, err := trimSpace.Call(dependenciestest.Default().Inject(context.Background()), testCase)
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
			result, err := title.Call(dependenciestest.Default().Inject(context.Background()), testCase)
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
			result, err := toUpper.Call(dependenciestest.Default().Inject(context.Background()), testCase)
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
			result, err := toLower.Call(dependenciestest.Default().Inject(context.Background()), testCase)
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

func TestTrimRight(t *testing.T) {
	testCases := []struct {
		name   string
		v      string
		cutset string
		want   string
	}{
		{
			name:   "Trailing dots",
			v:      "..Koala...",
			cutset: ".",
			want:   "..Koala",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			trim := generateDualArgStringFunction("trimRight", []string{stringArgV, cutset}, strings.TrimRight)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "cutset": values.NewString(tc.cutset)})
			result, err := trim.Call(dependenciestest.Default().Inject(context.Background()), testCase)
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

func TestTrimLeft(t *testing.T) {
	testCases := []struct {
		name   string
		v      string
		cutset string
		want   string
	}{
		{
			name:   "Trailing dots",
			v:      "..Koala...",
			cutset: ".",
			want:   "Koala...",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			trim := generateDualArgStringFunction("trimLeft", []string{stringArgV, cutset}, strings.TrimLeft)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "cutset": values.NewString(tc.cutset)})
			result, err := trim.Call(dependenciestest.Default().Inject(context.Background()), testCase)
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

func TestToTitle(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		want string
	}{
		{
			name: "lower case string",
			v:    "loud noises",
			want: "LOUD NOISES",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			title := generateSingleArgStringFunction("toTitle", strings.ToTitle)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v)})
			result, err := title.Call(dependenciestest.Default().Inject(context.Background()), testCase)
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

func TestHasSuffix(t *testing.T) {
	testCases := []struct {
		name   string
		v      string
		suffix string
		want   bool
	}{
		{
			name:   "String with suffix",
			v:      "test_suffix",
			suffix: "suffix",
			want:   true,
		},
		{
			name:   "String without suffix",
			v:      "test_suffi",
			suffix: "suffix",
			want:   false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			hasSuffix := generateDualArgStringFunctionReturnBool("hasSuffix", []string{stringArgV, suffix}, strings.HasSuffix)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "suffix": values.NewString(tc.suffix)})
			result, err := hasSuffix.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Bool()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %t, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestHasPrefix(t *testing.T) {
	testCases := []struct {
		name   string
		v      string
		prefix string
		want   bool
	}{
		{
			name:   "String with prefix",
			v:      "prefix_test",
			prefix: "prefix",
			want:   true,
		},
		{
			name:   "String without prefix",
			v:      "prefi_test",
			prefix: "prefix",
			want:   false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			hasPrefix := generateDualArgStringFunctionReturnBool("hasPrefix", []string{stringArgV, prefix}, strings.HasPrefix)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "prefix": values.NewString(tc.prefix)})
			result, err := hasPrefix.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Bool()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %t, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		name   string
		v      string
		substr string
		want   bool
	}{
		{
			name:   "Does contain substr",
			v:      "seafood",
			substr: "foo",
			want:   true,
		},
		{
			name:   "Does not contain substr",
			v:      "seafood",
			substr: "bar",
			want:   false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			containsStr := generateDualArgStringFunctionReturnBool("containsStr", []string{stringArgV, substr}, strings.Contains)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "substr": values.NewString(tc.substr)})
			result, err := containsStr.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Bool()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %t, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestContainsAny(t *testing.T) {
	testCases := []struct {
		name  string
		v     string
		chars string
		want  bool
	}{
		{
			name:  "Does containsAny",
			v:     "failure",
			chars: "u & i",
			want:  true,
		},
		{
			name:  "Does not containsAny",
			v:     "foo",
			chars: "",
			want:  false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			containsAny := generateDualArgStringFunctionReturnBool("containsAny", []string{stringArgV, chars}, strings.ContainsAny)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "chars": values.NewString(tc.chars)})
			result, err := containsAny.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Bool()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %t, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestEqualFold(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		t    string
		want bool
	}{
		{
			name: "Is Equal",
			v:    "Go",
			t:    "go",
			want: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			equalFold := generateDualArgStringFunctionReturnBool("equalFold", []string{stringArgV, stringArgT}, strings.EqualFold)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "t": values.NewString(tc.t)})
			result, err := equalFold.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Bool()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %t, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestCompare(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		t    string
		want int64
	}{
		{
			name: "a < b",
			v:    "a",
			t:    "b",
			want: -1,
		},
		{
			name: "a = a",
			v:    "a",
			t:    "a",
			want: 0,
		},
		{
			name: "b > a",
			v:    "b",
			t:    "a",
			want: 1,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			compare := generateDualArgStringFunctionReturnInt("compare", []string{stringArgV, stringArgT}, strings.Compare)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "t": values.NewString(tc.t)})
			result, err := compare.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Int()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestCount(t *testing.T) {
	testCases := []struct {
		name   string
		v      string
		substr string
		want   int64
	}{
		{
			name:   "countStr e's",
			v:      "cheese",
			substr: "e",
			want:   3,
		},
		{
			name:   "countStr nothing",
			v:      "five",
			substr: "",
			want:   5,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			countStr := generateDualArgStringFunctionReturnInt("countStr", []string{stringArgV, substr}, strings.Count)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "substr": values.NewString(tc.substr)})
			result, err := countStr.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Int()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestIndex(t *testing.T) {
	testCases := []struct {
		name   string
		v      string
		substr string
		want   int64
	}{
		{
			name:   "Exists",
			v:      "chicken",
			substr: "ken",
			want:   4,
		},
		{
			name:   "Does not exist",
			v:      "chicken",
			substr: "dmr",
			want:   -1,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			index := generateDualArgStringFunctionReturnInt("index", []string{stringArgV, substr}, strings.Index)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "substr": values.NewString(tc.substr)})
			result, err := index.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Int()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestIndexAny(t *testing.T) {
	testCases := []struct {
		name  string
		v     string
		chars string
		want  int64
	}{
		{
			name:  "Exists",
			v:     "chicken",
			chars: "aeiouy",
			want:  2,
		},
		{
			name:  "Does not exist",
			v:     "crwth",
			chars: "aeiouy",
			want:  -1,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			indexAny := generateDualArgStringFunctionReturnInt("indexAny", []string{stringArgV, chars}, strings.IndexAny)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "chars": values.NewString(tc.chars)})
			result, err := indexAny.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Int()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestLastIndex(t *testing.T) {
	testCases := []struct {
		name   string
		v      string
		substr string
		want   int64
	}{
		{
			name:   "Exists",
			v:      "go gopher",
			substr: "go",
			want:   3,
		},
		{
			name:   "Does not exist",
			v:      "go gopher",
			substr: "rodent",
			want:   -1,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lastIndex := generateDualArgStringFunctionReturnInt("lastIndex", []string{stringArgV, substr}, strings.LastIndex)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "substr": values.NewString(tc.substr)})
			result, err := lastIndex.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Int()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestLastIndexAny(t *testing.T) {
	testCases := []struct {
		name  string
		v     string
		chars string
		want  int64
	}{
		{
			name:  "Exists",
			v:     "go gopher",
			chars: "go",
			want:  4,
		},
		{
			name:  "Does exist",
			v:     "go gopher",
			chars: "rodent",
			want:  8,
		},
		{
			name:  "Fail",
			v:     "go gopher",
			chars: "fail",
			want:  -1,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lastIndexAny := generateDualArgStringFunctionReturnInt("lastIndexAny", []string{stringArgV, chars}, strings.LastIndexAny)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "chars": values.NewString(tc.chars)})
			result, err := lastIndexAny.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Int()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestIsDigit(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		want bool
	}{
		{
			name: "Is a digit",
			v:    "5",
			want: true,
		},
		{
			name: "Is not a digit",
			v:    "f",
			want: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			digit := generateUnicodeIsFunction("isDigit", unicode.IsDigit)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v)})
			result, err := digit.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Bool()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestIsLetter(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		want bool
	}{
		{
			name: "Is a letter",
			v:    "f",
			want: true,
		},
		{
			name: "Still a letter",
			v:    "F",
			want: true,
		},
		{
			name: "Is not a letter",
			v:    "5",
			want: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			is := generateUnicodeIsFunction("isLetter", unicode.IsLetter)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v)})
			result, err := is.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Bool()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestIsLower(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		want bool
	}{
		{
			name: "Is Lower",
			v:    "f",
			want: true,
		},
		{
			name: "Not letter",
			v:    "3",
			want: false,
		},
		{
			name: "Not lower",
			v:    "G",
			want: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			is := generateUnicodeIsFunction("isLower", unicode.IsLower)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v)})
			result, err := is.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Bool()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestIsUpper(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		want bool
	}{
		{
			name: "Is not Upper",
			v:    "f",
			want: false,
		},
		{
			name: "Not letter",
			v:    "3",
			want: false,
		},
		{
			name: "Upper",
			v:    "G",
			want: true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			is := generateUnicodeIsFunction("isUpper", unicode.IsUpper)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v)})
			result, err := is.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Bool()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestRepeat(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		i    int
		want string
	}{
		{
			name: "Banana - Ba",
			v:    "na",
			i:    2,
			want: "nana",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testValue := generateRepeat("repeat", []string{stringArgV, integer}, strings.Repeat)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "i": values.NewInt(int64(tc.i))})
			result, err := testValue.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Str()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestReplace(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		t    string
		u    string
		i    int
		want string
	}{
		{
			name: "Pig",
			v:    "oink oink oink",
			t:    "k",
			u:    "ky",
			i:    2,
			want: "oinky oinky oink",
		},
		{
			name: "Cow",
			v:    "oink oink oink",
			t:    "oink",
			u:    "moo",
			i:    -1,
			want: "moo moo moo",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testValue := generateReplace("replace", []string{stringArgV, stringArgT, stringArgU, integer}, strings.Replace)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v),
				"t": values.NewString(tc.t), "u": values.NewString(tc.u), "i": values.NewInt(int64(tc.i))})
			result, err := testValue.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Str()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestReplaceAll(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		t    string
		u    string
		want string
	}{
		{
			name: "Pig",
			v:    "oink oink oink",
			t:    "k",
			u:    "ky",
			want: "oinky oinky oinky",
		},
		{
			name: "Cow",
			v:    "oink oink oink",
			t:    "oink",
			u:    "moo",
			want: "moo moo moo",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testValue := generateReplaceAll("replaceAll", []string{stringArgV, stringArgT, stringArgU}, strings.ReplaceAll)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v),
				"t": values.NewString(tc.t), "u": values.NewString(tc.u)})
			result, err := testValue.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Str()

			if err != nil {
				t.Fatal(err)
			}

			if res != tc.want {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestSplit(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		t    string
		want values.Array
	}{
		{
			name: "Basic",
			v:    "a,b,c",
			t:    ",",
			want: values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), []values.Value{
				values.NewString("a"), values.NewString("b"), values.NewString("c")}),
		},
		{
			name: "Palindrome",
			v:    "a man a plan a canal panama",
			t:    "a ",
			want: values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), []values.Value{
				values.NewString(""), values.NewString("man "), values.NewString("plan "), values.NewString("canal panama")}),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testValue := generateSplit("split", []string{stringArgV, stringArgT}, strings.Split)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "t": values.NewString(tc.t)})
			result, err := testValue.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Array()

			if err != nil {
				t.Fatal(err)
			}

			if !res.Equal(tc.want) {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestSplitAfter(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		t    string
		want values.Array
	}{
		{
			name: "Basic",
			v:    "a,b,c",
			t:    ",",
			want: values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), []values.Value{
				values.NewString("a,"), values.NewString("b,"), values.NewString("c")}),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testValue := generateSplit("splitAfter", []string{stringArgV, stringArgT}, strings.SplitAfter)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "t": values.NewString(tc.t)})
			result, err := testValue.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Array()

			if err != nil {
				t.Fatal(err)
			}

			if !res.Equal(tc.want) {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestSplitN(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		t    string
		i    int
		want values.Array
	}{
		{
			name: "Basic",
			v:    "a,b,c",
			t:    ",",
			i:    2,
			want: values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), []values.Value{
				values.NewString("a"), values.NewString("b,c")}),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testValue := generateSplitN("splitN", []string{stringArgV, stringArgT, integer}, strings.SplitN)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "t": values.NewString(tc.t), "i": values.NewInt(int64(tc.i))})
			result, err := testValue.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Array()

			if err != nil {
				t.Fatal(err)
			}

			if !res.Equal(tc.want) {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestSplitAfterN(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		t    string
		i    int
		want values.Array
	}{
		{
			name: "Basic",
			v:    "a,b,c",
			t:    ",",
			i:    2,
			want: values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), []values.Value{
				values.NewString("a,"), values.NewString("b,c")}),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testValue := generateSplitN("splitAfterN", []string{stringArgV, stringArgT, integer}, strings.SplitAfterN)
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v), "t": values.NewString(tc.t), "i": values.NewInt(int64(tc.i))})
			result, err := testValue.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := result.Array()

			if err != nil {
				t.Fatal(err)
			}

			if !res.Equal(tc.want) {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})

	}
}

func TestJoinStr(t *testing.T) {
	fluxFunc := SpecialFns["joinStr"]
	arr := values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), []values.Value{
		values.NewString("a"), values.NewString("b"), values.NewString("c")})
	fluxArg := values.NewObjectWithValues(map[string]values.Value{"arr": arr, "v": values.NewString(", ")})
	want := strings.Join([]string{"a", "b", "c"}, ", ")
	got, err := fluxFunc.Call(dependenciestest.Default().Inject(context.Background()), fluxArg)
	if err != nil {
		t.Fatal(err)
	}
	if want != got.Str() {
		t.Errorf("input %f: expected %v, got %f", arr, want, got)
	}

}

func TestStrLength(t *testing.T) {
	testCases := []struct {
		name string
		v    string
		want int
	}{
		{
			name: "ASCII",
			v:    "abc",
			want: 3,
		},
		{
			name: "alphanumeric",
			v:    "CRJ34kf9",
			want: 8,
		},
		{
			name: "Blank",
			v:    "",
			want: 0,
		},
		{
			name: "Space",
			v:    "  ",
			want: 2,
		},
		{
			name: "Other Language",
			v:    "汉字",
			want: 2,
		},
		{
			name: "Space + Other",
			v:    "汉 字",
			want: 3,
		},
		{
			name: "Special Characters",
			v:    "latīna",
			want: 6,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testValue := strlen
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v)})
			result, err := testValue.Call(dependenciestest.Default().Inject(context.Background()), testCase)
			res := int(result.Int())

			if err != nil {
				t.Fatal(err)
			}

			if res != (tc.want) {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})
	}
}

func TestSubstring(t *testing.T) {
	testCases := []struct {
		name      string
		v         string
		start     int
		end       int
		want      string
		expectErr error
	}{
		{
			name:      "entire string",
			v:         "influx",
			start:     0,
			end:       6,
			want:      "influx",
			expectErr: errors.New("indices out of bounds"),
		},
		{
			name:      "simple substring",
			v:         "influx",
			start:     2,
			end:       5,
			want:      "flu",
			expectErr: errors.New("indices out of bounds"),
		},
		{
			name:      "chinese",
			v:         "汉字汉字汉字",
			start:     2,
			end:       5,
			want:      "汉字汉",
			expectErr: errors.New("indices out of bounds"),
		},
		{
			name:      "chinese and space",
			v:         "汉 字汉字  汉字",
			start:     4,
			end:       7,
			want:      "字  ",
			expectErr: errors.New("indices out of bounds"),
		},
		{
			name:      "alpha",
			v:         "ineedmesomeabcsoup",
			start:     -1,
			end:       7,
			want:      "",
			expectErr: errors.New("indices out of bounds"),
		},
		{
			name:      "beta",
			v:         "ineedmesomeabcsoup",
			start:     0,
			end:       3389,
			want:      "",
			expectErr: errors.New("indices out of bounds"),
		},
		{
			name:      "alphabet",
			v:         "ineedmesomeabcsoup",
			start:     -289,
			end:       23948,
			want:      "",
			expectErr: errors.New("indices out of bounds"),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testValue := substring
			testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString(tc.v),
				"start": values.NewInt(int64(tc.start)), "end": values.NewInt(int64(tc.end))})
			result, err := testValue.Call(dependenciestest.Default().Inject(context.Background()), testCase)

			if err != nil {
				if got, want := err.Error(), tc.expectErr.Error(); got != want {
					t.Errorf("unexpected error - want: %s, got: %s", want, got)
				}
				return
			}

			res := result.Str()

			if res != (tc.want) {
				t.Errorf("string function result %s expected: %v, got: %s", tc.name, tc.want, result)
			}
		})
	}
}

func BenchmarkSubstring(b *testing.B) {
	testValue := substring
	testCase := values.NewObjectWithValues(map[string]values.Value{"v": values.NewString("townsendapplebeepancake"),
		"start": values.NewInt(int64(0)), "end": values.NewInt(int64(5))})
	for i := 0; i < b.N; i++ {
		testValue.Call(dependenciestest.Default().Inject(context.Background()), testCase)
	}
}
