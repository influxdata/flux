package execute_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/values"
)

func TestNewWindow(t *testing.T) {
	want := execute.Window{
		Every:  values.ConvertDuration(time.Minute),
		Period: values.ConvertDuration(time.Minute),
		Offset: values.ConvertDuration(time.Second),
	}
	got := execute.NewWindow(values.ConvertDuration(time.Minute), values.ConvertDuration(time.Minute), values.ConvertDuration(time.Second))
	if !cmp.Equal(want, got) {
		t.Errorf("window different; -want/+got:\n%v\n", cmp.Diff(want, got))
	}

	// offset larger than "every" duration will be normalized
	want = execute.Window{
		Every:  values.ConvertDuration(time.Minute),
		Period: values.ConvertDuration(time.Minute),
		Offset: values.ConvertDuration(30 * time.Second),
	}
	got = execute.NewWindow(
		values.ConvertDuration(time.Minute),
		values.ConvertDuration(time.Minute),
		values.ConvertDuration(2*time.Minute+30*time.Second))
	if !cmp.Equal(want, got) {
		t.Errorf("window different; -want/+got:\n%v\n", cmp.Diff(want, got))
	}

	// Negative offset will be normalized
	want = execute.Window{
		Every:  values.ConvertDuration(time.Minute),
		Period: values.ConvertDuration(time.Minute),
		Offset: values.ConvertDuration(30 * time.Second),
	}
	got = execute.NewWindow(
		values.ConvertDuration(time.Minute),
		values.ConvertDuration(time.Minute),
		values.ConvertDuration(-2*time.Minute+30*time.Second))
	if !cmp.Equal(want, got) {
		t.Errorf("window different; -want/+got:\n%v\n", cmp.Diff(want, got))
	}
}

func TestWindow_GetEarliestBounds(t *testing.T) {
	var testcases = []struct {
		name string
		w    execute.Window
		t    execute.Time
		want execute.Bounds
	}{
		{
			name: "simple",
			w: execute.NewWindow(
				values.ConvertDuration(5*time.Minute),
				values.ConvertDuration(5*time.Minute),
				values.ConvertDuration(0)),
			t: execute.Time(6 * time.Minute),
			want: execute.Bounds{
				Start: execute.Time(5 * time.Minute),
				Stop:  execute.Time(10 * time.Minute),
			},
		},
		{
			name: "simple with offset",
			w: execute.NewWindow(
				values.ConvertDuration(5*time.Minute),
				values.ConvertDuration(5*time.Minute),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(5 * time.Minute),
			want: execute.Bounds{
				Start: execute.Time(30 * time.Second),
				Stop:  execute.Time(5*time.Minute + 30*time.Second),
			},
		},
		{
			name: "underlapping",
			w: execute.NewWindow(
				values.ConvertDuration(2*time.Minute),
				values.ConvertDuration(1*time.Minute),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(3 * time.Minute),
			want: execute.Bounds{
				Start: execute.Time(3*time.Minute + 30*time.Second),
				Stop:  execute.Time(4*time.Minute + 30*time.Second),
			},
		},
		{
			name: "underlapping not contained",
			w: execute.NewWindow(
				values.ConvertDuration(2*time.Minute),
				values.ConvertDuration(1*time.Minute),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(2*time.Minute + 45*time.Second),
			want: execute.Bounds{
				Start: execute.Time(3*time.Minute + 30*time.Second),
				Stop:  execute.Time(4*time.Minute + 30*time.Second),
			},
		},
		{
			name: "overlapping",
			w: execute.NewWindow(
				values.ConvertDuration(1*time.Minute),
				values.ConvertDuration(2*time.Minute),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(30 * time.Second),
			want: execute.Bounds{
				Start: execute.Time(-30 * time.Second),
				Stop:  execute.Time(1*time.Minute + 30*time.Second),
			},
		},
		{
			name: "partially overlapping",
			w: execute.NewWindow(
				values.ConvertDuration(1*time.Minute),
				values.ConvertDuration(3*time.Minute+30*time.Second),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(5*time.Minute + 45*time.Second),
			want: execute.Bounds{
				Start: execute.Time(3 * time.Minute),
				Stop:  execute.Time(6*time.Minute + 30*time.Second),
			},
		},
		{
			name: "partially overlapping (t on boundary)",
			w: execute.NewWindow(
				values.ConvertDuration(1*time.Minute),
				values.ConvertDuration(3*time.Minute+30*time.Second),
				values.ConvertDuration(30*time.Second)),
			t: execute.Time(5 * time.Minute),
			want: execute.Bounds{
				Start: execute.Time(2 * time.Minute),
				Stop:  execute.Time(5*time.Minute + 30*time.Second),
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.w.GetEarliestBounds(tc.t)
			if !cmp.Equal(tc.want, got) {
				t.Errorf("did not get expected bounds; -want/+got:\n%v\n", cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestWindow_GetOverlappingBounds(t *testing.T) {
	testcases := []struct {
		name string
		w    execute.Window
		b    execute.Bounds
		want []execute.Bounds
	}{
		{
			name: "simple",
			w: execute.Window{
				Every:  values.ConvertDuration(time.Minute),
				Period: values.ConvertDuration(time.Minute),
			},
			b: execute.Bounds{
				Start: execute.Time(5 * time.Minute),
				Stop:  execute.Time(8 * time.Minute),
			},
			want: []execute.Bounds{
				{Start: execute.Time(5 * time.Minute), Stop: execute.Time(6 * time.Minute)},
				{Start: execute.Time(6 * time.Minute), Stop: execute.Time(7 * time.Minute)},
				{Start: execute.Time(7 * time.Minute), Stop: execute.Time(8 * time.Minute)},
			},
		},
		{
			name: "simple with offset",
			w: execute.Window{
				Every:  values.ConvertDuration(time.Minute),
				Period: values.ConvertDuration(time.Minute),
				Offset: values.ConvertDuration(15 * time.Second),
			},
			b: execute.Bounds{
				Start: execute.Time(5 * time.Minute),
				Stop:  execute.Time(7 * time.Minute),
			},
			want: []execute.Bounds{
				{
					Start: execute.Time(4*time.Minute + 15*time.Second),
					Stop:  execute.Time(5*time.Minute + 15*time.Second),
				},
				{
					Start: execute.Time(5*time.Minute + 15*time.Second),
					Stop:  execute.Time(6*time.Minute + 15*time.Second),
				},
				{
					Start: execute.Time(6*time.Minute + 15*time.Second),
					Stop:  execute.Time(7*time.Minute + 15*time.Second),
				},
			},
		},
		{
			name: "underlapping, bounds in gap",
			w: execute.Window{
				Every:  values.ConvertDuration(2 * time.Minute),
				Period: values.ConvertDuration(time.Minute),
			},
			b: execute.Bounds{
				Start: execute.Time(30 * time.Second),
				Stop:  execute.Time(45 * time.Second),
			},
			want: []execute.Bounds{},
		},
		{
			name: "underlapping",
			w: execute.Window{
				Every:  values.ConvertDuration(2 * time.Minute),
				Period: values.ConvertDuration(time.Minute),
				Offset: values.ConvertDuration(30 * time.Second),
			},
			b: execute.Bounds{
				Start: execute.Time(time.Minute + 45*time.Second),
				Stop:  execute.Time(4*time.Minute + 35*time.Second),
			},
			want: []execute.Bounds{
				{
					Start: execute.Time(1*time.Minute + 30*time.Second),
					Stop:  execute.Time(2*time.Minute + 30*time.Second),
				},
				{
					Start: execute.Time(3*time.Minute + 30*time.Second),
					Stop:  execute.Time(4*time.Minute + 30*time.Second),
				},
			},
		},
		{
			name: "overlapping",
			w: execute.Window{
				Every:  values.ConvertDuration(1 * time.Minute),
				Period: values.ConvertDuration(2*time.Minute + 15*time.Second),
			},
			b: execute.Bounds{
				Start: execute.Time(10 * time.Minute),
				Stop:  execute.Time(12 * time.Minute),
			},
			want: []execute.Bounds{
				{
					Start: execute.Time(8*time.Minute + 45*time.Second),
					Stop:  execute.Time(11 * time.Minute),
				},
				{
					Start: execute.Time(9*time.Minute + 45*time.Second),
					Stop:  execute.Time(12 * time.Minute),
				},
				{
					Start: execute.Time(10*time.Minute + 45*time.Second),
					Stop:  execute.Time(13 * time.Minute),
				},
				{
					Start: execute.Time(11*time.Minute + 45*time.Second),
					Stop:  execute.Time(14 * time.Minute),
				},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.w.GetOverlappingBounds(tc.b)
			if !cmp.Equal(tc.want, got) {
				t.Errorf("got unexpected bounds; -want/+got:\n%v\n", cmp.Diff(tc.want, got))
			}
		})
	}
}
