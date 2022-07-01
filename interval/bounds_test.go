package interval

import (
	"testing"

	"github.com/influxdata/flux/values"
)

var tests = []struct {
	name string
	a, b Bounds
	want bool
}{
	{
		name: "edge overlap",
		a: Bounds{
			start: values.Time(0),
			stop:  values.Time(10),
		},
		b: Bounds{
			start: values.Time(10),
			stop:  values.Time(20),
		},

		want: false,
	},
	{
		name: "edge overlap sym",
		a: Bounds{
			start: values.Time(10),
			stop:  values.Time(20),
		},
		b: Bounds{
			start: values.Time(0),
			stop:  values.Time(10),
		},
		want: false,
	},
	{
		name: "single overlap",
		a: Bounds{
			start: values.Time(0),
			stop:  values.Time(10),
		},
		b: Bounds{
			start: values.Time(5),
			stop:  values.Time(15),
		},
		want: true,
	},
	{
		name: "no overlap sym",
		a: Bounds{
			start: values.Time(0),
			stop:  values.Time(10),
		},
		b: Bounds{
			start: values.Time(5),
			stop:  values.Time(15),
		},
		want: true,
	},
	{
		name: "double overlap (bounds contained)",
		a: Bounds{
			start: values.Time(10),
			stop:  values.Time(20),
		},
		b: Bounds{
			start: values.Time(14),
			stop:  values.Time(15),
		},
		want: true,
	},
	{
		name: "double overlap (bounds contained) sym",
		a: Bounds{
			start: values.Time(14),
			stop:  values.Time(15),
		},
		b: Bounds{
			start: values.Time(10),
			stop:  values.Time(20),
		},
		want: true,
	},
}

// Written to verify symmetrical behavior of interval.(Bounds).Overlaps
// Given two Bounds a and b, if a.Overlaps(b) then b.Overlaps(a).
//
// Cases:
// given two ranges [a1, a2), [b1, b2)
// a1 <= b1 <= a2 <= b2 -> true
// b1 <= a1 <= b2 <= a2 -> true
// a1 <= b1 <= b2 <= a2 -> true
// b2 <= a1 <= a2 <= b2 -> true
// a1 <= a2 <= b1 <= b2 -> false
// b1 <= b2 <= a1 <= a2 -> false
func TestBounds_Overlaps(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Overlaps(tt.b); got != tt.want {
				t.Errorf("Bounds.Overlaps() = %v, want %v", got, tt.want)
			}
		})
	}
}
