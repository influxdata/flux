package socket_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/socket"
	"github.com/influxdata/flux/stdlib/universe"

	_ "github.com/influxdata/flux/builtin" // We need to import the builtins for the tests to work.
)

func TestFromSocket_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from no args",
			Raw: `import "socket"
socket.from()`,
			WantErr: true,
		},
		{
			Name: "from wrong decoder",
			Raw: `import "socket"
socket.from(url: "url", decoder: "wrong")`,
			WantErr: true,
		},
		{
			Name: "from ok",
			Raw: `import "socket"
socket.from(url: "url", decoder: "line") |> range(start:-4h, stop:-2h) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "fromSocket0",
						Spec: &socket.FromSocketOpSpec{
							URL:     "url",
							Decoder: "line",
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
					{Parent: "fromSocket0", Child: "range1"},
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

func TestFromSocketOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"fromSocket","kind":"fromSocket","spec":{"url":"url","decoder":"csv"}}`)
	op := &flux.Operation{
		ID: "fromSocket",
		Spec: &socket.FromSocketOpSpec{
			URL:     "url",
			Decoder: "csv",
		},
	}
	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestFromSocketSource_Run(t *testing.T) {
	testCases := []struct {
		name  string
		spec  *socket.FromSocketProcedureSpec
		input string
		want  []*executetest.Table
	}{
		{
			name: "raw strings",
			spec: &socket.FromSocketProcedureSpec{Decoder: "line"},
			input: `this is
a line
socket
source
`,
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), "this is"},
					{execute.Time(1), "a line"},
					{execute.Time(2), "socket"},
					{execute.Time(3), "source"},
				},
			}},
		},
		{
			name: "csv",
			spec: &socket.FromSocketProcedureSpec{Decoder: "csv"},
			input: `#datatype,string,long,dateTime:RFC3339,string,string,double,boolean
#group,false,false,false,true,true,false,true
#default,,,,,,,
,result,table,_time,tag1,tag2,double,boolean
,,0,1970-01-01T00:00:00Z,a,b,0.42,true
,,0,1970-01-01T00:00:00Z,a,b,0.1,true
,,0,1970-01-01T00:00:00Z,a,b,-0.3,true
,,0,1970-01-01T00:00:00Z,a,b,10.0,true
,,0,1970-01-01T00:00:00Z,a,b,5.33,true
,,1,1970-01-01T00:00:00Z,b,b,0.42,true
,,1,1970-01-01T00:00:00Z,b,b,0.1,true
,,1,1970-01-01T00:00:00Z,b,b,-0.3,true
,,1,1970-01-01T00:00:00Z,b,b,10.0,true
,,1,1970-01-01T00:00:00Z,b,b,5.33,true
,,2,1970-01-01T00:00:00Z,b,b,0.42,false
,,2,1970-01-01T00:00:00Z,b,b,0.1,false
,,2,1970-01-01T00:00:00Z,b,b,-0.3,false
,,2,1970-01-01T00:00:00Z,b,b,10.0,false
,,2,1970-01-01T00:00:00Z,b,b,5.33,false
`,
			want: []*executetest.Table{
				{
					KeyCols: []string{"tag1", "tag2", "boolean"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "double", Type: flux.TFloat},
						{Label: "boolean", Type: flux.TBool},
					},
					Data: [][]interface{}{
						{execute.Time(0), "b", "b", 0.42, false},
						{execute.Time(0), "b", "b", 0.1, false},
						{execute.Time(0), "b", "b", -0.3, false},
						{execute.Time(0), "b", "b", 10.0, false},
						{execute.Time(0), "b", "b", 5.33, false},
					},
				},
				{
					KeyCols: []string{"tag1", "tag2", "boolean"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "double", Type: flux.TFloat},
						{Label: "boolean", Type: flux.TBool},
					},
					Data: [][]interface{}{
						{execute.Time(0), "a", "b", 0.42, true},
						{execute.Time(0), "a", "b", 0.1, true},
						{execute.Time(0), "a", "b", -0.3, true},
						{execute.Time(0), "a", "b", 10.0, true},
						{execute.Time(0), "a", "b", 5.33, true},
					},
				},
				{
					KeyCols: []string{"tag1", "tag2", "boolean"},
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "tag1", Type: flux.TString},
						{Label: "tag2", Type: flux.TString},
						{Label: "double", Type: flux.TFloat},
						{Label: "boolean", Type: flux.TBool},
					},
					Data: [][]interface{}{
						{execute.Time(0), "b", "b", 0.42, true},
						{execute.Time(0), "b", "b", 0.1, true},
						{execute.Time(0), "b", "b", -0.3, true},
						{execute.Time(0), "b", "b", 10.0, true},
						{execute.Time(0), "b", "b", 5.33, true},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			id := executetest.RandomDatasetID()
			d := executetest.NewDataset(id)
			c := execute.NewTableBuilderCache(executetest.UnlimitedAllocator)
			c.SetTriggerSpec(plan.DefaultTriggerSpec)
			r := ioutil.NopCloser(bytes.NewReader([]byte(tc.input)))
			ss, err := socket.NewSocketSource(tc.spec, r, &mock.AscendingTimeProvider{}, id)
			if err != nil {
				t.Fatal(err)
			}

			// Add `yield` in order to add `from` output tables to cache.
			ss.AddTransformation(executetest.NewYieldTransformation(d, c))
			ss.Run(context.Background())

			// Retrieve tables from cache.
			got, err := executetest.TablesFromCache(c)
			if err != nil {
				t.Fatal(err)
			}

			if len(got) < len(tc.want) {
				t.Errorf("wrong number of results want/got %d/%d", len(tc.want), len(got))
			}

			executetest.NormalizeTables(got)
			executetest.NormalizeTables(tc.want)

			if !cmp.Equal(tc.want, got, cmpopts.EquateNaNs()) {
				t.Errorf("unexpected tables -want/+got\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
