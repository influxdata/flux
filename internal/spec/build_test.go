package spec_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/execute"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/internal/operation"
	"github.com/influxdata/flux/internal/spec"
	"github.com/influxdata/flux/runtime"
)

func Benchmark_FromScript(b *testing.B) {
	query := `
import "influxdata/influxdb/monitor"
// Disable to the call to to since that isn't enabled
// in the flux repository.
option monitor.write = (tables=<-) => tables
check = from(bucket: "telegraf")
	|> range(start: -5m)
	|> mean()
	|> monitor.check(
		data: {tags: {}, _type: "default", _check_id: 101, _check_name: "test"},
		crit: (r) => r._value > 90,
		messageFn: (r) => "${r._value} is greater than 90",
	)
	|> monitor.stateChanges(toLevel: "crit")

// Multiple yield calls to the same table object so that
// we check whether we have a duplicate table object node
// to exercise that piece of code.
check |> yield(name: "checkResult")
check |> yield(name: "mean")
`
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := spec.FromScript(ctx, runtime.Default, time.Now(), query); err != nil {
			b.Fatal(err)
		}
	}
}

func TestFromEvaluation(t *testing.T) {
	ctx, deps := dependency.Inject(
		context.Background(),
		dependenciestest.Default(),
		execute.DefaultExecutionDependencies(),
	)
	defer deps.Finish()
	nowDefault := time.Unix(0, 0)

	type args struct {
		query      string
		now        time.Time
		skipYields bool
	}
	tests := []struct {
		name    string
		args    args
		want    *operation.Spec
		wantErr bool
	}{
		{
			name: "keep single trailing yield",
			args: args{
				query: `
				import "array"
				array.from(rows:[{a: 1}])
					|> yield(name: "a")
				`,
				now:        nowDefault,
				skipYields: false,
			},
			want: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "array.from0"},
					{ID: "yield1"},
				},
				Edges: []operation.Edge{
					{Parent: "array.from0", Child: "yield1"},
				},
			},
		},
		{
			name: "skip single trailing yield",
			args: args{
				query: `
				import "array"
				array.from(rows:[{a: 1}])
					|> yield(name: "a")
				`,
				now:        nowDefault,
				skipYields: true,
			},
			want: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "array.from0"},
				},
				// No edges since there's only a single node left
			},
		},
		{
			name: "keep multiple trailing yields",
			args: args{
				query: `
				import "array"
				array.from(rows:[{a: 1}])
					|> yield(name: "a")
					|> yield(name: "b")
					|> yield(name: "c")
				`,
				now:        nowDefault,
				skipYields: false,
			},
			want: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "array.from0"},
					{ID: "yield1"},
					{ID: "yield2"},
					{ID: "yield3"},
				},
				Edges: []operation.Edge{
					{Parent: "array.from0", Child: "yield1"},
					{Parent: "yield1", Child: "yield2"},
					{Parent: "yield2", Child: "yield3"},
				},
			},
		}, {
			name: "skip multiple trailing yields",
			args: args{
				query: `
				import "array"
				array.from(rows:[{a: 1}])
					|> yield(name: "a")
					|> yield(name: "b")
					|> yield(name: "c")
				`,
				now:        nowDefault,
				skipYields: true,
			},
			want: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "array.from0"},
				},
				// No edges since there's only a single node left
			},
		},
		{
			name: "keep interior yields",
			args: args{
				query: `
				import "array"
				array.from(rows:[{a: 1}])
					|> yield(name: "a")
					|> map(fn: (r) => r)
				`,
				now:        nowDefault,
				skipYields: false,
			},
			want: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "array.from0"},
					{ID: "yield1"},
					{ID: "map2"},
				},
				Edges: []operation.Edge{
					{Parent: "array.from0", Child: "yield1"},
					{Parent: "yield1", Child: "map2"},
				},
			},
		},
		{
			name: "skip interior yields",
			args: args{
				query: `
				import "array"
				array.from(rows:[{a: 1}])
					|> yield(name: "a")
					|> map(fn: (r) => r)
				`,
				now:        nowDefault,
				skipYields: true,
			},
			want: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "array.from0"},
					{ID: "map2"},
				},
				Edges: []operation.Edge{
					{Parent: "array.from0", Child: "map2"},
				},
			},
		},
		{
			name: "skip many interior yields",
			args: args{
				query: `
				import "array"
				array.from(rows:[{a: 1}])
					|> yield(name: "a")
					|> yield(name: "b")
					|> yield(name: "c")
					|> yield(name: "d")
					|> yield(name: "e")
					|> yield(name: "f")
					|> map(fn: (r) => r)
				`,
				now:        nowDefault,
				skipYields: true,
			},
			want: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "array.from0"},
					{ID: "map7"},
				},
				Edges: []operation.Edge{
					{Parent: "array.from0", Child: "map7"},
				},
			},
		},
		{
			name: "multiple ops keeping interior yields",
			args: args{
				query: `
				import "array"
				array.from(rows:[{a: 1}])
					|> yield(name: "a")
					|> yield(name: "b")
					|> map(fn: (r) => r)
					|> map(fn: (r) => r)
					|> yield(name: "c")
					|> map(fn: (r) => r)
				`,
				now:        nowDefault,
				skipYields: false,
			},
			want: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "array.from0"},
					{ID: "yield1"},
					{ID: "yield2"},
					{ID: "map3"},
					{ID: "map4"},
					{ID: "yield5"},
					{ID: "map6"},
				},
				Edges: []operation.Edge{
					{Parent: "array.from0", Child: "yield1"},
					{Parent: "yield1", Child: "yield2"},
					{Parent: "yield2", Child: "map3"},
					{Parent: "map3", Child: "map4"},
					{Parent: "map4", Child: "yield5"},
					{Parent: "yield5", Child: "map6"},
				},
			},
		},
		{
			name: "multiple ops skipping interior yields",
			args: args{
				query: `
				import "array"
				array.from(rows:[{a: 1}])
					|> yield(name: "a")
					|> yield(name: "b")
					|> map(fn: (r) => r)
					|> map(fn: (r) => r)
					|> yield(name: "c")
					|> map(fn: (r) => r)
				`,
				now:        nowDefault,
				skipYields: true,
			},
			want: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "array.from0"},
					{ID: "map3"},
					{ID: "map4"},
					{ID: "map6"},
				},
				Edges: []operation.Edge{
					{Parent: "array.from0", Child: "map3"},
					{Parent: "map3", Child: "map4"},
					{Parent: "map4", Child: "map6"},
				},
			},
		},
		{
			name: "multiple side effects when keeping yields",
			args: args{
				query: `
				import "array"
				import "sql"
				array.from(rows:[{a: 1}])
					|> sql.to(driverName: "sqlite3", dataSourceName: ":memory:", table: "test")
					|> sql.to(driverName: "sqlite3", dataSourceName: ":memory:", table: "test")
				`,
				now:        nowDefault,
				skipYields: false,
			},
			want: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "array.from0"},
					{ID: "toSQL1"},
					{ID: "toSQL2"},
				},
				Edges: []operation.Edge{
					{Parent: "array.from0", Child: "toSQL1"},
					{Parent: "toSQL1", Child: "toSQL2"},
				},
			},
		},
		{
			name: "multiple side effects when skipping yields",
			args: args{
				query: `
				import "array"
				import "sql"
				array.from(rows:[{a: 1}])
					|> sql.to(driverName: "sqlite3", dataSourceName: ":memory:", table: "test")
					|> sql.to(driverName: "sqlite3", dataSourceName: ":memory:", table: "test")
				`,
				now:        nowDefault,
				skipYields: true,
			},
			wantErr: true,
		},
		{
			name: "multiple side effects with trailing yield when skipping yields",
			args: args{
				query: `
				import "array"
				import "sql"
				array.from(rows:[{a: 1}])
					|> sql.to(driverName: "sqlite3", dataSourceName: ":memory:", table: "test")
					|> sql.to(driverName: "sqlite3", dataSourceName: ":memory:", table: "test")
					|> yield(name: "a")
				`,
				now:        nowDefault,
				skipYields: true,
			},
			wantErr: true,
		},
		{
			name: "multiple side effects with trailing yield when keeping yields",
			args: args{
				query: `
				import "array"
				import "sql"
				array.from(rows:[{a: 1}])
					|> sql.to(driverName: "sqlite3", dataSourceName: ":memory:", table: "test")
					|> sql.to(driverName: "sqlite3", dataSourceName: ":memory:", table: "test")
					|> yield(name: "a")
				`,
				now:        nowDefault,
				skipYields: false,
			},
			want: &operation.Spec{
				Operations: []*operation.Node{
					{ID: "array.from0"},
					{ID: "toSQL1"},
					{ID: "toSQL2"},
					{ID: "yield3"},
				},
				Edges: []operation.Edge{
					{Parent: "array.from0", Child: "toSQL1"},
					{Parent: "toSQL1", Child: "toSQL2"},
					{Parent: "toSQL2", Child: "yield3"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ses, _, err := runtime.Eval(ctx, tt.args.query)
			if err != nil {
				t.Fatal(err)
			}

			got, err := spec.FromEvaluation(ctx, ses, tt.args.now, tt.args.skipYields)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("FromEvaluation() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			gotOpIDs := make([]operation.NodeID, len(got.Operations))
			for _, o := range got.Operations {
				gotOpIDs = append(gotOpIDs, o.ID)
			}
			wantOpIDs := make([]operation.NodeID, len(tt.want.Operations))
			for _, o := range tt.want.Operations {
				wantOpIDs = append(wantOpIDs, o.ID)
			}

			if !reflect.DeepEqual(gotOpIDs, wantOpIDs) {
				t.Errorf("FromEvaluation() Operations \ngot:\n%v\nwant:\n%v", gotOpIDs, wantOpIDs)
			}
			if !reflect.DeepEqual(got.Edges, tt.want.Edges) {
				t.Errorf("FromEvaluation() Edges \ngot:\n%v\nwant:\n%v", got.Edges, tt.want.Edges)
			}
		})
	}
}
