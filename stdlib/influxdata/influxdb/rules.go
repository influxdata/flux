package influxdb

import (
	"context"

	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/stdlib/universe"
)

type FromRemoteRule struct{}

func (p FromRemoteRule) Name() string {
	return "influxdata/influxdb.FromRemoteRule"
}

func (p FromRemoteRule) Pattern() plan.Pattern {
	return plan.Pat(FromKind)
}

func (p FromRemoteRule) Rewrite(ctx context.Context, node plan.Node, nextNodeId *int) (plan.Node, bool, error) {
	spec := node.ProcedureSpec().(*FromProcedureSpec)
	if spec.Host == nil {
		return node, false, nil
	}

	return plan.CreatePhysicalNodeWithId(nextNodeId, "fromRemote", &FromRemoteProcedureSpec{
		FromProcedureSpec: spec,
	}), true, nil
}

type MergeRemoteRangeRule struct{}

func (p MergeRemoteRangeRule) Name() string {
	return "influxdata/influxdb.MergeRemoteRangeRule"
}

func (p MergeRemoteRangeRule) Pattern() plan.Pattern {
	return plan.Pat(universe.RangeKind, plan.Pat(FromRemoteKind))
}

func (p MergeRemoteRangeRule) Rewrite(ctx context.Context, node plan.Node, nextNodeId *int) (plan.Node, bool, error) {
	fromNode := node.Predecessors()[0]
	fromSpec := fromNode.ProcedureSpec().(*FromRemoteProcedureSpec)
	if fromSpec.Range != nil {
		return node, false, nil
	}

	rangeSpec := node.ProcedureSpec().(*universe.RangeProcedureSpec)
	newFromSpec := fromSpec.Copy().(*FromRemoteProcedureSpec)
	newFromSpec.Range = rangeSpec
	n, err := plan.MergeToPhysicalNode(node, fromNode, newFromSpec)
	if err != nil {
		return nil, false, err
	}
	return n, true, nil
}

type MergeRemoteFilterRule struct{}

func (p MergeRemoteFilterRule) Name() string {
	return "influxdata/influxdb.MergeRemoteFilterRule"
}

func (p MergeRemoteFilterRule) Pattern() plan.Pattern {
	return plan.Pat(universe.FilterKind, plan.Pat(FromRemoteKind))
}

func (p MergeRemoteFilterRule) Rewrite(ctx context.Context, node plan.Node, nextNodeId *int) (plan.Node, bool, error) {
	fromNode := node.Predecessors()[0]
	fromSpec := fromNode.ProcedureSpec().(*FromRemoteProcedureSpec)
	if fromSpec.Range == nil {
		return node, false, nil
	}

	fromSpec = fromSpec.Copy().(*FromRemoteProcedureSpec)
	fromSpec.Transformations = append(fromSpec.Transformations, node.ProcedureSpec())

	n, err := plan.MergeToPhysicalNode(node, fromNode, fromSpec)
	if err != nil {
		return nil, false, err
	}
	return n, true, nil
}

type BucketsRemoteRule struct{}

func (p BucketsRemoteRule) Name() string {
	return "influxdata/influxdb.BucketsRemoteRule"
}

func (p BucketsRemoteRule) Pattern() plan.Pattern {
	return plan.Pat(BucketsKind)
}

func (p BucketsRemoteRule) Rewrite(ctx context.Context, node plan.Node, nextNodeId *int) (plan.Node, bool, error) {
	spec := node.ProcedureSpec().(*BucketsProcedureSpec)
	if spec.Host == nil {
		return node, false, nil
	}

	return plan.CreatePhysicalNodeWithId(nextNodeId, "bucketsRemote", &BucketsRemoteProcedureSpec{
		BucketsProcedureSpec: spec,
	}), true, nil
}

// DefaultFromAttributes is used to inject default attributes
// for the various from attributes.
//
// This rule is not added by default. Each process must fill
// out the suitable defaults and add the rule on startup.
type DefaultFromAttributes struct {
	Org   *NameOrID
	Host  *string
	Token *string
}

func (d DefaultFromAttributes) Name() string {
	return "influxdata/influxdb.DefaultFromAttributes"
}

func (d DefaultFromAttributes) Pattern() plan.Pattern {
	return plan.Any()
}

func (d DefaultFromAttributes) Rewrite(ctx context.Context, n plan.Node, nextNodeId *int) (plan.Node, bool, error) {
	spec, ok := n.ProcedureSpec().(ProcedureSpec)
	if !ok {
		return n, false, nil
	}

	changed := false
	if spec.GetOrg() == nil && d.Org != nil {
		spec.SetOrg(d.Org)
		changed = true
	}
	if spec.GetToken() == nil && d.Token != nil {
		spec.SetToken(d.Token)
		changed = true
	}
	if spec.GetHost() == nil && d.Host != nil {
		spec.SetHost(d.Host)
		changed = true
	}
	return n, changed, nil
}
