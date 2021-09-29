package zoneinfo

import "time"

// ToLocalClock transposes the utc timestamp from a time in the UTC
// location to an equivalent clock time in the current location.
func (l *Location) ToLocalClock(utc int64) (local int64) {
	// This function does its math in seconds rather than nanoseconds.
	sec := utc / int64(time.Second)

	// Look for zone offset for sec, so we can adjust from UTC.
	// The lookup function expects the local time, so we pass sec in the
	// hope that it will not be too close to a zone transition.
	// If the adjusted time is not in the zone transition,
	// then we check the previous or next zone transition to see if
	// it is the appropriate one.
	// If there is no appropriate zone transition, the clock time does not
	// exist in that location and we truncate the time to the start or end
	// of the zone transition.
	_, offset, start, end, _ := l.lookup(sec)
	if offset != 0 {
		// The time may not be valid in our present zone offset.
		// If it is not, check to see if the previous or next zone offset
		// would be applicable.
		switch local := sec - int64(offset); {
		case local < start:
			_, offset, start, end, _ = l.lookup(start - 1)
			if local := sec - int64(offset); local < start || local >= end {
				// The time does not appear to exist. Truncate the interval.
				return end * int64(time.Second)
			}
		case local >= end:
			_, offset, start, end, _ = l.lookup(end)
			if local := sec - int64(offset); local < start || local >= end {
				// The time does not appear to exist. Truncate the interval.
				return start * int64(time.Second)
			}
		}
	}

	// In the above, we made an educated guess about which zone transition
	// we should use. For some clock times, there are multiple possible
	// translation times. While our current zone transition is valid, it
	// might not be the earliest timestamp for this clock time.
	//
	// We want to consistently get the earliest timestamp.
	// To do this, we inspect the previous zone transition to see if it
	// would also be valid if we were to use it instead.
	// If it is, we use the offset from that zone transition instead
	// of the one we found earlier.
	_, prevOffset, prevStart, prevEnd, _ := l.lookup(start - 1)
	if local := sec - int64(prevOffset); local >= prevStart && local < prevEnd {
		offset = prevOffset
	}
	return utc - int64(offset)*int64(time.Second)
}

// FromLocalClock transposes the local timestamp from a time in the location
// to an equivalent clock time in UTC.
func (l *Location) FromLocalClock(local int64) (utc int64) {
	sec := local / int64(time.Second)
	_, offset, _, _, _ := l.lookup(sec)
	return local + int64(offset)*int64(time.Second)
}
