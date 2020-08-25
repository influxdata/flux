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
	defer want.Release()

	if err := executetest.EqualResultIterators(want, got); err != nil {
		t.Fatal(err)
	}
}
