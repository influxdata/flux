package execute

import (
	"fmt"
	"sort"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/values"
)

type groupKey struct {
	cols   []flux.ColMeta
	values []values.Value
	sorted []int // maintains a list of the sorted indexes
}

func NewGroupKey(cols []flux.ColMeta, values []values.Value) flux.GroupKey {
	return newGroupKey(cols, values)
}

func newGroupKey(cols []flux.ColMeta, values []values.Value) *groupKey {
	sorted := make([]int, len(cols))
	for i := range cols {
		sorted[i] = i
	}
	sort.Slice(sorted, func(i, j int) bool {
		return cols[sorted[i]].Label < cols[sorted[j]].Label
	})
	return &groupKey{
		cols:   cols,
		values: values,
		sorted: sorted,
	}
}

func (k *groupKey) Cols() []flux.ColMeta {
	return k.cols
}
func (k *groupKey) Values() []values.Value {
	return k.values
}
func (k *groupKey) HasCol(label string) bool {
	return ColIdx(label, k.cols) >= 0
}
func (k *groupKey) LabelValue(label string) values.Value {
	if !k.HasCol(label) {
		return nil
	}
	return k.Value(ColIdx(label, k.cols))
}
func (k *groupKey) IsNull(j int) bool {
	return k.values[j].IsNull()
}
func (k *groupKey) Value(j int) values.Value {
	return k.values[j]
}
func (k *groupKey) ValueBool(j int) bool {
	return k.values[j].Bool()
}
func (k *groupKey) ValueUInt(j int) uint64 {
	return k.values[j].UInt()
}
func (k *groupKey) ValueInt(j int) int64 {
	return k.values[j].Int()
}
func (k *groupKey) ValueFloat(j int) float64 {
	return k.values[j].Float()
}
func (k *groupKey) ValueString(j int) string {
	return k.values[j].Str()
}
func (k *groupKey) ValueDuration(j int) Duration {
	return k.values[j].Duration()
}
func (k *groupKey) ValueTime(j int) Time {
	return k.values[j].Time()
}

func (k *groupKey) Equal(o flux.GroupKey) bool {
	return groupKeyEqual(k, o)
}

func (k *groupKey) Less(o flux.GroupKey) bool {
	return groupKeyLess(k, o)
}

func (k *groupKey) String() string {
	var b strings.Builder
	b.WriteRune('{')
	for j, c := range k.cols {
		if j != 0 {
			b.WriteRune(',')
		}
		fmt.Fprintf(&b, "%s=%v", c.Label, k.values[j])
	}
	b.WriteRune('}')
	return b.String()
}

func groupKeyEqual(a *groupKey, other flux.GroupKey) bool {
	b, ok := other.(*groupKey)
	if !ok {
		b = newGroupKey(other.Cols(), other.Values())
	}

	if len(a.cols) != len(b.cols) {
		return false
	}
	for i, idx := range a.sorted {
		jdx := b.sorted[i]
		if a.cols[idx] != b.cols[jdx] {
			return false
		}
		if anull, bnull := a.values[idx].IsNull(), b.values[jdx].IsNull(); anull && bnull {
			// Both key columns are null, consider them equal
			// So that rows are assigned to the same table.
			continue
		} else if anull || bnull {
			return false
		}

		switch a.cols[idx].Type {
		case flux.TBool:
			if a.ValueBool(idx) != b.ValueBool(jdx) {
				return false
			}
		case flux.TInt:
			if a.ValueInt(idx) != b.ValueInt(jdx) {
				return false
			}
		case flux.TUInt:
			if a.ValueUInt(idx) != b.ValueUInt(jdx) {
				return false
			}
		case flux.TFloat:
			if a.ValueFloat(idx) != b.ValueFloat(jdx) {
				return false
			}
		case flux.TString:
			if a.ValueString(idx) != b.ValueString(jdx) {
				return false
			}
		case flux.TTime:
			if a.ValueTime(idx) != b.ValueTime(jdx) {
				return false
			}
		}
	}
	return true
}

// groupKeyLess determines if the former key is lexicographically less than the
// latter.
func groupKeyLess(a *groupKey, other flux.GroupKey) bool {
	b, ok := other.(*groupKey)
	if !ok {
		b = newGroupKey(other.Cols(), other.Values())
	}

	min := len(a.sorted)
	if len(b.sorted) < min {
		min = len(b.sorted)
	}

	for i := 0; i < min; i++ {
		idx, jdx := a.sorted[i], b.sorted[i]
		if a.cols[idx].Label != b.cols[jdx].Label {
			// The labels at the current index are different
			// so whichever one is greater is the one missing
			// a value and the one missing a value is the less.
			// That causes this next conditional to look wrong.
			return a.cols[idx].Label > b.cols[jdx].Label
		}

		// The labels are identical. If the types are different,
		// then resolve the ordering based on the type.
		// TODO(jsternberg): Make this official in some way and part of the spec.
		if a.cols[idx].Type != b.cols[jdx].Type {
			return a.cols[idx].Type < b.cols[jdx].Type
		}

		// If a value is null, it is less than.
		if anull, bnull := a.values[idx].IsNull(), b.values[jdx].IsNull(); anull && bnull {
			continue
		} else if anull {
			return true
		} else if bnull {
			return false
		}

		// Neither value is null and they are the same type so compare.
		switch a.cols[idx].Type {
		case flux.TBool:
			if av, bv := a.ValueBool(idx), b.ValueBool(jdx); av != bv {
				return bv
			}
		case flux.TInt:
			if av, bv := a.ValueInt(idx), b.ValueInt(jdx); av != bv {
				return av < bv
			}
		case flux.TUInt:
			if av, bv := a.ValueUInt(idx), b.ValueUInt(jdx); av != bv {
				return av < bv
			}
		case flux.TFloat:
			if av, bv := a.ValueFloat(idx), b.ValueFloat(jdx); av != bv {
				return av < bv
			}
		case flux.TString:
			if av, bv := a.ValueString(idx), b.ValueString(jdx); av != bv {
				return av < bv
			}
		case flux.TTime:
			if av, bv := a.ValueTime(idx), b.ValueTime(jdx); av != bv {
				return av < bv
			}
		}
	}

	// In this case, min columns have been compared and found to be equal.
	// Whichever key has the greater number of columns is lexicographically
	// greater than the other.
	return len(a.sorted) < len(b.sorted)
}
