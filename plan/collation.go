package plan

import (
	"fmt"
)

const (
	CollationKey = "collation"
)

// CollationAttr is a physical attribute that describes the collation
// of the rows within a table. Note: the collation attribute does not
// say anything about how the tables in a stream are ordered.
type CollationAttr struct {
	Columns []string
	Desc    bool
}

var _ PhysicalAttr = (*CollationAttr)(nil)

func (ca *CollationAttr) Key() string { return CollationKey }

func (ca *CollationAttr) SuccessorsMustRequire() bool {
	return false
}

func (ca *CollationAttr) SatisfiedBy(attr PhysicalAttr) bool {
	gotCollation, ok := attr.(*CollationAttr)
	if !ok {
		return false
	}
	if ca.Desc != gotCollation.Desc {
		return false
	}

	if len(ca.Columns) > len(gotCollation.Columns) {
		return false
	}

	// Note that if we are looking for collation of [a, b] and we get a collation of [a, b, c]
	// the collation is still satisfied.
	for i, col := range ca.Columns {
		if gotCollation.Columns[i] != col {
			return false
		}
	}

	return true
}

func (ca *CollationAttr) String() string {
	return fmt.Sprintf("%v{Columns: %v, Desc: %v}", CollationKey, ca.Columns, ca.Desc)
}
