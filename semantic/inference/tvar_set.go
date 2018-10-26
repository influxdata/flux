package inference

type TvarSet []Tvar

func (a TvarSet) union(b TvarSet) TvarSet {
	union := make(TvarSet, len(a), len(a)+len(b))
	copy(union, a)
LOOP:
	for _, tv := range b {
		for _, tvu := range union {
			if tvu == tv {
				continue LOOP
			}
		}
		union = append(union, tv)
	}
	return union
}

func (a TvarSet) intersect(b TvarSet) TvarSet {
	intersect := make(TvarSet, 0, len(a)+len(b))
	for _, tva := range a {
		for _, tvb := range b {
			if tva == tvb {
				intersect = append(intersect, tva)
				break
			}
		}
	}
	return intersect
}
func (a TvarSet) hasIntersect(b TvarSet) bool {
	for _, tva := range a {
		for _, tvb := range b {
			if tva == tvb {
				return true
			}
		}
	}
	return false
}

func (a TvarSet) equal(b TvarSet) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (s TvarSet) copy() TvarSet {
	c := make(TvarSet, len(s))
	copy(c, s)
	return c
}
