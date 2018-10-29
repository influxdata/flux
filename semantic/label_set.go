package semantic

import "strings"

type LabelSet []string

func newLabelSet() LabelSet {
	return make(LabelSet, 0, 10)
}

var allLabels = LabelSet(nil)

func (s LabelSet) String() string {
	if s == nil {
		return "L"
	}
	if len(s) == 0 {
		return "∅"
	}
	var builder strings.Builder
	builder.WriteString("(")
	for i, l := range s {
		if i != 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(l)
	}
	builder.WriteString(")")
	return builder.String()
}

func (s LabelSet) contains(l string) bool {
	for _, lbl := range s {
		if l == lbl {
			return true
		}
	}
	return false
}

func (s LabelSet) union(o LabelSet) LabelSet {
	if s == nil {
		return s
	}
	union := make(LabelSet, len(s), len(s)+len(o))
	copy(union, s)
	for _, l := range o {
		if !union.contains(l) {
			union = append(union, l)
		}
	}
	return union
}

func (s LabelSet) intersect(o LabelSet) LabelSet {
	if s == nil {
		return o
	}
	if o == nil {
		return s
	}
	intersect := make(LabelSet, 0, len(s))
	for _, l := range s {
		if o.contains(l) {
			intersect = append(intersect, l)
		}
	}
	return intersect
}

func (a LabelSet) isSuperSet(b LabelSet) bool {
	if a == nil {
		return true
	}
	if b == nil {
		return false
	}
	for _, l := range b {
		if !a.contains(l) {
			return false
		}
	}
	return true
}

func (a LabelSet) isSubSet(b LabelSet) bool {
	if b == nil {
		return true
	}
	if a == nil {
		return false
	}
	for _, l := range a {
		if !b.contains(l) {
			return false
		}
	}
	return true
}

func (a LabelSet) equal(b LabelSet) bool {
	if len(a) != len(b) {
		return false
	}
	for _, l := range a {
		if !b.contains(l) {
			return false
		}
	}
	return true
}

func (s LabelSet) copy() LabelSet {
	if s == nil {
		return nil
	}
	c := make(LabelSet, len(s))
	copy(c, s)
	return c
}
