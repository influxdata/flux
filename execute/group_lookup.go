package execute

import (
	"sort"

	"github.com/influxdata/flux"
)

type GroupLookup struct {
	groups groupEntries

	//  range state
	rangeIdx int

	// Indicates whether the group needs to be sorted.
	needSort bool
}

type groupEntry struct {
	key   flux.GroupKey
	value interface{}
}

func NewGroupLookup() *GroupLookup {
	return &GroupLookup{
		groups: make(groupEntries, 0, 100),
	}
}

func (l *GroupLookup) findIdx(key flux.GroupKey) int {
	if l.needSort {
		sort.Sort(l.groups)
		l.needSort = false
	}

	i := sort.Search(len(l.groups), func(i int) bool {
		return !l.groups[i].key.Less(key)
	})
	if i < len(l.groups) && l.groups[i].key.Equal(key) {
		return i
	}
	return -1
}

func (l *GroupLookup) Lookup(key flux.GroupKey) (interface{}, bool) {
	if key == nil {
		return nil, false
	}
	i := l.findIdx(key)
	if i >= 0 {
		return l.groups[i].value, true
	}
	return nil, false
}

func (l *GroupLookup) Set(key flux.GroupKey, value interface{}) {
	i := l.findIdx(key)
	if i >= 0 {
		l.groups[i].value = value
	} else {
		// There is no need to sort the keys if key is the largest key.
		if !l.needSort && len(l.groups) > 0 {
			l.needSort = key.Less(l.groups[len(l.groups)-1].key)
		}

		l.groups = append(l.groups, groupEntry{
			key:   key,
			value: value,
		})
	}
}

func (l *GroupLookup) Delete(key flux.GroupKey) (v interface{}, found bool) {
	if key == nil {
		return
	}
	i := l.findIdx(key)
	found = i >= 0
	if found {
		if i <= l.rangeIdx {
			l.rangeIdx--
		}
		v = l.groups[i].value
		l.groups = append(l.groups[:i], l.groups[i+1:]...)
	}
	return
}

// Range will iterate over all groups keys in sorted order.
// Range must not be called within another call to Range.
// It is safe to call Set/Delete while ranging.
func (l *GroupLookup) Range(f func(key flux.GroupKey, value interface{})) {
	if l.needSort {
		sort.Sort(l.groups)
		l.needSort = false
	}

	for l.rangeIdx = 0; l.rangeIdx < len(l.groups); l.rangeIdx++ {
		entry := l.groups[l.rangeIdx]
		f(entry.key, entry.value)
	}
}

type groupEntries []groupEntry

func (p groupEntries) Len() int               { return len(p) }
func (p groupEntries) Less(i int, j int) bool { return p[i].key.Less(p[j].key) }
func (p groupEntries) Swap(i int, j int)      { p[i], p[j] = p[j], p[i] }
