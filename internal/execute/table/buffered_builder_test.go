package table_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/table/static"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/memory"
)

func TestBufferedBuilder_AppendTable(t *testing.T) {
	for _, tt := range []struct {
		name    string
		in      static.TableGroup
		want    static.Table
		wantErr string
	}{
		{
			name: "OneTable",
			in: static.TableGroup{
				static.StringKey("_measurement", "m0"),
				static.StringKey("_field", "f0"),
				static.StringKey("t0", "a"),
				static.Table{
					static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40),
					static.Ints("_value", 5, 3, nil, 9, 2),
				},
			},
			want: static.Table{
				static.StringKey("_measurement", "m0"),
				static.StringKey("_field", "f0"),
				static.StringKey("t0", "a"),
				static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40),
				static.Ints("_value", 5, 3, nil, 9, 2),
			},
		},
		{
			name: "Empty",
			in: static.TableGroup{
				static.StringKey("_measurement", "m0"),
				static.StringKey("_field", "f0"),
				static.StringKey("t0", "a"),
				static.Table{
					static.Times("_time"),
					static.Ints("_value"),
				},
			},
			want: static.Table{
				static.StringKey("_measurement", "m0"),
				static.StringKey("_field", "f0"),
				static.StringKey("t0", "a"),
				static.Times("_time"),
				static.Ints("_value"),
			},
		},
		{
			name: "TwoTables",
			in: static.TableGroup{
				static.StringKey("_measurement", "m0"),
				static.StringKey("_field", "f0"),
				static.Table{
					static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30),
					static.Ints("_value", 4, 8, 2, 7),
				},
				static.Table{
					static.Times("_time", "2020-01-01T00:00:40Z", 10, 20),
					static.Ints("_value", 3, 1, 9),
				},
			},
			want: static.Table{
				static.StringKey("_measurement", "m0"),
				static.StringKey("_field", "f0"),
				static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50, 60),
				static.Ints("_value", 4, 8, 2, 7, 3, 1, 9),
			},
		},
		{
			name: "FillNulls",
			in: static.TableGroup{
				static.StringKey("_measurement", "m0"),
				static.Table{
					static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
					static.Floats("f0", 3, 8, 2, 4, 11, 6),
					static.Ints("f1", 5, 9, 2, 3, 4, 10),
				},
				static.Table{
					static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
					static.Floats("f0", 18, 2, 7, 9, 4, 1),
				},
			},
			want: static.Table{
				static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50, "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				static.StringKey("_measurement", "m0"),
				static.Floats("f0", 3, 8, 2, 4, 11, 6, 18, 2, 7, 9, 4, 1),
				static.Ints("f1", 5, 9, 2, 3, 4, 10, nil, nil, nil, nil, nil, nil),
			},
		},
		{
			name: "BackfillNulls",
			in: static.TableGroup{
				static.StringKey("_measurement", "m0"),
				static.Table{
					static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
					static.Floats("f0", 3, 8, 2, 4, 11, 6),
				},
				static.Table{
					static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
					static.Floats("f0", 18, 2, 7, 9, 4, 1),
					static.Ints("f1", 5, 9, 2, 3, 4, 10),
				},
			},
			want: static.Table{
				static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50, "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				static.StringKey("_measurement", "m0"),
				static.Floats("f0", 3, 8, 2, 4, 11, 6, 18, 2, 7, 9, 4, 1),
				static.Ints("f1", nil, nil, nil, nil, nil, nil, 5, 9, 2, 3, 4, 10),
			},
		},
		{
			name: "BackfillEmpty",
			in: static.TableGroup{
				static.StringKey("_measurement", "m0"),
				static.Table{
					static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
					static.Floats("f0", 3, 8, 2, 4, 11, 6),
				},
				static.Table{
					static.Times("_time"),
					static.Floats("f0"),
					static.Ints("f1"),
				},
			},
			want: static.Table{
				static.StringKey("_measurement", "m0"),
				static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				static.Floats("f0", 3, 8, 2, 4, 11, 6),
				static.Ints("f1", nil, nil, nil, nil, nil, nil),
			},
		},
		{
			name: "ConflictingSchema",
			in: static.TableGroup{
				static.Table{
					static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
					static.Floats("_value", 3, 8, 2, 4, 11, 6),
				},
				static.Table{
					static.Times("_time", "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
					static.Ints("_value", 5, 9, 2, 3, 4, 10),
				},
			},
			wantErr: `schema collision detected: column "_value" is both of type int and float`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b *table.BufferedBuilder
			if err := tt.in.Do(func(tbl flux.Table) error {
				if b == nil {
					b = table.NewBufferedBuilder(tbl.Key(), memory.DefaultAllocator)
				}
				return b.AppendTable(tbl)
			}); err != nil {
				if want, got := tt.wantErr, err.Error(); want != got {
					t.Errorf("unexpected error -want/+got:\n\t- %s\n\t+ %s", want, got)
				}
				return
			}

			if tt.wantErr != "" {
				t.Fatal("expected error")
			}

			out, err := b.Table()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			want, got := tt.want, table.Iterator{out}

			if diff := table.Diff(want, got); diff != "" {
				t.Fatalf("unexpected diff -want/+got:\n%s", diff)
			}
		})
	}
}
