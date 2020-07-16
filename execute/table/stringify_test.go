package table_test

import (
	"testing"

	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/execute/table/static"
)

func TestStringify(t *testing.T) {
	in := static.Table{
		"_measurement": static.StringKey("m0"),
		"_field":       static.StringKey("f0"),
		"t0":           static.StringKey("a"),
		"_time":        static.Times("2020-01-01T00:00:00Z", 10, 20, 30, 40, 50),
		"_value":       static.Ints(6, 7, 4, 12, 3, 9),
	}
	got := table.Stringify(in)

	want := `# _field=f0,_measurement=m0,t0=a _time=time,_value=int
_field=f0,_measurement=m0,t0=a _time=2020-01-01T00:00:00Z,_value=6i
_field=f0,_measurement=m0,t0=a _time=2020-01-01T00:00:10Z,_value=7i
_field=f0,_measurement=m0,t0=a _time=2020-01-01T00:00:20Z,_value=4i
_field=f0,_measurement=m0,t0=a _time=2020-01-01T00:00:30Z,_value=12i
_field=f0,_measurement=m0,t0=a _time=2020-01-01T00:00:40Z,_value=3i
_field=f0,_measurement=m0,t0=a _time=2020-01-01T00:00:50Z,_value=9i
`
	if got != want {
		t.Errorf("unexpected string output -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}

func TestStringify_Empty(t *testing.T) {
	in := static.Table{
		"_measurement": static.StringKey("m0"),
		"_field":       static.StringKey("f0"),
		"t0":           static.StringKey("a"),
		"_time":        static.Times(),
		"_value":       static.Ints(),
	}
	got := table.Stringify(in)

	want := `# _field=f0,_measurement=m0,t0=a _time=time,_value=int
`
	if got != want {
		t.Errorf("unexpected string output -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}
