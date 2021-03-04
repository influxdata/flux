package csv_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	_ "github.com/influxdata/flux/fluxinit/static" // We need to init flux for the tests to work.
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/csv"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
)

func TestFromCSV_ReturnSingleResult(t *testing.T) {
	tests := []struct {
		name    string
		spec    *csv.FromCSVProcedureSpec
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "happy path",
			spec: &csv.FromCSVProcedureSpec{
				CSV: `#datatype,string,long,string,string,string,string,double,dateTime:RFC3339
#group,false,false,true,true,true,true,false,false
#default,cpu,,,,,,,
,result,table,_field,_measurement,cpu,host,_value,_time
,,0,usage_guest,cpu,cpu-total,ip-192-168-1-16.ec2.internal,0,2020-10-09T22:18:00Z
`,
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_field", "_measurement", "cpu", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_field", Type: flux.TString},
						{Label: "_measurement", Type: flux.TString},
						{Label: "cpu", Type: flux.TString},
						{Label: "host", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_time", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{
							"usage_guest",
							"cpu",
							"cpu-total",
							"ip-192-168-1-16.ec2.internal",
							0.0,
							values.ConvertTime(time.Date(2020, 10, 9, 22, 18, 0, 0, time.UTC)),
						},
					},
				},
			},
		},
		{
			name: "same result ids",
			spec: &csv.FromCSVProcedureSpec{
				CSV: `#datatype,string,long,string,string,string,string,double,dateTime:RFC3339
#group,false,false,true,true,true,true,false,false
#default,cpu,,,,,,,
,result,table,_field,_measurement,cpu,host,_value,_time
,,0,usage_guest,cpu,cpu-total,ip-192-168-1-16.ec2.internal,0,2020-10-09T22:18:00Z

#datatype,string,long,string,string,string,string,string,string,string,double,dateTime:RFC3339
#group,false,false,true,true,true,true,true,true,true,false,false
#default,cpu,,,,,,,,,,
,result,table,_field,_measurement,device,fstype,host,mode,path,_value,_time
,,3,inodes_total,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4878333294,2020-10-09T22:19:00Z
`,
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"_field", "_measurement", "cpu", "host"},
					ColMeta: []flux.ColMeta{
						{Label: "_field", Type: flux.TString},
						{Label: "_measurement", Type: flux.TString},
						{Label: "cpu", Type: flux.TString},
						{Label: "host", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "_time", Type: flux.TTime},
					},
					Data: [][]interface{}{
						{
							"usage_guest",
							"cpu",
							"cpu-total",
							"ip-192-168-1-16.ec2.internal",
							0.0,
							values.ConvertTime(time.Date(2020, 10, 9, 22, 18, 0, 0, time.UTC)),
						},
					},
				},
				{
					KeyCols: []string{"_field", "_measurement", "device", "fstype", "host", "mode", "path"},
					ColMeta: []flux.ColMeta{
						{Label: "_field", Type: flux.TString},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
						{Label: "device", Type: flux.TString},
						{Label: "fstype", Type: flux.TString},
						{Label: "host", Type: flux.TString},
						{Label: "mode", Type: flux.TString},
						{Label: "path", Type: flux.TString},
					},
					Data: [][]interface{}{
						{
							"inodes_total",
							"disk",
							values.ConvertTime(time.Date(2020, 10, 9, 22, 19, 0, 0, time.UTC)),
							4878333294.0,
							"disk1s1",
							"apfs",
							"ip-192-168-1-16.ec2.internal",
							"rw",
							"/System/Volumes/Data",
						},
					},
				},
			},
		},
		{
			name: "different result ids",
			spec: &csv.FromCSVProcedureSpec{
				CSV: `#datatype,string,long,string,string,string,string,double,dateTime:RFC3339
#group,false,false,true,true,true,true,false,false
#default,cpu,,,,,,,
,result,table,_field,_measurement,cpu,host,_value,_time
,,0,usage_guest,cpu,cpu-total,ip-192-168-1-16.ec2.internal,0,2020-10-09T22:18:00Z

#datatype,string,long,string,string,string,string,string,string,string,double,dateTime:RFC3339
#group,false,false,true,true,true,true,true,true,true,false,false
#default,disk,,,,,,,,,,
,result,table,_field,_measurement,device,fstype,host,mode,path,_value,_time
,,3,inodes_free,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4878333294,2020-10-09T22:18:00Z
,,3,inodes_free,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4878333286,2020-10-09T22:19:00Z
,,3,inodes_free,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4878333253.4,2020-10-09T22:19:44.191958Z
,,4,inodes_total,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4882452840,2020-10-09T22:18:00Z
,,4,inodes_total,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4882452840,2020-10-09T22:19:00Z
,,4,inodes_total,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4882452840,2020-10-09T22:19:44.191958Z
,,5,inodes_used,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4119546,2020-10-09T22:18:00Z
,,5,inodes_used,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4119554,2020-10-09T22:19:00Z
,,5,inodes_used,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4119586.6,2020-10-09T22:19:44.191958Z
`,
			},
			wantErr: errors.New(
				codes.FailedPrecondition,
				"error in csv.from(): csv.from() can only parse 1 result",
			),
		},
		{
			name: "no results",
			spec: &csv.FromCSVProcedureSpec{
				CSV: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executetest.RunSourceHelper(t,
				test.want,
				test.wantErr,
				func(id execute.DatasetID) execute.Source {
					a := mock.AdministrationWithContext(context.Background())
					s, err := csv.CreateSource(test.spec, id, a)
					if err != nil {
						t.Fatal(err)
					}
					return s
				},
			)
		})
	}
}

func TestFromCSV_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "from no args",
			Raw:     `import "csv" csv.from()`,
			WantErr: true,
		},
		{
			Name:    "from conflicting args",
			Raw:     `import "csv" csv.from(csv:"d", file:"b")`,
			WantErr: true,
		},
		{
			Name:    "from repeat arg",
			Raw:     `import "csv" csv.from(csv:"telegraf", csv:"oops")`,
			WantErr: true,
		},
		{
			Name:    "from",
			Raw:     `import "csv" csv.from(csv:"telegraf", chicken:"what is this?")`,
			WantErr: true,
		},
		{
			Name: "fromCSV text",
			Raw:  `import "csv" csv.from(csv: "1,2") |> range(start:-4h, stop:-2h) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "fromCSV0",
						Spec: &csv.FromCSVOpSpec{
							CSV:  "1,2",
							Mode: "annotations",
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start: flux.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: flux.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "fromCSV0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestFromCSVOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"fromCSV","kind":"fromCSV","spec":{"csv":"1,2","mode":"annotations"}}`)
	op := &flux.Operation{
		ID: "fromCSV",
		Spec: &csv.FromCSVOpSpec{
			CSV:  "1,2",
			Mode: "annotations",
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFromCSV_Run(t *testing.T) {
	spec := &csv.FromCSVProcedureSpec{
		CSV: `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,false,true,true,false
#default,_result,,,,,,,
,result,table,_start,_stop,_time,_measurement,host,_value
,,0,2018-04-17T00:00:00Z,2018-04-17T00:05:00Z,2018-04-17T00:00:00Z,cpu,A,42
,,0,2018-04-17T00:00:00Z,2018-04-17T00:05:00Z,2018-04-17T00:00:01Z,cpu,A,43
,,1,2018-04-17T00:05:00Z,2018-04-17T00:10:00Z,2018-04-17T00:06:00Z,mem,A,52
,,1,2018-04-17T00:05:00Z,2018-04-17T00:10:00Z,2018-04-17T00:07:01Z,mem,A,53
`,
	}
	want := []*executetest.Table{
		{
			KeyCols: []string{"_start", "_stop", "_measurement", "host"},
			ColMeta: []flux.ColMeta{
				{Label: "_start", Type: flux.TTime},
				{Label: "_stop", Type: flux.TTime},
				{Label: "_time", Type: flux.TTime},
				{Label: "_measurement", Type: flux.TString},
				{Label: "host", Type: flux.TString},
				{Label: "_value", Type: flux.TFloat},
			},
			Data: [][]interface{}{
				{
					values.ConvertTime(time.Date(2018, 4, 17, 0, 0, 0, 0, time.UTC)),
					values.ConvertTime(time.Date(2018, 4, 17, 0, 5, 0, 0, time.UTC)),
					values.ConvertTime(time.Date(2018, 4, 17, 0, 0, 0, 0, time.UTC)),
					"cpu",
					"A",
					42.0,
				},
				{
					values.ConvertTime(time.Date(2018, 4, 17, 0, 0, 0, 0, time.UTC)),
					values.ConvertTime(time.Date(2018, 4, 17, 0, 5, 0, 0, time.UTC)),
					values.ConvertTime(time.Date(2018, 4, 17, 0, 0, 1, 0, time.UTC)),
					"cpu",
					"A",
					43.0,
				},
			},
		},
		{
			KeyCols: []string{"_start", "_stop", "_measurement", "host"},
			ColMeta: []flux.ColMeta{
				{Label: "_start", Type: flux.TTime},
				{Label: "_stop", Type: flux.TTime},
				{Label: "_time", Type: flux.TTime},
				{Label: "_measurement", Type: flux.TString},
				{Label: "host", Type: flux.TString},
				{Label: "_value", Type: flux.TFloat},
			},
			Data: [][]interface{}{
				{
					values.ConvertTime(time.Date(2018, 4, 17, 0, 5, 0, 0, time.UTC)),
					values.ConvertTime(time.Date(2018, 4, 17, 0, 10, 0, 0, time.UTC)),
					values.ConvertTime(time.Date(2018, 4, 17, 0, 6, 0, 0, time.UTC)),
					"mem",
					"A",
					52.0,
				},
				{
					values.ConvertTime(time.Date(2018, 4, 17, 0, 5, 0, 0, time.UTC)),
					values.ConvertTime(time.Date(2018, 4, 17, 0, 10, 0, 0, time.UTC)),
					values.ConvertTime(time.Date(2018, 4, 17, 0, 7, 1, 0, time.UTC)),
					"mem",
					"A",
					53.0,
				},
			},
		},
	}
	executetest.RunSourceHelper(t,
		want,
		nil,
		func(id execute.DatasetID) execute.Source {
			a := mock.AdministrationWithContext(context.Background())
			s, err := csv.CreateSource(spec, id, a)
			if err != nil {
				t.Fatal(err)
			}
			return s
		},
	)
}

func TestFromCSV_RunCancel(t *testing.T) {
	var csvTextBuilder strings.Builder
	csvTextBuilder.WriteString(`#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,false,true,true,false
#default,_result,,,,,,,
,result,table,_start,_stop,_time,_measurement,host,_value
`)
	for i := 0; i < 1500; i++ {
		// The csv must contain over 1000 rows so that we are triggering multiple buffers.
		csvTextBuilder.WriteString(",,0,2018-04-17T00:00:00Z,2018-04-17T00:05:00Z,2018-04-17T00:00:00Z,cpu,A,42\n")
	}
	spec := &csv.FromCSVProcedureSpec{
		CSV: csvTextBuilder.String(),
	}

	id := executetest.RandomDatasetID()
	a := mock.AdministrationWithContext(context.Background())
	s, err := csv.CreateSource(spec, id, a)
	if err != nil {
		t.Fatal(err)
	}
	// Add a do-nothing transformation which never reads our table.
	// We want to produce a table and send it with the expectation
	// that our source might not read it.
	s.AddTransformation(&noopTransformation{})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	// Canceling the context should free the runner.
	cancel()

	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		t.Fatal("csv.from did not cancel when the context was terminated")
	case <-done:
	}
}

type noopTransformation struct {
	execute.ExecutionNode
}

func (n *noopTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error { return nil }
func (n *noopTransformation) Process(id execute.DatasetID, tbl flux.Table) error         { return nil }
func (n *noopTransformation) UpdateWatermark(id execute.DatasetID, t execute.Time) error { return nil }
func (n *noopTransformation) UpdateProcessingTime(id execute.DatasetID, t execute.Time) error {
	return nil
}
func (n *noopTransformation) Finish(id execute.DatasetID, err error) {}
