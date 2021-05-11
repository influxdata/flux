package runtime_test

import (
	"testing"

	"github.com/influxdata/flux/runtime"
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
			got, _ := runtime.LookupBuiltinType(testCase.path, testCase.id)
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
			want: "(?csv: string, ?file: string, ?mode: string) => [A]",
		},
		{
			path: "date",
			id:   "nanosecond",
			name: "lookup date.nanosecond",
			want: "(t: A) => int",
		},
		{
			path: "date",
			id:   "truncate",
			name: "lookup date.truncate",
			want: "(t: A, unit: duration) => time",
		},
		{
			path: "experimental/bigtable",
			id:   "from",
			name: "lookup experimental/bigtable.from",
			want: "(instance: string, project: string, table: string, token: string) => [A]",
		},
		{
			path: "experimental",
			id:   "to",
			name: "lookup experimental.to",
			want: "(?bucket: string, ?bucketID: string, ?host: string, ?org: string, ?orgID: string, <-tables: [A], ?token: string) => [A]",
		},
		{
			path: "http",
			id:   "post",
			name: "lookup http.post",
			want: "(?data: bytes, ?headers: A, url: string) => int",
		},
		{
			path: "influxdata/influxdb/secrets",
			id:   "get",
			name: "lookup influxdata/influxdb/secrets.get",
			want: "(key: string) => string",
		},
		{
			path: "json",
			id:   "encode",
			name: "lookup json.encode",
			want: "(v: A) => bytes",
		},
		{
			path: "strings",
			id:   "title",
			name: "lookup strings.title",
			want: "(v: string) => string",
		},
		{
			path: "system",
			id:   "time",
			name: "lookup system.time",
			want: "() => time",
		},
		{
			path: "universe",
			id:   "bool",
			name: "lookup universe.bool",
			want: "(v: A) => bool",
		},
		{
			path: "internal/promql",
			id:   "changes",
			name: "lookup internal/promql.changes",
			want: "(<-tables: [{A with _value: float}]) => [{B with _value: float}]",
		},
		{
			path: "sql",
			id:   "to",
			name: "lookup sql.to",
			want: "(?batchSize: int, dataSourceName: string, driverName: string, table: string, <-tables: [A]) => [A]",
		},
		{
			path: "testing",
			id:   "assertEmpty",
			name: "lookup testing.assertEmpty",
			want: "(<-tables: [A]) => [A]",
		},
		{
			path: "universe",
			id:   "filter",
			name: "lookup universe.filter",
			want: "(fn: (r: A) => bool, ?onEmpty: string, <-tables: [A]) => [A]",
		},
		{
			path: "universe",
			id:   "map",
			name: "lookup universe.map",
			want: "(fn: (r: A) => B, ?mergeKey: bool, <-tables: [A]) => [B]",
		},
	} {

		t.Run(testCase.name, func(t *testing.T) {
			got, _ := runtime.LookupBuiltinType(testCase.path, testCase.id)
			if want, got := testCase.want, got.CanonicalString(); want != got {
				t.Fatalf("unexpected result -want/+got\n\t- %s\n\t+ %s", want, got)
			}
		})
	}
}
