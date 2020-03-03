package influxdb_test

import (
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values/valuestest"
)

func TestFromRemoteRule_WithHost(t *testing.T) {
	fromSpec := influxdb.FromProcedureSpec{
		Org:    &influxdb.NameOrID{Name: "influxdata"},
		Bucket: influxdb.NameOrID{Name: "telegraf"},
		Host:   stringPtr("http://localhost:9999"),
	}
	rangeSpec := universe.RangeProcedureSpec{
		Bounds: flux.Bounds{
			Start: flux.Time{
				IsRelative: true,
				Relative:   -time.Minute,
			},
			Stop: flux.Time{
				IsRelative: true,
			},
		},
	}

	tc := plantest.RuleTestCase{
		Name: "with host",
		Rules: []plan.Rule{
			influxdb.FromRemoteRule{},
		},
		Before: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreateLogicalNode("from", &fromSpec),
				plan.CreateLogicalNode("range", &rangeSpec),
			},
			Edges: [][2]int{{0, 1}},
		},
		After: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreatePhysicalNode("fromRemote", &influxdb.FromRemoteProcedureSpec{
					FromProcedureSpec: &fromSpec,
				}),
				plan.CreateLogicalNode("range", &rangeSpec),
			},
			Edges: [][2]int{{0, 1}},
		},
	}
	plantest.PhysicalRuleTestHelper(t, &tc)
}

func TestFromRemoteRule_WithoutHost(t *testing.T) {
	fromSpec := influxdb.FromProcedureSpec{
		Bucket: influxdb.NameOrID{Name: "telegraf"},
	}
	rangeSpec := universe.RangeProcedureSpec{
		Bounds: flux.Bounds{
			Start: flux.Time{
				IsRelative: true,
				Relative:   -time.Minute,
			},
			Stop: flux.Time{
				IsRelative: true,
			},
		},
	}

	tc := plantest.RuleTestCase{
		Name: "without host",
		Rules: []plan.Rule{
			influxdb.FromRemoteRule{},
		},
		Before: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreateLogicalNode("from", &fromSpec),
				plan.CreateLogicalNode("range", &rangeSpec),
			},
			Edges: [][2]int{{0, 1}},
		},
		NoChange: true,
	}
	plantest.PhysicalRuleTestHelper(t, &tc)
}

func TestFromRemoteRule_WithoutHostValidation(t *testing.T) {
	fromSpec := influxdb.FromProcedureSpec{
		Bucket: influxdb.NameOrID{Name: "telegraf"},
	}
	rangeSpec := universe.RangeProcedureSpec{
		Bounds: flux.Bounds{
			Start: flux.Time{
				IsRelative: true,
				Relative:   -time.Minute,
			},
			Stop: flux.Time{
				IsRelative: true,
			},
		},
	}

	tc := plantest.RuleTestCase{
		Name: "without host validation",
		Rules: []plan.Rule{
			influxdb.FromRemoteRule{},
		},
		Before: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreateLogicalNode("from", &fromSpec),
				plan.CreateLogicalNode("range", &rangeSpec),
			},
			Edges: [][2]int{{0, 1}},
		},
		ValidateError: errors.New(codes.Internal, "from requires a remote host to be specified"),
	}
	plantest.PhysicalRuleTestHelper(t, &tc)
}

func TestFromRemoteRule_WithoutOrgValidation(t *testing.T) {
	fromSpec := influxdb.FromProcedureSpec{
		Bucket: influxdb.NameOrID{Name: "telegraf"},
		Host:   stringPtr("http://localhost:9999"),
	}
	rangeSpec := universe.RangeProcedureSpec{
		Bounds: flux.Bounds{
			Start: flux.Time{
				IsRelative: true,
				Relative:   -time.Minute,
			},
			Stop: flux.Time{
				IsRelative: true,
			},
		},
	}

	tc := plantest.RuleTestCase{
		Name: "without org validation",
		Rules: []plan.Rule{
			influxdb.FromRemoteRule{},
		},
		Before: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreateLogicalNode("from", &fromSpec),
				plan.CreateLogicalNode("range", &rangeSpec),
			},
			Edges: [][2]int{{0, 1}},
		},
		ValidateError: errors.New(codes.Invalid, "reading from a remote host requires an organization to be set"),
	}
	plantest.PhysicalRuleTestHelper(t, &tc)
}

func TestFromRemoteRule_WithoutRangeValidation(t *testing.T) {
	fromSpec := influxdb.FromProcedureSpec{
		Org:    &influxdb.NameOrID{Name: "influxdata"},
		Bucket: influxdb.NameOrID{Name: "telegraf"},
		Host:   stringPtr("http://localhost:9999"),
	}

	tc := plantest.RuleTestCase{
		Name: "without range validation",
		Rules: []plan.Rule{
			influxdb.FromRemoteRule{},
		},
		Before: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreateLogicalNode("from", &fromSpec),
			},
		},
		ValidateError: errors.New(codes.Invalid, "cannot submit unbounded read to \"telegraf\"; try bounding 'from' with a call to 'range'"),
	}
	plantest.PhysicalRuleTestHelper(t, &tc)
}

func TestMergeRemoteRangeRule(t *testing.T) {
	fromSpec := influxdb.FromProcedureSpec{
		Bucket: influxdb.NameOrID{Name: "telegraf"},
		Host:   stringPtr("http://localhost:9999"),
	}
	rangeSpec := universe.RangeProcedureSpec{
		Bounds: flux.Bounds{
			Start: flux.Time{
				IsRelative: true,
				Relative:   -time.Minute,
			},
			Stop: flux.Time{
				IsRelative: true,
			},
		},
	}

	tc := plantest.RuleTestCase{
		Name: "MergeRemoteRange",
		Rules: []plan.Rule{
			influxdb.FromRemoteRule{},
			influxdb.MergeRemoteRangeRule{},
		},
		Before: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreateLogicalNode("from", &fromSpec),
				plan.CreateLogicalNode("range", &rangeSpec),
			},
			Edges: [][2]int{{0, 1}},
		},
		After: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreatePhysicalNode("merged_fromRemote_range", &influxdb.FromRemoteProcedureSpec{
					FromProcedureSpec: &fromSpec,
					Range:             &rangeSpec,
				}),
			},
		},
	}
	plantest.PhysicalRuleTestHelper(t, &tc)
}

func TestMergeRemoteFilterRule(t *testing.T) {
	fromSpec := influxdb.FromProcedureSpec{
		Bucket: influxdb.NameOrID{Name: "telegraf"},
		Host:   stringPtr("http://localhost:9999"),
	}
	rangeSpec := universe.RangeProcedureSpec{
		Bounds: flux.Bounds{
			Start: flux.Time{
				IsRelative: true,
				Relative:   -time.Minute,
			},
			Stop: flux.Time{
				IsRelative: true,
			},
		},
	}
	filterSpec := universe.FilterProcedureSpec{
		Fn: interpreter.ResolvedFunction{
			Fn:    executetest.FunctionExpression(t, `(r) => r._value > 0.0`),
			Scope: valuestest.Scope(),
		},
	}

	tc := plantest.RuleTestCase{
		Name: "MergeRemoteRange",
		Rules: []plan.Rule{
			influxdb.FromRemoteRule{},
			influxdb.MergeRemoteRangeRule{},
			influxdb.MergeRemoteFilterRule{},
		},
		Before: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreateLogicalNode("from", &fromSpec),
				plan.CreateLogicalNode("range", &rangeSpec),
				plan.CreateLogicalNode("filter", &filterSpec),
			},
			Edges: [][2]int{
				{0, 1},
				{1, 2},
			},
		},
		After: &plantest.PlanSpec{
			Nodes: []plan.Node{
				plan.CreatePhysicalNode("merged_fromRemote_range_filter", &influxdb.FromRemoteProcedureSpec{
					FromProcedureSpec: &fromSpec,
					Range:             &rangeSpec,
					Transformations: []plan.ProcedureSpec{
						&filterSpec,
					},
				}),
			},
		},
	}
	plantest.PhysicalRuleTestHelper(t, &tc)
}

func TestDefaultFromAttributes(t *testing.T) {
	for _, tc := range []plantest.RuleTestCase{
		{
			Name: "all defaults",
			Rules: []plan.Rule{
				influxdb.DefaultFromAttributes{
					Org:   &influxdb.NameOrID{Name: "influxdata"},
					Host:  stringPtr("http://localhost:9999"),
					Token: stringPtr("mytoken"),
				},
			},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("from", &influxdb.FromProcedureSpec{
						Bucket: influxdb.NameOrID{Name: "telegraf"},
					}),
				},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("from", &influxdb.FromProcedureSpec{
						Org:    &influxdb.NameOrID{Name: "influxdata"},
						Bucket: influxdb.NameOrID{Name: "telegraf"},
						Host:   stringPtr("http://localhost:9999"),
						Token:  stringPtr("mytoken"),
					}),
				},
			},
		},
		{
			Name: "no defaults",
			Rules: []plan.Rule{
				influxdb.DefaultFromAttributes{
					Org:   &influxdb.NameOrID{Name: "influxdata"},
					Host:  stringPtr("http://localhost:9999"),
					Token: stringPtr("mytoken"),
				},
			},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("from", &influxdb.FromProcedureSpec{
						Org:    &influxdb.NameOrID{Name: "alternate_org"},
						Bucket: influxdb.NameOrID{Name: "telegraf"},
						Host:   stringPtr("http://mysupersecretserver:9999"),
						Token:  stringPtr("differenttoken"),
					}),
				},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("from", &influxdb.FromProcedureSpec{
						Org:    &influxdb.NameOrID{Name: "alternate_org"},
						Bucket: influxdb.NameOrID{Name: "telegraf"},
						Host:   stringPtr("http://mysupersecretserver:9999"),
						Token:  stringPtr("differenttoken"),
					}),
				},
			},
		},
		{
			Name: "with remote from",
			Rules: []plan.Rule{
				influxdb.FromRemoteRule{},
				influxdb.DefaultFromAttributes{
					Host: stringPtr("http://localhost:9999"),
				},
			},
			Before: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreateLogicalNode("from", &influxdb.FromProcedureSpec{
						Bucket: influxdb.NameOrID{Name: "telegraf"},
					}),
				},
			},
			After: &plantest.PlanSpec{
				Nodes: []plan.Node{
					plan.CreatePhysicalNode("fromRemote", &influxdb.FromRemoteProcedureSpec{
						FromProcedureSpec: &influxdb.FromProcedureSpec{
							Bucket: influxdb.NameOrID{Name: "telegraf"},
							Host:   stringPtr("http://localhost:9999"),
						},
					}),
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			plantest.PhysicalRuleTestHelper(t, &tc)
		})
	}
}
