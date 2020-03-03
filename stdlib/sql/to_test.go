package sql_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin" // We need to import the builtins for the tests to work.
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	fsql "github.com/influxdata/flux/stdlib/sql"
	"github.com/influxdata/flux/values"
	_ "github.com/mattn/go-sqlite3"
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
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "toSQL1",
						Spec: &fsql.ToSQLOpSpec{
							DriverName:     "sqlmock",
							DataSourceName: "root@/db",
							Table:          "TestTable",
							BatchSize:      fsql.DefaultBatchSize,
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
		Table        []*executetest.Table
		ColumnNames  []string
		ValueStrings [][]string
		ValueArgs    [][]interface{}
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
				ColumnNames:  []string{"_time", "_measurement", "_value", "fred"},
				ValueStrings: [][]string{{"(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)"}},
				ValueArgs: [][]interface{}{{
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
				ColumnNames:  []string{"_time", "_measurement", "_value", "fred"},
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
				ColumnNames:  []string{"_time", "_measurement", "_value", "fred"},
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
				ColumnNames:  []string{"_time", "_measurement", "_value", "fred"},
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

			transformation, err := fsql.NewToSQLTransformation(d, dependenciestest.Default(), c, tc.spec)
			if err != nil {
				t.Fatal(err)
			}

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

func TestToSql_NewTransformation(t *testing.T) {
	test := executetest.TfUrlValidationTest{
		CreateFn: func(d execute.Dataset, deps flux.Dependencies, cache execute.TableBuilderCache,
			spec plan.ProcedureSpec) (execute.Transformation, error) {
			return fsql.NewToSQLTransformation(d, deps, cache, spec.(*fsql.ToSQLProcedureSpec))
		},
		Cases: []executetest.TfUrlValidationTestCase{
			{
				Name: "ok mysql",
				Spec: &fsql.ToSQLProcedureSpec{
					Spec: &fsql.ToSQLOpSpec{
						DriverName:     "mysql",
						DataSourceName: "username:password@tcp(localhost:12345)/dbname?param=value",
					},
				},
				WantErr: "connection refused",
			}, {
				Name: "ok postgres",
				Spec: &fsql.ToSQLProcedureSpec{
					Spec: &fsql.ToSQLOpSpec{
						DriverName:     "postgres",
						DataSourceName: "postgres://pqgotest:password@localhost:12345/pqgotest?sslmode=verify-full",
					},
				},
				WantErr: "connection refused",
			}, {
				Name: "invalid driver",
				Spec: &fsql.ToSQLProcedureSpec{
					Spec: &fsql.ToSQLOpSpec{
						DriverName:     "voltdb",
						DataSourceName: "voltdb://pqgotest:password@localhost:12345/pqgotest?sslmode=verify-full",
					},
				},
				WantErr: "sql driver voltdb not supported",
			}, {
				Name: "no such host",
				Spec: &fsql.ToSQLProcedureSpec{
					Spec: &fsql.ToSQLOpSpec{
						DriverName:     "mysql",
						DataSourceName: "username:password@tcp(notfound:12345)/dbname?param=value",
					},
				},
				WantErr: "no such host",
			}, {
				Name: "private ip",
				Spec: &fsql.ToSQLProcedureSpec{
					Spec: &fsql.ToSQLOpSpec{
						DriverName:     "mysql",
						DataSourceName: "username:password@tcp(localhost:12345)/dbname?param=value",
					},
				},
				Validator: url.PrivateIPValidator{},
				WantErr:   "url is not valid, it connects to a private IP",
			},
		},
	}
	test.Run(t)
}
func TestSqlite3To(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from with database",
			Raw:  `import "sql" from(bucket: "mybucket") |> sql.to(driverName:"sqlite3", dataSourceName:"file::memory:", table:"TestTable", batchSize:10000)`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "toSQL1",
						Spec: &fsql.ToSQLOpSpec{
							DriverName:     "sqlite3",
							DataSourceName: "file::memory:",
							Table:          "TestTable",
							BatchSize:      10000,
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

func TestToSQLite3_Process(t *testing.T) {
	driverName := "sqlite3"
	// use the in-memory mode - so we can test the functionality of the "type interactions" between driver and flux without needing an underlying FS
	dsn := "file::memory:"
	_, _, _ = sqlmock.NewWithDSN(dsn)
	type wanted struct {
		Table        []*executetest.Table
		ColumnNames  []string
		ValueStrings [][]string
		ValueArgs    [][]interface{}
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
					BatchSize:      10000,
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
				ColumnNames:  []string{"_time", "_measurement", "_value", "fred"},
				ValueStrings: [][]string{{"(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)", "(?,?,?,?)"}},
				ValueArgs: [][]interface{}{{
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
					BatchSize:      10000,
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
				ColumnNames:  []string{"_time", "_measurement", "_value", "fred"},
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
					BatchSize:      10000,
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
				ColumnNames:  []string{"_time", "_measurement", "_value", "fred"},
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
					BatchSize:      10000,
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
				ColumnNames:  []string{"_time", "_measurement", "_value", "fred"},
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

			transformation, err := fsql.NewToSQLTransformation(d, dependenciestest.Default(), c, tc.spec)
			if err != nil {
				t.Fatal(err)
			}

			a := tc.data
			colNames, valStrings, valArgs, err := fsql.CreateInsertComponents(transformation, a)
			if tc.name == "coltable with bool" {
				// sqlite doesn't have a BOOL type, so let user know, do not perform implicit type conversion
				if err == nil {
					t.Fatal(err)
				}
				if err.Error() != "SQLite does not support column type bool" {
					t.Fatal(err)
				}
			} else {
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
			}
		})
	}
}
