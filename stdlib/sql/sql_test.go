package sql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

type MockRowReader struct {
	Cursor      *sql.Rows
	columns     []interface{}
	columnTypes []flux.ColType
	columnNames []string
	row         int
}

func (m *MockRowReader) Next() bool {
	if m.row < 2 {
		m.row++
		return true
	}
	return false
}

func (m *MockRowReader) GetNextRow() ([]values.Value, error) {
	timestamp, _ := values.ParseTime("2019-06-03 13:59:01")
	if m.row == 1 {
		return []values.Value{values.NewInt(42), values.NewFloat(42.0), values.NewBool(true), values.NewTime(timestamp)}, nil
	} else if m.row == 2 {
		return []values.Value{
			values.NewNull(flux.SemanticType(flux.TInt)),
			values.NewNull(flux.SemanticType(flux.TFloat)),
			values.NewNull(flux.SemanticType(flux.TBool)),
			values.NewNull(flux.SemanticType(flux.TTime))}, nil
	}
	return nil, fmt.Errorf("no more rows")
}

func (m *MockRowReader) InitColumnNames(s []string) {
	m.columnNames = s
}

func (m *MockRowReader) InitColumnTypes(c []*sql.ColumnType) {
	m.columnTypes = []flux.ColType{flux.TInt, flux.TFloat, flux.TBool, flux.TTime}
}

func (m *MockRowReader) ColumnNames() []string {
	return []string{"int", "float", "bool", "timestamp"}
}

func (m *MockRowReader) ColumnTypes() []flux.ColType {
	return m.columnTypes
}

func (m *MockRowReader) SetColumns(i []interface{}) {
	m.columns = i
}

func (m *MockRowReader) Close() error {
	return nil
}

func TestFromRowReader(t *testing.T) {
	t.Run("Mock RowReader", func(t *testing.T) {

		var rr execute.RowReader = &MockRowReader{row: 0}
		rr.(*MockRowReader).InitColumnTypes(nil)
		alloc := &memory.Allocator{}
		table, err := read(context.Background(), rr, alloc)
		if err != nil {
			t.Fatal(err)
		}

		timestamp, _ := values.ParseTime("2019-06-03 13:59:01")

		want := &executetest.Table{
			ColMeta: []flux.ColMeta{
				{Label: "int", Type: flux.TInt},
				{Label: "float", Type: flux.TFloat},
				{Label: "bool", Type: flux.TBool},
				{Label: "timestamp", Type: flux.TTime},
			},
			Data: [][]interface{}{
				{int64(42), 42.0, true, timestamp},
				{nil, nil, nil, nil},
			},
		}

		firstRow := values.NewObject()
		firstRow.Set("int", values.NewInt(42))
		firstRow.Set("float", values.NewFloat(42.0))
		firstRow.Set("bool", values.NewBool(true))
		firstRow.Set("timestamp", values.NewTime(timestamp))

		secondRow := values.NewObject()
		secondRow.Set("int", values.NewNull(flux.SemanticType(flux.TInt)))
		secondRow.Set("float", values.NewNull(flux.SemanticType(flux.TFloat)))
		secondRow.Set("bool", values.NewNull(flux.SemanticType(flux.TBool)))
		secondRow.Set("timestamp", values.NewNull(flux.SemanticType(flux.TTime)))

		if !cmp.Equal(want.Cols(), table.Cols()) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(want.Cols(), table.Cols()))
		}
		if !cmp.Equal(want.Key(), table.Key()) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(want.Key(), table.Key()))
		}
		if !cmp.Equal([]flux.ColMeta(nil), table.Key().Cols()) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff([]flux.ColMeta(nil), table.Key().Cols()))
		}

		buffer := execute.NewColListTableBuilder(table.Key(), executetest.UnlimitedAllocator)
		if err := execute.AddTableCols(table, buffer); err != nil {
			t.Fatal(err)
		}
		if err := execute.AppendTable(table, buffer); err != nil {
			t.Fatal(err)
		}

		wantBuffer := execute.NewColListTableBuilder(want.Key(), executetest.UnlimitedAllocator)
		if err := execute.AddTableCols(want, wantBuffer); err != nil {
			t.Fatal(err)
		}
		if err := execute.AppendTable(want, wantBuffer); err != nil {
			t.Fatal(err)
		}

		for i := 0; i < 2; i++ {
			want := wantBuffer.GetRow(i)
			got := buffer.GetRow(i)
			// the second row has a lot of nil values.Value which cannot pass values.Value.Equals() check.
			if !(i == 0 && got.Equal(want)) &&
				!(i == 1 && got.(fmt.Stringer).String() == want.(fmt.Stringer).String()) {
				t.Fatalf("unexpected result -want/+got:\n%s", cmp.Diff(want, got))
			}
		}

	})
}

func TestMySqlParsing(t *testing.T) {
	// here we want to build a mocked representation of what's in our MySql db, and then run our RowReader over it, then verify that the results
	// are as expected.
	// NOTE: no meaningful test for reading bools, because the DB doesn't support them, and we already know that we can read INT types
	testCases := []struct {
		name       string
		columnName string
		data       *sql.Rows
		want       [][]values.Value
	}{
		{
			name:       "ints",
			columnName: "_int",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(int64(6)).AddRow(int64(1)).AddRow(int64(643)).AddRow(int64(42)).AddRow(int64(1283))),
			want:       [][]values.Value{{values.NewInt(6)}, {values.NewInt(1)}, {values.NewInt(643)}, {values.NewInt(42)}, {values.NewInt(1283)}},
		},
		{
			name:       "floats",
			columnName: "_float",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(float64(6)).AddRow(float64(1)).AddRow(float64(643)).AddRow(float64(42)).AddRow(float64(1283))),
			want:       [][]values.Value{{values.NewFloat(6)}, {values.NewFloat(1)}, {values.NewFloat(643)}, {values.NewFloat(42)}, {values.NewFloat(1283)}},
		},
		{
			name:       "strings",
			columnName: "_string",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(string("6")).AddRow(string("1")).AddRow(string("643")).AddRow(string("42")).AddRow(string("1283"))),
			want:       [][]values.Value{{values.NewString("6")}, {values.NewString("1")}, {values.NewString("643")}, {values.NewString("42")}, {values.NewString("1283")}},
		},
		{
			name:       "datetime",
			columnName: "_datetime",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(createTestTimes()[0].(time.Time)).AddRow(createTestTimes()[1].(time.Time)).AddRow(createTestTimes()[2].(time.Time)).AddRow(createTestTimes()[3].(time.Time)).AddRow(createTestTimes()[4].(time.Time))),
			want: [][]values.Value{
				{values.NewTime(values.ConvertTime(createTestTimes()[0].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[1].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[2].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[3].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[4].(time.Time)))},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			TestReader, err := NewMySQLRowReader(tc.data)
			if !cmp.Equal(nil, err) {
				t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
			}
			i := 0
			for TestReader.Next() {
				row, _ := TestReader.GetNextRow()
				if !cmp.Equal(tc.want[i], row) {
					t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(tc.want[i], row))
				}
				i++
			}
		})
	}
}

func TestPostgresParsing(t *testing.T) {
	// here we want to build a mocked representation of what's in our Postgres db, and then run our RowReader over it, then verify that the results
	// are as expected
	testCases := []struct {
		name       string
		columnName string
		data       *sql.Rows
		want       [][]values.Value
	}{
		{
			name:       "ints",
			columnName: "_int",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(int64(6)).AddRow(int64(1)).AddRow(int64(643)).AddRow(int64(42)).AddRow(int64(1283))),
			want:       [][]values.Value{{values.NewInt(6)}, {values.NewInt(1)}, {values.NewInt(643)}, {values.NewInt(42)}, {values.NewInt(1283)}},
		},
		{
			name:       "floats",
			columnName: "_float",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(float64(6)).AddRow(float64(1)).AddRow(float64(643)).AddRow(float64(42)).AddRow(float64(1283))),
			want:       [][]values.Value{{values.NewFloat(6)}, {values.NewFloat(1)}, {values.NewFloat(643)}, {values.NewFloat(42)}, {values.NewFloat(1283)}},
		},
		{
			name:       "strings",
			columnName: "_string",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(string("6")).AddRow(string("1")).AddRow(string("643")).AddRow(string("42")).AddRow(string("1283"))),
			want:       [][]values.Value{{values.NewString("6")}, {values.NewString("1")}, {values.NewString("643")}, {values.NewString("42")}, {values.NewString("1283")}},
		},
		{
			name:       "bools",
			columnName: "_bools",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(bool(true)).AddRow(bool(false)).AddRow(bool(true)).AddRow(bool(false)).AddRow(bool(true))),
			want:       [][]values.Value{{values.NewBool(true)}, {values.NewBool(false)}, {values.NewBool(true)}, {values.NewBool(false)}, {values.NewBool(true)}},
		},
		{
			name:       "datetime",
			columnName: "_datetime",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(createTestTimes()[0].(time.Time)).AddRow(createTestTimes()[1].(time.Time)).AddRow(createTestTimes()[2].(time.Time)).AddRow(createTestTimes()[3].(time.Time)).AddRow(createTestTimes()[4].(time.Time))),
			want: [][]values.Value{
				{values.NewTime(values.ConvertTime(createTestTimes()[0].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[1].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[2].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[3].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[4].(time.Time)))},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			TestReader, err := NewPostgresRowReader(tc.data)
			if !cmp.Equal(nil, err) {
				t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
			}
			i := 0
			for TestReader.Next() {
				row, _ := TestReader.GetNextRow()
				if !cmp.Equal(tc.want[i], row) {
					t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(tc.want[i], row))
				}
				i++
			}
		})
	}
}

func TestSQLiteParsing(t *testing.T) {
	// here we want to build a mocked representation of what's in our SQLite db, and then run our RowReader over it, then verify that the results
	// are as expected.
	// NOTE: no meaningful test for reading bools, because the DB doesn't support them, and we already know that we can read INT types
	testCases := []struct {
		name       string
		columnName string
		data       *sql.Rows
		want       [][]values.Value
	}{
		{
			name:       "ints",
			columnName: "_int",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(int64(6)).AddRow(int64(1)).AddRow(int64(643)).AddRow(int64(42)).AddRow(int64(1283))),
			want:       [][]values.Value{{values.NewInt(6)}, {values.NewInt(1)}, {values.NewInt(643)}, {values.NewInt(42)}, {values.NewInt(1283)}},
		},
		{
			name:       "floats",
			columnName: "_float",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(float64(6)).AddRow(float64(1)).AddRow(float64(643)).AddRow(float64(42)).AddRow(float64(1283))),
			want:       [][]values.Value{{values.NewFloat(6)}, {values.NewFloat(1)}, {values.NewFloat(643)}, {values.NewFloat(42)}, {values.NewFloat(1283)}},
		},
		{
			name:       "strings",
			columnName: "_string",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(string("6")).AddRow(string("1")).AddRow(string("643")).AddRow(string("42")).AddRow(string("1283"))),
			want:       [][]values.Value{{values.NewString("6")}, {values.NewString("1")}, {values.NewString("643")}, {values.NewString("42")}, {values.NewString("1283")}},
		},
		{
			name:       "datetime",
			columnName: "_datetime",
			data:       mockRowsToSQLRows(sqlmock.NewRows([]string{"column"}).AddRow(createTestTimes()[0].(time.Time)).AddRow(createTestTimes()[1].(time.Time)).AddRow(createTestTimes()[2].(time.Time)).AddRow(createTestTimes()[3].(time.Time)).AddRow(createTestTimes()[4].(time.Time))),
			want: [][]values.Value{
				{values.NewTime(values.ConvertTime(createTestTimes()[0].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[1].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[2].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[3].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[4].(time.Time)))},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			TestReader, err := NewSqliteRowReader(tc.data)
			if !cmp.Equal(nil, err) {
				t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
			}
			i := 0
			for TestReader.Next() {
				row, _ := TestReader.GetNextRow()
				if !cmp.Equal(tc.want[i], row) {
					t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(tc.want[i], row))
				}
				i++
			}
		})
	}
}

func createTestTimes() []interface{} {
	str := []string{"2019-06-03 13:59:00",
		"2019-06-03 13:59:01",
		"2019-06-03 13:59:02",
		"2019-06-03 13:59:03",
		"2019-06-03 13:59:04",
		"2019-06-03 13:59:05",
		"2019-06-03 13:59:06",
		"2019-06-03 13:59:07"}

	a := make([]interface{}, len(str))
	for i, b := range str {
		t, _ := time.Parse(layout, string(b))
		a[i] = t
	}
	return a
}

// kind of abusing the functionality here, but it works well for our purpose
func mockRowsToSQLRows(mockedRows *sqlmock.Rows) *sql.Rows {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("select").WillReturnRows(mockedRows)
	// the following basically does a type cast to what we need
	rows, _ := db.Query("select")
	return rows
}
