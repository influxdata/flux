package table_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table/static"
	"github.com/influxdata/flux/internal/execute/table"
)

func TestMask(t *testing.T) {
	want := static.TableGroup{
		static.Times("_time", "2020-07-31T09:00:00Z", 10, 20, 30, 40, 50),
		static.Ints("_value", 0, 1, 2, 3, 4, 5),
		static.TableList{
			static.StringKeys("t0", "a-0", "a-1", "a-2"),
		},
	}

	in := static.TableGroup{
		static.StringKey("_measurement", "m0"),
		static.StringKey("_field", "f0"),
		static.TimeKey("_start", "2020-07-31T09:00:00Z"),
		static.TimeKey("_stop", "2020-07-31T10:00:00Z"),
	}
	in = append(in, want...)

	var got table.Iterator
	if err := in.Do(func(tbl flux.Table) error {
		maskTable := table.Mask(tbl, []string{"_measurement", "_field", "_start", "_stop"})
		cpy, err := execute.CopyTable(maskTable)
		if err != nil {
			return err
		}
		got = append(got, cpy)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if diff := table.Diff(want, got); diff != "" {
		t.Fatalf("unexpected diff -want/+got:\n%s", diff)
	}
}
