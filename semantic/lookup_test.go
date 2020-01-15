package semantic_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	semantic "github.com/influxdata/flux/semantic"
)

func TestLookupSimpleTypes(t *testing.T) {
	for _, testCase := range []struct {
		path string
		id   string
		name string
		want string
	}{
		{path: "math", id: "pi", name: "lookup math.pi", want: "float"},
		{path: "math", id: "maxint", name: "lookup math.maxint", want: "int"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got, _ := semantic.LookupBuiltinType(testCase.path, testCase.id)
			if want, got := testCase.want, got.String(); want != got {
				t.Fatalf("unexpected result -want/+got\n\t- %s\n\t+ %s", want, got)
			}
		})
	}
}

// TestLookupComplexTypes has a list of several test cases. This is not
// meant to be an exhaustive list of all builtins from the stdlib, but it
// is meant to cover various PolyTypes and touch most packages from the stdlib.
// It is not necessary to update this test with every future addition to the stdlib.
func TestLookupComplexTypes(t *testing.T) {
	for _, testCase := range []struct {
		path string
		id   string
		name string
		want string
	}{
		{
			path: "csv",
			id:   "from",
			name: "lookup csv.from",
			want: "(?csv: string, ?file: string) -> [t0]",
		},
		{
			path: "date",
			id:   "nanosecond",
			name: "lookup date.nanosecond",
			want: "(t: time) -> int",
		},
		{
			path: "date",
			id:   "truncate",
			name: "lookup date.truncate",
			want: "(t: time, unit: duration) -> time",
		},
		{
			path: "experimental/bigtable",
			id:   "from",
			name: "lookup experimental/bigtable.from",
			want: "(instance: string, project: string, table: string, token: string) -> [t0]",
		},
		{
			path: "experimental",
			id:   "to",
			name: "lookup experimental.to",
			want: "(?bucket: string, ?bucketID: string, ?host: string, ?org: string, ?orgID: string, <-tables: [t0], ?token: string) -> [t0]",
		},
		{
			path: "http",
			id:   "post",
			name: "lookup http.post",
			want: "(?data: bytes, ?headers: t0, url: string) -> int",
		},
		{
			path: "influxdata/influxdb/secrets",
			id:   "get",
			name: "lookup influxdata/influxdb/secrets.get",
			want: "(key: string) -> string",
		},
		{
			path: "json",
			id:   "encode",
			name: "lookup json.encode",
			want: "(v: t0) -> bytes",
		},
		{
			path: "strings",
			id:   "title",
			name: "lookup strings.title",
			want: "(v: string) -> string",
		},
		{
			path: "system",
			id:   "time",
			name: "lookup system.time",
			want: "() -> time",
		},
		{
			path: "universe",
			id:   "bool",
			name: "lookup universe.bool",
			want: "(v: t0) -> bool",
		},
		{
			path: "internal/promql",
			id:   "changes",
			name: "lookup internal/promql.changes",
			want: "(<-tables: [{_value: float | t0}]) -> [{_value: float | t1}]",
		},
		{
			path: "sql",
			id:   "to",
			name: "lookup sql.to",
			want: "(?batchSize: int, dataSourceName: string, driverName: string, table: string, <-tables: [t0]) -> [t0]",
		},
		{
			path: "testing",
			id:   "assertEmpty",
			name: "lookup testing.assertEmpty",
			want: "(<-tables: [t0]) -> [t0]",
		},
		{
			path: "universe",
			id:   "filter",
			name: "lookup universe.filter",
			want: "(fn: (r: t0) -> bool, ?onEmpty: string, <-tables: [t0]) -> [t0]",
		},
		{
			path: "universe",
			id:   "map",
			name: "lookup universe.map",
			want: "(fn: (r: t0) -> t1, ?mergeKey: bool, <-tables: [t0]) -> [t1]",
		},
	} {

		t.Run(testCase.name, func(t *testing.T) {
			got, _ := semantic.LookupBuiltinType(testCase.path, testCase.id)
			if want, got := testCase.want, canonicalizeType(got); want != got {
				t.Fatalf("unexpected result -want/+got\n\t- %s\n\t+ %s", want, got)
			}
		})
	}
}

var tvarRegexp *regexp.Regexp = regexp.MustCompile("t[0-9]+")

// canonicalizeType returns a string representation of the given type
// that reindexes in the numbers in type variables starting from zero.
func canonicalizeType(mt semantic.MonoType) string {
	counter := 0
	tvm := make(map[uint64]int)
	err := mt.GetCanonicalMapping(&counter, tvm)

	if err != nil {
		panic(err)
	}
	tstr := mt.String()
	return tvarRegexp.ReplaceAllStringFunc(tstr, func(in string) string {
		n, err := strconv.Atoi(in[1:])
		if err != nil {
			panic(err)
		}
		nn, ok := tvm[uint64(n)]
		if !ok {
			panic(fmt.Sprintf("could not find tvar mapping for %v", in))
		}
		return fmt.Sprintf("t%v", nn)
	})
}
