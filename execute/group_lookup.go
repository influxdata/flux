package execute

import (
	"github.com/mvn-trinhnguyen2-dn/flux/internal/execute/groupkey"
)

type GroupLookup = groupkey.Lookup

// NewGroupLookup constructs a GroupLookup.
func NewGroupLookup() *GroupLookup {
	return groupkey.NewLookup()
}

type RandomAccessGroupLookup = groupkey.RandomAccessLookup

// NewRandomAccessGroupLookup constructs a RandomAccessGroupLookup.
func NewRandomAccessGroupLookup() *RandomAccessGroupLookup {
	return groupkey.NewRandomAccessLookup()
}
