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
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/metadata"
	"github.com/influxdata/flux/mock"
	"github.com/opentracing/opentracing-go"
)

// Simulates setting the profilers option in flux to "operator"
func configureOperatorProfiler(ctx context.Context) *execute.OperatorProfiler {
	profilerNames := []string{"operator"}

	execOptsConfig := lang.ExecOptsConfig{}
	execOptsConfig.ConfigureProfiler(ctx, profilerNames)

	deps := execute.GetExecutionDependencies(ctx)
	return deps.ExecutionOptions.OperatorProfiler
}

func TestOperatorProfiler_GetResult(t *testing.T) {
	// Create a base execution dependencies.
	deps := execute.DefaultExecutionDependencies()
	ctx := deps.Inject(context.Background())

	// Add operator profiler to context
	p := configureOperatorProfiler(ctx)

	// Build the "want" table.
	// This table is built dynamically because the table includes time data which changes
	// every time the test is run.
	var wantStr bytes.Buffer
	// Need to have the goroutines to sync on their writes to the buffer.
	wantStr.WriteString(`
#datatype,string,long,string,string,string,long,long,long,long,double
#group,false,false,true,false,false,false,false,false,false,false
#default,_profiler,,,,,,,,,
,result,table,_measurement,Type,Label,Count,MinDuration,MaxDuration,DurationSum,MeanDuration
`)
	count := 4
	wg := sync.WaitGroup{}
	wg.Add(count)
	fn := func(opType string, label string, ctx context.Context, offset int) {
		st := time.Date(2020, 10, 14, 12, 30, 0, 0, time.UTC)
		_, span := execute.StartSpanFromContext(ctx, opType, label, opentracing.StartTime(st))
		profilerSpan := span.(*execute.OperatorProfilingSpan)
		// Finish() will write the data to the profiler
		// In Flux runtime, this is called when an execution node finishes execution
		profilerSpan.FinishWithOptions(opentracing.FinishOptions{
			FinishTime: time.Date(2020, 10, 14, 12, 30, 0, 1000+offset, time.UTC),
		})
		wg.Done()
	}
	for i := 0; i < count; i++ {
		op := fmt.Sprintf("type%d", i%2)
		go fn(op, "lab0", ctx, 100+i*i)
		// Waiting between each loop seems to ensure better consistency of row order
		// for final table. However, this is still not guaranteed, becausethe result
		// aggregates are stored in a map, and thus order is not 100% guaranteed.
		time.Sleep(100 * time.Millisecond)
	}
	wantStr.WriteString(fmt.Sprintf(",,0,profiler/operator,%s,%s,%d,%d,%d,%d,%f\n",
		"type0",
		"lab0",
		2,
		1100,
		1104,
		2204,
		1102.0,
	))
	wantStr.WriteString(fmt.Sprintf(",,0,profiler/operator,%s,%s,%d,%d,%d,%d,%f\n",
		"type1",
		"lab0",
		2,
		1101,
		1109,
		2210,
		1105.0,
	))
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

func TestOperatorProfiler_GroupByLabel(t *testing.T) {
	// Create an operator profiler
	p := execute.AllProfilers["operator"]()
	// And inject it to the context.
	ctx := context.WithValue(context.Background(), execute.OperatorProfilerContextKey, p)
	// Build the "want" table.
	// This table is built dynamically because the table includes time data which changes
	// every time the test is run.
	var wantStr bytes.Buffer
	// Need to have the goroutines to sync on their writes to the buffer.
	wantStr.WriteString(`
#datatype,string,long,string,string,string,long,long,long,long,double
#group,false,false,true,false,false,false,false,false,false,false
#default,_profiler,,,,,,,,,
,result,table,_measurement,Type,Label,Count,MinDuration,MaxDuration,DurationSum,MeanDuration
`)
	count := 4
	wg := sync.WaitGroup{}
	wg.Add(count)
	fn := func(opType string, label string, ctx context.Context, offset int) {
		st := time.Date(2020, 10, 14, 12, 30, 0, 0, time.UTC)
		_, span := execute.StartSpanFromContext(ctx, opType, label, opentracing.StartTime(st))
		profilerSpan := span.(*execute.OperatorProfilingSpan)
		// Finish() will write the data to the profiler
		// In Flux runtime, this is called when an execution node finishes execution
		profilerSpan.FinishWithOptions(opentracing.FinishOptions{
			FinishTime: time.Date(2020, 10, 14, 12, 30, 0, 1000+offset, time.UTC),
		})
		wg.Done()
	}
	for i := 0; i < count; i++ {
		label := fmt.Sprintf("lab%d", i%2)
		go fn("type0", label, ctx, 100+i*i)
		// Waiting between each loop seems to ensure better consistency of row order
		// for final table. However, this is still not guaranteed, becausethe result
		// aggregates are stored in a map, and thus order is not 100% guaranteed.
		time.Sleep(100 * time.Millisecond)
	}
	wantStr.WriteString(fmt.Sprintf(",,0,profiler/operator,%s,%s,%d,%d,%d,%d,%f\n",
		"type0",
		"lab0",
		2,
		1100,
		1104,
		2204,
		1102.0,
	))
	wantStr.WriteString(fmt.Sprintf(",,0,profiler/operator,%s,%s,%d,%d,%d,%d,%f\n",
		"type0",
		"lab1",
		2,
		1101,
		1109,
		2210,
		1105.0,
	))
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
