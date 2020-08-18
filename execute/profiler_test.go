package execute_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/metadata"
	"github.com/influxdata/flux/mock"
)

func TestFluxStatisticsProfiler_GetResult(t *testing.T) {
	p := &execute.FluxStatisticsProfiler{}
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
#datatype,string,long,string,string,string
#group,false,false,true,false,false
#default,_profiler,,,,
,result,table,_measurement,_field,_value
,,0,profiler/FluxStatistics,CompileDuration,2ns
,,0,profiler/FluxStatistics,Concurrency,7
,,0,profiler/FluxStatistics,ExecuteDuration,6ns
,,0,profiler/FluxStatistics,MaxAllocated,8
,,0,profiler/FluxStatistics,PlanDuration,4ns
,,0,profiler/FluxStatistics,QueueDuration,3ns
,,0,profiler/FluxStatistics,RequeueDuration,5ns
,,0,profiler/FluxStatistics,RuntimeErrors,"1
2"
,,0,profiler/FluxStatistics,TotalAllocated,9
,,0,profiler/FluxStatistics,TotalDuration,1ns
,,0,profiler/FluxStatistics,flux/query-plan,"query plan"
,,0,profiler/FluxStatistics,influxdb/scanned-bytes,10
,,0,profiler/FluxStatistics,influxdb/scanned-values,11
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
	defer want.Release()

	if err := executetest.EqualResultIterators(want, got); err != nil {
		t.Fatal(err)
	}
}
