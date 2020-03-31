package lang_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
)

func runQuery(script string, opts ...lang.CompileOption) (flux.Query, error) {
	program, err := lang.Compile(script, time.Unix(0, 0), opts...)
	if err != nil {
		return nil, err
	}
	ctx := executetest.NewTestExecuteDependencies().Inject(context.Background())
	q, err := program.Start(ctx, &memory.Allocator{})
	if err != nil {
		return nil, err
	}
	return q, nil
}

var validScript = `
import "csv"

data = "
#datatype,string,long,string,long
#group,false,false,false,true
#default,_result,,,
,result,table,tag,value
,,0,a,10
,,0,a,10
,,1,b,20
,,1,b,20
,,2,c,30
,,2,c,30
,,3,d,40
,,3,d,40
"

csv.from(csv: data)
	|> filter(fn: (r) => true)
	|> map(fn: (r) => r)
	|> limit(n: 100)
	|> yield(name: "res")`

func TestQuery_Results(t *testing.T) {
	q, err := runQuery(validScript)
	if err != nil {
		t.Fatalf("unexpected error while creating query: %s", err)
	}

	// gather counts
	var resCount, tableCount, totRows int
	for res := range q.Results() {
		resCount++
		if err := res.Tables().Do(func(tbl flux.Table) error {
			tableCount++
			return tbl.Do(func(cr flux.ColReader) error {
				tags := cr.Strings(0)
				for i := 0; i < tags.Len(); i++ {
					totRows++
				}
				return nil
			})
		}); err != nil {
			t.Fatalf("unexpected error while iterating over tables: %s", err)
		}
	}

	// compare expected counts
	if resCount != 1 {
		t.Errorf("got %d results instead of %d", resCount, 1)
	}
	if tableCount != 4 {
		t.Errorf("got %d tables instead of %d", tableCount, 4)
	}
	if totRows != 8 {
		t.Errorf("got %d rows instead of %d", totRows, 8)
	}

	// release query resources
	q.Done()

	if q.Err() != nil {
		t.Fatalf("unexpected error from query execution: %s", q.Err())
	}
}

func TestQuery_Stats(t *testing.T) {
	t.Skip("stats are updated by the controller, running a standalone query won't update them")

	q, err := runQuery(validScript)
	if err != nil {
		t.Fatalf("unexpected error while creating query: %s", err)
	}

	// consume results
	for res := range q.Results() {
		if err := res.Tables().Do(func(tbl flux.Table) error {
			return tbl.Do(func(cr flux.ColReader) error {
				return nil
			})
		}); err != nil {
			t.Fatalf("unexpected error while iterating over tables: %s", err)
		}
	}

	stats := q.Statistics()
	if stats.TotalDuration <= 0 {
		t.Errorf("unexpected total duration: %v", stats.CompileDuration)
	}
	if stats.CompileDuration <= 0 {
		t.Errorf("unexpected compile duration: %v", stats.CompileDuration)
	}
	if stats.QueueDuration <= 0 {
		t.Errorf("unexpected queue duration: %v", stats.CompileDuration)
	}
	if stats.PlanDuration <= 0 {
		t.Errorf("unexpected plan duration: %v", stats.CompileDuration)
	}
	if stats.ExecuteDuration <= 0 {
		t.Errorf("unexpected execute duration: %v", stats.CompileDuration)
	}
	if stats.Concurrency <= 0 {
		t.Errorf("unexpected concurrency: %v", stats.CompileDuration)
	}
	if stats.MaxAllocated <= 0 {
		t.Errorf("unexpected max allocated: %v", stats.CompileDuration)
	}
}

func TestQuery_ProfilingStats(t *testing.T) {
	q, err := runQuery(validScript, lang.WithExecuteOptions(execute.WithProfiling()))
	if err != nil {
		t.Fatalf("unexpected error while creating query: %s", err)
	}

	// consume results
	for res := range q.Results() {
		if err := res.Tables().Do(func(tbl flux.Table) error {
			return tbl.Do(func(cr flux.ColReader) error {
				return nil
			})
		}); err != nil {
			t.Fatalf("unexpected error while iterating over tables: %s", err)
		}
	}

	// The statistics are not complete until Done is called.
	q.Done()
	if q.Err() != nil {
		t.Fatalf("unexpected error from query execution: %s", q.Err())
	}

	stats := q.Statistics()
	if len(stats.Metadata) == 0 {
		t.Fatal("expected some metadata in stats, got none")
	}
	checkMeta := func(k string, m execute.TableMetadata) {
		if m.Start.After(time.Now()) {
			t.Errorf("inconsistent start time: %s -> %v", k, m)
		}
		if m.Stop.After(time.Now()) {
			t.Errorf("inconsistent stop time: %s -> %v", k, m)
		}
		if m.Stop.Before(m.Start) {
			t.Errorf("stop is before start: %s -> %v", k, m)
		}
		if m.NoRows <= 0 {
			t.Errorf("unexpected value for number of rows: %s -> %v", k, m)
		}
		if m.NoRows <= 0 {
			t.Errorf("unexpected value for number of values: %s -> %v", k, m)
		}
		if m.RowsSec <= 0.0 {
			t.Errorf("unexpected value for rows per second: %s -> %v", k, m)
		}
		if m.ValuesSec <= 0.0 {
			t.Errorf("unexpected value for values per second: %s -> %v", k, m)
		}
	}
	for k, vs := range stats.Metadata {
		if strings.HasPrefix(k, "profiling") {
			if l := len(vs); l > 1 {
				t.Errorf("expected only 1 profiling entry per key, got %d", l)
			}
			switch v := vs[0].(type) {
			case map[string]execute.TableMetadata:
				for sk, p := range v {
					checkMeta(fmt.Sprintf("%s -> %s", k, sk), p)
				}
			case map[string][]execute.TableMetadata:
				for sk, p := range v {
					for _, m := range p {
						checkMeta(fmt.Sprintf("%s -> %s", k, sk), m)
					}
				}
			default:
				t.Errorf("unexpected type: %T", v)
			}
		}
	}
}

func TestQuery_RuntimeError(t *testing.T) {
	var invalidScript = `
import "csv"

data = "
#datatype,string,long,long,string
#group,false,false,false,true
#default,_result,,,
,result,table,value,tag
,,0,10,a
"

csv.from(csv: data) |> map(fn: (r) => r.nonexistent)`

	q, err := runQuery(invalidScript)
	if err != nil {
		t.Fatalf("unexpected error while creating query: %s", err)
	}

	// consume and check for error
	for res := range q.Results() {
		if err := res.Tables().Do(func(tbl flux.Table) error {
			return tbl.Do(func(cr flux.ColReader) error {
				// does nothing
				return nil
			})
		}); err == nil {
			t.Fatal("expected error from accessing wrong property, got none")
		}
	}

	if q.Err() != nil {
		t.Fatalf("unexpected error from query execution: %s", q.Err())
	}
}
