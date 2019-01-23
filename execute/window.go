package execute

type Window struct {
	Every  Duration
	Period Duration
	Offset Duration
}

// NewWindow creates a window with the given parameters,
// and normalizes the offset to a small positive duration.
func NewWindow(every, period, offset Duration) Window {
	// Normalize the offset to a small positive duration
	if offset < 0 {
		offset += every * ((offset / -every) + 1)
	} else if offset > every {
		offset -= every * (offset / every)
	}

	return Window{
		Every:  every,
		Period: period,
		Offset: offset,
	}
}

// GetEarliestBounds returns the bounds for the earliest window bounds
// that contains the given time t.  For underlapping windows that
// do not contain time t, the window directly after time t will be returned.
func (w Window) GetEarliestBounds(t Time) Bounds {
	// translate to not-offset coordinate
	t = t.Add(-w.Offset)

	stop := t.Truncate(w.Every).Add(w.Every)

	// translate to offset coordinate
	stop = stop.Add(w.Offset)

	start := stop.Add(-w.Period)
	return Bounds{
		Start: start,
		Stop:  stop,
	}
}

// GetOverlappingBounds returns a slice of bounds for each window
// that overlaps the input bounds b.
func (w Window) GetOverlappingBounds(b Bounds) []Bounds {
	if b.IsEmpty() {
		return []Bounds{}
	}

	c := (b.Duration() / w.Every) + (w.Period / w.Every)
	bs := make([]Bounds, 0, c)

	bi := w.GetEarliestBounds(b.Start)
	for bi.Start < b.Stop {
		bs = append(bs, bi)
		bi.Start = bi.Start.Add(w.Every)
		bi.Stop = bi.Stop.Add(w.Every)
	}

	return bs
}
