package lang_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/influxdata/flux/builtin"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/pkg/errors"
)

func runQuery(script string) (flux.Query, error) {
	program, err := lang.Compile(script, time.Unix(0, 0))
	if err != nil {
		return nil, err
	}
	q, err := program.Start(context.Background(), &memory.Allocator{})
	if err != nil {
		return nil, err
	}
	return q, nil
}

var validScript = `
import "csv"

data = "
#datatype,string,long,long,string
#group,false,false,false,true
#default,_result,,,
,result,table,value,tag
,,0,10,a
,,0,10,a
,,1,20,b
,,1,20,b
,,2,30,c
,,2,30,c
,,3,40,d
,,3,40,d
"

csv.from(csv: data) |> yield(name: "res")`

func TestQuery_Results(t *testing.T) {
	q, err := runQuery(validScript)
	if err != nil {
		t.Fatal(errors.Wrap(err, "unexpected error while creating query"))
	}

	// gather counts
	var resCount, tableCount, totRows int
	for res := range q.Results() {
		resCount++
		if err := res.Tables().Do(func(tbl flux.Table) error {
			tableCount++
			return tbl.Do(func(cr flux.ColReader) error {
				tags := cr.Strings(1)
				for i := 0; i < tags.Len(); i++ {
					totRows++
				}
				return nil
			})
		}); err != nil {
			t.Fatal(errors.Wrap(err, "unexpected error while iterating over tables"))
		}
	}

	// com,pare expected counts
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
		t.Fatal(errors.Wrap(q.Err(), "unexpected error from query execution"))
	}
}

func TestQuery_Stats(t *testing.T) {
	t.Skip("stats are updated by the controller, running a standalone query won't update them")

	q, err := runQuery(validScript)
	if err != nil {
		t.Fatal(errors.Wrap(err, "unexpected error while creating query"))
	}

	// consume results
	for res := range q.Results() {
		if err := res.Tables().Do(func(tbl flux.Table) error {
			return tbl.Do(func(cr flux.ColReader) error {
				return nil
			})
		}); err != nil {
			t.Fatal(errors.Wrap(err, "unexpected error while iterating over tables"))
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
		t.Fatal(errors.Wrap(err, "unexpected error while creating query"))
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
		t.Fatal(errors.Wrap(q.Err(), "unexpected error from query execution"))
	}
}
