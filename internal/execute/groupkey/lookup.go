package groupkey

import (
	"sort"

	"github.com/influxdata/flux"
)

// Lookup is a container that maps group keys to a value.
//
// The Lookup container is optimized for appending values in
// order and iterating over them in the same order. The Lookup
// will always have a deterministic order for the Range call, but that
// order may be influenced by the order that inserts happen.
//
// At the current moment, the Lookup maintains the groups in sorted
// order although future implementations may change that.
//
// To optimize inserts, the lookup is kept in an array of arrays. The first
// layer keeps a group of sorted key groups and each of these groups maintains
// their own sorted list of keys. Each time a new key is added, it is appended
// to the end of one of the key lists. If a key needs to be added in the middle
// of a list, the list is split into two so that the key can be appended.
// The index of the last list to be used is maintained so that future inserts
// can skip past the first search for the key list and an insert can be done in
// constant time. Similarly, a lookup for a key that was just inserted will also
// be in constant time with the worst case time being O(n log n).
type Lookup struct {
	// groups contains groups of group keys in sorted order.
	// These are optimized for appending access.
	groups []*groupKeyList

	// lastIndex contains the last group that an entry was
	// found in or appended to. This is used to optimize appending.
	lastIndex int

	// nextID is the next id that will be assigned to a key group.
	nextID int
}

// groupKeyList is a group of keys in sorted order.
type groupKeyList struct {
	id       int // unique id for the key group within the group lookup
	elements []groupKeyListElement
	deleted  int
}

type groupKeyListElement struct {
	key     flux.GroupKey
	value   interface{}
	deleted bool
}

func (kg *groupKeyList) First() flux.GroupKey {
	return kg.elements[0].key
}

func (kg *groupKeyList) Last() flux.GroupKey {
	return kg.elements[len(kg.elements)-1].key
}

func (kg *groupKeyList) set(i int, value interface{}) {
	if kg.elements[i].deleted {
		kg.elements[i].deleted = false
		kg.deleted--
	}
	kg.elements[i].value = value
}

func (kg *groupKeyList) delete(i int) {
	kg.elements[i].value = nil
	kg.elements[i].deleted = true
	kg.deleted++
}

// Index determines the location of this key within the key group.
// It returns -1 if this key does not exist within the group.
// It will return -1 if the entry is present, but deleted.
func (kg *groupKeyList) Index(key flux.GroupKey) int {
	i := kg.InsertAt(key)
	if i >= len(kg.elements) || kg.elements[i].deleted || !kg.elements[i].key.Equal(key) {
		return -1
	}
	return i
}

// InsertAt will return the index where this key should be inserted.
// If this key would be inserted before the first element, this will
// return 0. If the element exists, then this will return the index
// where that element is located. If the key should be inserted at the
// end of the array, it will return an index the size of the array.
func (kg *groupKeyList) InsertAt(key flux.GroupKey) int {
	if kg.Last().Less(key) {
		return len(kg.elements)
	}
	return sort.Search(len(kg.elements), func(i int) bool {
		return !kg.elements[i].key.Less(key)
	})
}

func (kg *groupKeyList) At(i int) interface{} {
	return kg.elements[i].value
}

// NewLookup constructs a Lookup.
func NewLookup() *Lookup {
	return &Lookup{
		lastIndex: -1,
		nextID:    1,
	}
}

// Lookup will retrieve the value associated with the given key if it exists.
func (l *Lookup) Lookup(key flux.GroupKey) (interface{}, bool) {
	if key == nil || len(l.groups) == 0 {
		return nil, false
	}

	group := l.lookupGroup(key)
	if group == -1 {
		return nil, false
	}

	i := l.groups[group].Index(key)
	if i != -1 {
		return l.groups[group].At(i), true
	}
	return nil, false
}

// LookupOrCreate will retrieve the value associated with the given key or,
// if it does not exist, will invoke the function to create one and set
// it in the group lookup.
func (l *Lookup) LookupOrCreate(key flux.GroupKey, fn func() interface{}) interface{} {
	group, ok := l.Lookup(key)
	if !ok {
		group = fn()
		l.Set(key, group)
	}
	return group
}

// Set will set the value for the given key. It will overwrite an existing value.
func (l *Lookup) Set(key flux.GroupKey, value interface{}) {
	group := l.lookupGroup(key)
	l.createOrSetInGroup(group, key, value)
}

// Clear will clear the group lookup and reset it to contain nothing.
func (l *Lookup) Clear() {
	l.lastIndex = -1
	l.nextID = 1
	l.groups = nil
}

// lookupGroup finds the group index where this key would be located if it were to
// be found or inserted. If no suitable group can be found, then this will return -1
// which indicates that a group has to be created at index 0.
func (l *Lookup) lookupGroup(key flux.GroupKey) int {
	if l.lastIndex >= 0 {
		kg := l.groups[l.lastIndex]
		if !key.Less(kg.First()) {
			// If the next group doesn't exist or has a first value that is
			// greater than this key, then we can return the last index and
			// avoid performing a binary search.
			if l.lastIndex == len(l.groups)-1 || key.Less(l.groups[l.lastIndex+1].First()) {
				return l.lastIndex
			}
		}
	}

	// Find the last group where the first key is less than or equal
	// than the key we are looking for. This means we need to search for
	// the first group where the first key is greater than the key we are setting
	// and use the group before that one.
	index := sort.Search(len(l.groups), func(i int) bool {
		return key.Less(l.groups[i].First())
	}) - 1
	if index >= 0 {
		l.lastIndex = index
	}
	return index
}

// createOrSetInGroup will overwrite or insert a key into the group with the associated value.
// If the key needs to be inserted into the middle of the array, it splits the array into
// two different groups so that the value is always appended to the end of a group to optimize
// future inserts.
func (l *Lookup) createOrSetInGroup(index int, key flux.GroupKey, value interface{}) {
	// If this index is at -1, then we are inserting a value with a smaller key
	// than every group and we need to create a new group to insert it at the
	// beginning.
	if index == -1 {
		l.groups = append(l.groups, nil)
		copy(l.groups[1:], l.groups[:])
		l.groups[0] = l.newKeyGroup([]groupKeyListElement{
			{key: key, value: value},
		})
		l.lastIndex = 0
		return
	}

	kg := l.groups[index]

	// Find the location where this should be inserted.
	i := kg.InsertAt(key)

	// If this should be inserted after the last element, do it and leave.
	if i == len(kg.elements) {
		kg.elements = append(kg.elements, groupKeyListElement{
			key:   key,
			value: value,
		})
		return
	} else if kg.elements[i].key.Equal(key) {
		// If the entry already exists at this index, set the value.
		kg.set(i, value)
		return
	}

	// We have to split this entry into two new elements. First, we start
	// by creating space for the new entry.
	l.groups = append(l.groups, nil)
	copy(l.groups[index+2:], l.groups[index+1:])
	// Construct the new group entry and copy the end of the slice
	// into the new key group.
	l.groups[index+1] = func() *groupKeyList {
		// TODO(rockstar): A nice optimization here would be to prevent
		// the deleted items from being copied. However, this entire function
		// needs to be refactored to support that, as it's possible that *all*
		// the elements have been deleted, so no split is needed.
		// Moving currently deleted elements out of this key group, the deleted
		// count must be decremented.
		for _, item := range kg.elements[i:] {
			if item.deleted {
				kg.deleted--
			}
		}

		entries := make([]groupKeyListElement, len(kg.elements[i:]))
		copy(entries, kg.elements[i:])

		return l.newKeyGroup(entries)
	}()
	// Use a slice on the key group elements to remove the extra elements.
	// Then append the new key group entry.
	kg.elements = kg.elements[:i:cap(kg.elements)]
	kg.elements = append(kg.elements, groupKeyListElement{
		key:   key,
		value: value,
	})
}

// newKeyGroup will construct a new groupKeyList with the next available id. The
// ids are used for detecting if a group has been deleted during a call to Range.
func (l *Lookup) newKeyGroup(entries []groupKeyListElement) *groupKeyList {
	id := l.nextID
	l.nextID++
	return &groupKeyList{
		id:       id,
		elements: entries,
	}
}

// Delete will remove the key from this Lookup. It will return the same
// thing as a call to Lookup.
func (l *Lookup) Delete(key flux.GroupKey) (v interface{}, found bool) {
	if key == nil {
		return
	}

	group := l.lookupGroup(key)
	if group == -1 {
		return nil, false
	}

	kg := l.groups[group]
	i := kg.Index(key)
	if i == -1 {
		return nil, false
	}
	v = kg.At(i)
	kg.delete(i)
	if len(kg.elements) == kg.deleted {
		// All elements in this have been deleted so delete this node.
		copy(l.groups[group:], l.groups[group+1:])
		l.groups = l.groups[: len(l.groups)-1 : cap(l.groups)]
		l.lastIndex = -1
	}
	return v, true
}

// Range will iterate over all groups keys in a stable ordering.
// Range must not be called within another call to Range.
// It is safe to call Set/Delete while ranging.
func (l *Lookup) Range(f func(key flux.GroupKey, value interface{})) {
	for i := 0; i < len(l.groups); {
		kg := l.groups[i]
		for j := 0; j < len(kg.elements); j++ {
			entry := kg.elements[j]
			if entry.deleted {
				continue
			}
			f(entry.key, entry.value)
		}
		if i < len(l.groups) && l.groups[i].id == kg.id {
			i++
		}
	}
}

// RandomAccessLookup is a GroupLookup container that is optimized
// for random access.
type RandomAccessLookup struct {
	elements []*groupLookupElement
	index    map[uint64]*groupLookupElement
}

type groupLookupElement struct {
	Key     flux.GroupKey
	Value   interface{}
	Next    *groupLookupElement
	Deleted bool
}

// NewRandomAccessLookup constructs a RandomAccessLookup.
func NewRandomAccessLookup() *RandomAccessLookup {
	return &RandomAccessLookup{
		index: make(map[uint64]*groupLookupElement),
	}
}

func (l *RandomAccessLookup) idForKey(key flux.GroupKey) uint64 {
	k, ok := key.(*groupKey)
	if !ok {
		k = newGroupKey(key.Cols(), key.Values())
	}
	return k.hash64()
}

// Lookup will retrieve the value associated with the given key if it exists.
func (l *RandomAccessLookup) Lookup(key flux.GroupKey) (interface{}, bool) {
	id := l.idForKey(key)
	e, ok := l.index[id]
	if !ok {
		return nil, false
	}
	for ; e != nil; e = e.Next {
		if !e.Deleted && key.Equal(e.Key) {
			return e.Value, true
		}
	}
	return nil, false
}

// LookupOrCreate will retrieve the value associated with the given key or,
// if it does not exist, will invoke the function to create one and set
// it in the group lookup.
func (l *RandomAccessLookup) LookupOrCreate(key flux.GroupKey, fn func() interface{}) interface{} {
	value, ok := l.Lookup(key)
	if !ok {
		value = fn()
		l.Set(key, value)
	}
	return value
}

// Set will set the value for the given key. It will overwrite an existing value.
func (l *RandomAccessLookup) Set(key flux.GroupKey, value interface{}) {
	id := l.idForKey(key)
	e, ok := l.index[id]
	if !ok {
		e = &groupLookupElement{
			Key: key,
		}
		l.index[id] = e
		l.elements = append(l.elements, e)
	} else if !key.Equal(e.Key) {
		// The present entry doesn't match the group key.
		// This indicates a hash conflict so try to find
		// the group key or add it.
		for ; e.Next != nil; e = e.Next {
			if key.Equal(e.Next.Key) {
				break
			}
		}

		if e.Next == nil {
			e.Next = &groupLookupElement{
				Key: key,
			}
			l.elements = append(l.elements, e.Next)
		}
		e = e.Next
	}
	e.Value = value
	e.Deleted = false
}

// Clear will clear the group lookup and reset it to contain nothing.
func (l *RandomAccessLookup) Clear() {
	l.elements = nil
	l.index = make(map[uint64]*groupLookupElement)
}

// Delete will remove the key from this GroupLookup. It will return the same
// thing as a call to Lookup.
func (l *RandomAccessLookup) Delete(key flux.GroupKey) (v interface{}, found bool) {
	if key == nil {
		return
	}

	id := l.idForKey(key)
	e, ok := l.index[id]
	if !ok {
		return nil, false
	}
	for ; e != nil; e = e.Next {
		if !e.Deleted && key.Equal(e.Key) {
			e.Deleted = true
			return e.Value, true
		}
	}
	return nil, false
}

// Range will iterate over all groups keys in a stable ordering.
// Range must not be called within another call to Range.
// It is safe to call Set/Delete while ranging.
func (l *RandomAccessLookup) Range(f func(key flux.GroupKey, value interface{})) {
	for _, e := range l.elements {
		if e.Deleted {
			continue
		}
		f(e.Key, e.Value)
	}
}
