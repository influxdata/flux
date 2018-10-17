package planner_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/influxdata/flux/planner"
	"github.com/influxdata/flux/planner/plantest"
	"github.com/influxdata/flux/values"
)

func TestBoundsIntersect(t *testing.T) {
	now := time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		now  time.Time
		a, b *planner.Bounds
		want *planner.Bounds
	}{
		{
			name: "contained",
			a: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-1 * time.Hour)),
				Stop:  values.ConvertTime(now),
			},
			b: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-30 * time.Minute)),
				Stop:  values.ConvertTime(now),
			},
			want: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-30 * time.Minute)),
				Stop:  values.ConvertTime(now),
			},
		},
		{
			name: "contained sym",
			now:  time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC),
			a: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-30 * time.Minute)),
				Stop:  values.ConvertTime(now),
			},
			b: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-1 * time.Hour)),
				Stop:  values.ConvertTime(now),
			},
			want: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-30 * time.Minute)),
				Stop:  values.ConvertTime(now),
			},
		},
		{
			name: "no overlap",
			now:  time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC),
			a: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-1 * time.Hour)),
				Stop:  values.ConvertTime(now),
			},
			b: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-3 * time.Hour)),
				Stop:  values.ConvertTime(now.Add(-2 * time.Hour)),
			},
			want: planner.EmptyBounds,
		},
		{
			name: "overlap",
			now:  time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC),
			a: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-1 * time.Hour)),
				Stop:  values.ConvertTime(now),
			},
			b: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-2 * time.Hour)),
				Stop:  values.ConvertTime(now.Add(-30 * time.Minute)),
			},
			want: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-1 * time.Hour)),
				Stop:  values.ConvertTime(now.Add(-30 * time.Minute)),
			},
		},
		{
			name: "absolute times",
			a: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 1, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(time.Date(2018, time.January, 1, 0, 3, 0, 0, time.UTC)),
			},
			b: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 4, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(time.Date(2018, time.January, 1, 0, 5, 0, 0, time.UTC)),
			},
			want: planner.EmptyBounds,
		},
		{
			name: "intersect with empty returns empty",
			now:  time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC),
			a: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 15, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(now),
			},
			b:    planner.EmptyBounds,
			want: planner.EmptyBounds,
		},
		{
			name: "intersect with empty returns empty sym",
			now:  time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC),
			a:    planner.EmptyBounds,
			b: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 15, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(now),
			},
			want: planner.EmptyBounds,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Intersect(tt.b)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("unexpected bounds -want/+got:\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestBounds_Union(t *testing.T) {
	now := time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		now  time.Time
		a, b *planner.Bounds
		want *planner.Bounds
	}{
		{
			name: "basic case",
			a: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 1, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(time.Date(2018, time.January, 1, 0, 3, 0, 0, time.UTC)),
			},
			b: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 2, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(time.Date(2018, time.January, 1, 0, 4, 0, 0, time.UTC)),
			},
			want: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 1, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(time.Date(2018, time.January, 1, 0, 4, 0, 0, time.UTC)),
			},
		},
		{
			name: "union with empty returns empty",
			a: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 15, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(now),
			},
			b:    planner.EmptyBounds,
			want: planner.EmptyBounds,
		},
		{
			name: "union with empty returns empty sym",
			now:  time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC),
			a:    planner.EmptyBounds,
			b: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 15, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(now),
			},
			want: planner.EmptyBounds,
		},
		{
			name: "no overlap",
			now:  time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC),
			a: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 15, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(time.Date(2018, time.January, 1, 0, 20, 0, 0, time.UTC)),
			},
			b: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 45, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(time.Date(2018, time.January, 1, 0, 50, 0, 0, time.UTC)),
			},
			want: &planner.Bounds{
				Start: values.ConvertTime(time.Date(2018, time.January, 1, 0, 15, 0, 0, time.UTC)),
				Stop:  values.ConvertTime(time.Date(2018, time.January, 1, 0, 50, 0, 0, time.UTC)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Union(tt.b)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("unexpected bounds -want/+got:\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestBounds_IsEmpty(t *testing.T) {
	now := time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		now    time.Time
		bounds *planner.Bounds
		want   bool
	}{
		{
			name: "empty bounds / start == stop",
			now:  time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC),
			bounds: &planner.Bounds{
				Start: values.ConvertTime(now),
				Stop:  values.ConvertTime(now),
			},
			want: true,
		},
		{
			name: "empty bounds / absolute now == relative now",
			now:  time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC),
			bounds: &planner.Bounds{
				Start: values.ConvertTime(now),
				Stop:  values.ConvertTime(time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC)),
			},
			want: true,
		},
		{
			name: "start > stop",
			now:  time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC),
			bounds: &planner.Bounds{
				Start: values.ConvertTime(now.Add(time.Hour)),
				Stop:  values.ConvertTime(now),
			},
			want: true,
		},
		{
			name: "start < stop",
			now:  time.Date(2018, time.August, 14, 11, 0, 0, 0, time.UTC),
			bounds: &planner.Bounds{
				Start: values.ConvertTime(now.Add(-1 * time.Hour)),
				Stop:  values.ConvertTime(now),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bounds.IsEmpty()
			if got != tt.want {
				t.Errorf("unexpected result for bounds.IsEmpty(): got %t, want %t", got, tt.want)
			}
		})
	}
}

// A BoundsAwareProcedureSpec that intersects its bounds with its predecessors' bounds
type mockBoundsIntersectProcedureSpec struct {
	planner.DefaultCost
	bounds *planner.Bounds
}

func (m *mockBoundsIntersectProcedureSpec) Kind() planner.ProcedureKind {
	return "mock-intersect-bounds"
}

func (m *mockBoundsIntersectProcedureSpec) Copy() planner.ProcedureSpec {
	return &mockBoundsIntersectProcedureSpec{}
}

func (m *mockBoundsIntersectProcedureSpec) TimeBounds(predecessorBounds *planner.Bounds) *planner.Bounds {
	if predecessorBounds != nil {
		return predecessorBounds.Intersect(m.bounds)
	}
	return m.bounds
}

// A BoundsAwareProcedureSpec that shifts its predecessors' bounds
type mockBoundsShiftProcedureSpec struct {
	planner.DefaultCost
	by values.Duration
}

func (m *mockBoundsShiftProcedureSpec) Kind() planner.ProcedureKind {
	return "mock-shift-bounds"
}

func (m *mockBoundsShiftProcedureSpec) Copy() planner.ProcedureSpec {
	return &mockBoundsShiftProcedureSpec{}
}

func (m *mockBoundsShiftProcedureSpec) TimeBounds(predecessorBounds *planner.Bounds) *planner.Bounds {
	if predecessorBounds != nil {
		return predecessorBounds.Shift(m.by)
	}
	return nil
}

// Create a PlanNode with id and mockBoundsIntersectProcedureSpec
func makeBoundsNode(id string, bounds *planner.Bounds) planner.PlanNode {
	return planner.CreatePhysicalNode(planner.NodeID(id),
		&mockBoundsIntersectProcedureSpec{
			bounds: bounds,
		})
}

// Create a PlanNode with id and mockBoundsShiftProcedureSpec
func makeShiftNode(id string, duration values.Duration) planner.PlanNode {
	return planner.CreateLogicalNode(planner.NodeID(id),
		&mockBoundsShiftProcedureSpec{
			by: duration,
		})
}

func bounds(start, stop int) *planner.Bounds {
	return &planner.Bounds{
		Start: values.Time(start),
		Stop:  values.Time(stop),
	}
}

// Test that bounds are propagated up through the plan correctly
func TestBounds_ComputePlanBounds(t *testing.T) {
	tests := []struct {
		// Name of test
		name string
		// Nodes and edges defining plan
		spec *plantest.PhysicalPlanSpec
		// Map from node ID to the expected bounds for that node
		want map[planner.NodeID]*planner.Bounds
	}{
		{
			name: "no bounds",
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					makeNode("0"),
				},
			},
			want: map[planner.NodeID]*planner.Bounds{
				"0": nil,
			},
		},
		{
			name: "single time bounds",
			// 0 -> 1 -> 2 -> 3 -> 4
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					makeNode("0"),
					makeNode("1"),
					makeBoundsNode("2", bounds(5, 10)),
					makeNode("3"),
					makeNode("4"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
				},
			},
			want: map[planner.NodeID]*planner.Bounds{
				"0": nil,
				"1": nil,
				"2": bounds(5, 10),
				"3": bounds(5, 10),
				"4": bounds(5, 10)},
		},
		{
			name: "multiple intersect time bounds",
			// 0 -> 1 -> 2 -> 3 -> 4
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					makeNode("0"),
					makeBoundsNode("1", bounds(5, 10)),
					makeNode("2"),
					makeBoundsNode("3", bounds(7, 11)),
					makeNode("4"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
				},
			},
			want: map[planner.NodeID]*planner.Bounds{
				"0": nil,
				"1": bounds(5, 10),
				"2": bounds(5, 10),
				"3": bounds(7, 10),
				"4": bounds(7, 10)},
		},
		{
			name: "shift nil time bounds",
			// 0 -> 1 -> 2
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					makeNode("0"),
					makeShiftNode("1", values.Duration(5)),
					makeNode("2"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			want: map[planner.NodeID]*planner.Bounds{
				"0": nil,
				"1": nil,
				"2": nil,
			},
		},
		{
			name: "shift bounds after intersecting bounds",
			// 0 -> 1 -> 2 -> 3 -> 4
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					makeNode("0"),
					makeBoundsNode("1", bounds(5, 10)),
					makeNode("2"),
					makeShiftNode("3", values.Duration(5)),
					makeNode("4"),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
					{2, 3},
					{3, 4},
				},
			},
			want: map[planner.NodeID]*planner.Bounds{
				"0": nil,
				"1": bounds(5, 10),
				"2": bounds(5, 10),
				"3": bounds(10, 15),
				"4": bounds(10, 15)},
		},
		{
			name: "join",
			//   2
			//  / \
			// 0   1
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					makeBoundsNode("0", bounds(5, 10)),
					makeBoundsNode("1", bounds(12, 20)),
					makeNode("2"),
				},
				Edges: [][2]int{
					{0, 2},
					{1, 2},
				},
			},
			want: map[planner.NodeID]*planner.Bounds{
				"0": bounds(5, 10),
				"1": bounds(12, 20),
				"2": bounds(5, 20),
			},
		},
		{
			name: "yields",
			// 3   4
			//  \ /
			//   1   2
			//    \ /
			//     0
			spec: &plantest.PhysicalPlanSpec{
				Nodes: []planner.PlanNode{
					makeNode("0"),
					makeBoundsNode("1", bounds(5, 10)),
					makeNode("2"),
					makeNode("3"),
					makeNode("4"),
				},
				Edges: [][2]int{
					{0, 1},
					{0, 2},
					{1, 3},
					{1, 4},
				},
			},
			want: map[planner.NodeID]*planner.Bounds{
				"0": nil,
				"1": bounds(5, 10),
				"2": nil,
				"3": bounds(5, 10),
				"4": bounds(5, 10),
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Create plan from spec
			thePlan := plantest.CreatePhysicalPlanSpec(tc.spec)

			// Method used to compute the bounds at each node
			if err := thePlan.BottomUpWalk(planner.ComputeBounds); err != nil {
				t.Fatal(err)
			}

			// Map NodeID -> Bounds
			got := make(map[planner.NodeID]*planner.Bounds)
			thePlan.BottomUpWalk(func(n planner.PlanNode) error {
				got[n.ID()] = n.Bounds()
				return nil
			})

			if !cmp.Equal(tc.want, got) {
				t.Errorf("Did not get expected time bounds, -want/+got:\n%v", cmp.Diff(tc.want, got))
			}
		})
	}
}
