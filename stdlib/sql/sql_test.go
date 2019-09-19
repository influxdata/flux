package sql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
	"github.com/stretchr/testify/assert"
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
				{int64(42), float64(42.0), true, timestamp},
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
			assert.Equal(t, wantBuffer.GetRow(i), buffer.GetRow(i))
		}

	})
}

func TestMySQLParsing(t *testing.T) {
	testCases := []struct {
		name       string
		columnName string
		columnType flux.ColType
		data       [][]uint8
		want       [][]values.Value
	}{
		{
			name:       "ints",
			columnName: "_int",
			columnType: flux.TInt,
			data:       stringSliceToByteArrays([]string{"6", "1", "643", "42", "1283", "4", "0", "18"}),
			want:       [][]values.Value{{values.NewInt(6)}, {values.NewInt(1)}, {values.NewInt(643)}, {values.NewInt(42)}, {values.NewInt(1283)}, {values.NewInt(4)}, {values.NewInt(0)}, {values.NewInt(18)}},
		},
		{
			name:       "floats",
			columnName: "_float",
			columnType: flux.TFloat,
			data:       stringSliceToByteArrays([]string{"6", "1", "643", "42", "1283", "4", "0", "18"}),
			want:       [][]values.Value{{values.NewFloat(6)}, {values.NewFloat(1)}, {values.NewFloat(643)}, {values.NewFloat(42)}, {values.NewFloat(1283)}, {values.NewFloat(4)}, {values.NewFloat(0)}, {values.NewFloat(18)}},
		},
		{
			name:       "strings",
			columnName: "_string",
			columnType: flux.TString,
			data:       stringSliceToByteArrays([]string{"6", "1", "643", "42", "1283", "4", "0", "18"}),
			want:       [][]values.Value{{values.NewString("6")}, {values.NewString("1")}, {values.NewString("643")}, {values.NewString("42")}, {values.NewString("1283")}, {values.NewString("4")}, {values.NewString("0")}, {values.NewString("18")}},
		},
		{
			name:       "datetime",
			columnName: "_datetime",
			columnType: flux.TTime,
			data: stringSliceToByteArrays([]string{
				"2019-06-03 13:59:00",
				"2019-06-03 13:59:01",
				"2019-06-03 13:59:02",
				"2019-06-03 13:59:03",
				"2019-06-03 13:59:04",
				"2019-06-03 13:59:05",
				"2019-06-03 13:59:06",
				"2019-06-03 13:59:07"}),
			want: [][]values.Value{
				{values.NewTime(values.ConvertTime(createTestTimes()[0].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[1].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[2].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[3].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[4].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[5].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[6].(time.Time)))},
				{values.NewTime(values.ConvertTime(createTestTimes()[7].(time.Time)))},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			TestReader := &MySQLRowReader{}

			TestReader.NextFunc = func() func() bool {
				i := 0
				vals := make([]interface{}, len(tc.data))
				for i, v := range tc.data {
					vals[i] = v
				}
				return func() bool {
					if i < len(tc.data) {
						TestReader.SetColumns([]interface{}{vals[i]})
						i++
						return true
					}
					return false
				}
			}()
			TestReader.InitColumnNames([]string{tc.columnName})
			TestReader.SetColumnTypes([]flux.ColType{tc.columnType})

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

func stringSliceToByteArrays(s []string) [][]byte {
	array := make([][]byte, len(s))

	for i := range s {
		b := []byte(s[i])
		array[i] = b
	}

	return array
}
