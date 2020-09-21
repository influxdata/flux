package execute_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/metadata"
	"github.com/influxdata/flux/mock"
)

func TestOperatorProfiler_GetResult(t *testing.T) {
	// Create an operator profiler
	p := execute.AllProfilers["operator"]()
	// And inject it to the context.
	ctx := context.WithValue(context.Background(), execute.OperatorProfilerContextKey, p)
	// Build the "want" table.
	// This table is built dynamically because the table includes time data which changes
	// every time the test is run.
	var wantStr bytes.Buffer
	// Need to have the goroutines to sync on their writes to the buffer.
	var mu sync.Mutex
	wantStr.WriteString(`
#datatype,string,long,string,dateTime:RFC3339,dateTime:RFC3339,string,string,long
#group,false,false,true,false,false,false,false,false
#default,_profiler,,,,,,,
,result,table,_measurement,_start,_stop,Type,Label,Duration
`)
	// Unfortunately, the operator profiler result is only grouped on _measurement, it cannot
	// ensure a deterministic row order with our executetest.EqualResultIterators
	// Therefore, currently I only set this goroutine count to 2.
	count := 2
	wg := sync.WaitGroup{}
	wg.Add(count)
	fn := func(label string, ctx context.Context) {
		_, span := execute.StartSpanFromContext(ctx, "tf", label)
		profilerSpan := span.(*execute.OperatorProfilingSpan)
		// Finish() will write the data to the profiler
		// In Flux runtime, this is called when an execution node finishes execution
		profilerSpan.Finish()
		mu.Lock()
		// Write the expected result from raw span data.
		wantStr.WriteString(fmt.Sprintf(",,0,profiler/operator,%s,%s,tf,%s,%d\n",
			profilerSpan.Result.Start.Format(time.RFC3339Nano),
			profilerSpan.Result.Stop.Format(time.RFC3339Nano),
			profilerSpan.Result.Label,
			profilerSpan.Result.Stop.Sub(profilerSpan.Result.Start).Nanoseconds()))
		mu.Unlock()
		wg.Done()
	}
	for i := 0; i < count; i++ {
		go fn(fmt.Sprintf("op%d", i), ctx)
	}
	wg.Wait()
	// Wait a bit for the profiling results to be added.
	// In the query code path this is guaranteed because we only access the result
	// after the query finishes execution AND its result tables are read and encoded.
	time.Sleep(100 * time.Millisecond)
	tbl, err := p.GetResult(nil, &memory.Allocator{})
	if err != nil {
		t.Error(err)
	}
	result := table.NewProfilerResult(tbl)
	got := flux.NewSliceResultIterator([]flux.Result{&result})
	dec := csv.NewMultiResultDecoder(csv.ResultDecoderConfig{})
	want, e := dec.Decode(ioutil.NopCloser(strings.NewReader(wantStr.String())))
	if e != nil {
		t.Error(err)
	}
	if err := executetest.EqualResultIterators(want, got); err != nil {
		t.Fatal(err)
	}
}

func TestQueryProfiler_GetResult(t *testing.T) {
	p := &execute.QueryProfiler{}
	q := &mock.Query{}
	q.SetStatistics(flux.Statistics{
		TotalDuration:   1,
		CompileDuration: 2,
		QueueDuration:   3,
		PlanDuration:    4,
		RequeueDuration: 5,
		ExecuteDuration: 6,
		Concurrency:     7,
		MaxAllocated:    8,
		TotalAllocated:  9,
		RuntimeErrors:   []string{"1", "2"},
		Metadata: metadata.Metadata{
			"influxdb/scanned-bytes":  []interface{}{10},
			"influxdb/scanned-values": []interface{}{11},
			"flux/query-plan":         []interface{}{"query plan"},
		},
	})
	wantStr := `
#datatype,string,long,string,long,long,long,long,long,long,long,long,long,string,string,long,long
#group,false,false,true,false,false,false,false,false,false,false,false,false,false,false,false,false
#default,_profiler,,,,,,,,,,,,,,,
,result,table,_measurement,TotalDuration,CompileDuration,QueueDuration,PlanDuration,RequeueDuration,ExecuteDuration,Concurrency,MaxAllocated,TotalAllocated,RuntimeErrors,flux/query-plan,influxdb/scanned-bytes,influxdb/scanned-values
,,0,profiler/query,1,2,3,4,5,6,7,8,9,"1
2","query plan",10,11
`
	q.Done()
	tbl, err := p.GetResult(q, &memory.Allocator{})
	if err != nil {
		t.Error(err)
	}
	result := table.NewProfilerResult(tbl)
	got := flux.NewSliceResultIterator([]flux.Result{&result})
	dec := csv.NewMultiResultDecoder(csv.ResultDecoderConfig{})
	want, e := dec.Decode(ioutil.NopCloser(strings.NewReader(wantStr)))
	if e != nil {
		t.Error(err)
	}
	if err := executetest.EqualResultIterators(want, got); err != nil {
		t.Fatal(err)
	}
}
