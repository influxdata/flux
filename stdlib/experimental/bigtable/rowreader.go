package bigtable

import (
	"fmt"

	"cloud.google.com/go/bigtable"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"golang.org/x/net/context"
)

type BigtableRowReader struct {
	rowsByFamily map[string][]FamilyRow
	families     []string
	familyIndex  int

	columnIndices map[string]int
	rowIndex      int
	// We need columnNames because maps aren't ordered in Go
	columnNames []string
}

func (m *BigtableRowReader) Next() bool {
	m.rowIndex++
	return m.rowIndex < len(m.rowsByFamily[m.families[m.familyIndex]])
}

func (m *BigtableRowReader) GetNextRow() ([]values.Value, error) {
	r := m.rowsByFamily[m.families[m.familyIndex]][m.rowIndex]

	if len(r) <= 0 {
		return nil, nil
	}

	// Start off with the row name and time
	rowValues := make([]values.Value, len(m.columnIndices))
	rowValues[0] = values.NewString(r[0].Row)
	rowValues[1] = values.NewTime(values.ConvertTime(r[0].Timestamp.Time()))
	rowValues[2] = values.NewString(m.families[m.familyIndex])

	// Keeps track of which columnIndices this row includes
	encountered := make([]bool, len(m.columnIndices))
	// Row key and time have already been appended
	encountered[0], encountered[1], encountered[2] = true, true, true

	for _, item := range r {
		label := item.Column
		idx, ok := m.columnIndices[label]
		if !ok {
			idx = len(m.columnIndices)
			m.columnIndices[label] = idx
			m.columnNames = append(m.columnNames, label)
			encountered = append(encountered, true)
			rowValues = append(rowValues, values.NewString(string(item.Value)))
		} else {
			encountered[idx] = true
			rowValues[idx] = values.NewString(string(item.Value))
		}
	}

	for _, idx := range m.columnIndices {
		if !encountered[idx] {
			rowValues[idx] = values.NewNull(semantic.BasicString)
		}
	}

	return rowValues, nil
}

func (m *BigtableRowReader) ColumnNames() []string {
	return m.columnNames
}

func (m *BigtableRowReader) ColumnTypes() []flux.ColType {
	return nil
}

func (m *BigtableRowReader) SetColumns([]interface{}) {}

func (m *BigtableRowReader) Close() error { return nil }

func NewBigtableRowReader(ctx context.Context, c *BigtableDecoder) (execute.RowReader, error) {
	reader := &BigtableRowReader{
		rowsByFamily: make(map[string][]FamilyRow),
		families:     make([]string, 0),
	}

	reader.familyIndex = -1

	// bigtable.InfiniteRange(start) returns the RowRange consisting of all keys at least as large as start
	// "" is the lexicographically smallest byte string - gets all rows
	if c.spec.RowSet == nil {
		c.spec.RowSet = bigtable.InfiniteRange("")
	}
	c.spec.ReadOptions = append(c.spec.ReadOptions, bigtable.RowFilter(c.spec.Filter))
	if err := c.tbl.ReadRows(ctx, c.spec.RowSet, func(r bigtable.Row) bool {
		for family := range r {
			if _, ok := reader.rowsByFamily[family]; !ok {
				reader.rowsByFamily[family] = make([]FamilyRow, 0)
				reader.families = append(reader.families, family)
			}
			reader.rowsByFamily[family] = append(reader.rowsByFamily[family], r[family])
		}
		return true
	}, c.spec.ReadOptions...); err != nil {
		return nil, err
	}

	return reader, nil
}

func (m *BigtableRowReader) nextFamily() (bool, error) {
	m.familyIndex++
	m.rowIndex = -1
	m.columnIndices = map[string]int{"rowKey": 0, execute.DefaultTimeColLabel: 1, "family": 2}
	m.columnNames = []string{"rowKey", execute.DefaultTimeColLabel, "family"}

	if len(m.rowsByFamily) == 0 {
		return false, fmt.Errorf("no rows found")
	}
	return m.familyIndex < len(m.rowsByFamily), nil
}

func (m *BigtableRowReader) currentFamily() string {
	return m.families[m.familyIndex]
}
