package execute_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/csv"
	"github.com/InfluxCommunity/flux/execute"
	"github.com/InfluxCommunity/flux/execute/executetest"
	"github.com/InfluxCommunity/flux/execute/table"
	"github.com/InfluxCommunity/flux/lang"
	"github.com/InfluxCommunity/flux/memory"
	"github.com/InfluxCommunity/flux/metadata"
	"github.com/InfluxCommunity/flux/mock"
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
	var wantStr bytes.Buffer
	wantStr.WriteString(`
#datatype,string,long,string,string,string,long,long,long,long,double
#group,false,false,true,false,false,false,false,false,false,false
#default,_profiler,,,,,,,,,
,result,table,_measurement,Type,Label,Count,MinDuration,MaxDuration,DurationSum,MeanDuration
`)
	fmt.Fprintf(&wantStr, ",,0,profiler/operator,%s,%s,%d,%d,%d,%d,%f\n",
		"type0", "lab0", 4, 1000, 1606, 5212, 1303.0,
	)
	fmt.Fprintf(&wantStr, ",,0,profiler/operator,%s,%s,%d,%d,%d,%d,%f\n",
		"type1", "lab0", 4, 1101, 1707, 5616, 1404.0,
	)
	fmt.Fprintf(&wantStr, ",,0,profiler/operator,%s,%s,%d,%d,%d,%d,%f\n",
		"type0", "lab1", 4, 1808, 2414, 8444, 2111.0,
	)
	fmt.Fprintf(&wantStr, ",,0,profiler/operator,%s,%s,%d,%d,%d,%d,%f\n",
		"type1", "lab1", 4, 1909, 2515, 8848, 2212.0,
	)
	count := 16

	stats := flux.Statistics{
		Profiles: make([]flux.TransportProfile, 0, 4),
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			stats.Profiles = append(stats.Profiles, flux.TransportProfile{
				NodeType: fmt.Sprintf("type%d", i),
				Label:    fmt.Sprintf("lab%d", j),
			})
		}
	}

	fn := func(profile *flux.TransportProfile, offset int) {
		st := time.Date(2020, 10, 14, 12, 30, 0, 0, time.UTC)
		span := profile.StartSpan(st)
		span.FinishWithTime(time.Date(2020, 10, 14, 12, 30, 0, 1000+offset, time.UTC))
	}

	// Write profiles for the various different transports.
	for i := 0; i < count; i++ {
		profile := &stats.Profiles[i%2*2+i/8]
		fn(profile, 100*i+i)
	}

	q := &mock.Query{}
	q.SetStatistics(stats)
	q.Done()

	tbl, err := p.GetSortedResult(q, &memory.ResourceAllocator{}, false, "MeanDuration")
	if err != nil {
		t.Error(err)
	}
	result := table.NewProfilerResult(tbl)
	got := flux.NewSliceResultIterator([]flux.Result{&result})
	dec := csv.NewMultiResultDecoder(csv.ResultDecoderConfig{})
	want, e := dec.Decode(io.NopCloser(strings.NewReader(wantStr.String())))
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
2","query plan
",10,11
`
	q.Done()
	tbl, err := p.GetResult(q, &memory.ResourceAllocator{})
	if err != nil {
		t.Error(err)
	}
	result := table.NewProfilerResult(tbl)
	got := flux.NewSliceResultIterator([]flux.Result{&result})
	dec := csv.NewMultiResultDecoder(csv.ResultDecoderConfig{})
	want, e := dec.Decode(io.NopCloser(strings.NewReader(wantStr)))
	if e != nil {
		t.Error(err)
	}
	if err := executetest.EqualResultIterators(want, got); err != nil {
		t.Fatal(err)
	}
}
