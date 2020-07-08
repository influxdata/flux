package table_test

import (
	"testing"

	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/internal/execute/table/static"
	"github.com/influxdata/flux/memory"
)

func TestBufferedBuilder_AppendTable(t *testing.T) {
	for _, tt := range []struct {
		name    string
		in      []static.Table
		want    static.Table
		wantErr string
	}{
		{
			name: "OneTable",
			in: []static.Table{{
				"_measurement": static.StringKey("m0"),
				"_field":       static.StringKey("f0"),
				"t0":           static.StringKey("a"),
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40),
				"_value":       static.Ints(5, 3, nil, 9, 2),
			}},
			want: static.Table{
				"_measurement": static.StringKey("m0"),
				"_field":       static.StringKey("f0"),
				"t0":           static.StringKey("a"),
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40),
				"_value":       static.Ints(5, 3, nil, 9, 2),
			},
		},
		{
			name: "Empty",
			in: []static.Table{{
				"_measurement": static.StringKey("m0"),
				"_field":       static.StringKey("f0"),
				"t0":           static.StringKey("a"),
				"_time":        static.Times(),
				"_value":       static.Ints(),
			}},
			want: static.Table{
				"_measurement": static.StringKey("m0"),
				"_field":       static.StringKey("f0"),
				"t0":           static.StringKey("a"),
				"_time":        static.Times(),
				"_value":       static.Ints(),
			},
		},
		{
			name: "TwoTables",
			in: []static.Table{{
				"_measurement": static.StringKey("m0"),
				"_field":       static.StringKey("f0"),
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30),
				"_value":       static.Ints(4, 8, 2, 7),
			}, {
				"_measurement": static.StringKey("m0"),
				"_field":       static.StringKey("f0"),
				"_time":        static.Times("2020-01-01T00:00:40Z", 10, 20),
				"_value":       static.Ints(3, 1, 9),
			}},
			want: static.Table{
				"_measurement": static.StringKey("m0"),
				"_field":       static.StringKey("f0"),
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50, 60),
				"_value":       static.Ints(4, 8, 2, 7, 3, 1, 9),
			},
		},
		{
			name: "FillNulls",
			in: []static.Table{{
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				"_measurement": static.StringKey("m0"),
				"f0":           static.Floats(3, 8, 2, 4, 11, 6),
				"f1":           static.Ints(5, 9, 2, 3, 4, 10),
			}, {
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				"_measurement": static.StringKey("m0"),
				"f0":           static.Floats(18, 2, 7, 9, 4, 1),
			}},
			want: static.Table{
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50, "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				"_measurement": static.StringKey("m0"),
				"f0":           static.Floats(3, 8, 2, 4, 11, 6, 18, 2, 7, 9, 4, 1),
				"f1":           static.Ints(5, 9, 2, 3, 4, 10, nil, nil, nil, nil, nil, nil),
			},
		},
		{
			name: "BackfillNulls",
			in: []static.Table{{
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				"_measurement": static.StringKey("m0"),
				"f0":           static.Floats(3, 8, 2, 4, 11, 6),
			}, {
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				"_measurement": static.StringKey("m0"),
				"f0":           static.Floats(18, 2, 7, 9, 4, 1),
				"f1":           static.Ints(5, 9, 2, 3, 4, 10),
			}},
			want: static.Table{
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50, "2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				"_measurement": static.StringKey("m0"),
				"f0":           static.Floats(3, 8, 2, 4, 11, 6, 18, 2, 7, 9, 4, 1),
				"f1":           static.Ints(nil, nil, nil, nil, nil, nil, 5, 9, 2, 3, 4, 10),
			},
		},
		{
			name: "BackfillEmpty",
			in: []static.Table{{
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				"_measurement": static.StringKey("m0"),
				"f0":           static.Floats(3, 8, 2, 4, 11, 6),
			}, {
				"_time":        static.Times(),
				"_measurement": static.StringKey("m0"),
				"f0":           static.Floats(),
				"f1":           static.Ints(),
			}},
			want: static.Table{
				"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				"_measurement": static.StringKey("m0"),
				"f0":           static.Floats(3, 8, 2, 4, 11, 6),
				"f1":           static.Ints(nil, nil, nil, nil, nil, nil),
			},
		},
		{
			name: "ConflictingSchema",
			in: []static.Table{{
				"_time":  static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				"_value": static.Floats(3, 8, 2, 4, 11, 6),
			}, {
				"_time":  static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
				"_value": static.Ints(5, 9, 2, 3, 4, 10),
			}},
			wantErr: `schema collision detected: column "_value" is both of type int and float`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b := table.NewBufferedBuilder(tt.in[0].Key(), memory.DefaultAllocator)
			for _, tbl := range tt.in {
				if err := b.AppendTable(tbl); err != nil {
					if want, got := tt.wantErr, err.Error(); want != got {
						t.Errorf("unexpected error -want/+got:\n\t- %q\n\t+ %q", want, got)
					}
					return
				}
			}

			if tt.wantErr != "" {
				t.Fatal("expected error")
			}

			out, err := b.Table()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if diff := table.Diff(t, tt.want, out); diff != "" {
				t.Fatalf("unexpected diff -want/+got:\n%s", diff)
			}
		})
	}
}
