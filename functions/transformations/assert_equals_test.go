package transformations_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/plan"
	"sort"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/querytest"
)

func TestAssertEqualsOperation_Marshaling(t *testing.T) {
	data := []byte(`{"id":"assertEquals","kind":"assertEquals","spec":{"name":"simple"}}`)
	op := &flux.Operation{
		ID: "assertEquals",
		Spec: &transformations.AssertEqualsOpSpec{
			Name: "simple",
		},
	}

	querytest.OperationMarshalingTestHelper(t, data, op)
}

func TestAssertEquals_Process(t *testing.T) {

	testCases := []struct {
		skip  bool
		name  string
		spec  *transformations.AssertEqualsProcedureSpec
		data0 []*executetest.Table // data from parent 0
		data1 []*executetest.Table // data from parent 1
		want  []*executetest.Table
	}{
		{
			name: "simple equality",
			spec: &transformations.AssertEqualsProcedureSpec{
				DefaultCost: plan.DefaultCost{},
				Name:        "simple",
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
			want: []*executetest.Table{
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
			c.SetTriggerSpec(execute.DefaultTriggerSpec)
			jt := transformations.NewAssertEqualsTransformation(d, c, tc.spec, parents[0], parents[1])

			executetest.NormalizeTables(tc.data0)
			executetest.NormalizeTables(tc.data1)
			l := len(tc.data0)
			if len(tc.data1) > l {
				l = len(tc.data1)
			}
			for i := 0; i < l; i++ {
				if i < len(tc.data0) {
					if err := jt.Process(parents[0], tc.data0[i]); err != nil {
						t.Fatal(err)
					}
				}
				if i < len(tc.data1) {
					if err := jt.Process(parents[1], tc.data1[i]); err != nil {
						t.Fatal(err)
					}
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
