package sql_test

import (
	"github.com/influxdata/flux/plan"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	flux "github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin" // We need to import the builtins for the tests to work.
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	fsql "github.com/influxdata/flux/stdlib/sql"
	"github.com/influxdata/flux/values"
)

func TestSqlTo(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with database",
			Raw:  `import "sql" from(bucket: "mybucket") |> sql.to(driverName:"sqlmock", dataSourceName:"root@/db", table:"TestTable")`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: "mybucket",
						},
					},
					{
						ID: "toSQL1",
						Spec: &fsql.ToSQLOpSpec{
							DriverName:     "sqlmock",
							DataSourceName: "root@/db",
							Table:          "TestTable",
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "toSQL1"},
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

func TestToSQL_Process(t *testing.T) {
	driverName := "sqlmock"
	dsn := "root@/db"
	_, _, _ = sqlmock.NewWithDSN(dsn)
	type wanted struct {
		Table  []*executetest.Table
		ColumnNames []string
		ValueStrings [][]string
		ValueArgs [][]interface{}
	}
	testCases := []struct {
		name string
		spec *fsql.ToSQLProcedureSpec
		data flux.Table
		want wanted
	}{
		{
			name: "coltable with name in _measurement",
			spec: &fsql.ToSQLProcedureSpec{
				Spec: &fsql.ToSQLOpSpec{
					DriverName:     driverName,
					DataSourceName: dsn,
					Table:          "TestTable2",
				},
			},
			data: executetest.MustCopyTable(&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TFloat},
					{Label: "fred", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), "a", 2.0, "one"},
					{execute.Time(21), "a", 2.0, "one"},
					{execute.Time(21), "b", 1.0, "seven"},
					{execute.Time(31), "a", 3.0, "nine"},
					{execute.Time(41), "c", 4.0, "elevendyone"},
				},
			}),
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_value", Type: flux.TFloat},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", 2.0, "one"},
						{execute.Time(21), "a", 2.0, "one"},
						{execute.Time(21), "b", 1.0, "seven"},
						{execute.Time(31), "a", 3.0, "nine"},
						{execute.Time(41), "c", 4.0, "elevendyone"},
					},
				}},
				ColumnNames: []string{"_time", "_measurement", "_value", "fred" },
				ValueStrings: [][]string{{"(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)"}},
				ValueArgs:[][]interface{}{{
					values.Time(int64(execute.Time(11))).Time(), "a", 2.0, "one",
					values.Time(int64(execute.Time(21))).Time(), "a", 2.0, "one",
					values.Time(int64(execute.Time(21))).Time(), "b", 1.0, "seven",
					values.Time(int64(execute.Time(31))).Time(), "a", 3.0, "nine",
					values.Time(int64(execute.Time(41))).Time(), "c", 4.0, "elevendyone"}},
			},
		},
		{
			name: "coltable with ints",
			spec: &fsql.ToSQLProcedureSpec{
				Spec: &fsql.ToSQLOpSpec{
					DriverName:     driverName,
					DataSourceName: dsn,
					Table:          "TestTable2",
				},
			},
			data: executetest.MustCopyTable(&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TInt},
					{Label: "fred", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), "a", int64(2), "one"},
					{execute.Time(21), "a", int64(2), "one"},
					{execute.Time(21), "b", int64(1), "seven"},
					{execute.Time(31), "a", int64(3), "nine"},
					{execute.Time(41), "c", int64(4), "elevendyone"},
				},
			}),
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_value", Type: flux.TInt},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", int64(2), "one"},
						{execute.Time(21), "a", int64(2), "one"},
						{execute.Time(21), "b", int64(1), "seven"},
						{execute.Time(31), "a", int64(3), "nine"},
						{execute.Time(41), "c", int64(4), "elevendyone"},
					},
				}},
				ColumnNames: []string{"_time", "_measurement", "_value", "fred" },
				ValueStrings: [][]string{{"(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)"}},
				ValueArgs: [][]interface{}{{
					values.Time(int64(execute.Time(11))).Time(), "a", int64(2), "one",
					values.Time(int64(execute.Time(21))).Time(), "a", int64(2), "one",
					values.Time(int64(execute.Time(21))).Time(), "b", int64(1), "seven",
					values.Time(int64(execute.Time(31))).Time(), "a", int64(3), "nine",
					values.Time(int64(execute.Time(41))).Time(), "c", int64(4), "elevendyone"}},
			},
		},
		{
			name: "coltable with uints",
			spec: &fsql.ToSQLProcedureSpec{
				Spec: &fsql.ToSQLOpSpec{
					DriverName:     driverName,
					DataSourceName: dsn,
					Table:          "TestTable2",
				},
			},
			data: executetest.MustCopyTable(&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TUInt},
					{Label: "fred", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), "a", uint64(2), "one"},
					{execute.Time(21), "a", uint64(2), "one"},
					{execute.Time(21), "b", uint64(1), "seven"},
					{execute.Time(31), "a", uint64(3), "nine"},
					{execute.Time(41), "c", uint64(4), "elevendyone"},
				},
			}),
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_value", Type: flux.TUInt},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", uint64(2), "one"},
						{execute.Time(21), "a", uint64(2), "one"},
						{execute.Time(21), "b", uint64(1), "seven"},
						{execute.Time(31), "a", uint64(3), "nine"},
						{execute.Time(41), "c", uint64(4), "elevendyone"},
					},
				}},
				ColumnNames: []string{"_time", "_measurement", "_value", "fred" },
				ValueStrings: [][]string{{"(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)"}},
				ValueArgs: [][]interface{}{{
					values.Time(int64(execute.Time(11))).Time(), "a", uint64(2), "one",
					values.Time(int64(execute.Time(21))).Time(), "a", uint64(2), "one",
					values.Time(int64(execute.Time(21))).Time(), "b", uint64(1), "seven",
					values.Time(int64(execute.Time(31))).Time(), "a", uint64(3), "nine",
					values.Time(int64(execute.Time(41))).Time(), "c", uint64(4), "elevendyone"}},
			},
		},
		{
			name: "coltable with bool",
			spec: &fsql.ToSQLProcedureSpec{
				Spec: &fsql.ToSQLOpSpec{
					DriverName:     driverName,
					DataSourceName: dsn,
					Table:          "TestTable2",
				},
			},
			data: executetest.MustCopyTable(&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "_time", Type: flux.TTime},
					{Label: "_measurement", Type: flux.TString},
					{Label: "_value", Type: flux.TBool},
					{Label: "fred", Type: flux.TString},
				},
				Data: [][]interface{}{
					{execute.Time(11), "a", true, "one"},
					{execute.Time(21), "a", true, "one"},
					{execute.Time(21), "b", false, "seven"},
					{execute.Time(31), "a", true, "nine"},
					{execute.Time(41), "c", false, "elevendyone"},
				},
			}),
			want: wanted{
				Table: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_measurement", Type: flux.TString},
						{Label: "_value", Type: flux.TBool},
						{Label: "fred", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(11), "a", true, "one"},
						{execute.Time(21), "a", true, "one"},
						{execute.Time(21), "b", false, "seven"},
						{execute.Time(31), "a", true, "nine"},
						{execute.Time(41), "c", false, "elevendyone"},
					},
				}},
				ColumnNames: []string{"_time", "_measurement", "_value", "fred" },
				ValueStrings: [][]string{{"(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)"}},
				ValueArgs: [][]interface{}{{
					values.Time(int64(execute.Time(11))).Time(), "a", true, "one",
					values.Time(int64(execute.Time(21))).Time(), "a", true, "one",
					values.Time(int64(execute.Time(21))).Time(), "b", false, "seven",
					values.Time(int64(execute.Time(31))).Time(), "a", true, "nine",
					values.Time(int64(execute.Time(41))).Time(), "c", false, "elevendyone"}},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := executetest.NewDataset(executetest.RandomDatasetID())
			c := execute.NewTableBuilderCache(executetest.UnlimitedAllocator)
			c.SetTriggerSpec(plan.DefaultTriggerSpec)

			transformation := fsql.NewToSQLTransformation(d, c, tc.spec)
			a := tc.data
			colNames, valStrings, valArgs, err := fsql.CreateInsertComponents(transformation, a)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(tc.want.ColumnNames, colNames, cmpopts.EquateNaNs()) {
				t.Log(cmp.Diff(tc.want.ColumnNames, colNames))
				t.Fail()
			}
			if !cmp.Equal(tc.want.ValueStrings, valStrings, cmpopts.EquateNaNs()) {
				t.Log(cmp.Diff(tc.want.ValueStrings, valStrings))
				t.Fail()
			}
			if !cmp.Equal(tc.want.ValueArgs, valArgs, cmpopts.EquateNaNs()) {
				t.Log(cmp.Diff(tc.want.ValueArgs, valArgs))
				t.Fail()
			}
		})
	}
}
