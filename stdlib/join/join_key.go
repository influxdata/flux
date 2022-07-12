package join

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/values"
)

type joinKey struct {
	columns []flux.ColMeta
	values  []values.Value
}

func newJoinKey(cols []flux.ColMeta, vals []values.Value) joinKey {
	return joinKey{
		columns: cols,
		values:  vals,
	}
}

func joinKeyFromRow(cols []flux.ColMeta, chunk table.Chunk, idx int) joinKey {
	vals := make([]values.Value, 0, len(chunk.Cols()))
	buf := chunk.Buffer()
	for _, col := range cols {
		ci := chunk.Index(col.Label)
		var v values.Value = values.Null
		if ci >= 0 {
			v = execute.ValueForRow(&buf, idx, ci)
		}

		v.Retain()
		vals = append(vals, v)
	}
	return newJoinKey(cols, vals)
}

func (k *joinKey) equal(other joinKey) bool {
	if len(k.columns) != len(other.columns) {
		return false
	}

	for i := 0; i < len(k.columns); i++ {
		if !k.values[i].Equal(other.values[i]) {
			return false
		}
	}
	return true
}

// Determines if `k` is lexicographically less than `other`
func (k *joinKey) less(other joinKey) bool {
	a, b := k, other
	for i := 0; i < len(k.values); i++ {
		if b.values[i].IsNull() {
			return true
		} else if a.values[i].IsNull() {
			return false
		}

		switch a.columns[i].Type {
		case flux.TBool:
			if av, bv := a.values[i].Bool(), b.values[i].Bool(); av != bv {
				return bv
			}
		case flux.TInt:
			if av, bv := a.values[i].Int(), b.values[i].Int(); av != bv {
				return av < bv
			}
		case flux.TUInt:
			if av, bv := a.values[i].UInt(), b.values[i].UInt(); av != bv {
				return av < bv
			}
		case flux.TFloat:
			if av, bv := a.values[i].Float(), b.values[i].Float(); av != bv {
				return av < bv
			}
		case flux.TString:
			if av, bv := a.values[i].Str(), b.values[i].Str(); av != bv {
				return av < bv
			}
		case flux.TTime:
			if av, bv := a.values[i].Time(), b.values[i].Time(); av != bv {
				return av < bv
			}
		}
	}
	return false
}

func (k *joinKey) str() string {
	keyString := "["
	for i, col := range k.columns {
		keyString = fmt.Sprintf("%s %s:%s=%v ", keyString, col.Label, col.Type, k.values[i])
	}
	keyString = fmt.Sprintf("%s]", keyString)
	return keyString
}
