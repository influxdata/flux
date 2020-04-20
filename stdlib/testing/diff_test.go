package testing_test

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/plan"
	fluxtesting "github.com/influxdata/flux/stdlib/testing"
)

func TestDiff_Process(t *testing.T) {

	testCases := []struct {
		skip    bool
		name    string
		spec    *fluxtesting.DiffProcedureSpec
		data0   []*executetest.Table // data from parent 0
		data1   []*executetest.Table // data from parent 1
		want    []*executetest.Table
		wantErr bool
	}{
		{
			name: "same table",
			spec: &fluxtesting.DiffProcedureSpec{
				DefaultCost: plan.DefaultCost{},
			},
			data0: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), 3.0},
					},
				},
			},
			data1: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), 3.0},
					},
				},
			},
			want: []*executetest.Table(nil),
		},
		{
			name: "different values",
			spec: &fluxtesting.DiffProcedureSpec{
				DefaultCost: plan.DefaultCost{},
			},
			data0: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), 3.0},
					},
				},
			},
			data1: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 3.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), 1.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_diff", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"-", execute.Time(1), 1.0},
						{"+", execute.Time(1), 3.0},
						{"-", execute.Time(3), 3.0},
						{"+", execute.Time(3), 1.0},
					},
				},
			},
		},
		{
			name: "mismatched size",
			spec: &fluxtesting.DiffProcedureSpec{
				DefaultCost: plan.DefaultCost{},
			},
			data0: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), 3.0},
					},
				},
			},
			data1: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_diff", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"-", execute.Time(3), 3.0},
					},
				},
			},
		},
		{
			name: "missing from got",
			spec: &fluxtesting.DiffProcedureSpec{
				DefaultCost: plan.DefaultCost{},
			},
			data0: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), 3.0},
					},
				},
			},
			data1: []*executetest.Table{},
			want: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_diff", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"-", execute.Time(1), 1.0},
						{"-", execute.Time(2), 2.0},
						{"-", execute.Time(3), 3.0},
					},
				},
			},
		},
		{
			name: "missing from want",
			spec: &fluxtesting.DiffProcedureSpec{
				DefaultCost: plan.DefaultCost{},
			},
			data0: []*executetest.Table{},
			data1: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), 3.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_diff", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"+", execute.Time(1), 1.0},
						{"+", execute.Time(2), 2.0},
						{"+", execute.Time(3), 3.0},
					},
				},
			},
		},
		{
			name: "float64 comparison large epsilon",
			spec: &fluxtesting.DiffProcedureSpec{
				DefaultCost: plan.DefaultCost{},
				Epsilon:     1e-6,
			},
			data0: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), math.Inf(1)},
					},
				},
			},
			data1: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.000001},
						{execute.Time(2), 2.0},
						{execute.Time(3), math.Inf(1)},
					},
				},
			},
			want: []*executetest.Table(nil),
		},
		{
			name: "float64 comparison default epsilon",
			spec: &fluxtesting.DiffProcedureSpec{
				DefaultCost: plan.DefaultCost{},
			},
			data0: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.0},
						{execute.Time(2), 2.0},
						{execute.Time(3), math.Inf(1)},
					},
				},
			},
			data1: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{execute.Time(1), 1.000001},
						{execute.Time(2), 2.0},
						{execute.Time(3), math.Inf(1)},
					},
				},
			},
			want: []*executetest.Table{
				{
					ColMeta: []flux.ColMeta{
						{Label: "_diff", Type: flux.TString},
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{"-", execute.Time(1), 1.0},
						{"+", execute.Time(1), 1.000001},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}

			id0 := executetest.RandomDatasetID()
			id1 := executetest.RandomDatasetID()

			parents := []execute.DatasetID{
				execute.DatasetID(id0),
				execute.DatasetID(id1),
			}

			d := executetest.NewDataset(executetest.RandomDatasetID())
			c := execute.NewTableBuilderCache(executetest.UnlimitedAllocator)
			c.SetTriggerSpec(plan.DefaultTriggerSpec)
			jt := fluxtesting.NewDiffTransformation(d, c, tc.spec, parents[0], parents[1], executetest.UnlimitedAllocator)

			executetest.NormalizeTables(tc.data0)
			executetest.NormalizeTables(tc.data1)
			l := len(tc.data0)
			if len(tc.data1) > l {
				l = len(tc.data1)
			}
			var err error
			for i := 0; i < l; i++ {
				if i < len(tc.data0) {
					if err = jt.Process(parents[0], tc.data0[i]); err != nil {
						break
					}
				}
				if i < len(tc.data1) {
					if err = jt.Process(parents[1], tc.data1[i]); err != nil {
						break
					}
				}
			}
			jt.Finish(parents[0], err)
			jt.Finish(parents[1], err)

			if tc.wantErr {
				if err == nil {
					t.Fatal(fmt.Errorf("case %s expected an error, got none", tc.name))
				} else {
					return
				}
			}

			got, err := executetest.TablesFromCache(c)
			if err != nil {
				t.Fatal(err)
			}

			executetest.NormalizeTables(got)
			executetest.NormalizeTables(tc.want)

			sort.Sort(executetest.SortedTables(got))
			sort.Sort(executetest.SortedTables(tc.want))

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected tables -want/+got\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
