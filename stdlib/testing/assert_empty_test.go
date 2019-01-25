package testing_test

import (
	"errors"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	fluxtesting "github.com/influxdata/flux/stdlib/testing"
)

func TestAssertEmpty_Process(t *testing.T) {
	testCases := []struct {
		name    string
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "no tables",
			data: []flux.Table{},
			want: []*executetest.Table(nil),
		},
		{
			name: "all empty tables",
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{},
			}},
			want: []*executetest.Table(nil),
		},
		{
			name: "non-empty table",
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"t1"},
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_value", Type: flux.TFloat},
					{Label: "t1", Type: flux.TString},
					{Label: "t2", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(0), 7.0, "a", "y"},
				},
			}},
			wantErr: errors.New("found 1 tables that were not empty"),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				tc.wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					return fluxtesting.NewAssertEmptyTransformation(d, c)
				},
			)
		})
	}
}
