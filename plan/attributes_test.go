package plan_test

import (
	"fmt"
	"testing"

	"github.com/InfluxCommunity/flux/plan"
	"github.com/InfluxCommunity/flux/plan/plantest"
	"github.com/stretchr/testify/require"
)

const mockAttrKey = "mock-attr"

type mockAttr struct {
	successorMustRequire bool
	v                    int
}

func (m *mockAttr) String() string              { return fmt.Sprintf("%v{v: %v}", mockAttrKey, m.v) }
func (m *mockAttr) Key() string                 { return mockAttrKey }
func (m *mockAttr) SuccessorsMustRequire() bool { return m.successorMustRequire }
func (m *mockAttr) SatisfiedBy(attr plan.PhysicalAttr) bool {
	other, ok := attr.(*mockAttr)
	if !ok {
		return false
	}
	return m.v == other.v
}

func TestCheckRequiredAttributes(t *testing.T) {
	tcs := []struct {
		name  string
		input *plantest.PlanSpec
		err   string
	}{
		{
			name: "valid",
			input: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalNode("has-attr", plantest.MockProcedureSpec{
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{
								mockAttrKey: &mockAttr{},
							}
						},
					}),
					plantest.CreatePhysicalNode("passthru", plantest.MockProcedureSpec{
						PassThroughAttributeFn: func(attrKey string) bool { return attrKey == mockAttrKey },
					}),
					plantest.CreatePhysicalNode("require-attr", plantest.MockProcedureSpec{
						RequiredAttributesFn: func() []plan.PhysicalAttributes {
							return []plan.PhysicalAttributes{
								{
									mockAttrKey: &mockAttr{},
								},
							}
						},
					}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
		},
		{
			name: "wrong number of required attributes",
			input: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalNode("pred0", plantest.MockProcedureSpec{
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{
								mockAttrKey: &mockAttr{},
							}
						},
					}),
					plantest.CreatePhysicalNode("pred1", plantest.MockProcedureSpec{
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{
								mockAttrKey: &mockAttr{},
							}
						},
					}),
					plantest.CreatePhysicalNode("require-attr", plantest.MockProcedureSpec{
						RequiredAttributesFn: func() []plan.PhysicalAttributes {
							return []plan.PhysicalAttributes{
								{
									mockAttrKey: &mockAttr{},
								},
							}
						},
					}),
				},
				Edges: [][2]int{
					{0, 2},
					{1, 2},
				},
			},
			err: "node has 2 predecessors but has 1 sets of required attributes",
		},
		{
			name: "missing required attributes",
			input: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalNode("no-attr", plantest.MockProcedureSpec{}),
					plantest.CreatePhysicalNode("passthru", plantest.MockProcedureSpec{
						PassThroughAttributeFn: func(attrKey string) bool { return attrKey == mockAttrKey },
					}),
					plantest.CreatePhysicalNode("require-attr", plantest.MockProcedureSpec{
						RequiredAttributesFn: func() []plan.PhysicalAttributes {
							return []plan.PhysicalAttributes{
								{
									mockAttrKey: &mockAttr{},
								},
							}
						},
					}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			err: `attribute "mock-attr", required by "require-attr", is missing from predecessor "no-attr"`,
		},
		{
			name: "logical node does not provide attributes",
			input: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("logical", plantest.MockProcedureSpec{}),
					plantest.CreatePhysicalNode("passthru", plantest.MockProcedureSpec{
						PassThroughAttributeFn: func(attrKey string) bool { return attrKey == mockAttrKey },
					}),
					plantest.CreatePhysicalNode("require-attr", plantest.MockProcedureSpec{
						RequiredAttributesFn: func() []plan.PhysicalAttributes {
							return []plan.PhysicalAttributes{
								{
									mockAttrKey: &mockAttr{},
								},
							}
						},
					}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			err: `attribute "mock-attr", required by "require-attr", ` +
				`is missing from predecessor "logical" which is a logical node`,
		},
		{
			name: "attribute present but not satisfied",
			input: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalNode("has-attr", plantest.MockProcedureSpec{
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{
								mockAttrKey: &mockAttr{v: 1},
							}
						},
					}),
					plantest.CreatePhysicalNode("passthru", plantest.MockProcedureSpec{
						PassThroughAttributeFn: func(attrKey string) bool { return attrKey == mockAttrKey },
					}),
					plantest.CreatePhysicalNode("require-attr", plantest.MockProcedureSpec{
						RequiredAttributesFn: func() []plan.PhysicalAttributes {
							return []plan.PhysicalAttributes{
								{
									mockAttrKey: &mockAttr{v: 0},
								},
							}
						},
					}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			err: `node "require-attr" requires attribute mock-attr{v: 0}, which is not satisfied ` +
				`by predecessor "passthru", which has attribute mock-attr{v: 1}`,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			spec := plantest.CreatePlanSpec(tc.input)
			err := spec.BottomUpWalk(func(node plan.Node) error {
				if pn, ok := node.(*plan.PhysicalPlanNode); ok {
					return plan.CheckRequiredAttributes(pn)
				}
				return nil
			})
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				if err == nil {
					t.Fatalf("expected error %q but did not get an error", tc.err)
				}
				require.Equal(t, tc.err, err.Error())
			}
		})
	}
}

func TestCheckSuccessorsMustRequire(t *testing.T) {
	tcs := []struct {
		name  string
		input *plantest.PlanSpec
		err   string
	}{
		{
			name: "valid",
			input: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalNode("successor-must-require", plantest.MockProcedureSpec{
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{
								mockAttrKey: &mockAttr{successorMustRequire: true},
							}
						},
					}),
					plantest.CreatePhysicalNode("passthru", plantest.MockProcedureSpec{
						PassThroughAttributeFn: func(attrKey string) bool { return attrKey == mockAttrKey },
					}),
					plantest.CreatePhysicalNode("requires-attr", plantest.MockProcedureSpec{
						RequiredAttributesFn: func() []plan.PhysicalAttributes {
							return []plan.PhysicalAttributes{
								{
									mockAttrKey: &mockAttr{successorMustRequire: true},
								},
							}
						},
					}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
		},
		{
			name: "attr not required by successor",
			input: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalNode("successor-does-not-require", plantest.MockProcedureSpec{
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{
								mockAttrKey: &mockAttr{successorMustRequire: false},
							}
						},
					}),
					plantest.CreatePhysicalNode("passthru", plantest.MockProcedureSpec{
						PassThroughAttributeFn: func(attrKey string) bool { return attrKey == mockAttrKey },
					}),
					plantest.CreatePhysicalNode("does-not-require", plantest.MockProcedureSpec{}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
		},
		{
			name: "no successor",
			input: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalNode("successor-must-require", plantest.MockProcedureSpec{
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{
								mockAttrKey: &mockAttr{successorMustRequire: true},
							}
						},
					}),
				},
				Edges: [][2]int{},
			},
			err: `node "successor-must-require" provides attribute mock-attr that must be required but has no successors to require it`,
		},
		{
			name: "no requiring successor",
			input: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalNode("successor-must-require", plantest.MockProcedureSpec{
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{
								mockAttrKey: &mockAttr{successorMustRequire: true},
							}
						},
					}),
					plantest.CreatePhysicalNode("passthru", plantest.MockProcedureSpec{
						PassThroughAttributeFn: func(attrKey string) bool { return attrKey == mockAttrKey },
					}),
					plantest.CreatePhysicalNode("does-not-require-attr", plantest.MockProcedureSpec{}),
				},
				Edges: [][2]int{
					{0, 1},
					{1, 2},
				},
			},
			err: `plan node "successor-must-require" has attribute "mock-attr" that must be required by successors, ` +
				`but it is not required or propagated by successor "does-not-require-attr"`,
		},
		{
			name: "successor is logical node",
			input: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalNode("successor-must-require", plantest.MockProcedureSpec{
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{
								mockAttrKey: &mockAttr{successorMustRequire: true},
							}
						},
					}),
					plan.CreateLogicalNode("logical", plantest.MockProcedureSpec{}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			err: `plan node "successor-must-require" has attribute "mock-attr" ` +
				`that must be required by successors, but it is not required or propagated ` +
				`by successor "logical" which is a logical node`,
		},
		{
			name: "no requiring successor passthru",
			input: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plantest.CreatePhysicalNode("successor-must-require", plantest.MockProcedureSpec{
						OutputAttributesFn: func() plan.PhysicalAttributes {
							return plan.PhysicalAttributes{
								mockAttrKey: &mockAttr{successorMustRequire: true},
							}
						},
					}),
					plantest.CreatePhysicalNode("passthru", plantest.MockProcedureSpec{
						PassThroughAttributeFn: func(attrKey string) bool { return attrKey == mockAttrKey },
					}),
				},
				Edges: [][2]int{
					{0, 1},
				},
			},
			err: `plan node "successor-must-require" has attribute "mock-attr" that must be required by successors, ` +
				`but no successors require it`,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			spec := plantest.CreatePlanSpec(tc.input)
			err := spec.BottomUpWalk(func(node plan.Node) error {
				if pn, ok := node.(*plan.PhysicalPlanNode); ok {
					return plan.CheckSuccessorsMustRequire(pn)
				}
				return nil
			})
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				if err == nil {
					t.Fatalf("expected error %q but did not get an error", tc.err)
				}
				require.Equal(t, tc.err, err.Error())
			}
		})
	}
}
