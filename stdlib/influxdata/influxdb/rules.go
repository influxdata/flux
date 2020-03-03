package influxdb

import (
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

func (p FromRemoteRule) Rewrite(node plan.Node) (plan.Node, bool, error) {
	spec := node.ProcedureSpec().(*FromProcedureSpec)
	if spec.Host == nil {
		return node, false, nil
	}

	return plan.CreatePhysicalNode("fromRemote", &FromRemoteProcedureSpec{
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

func (p MergeRemoteRangeRule) Rewrite(node plan.Node) (plan.Node, bool, error) {
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

func (p MergeRemoteFilterRule) Rewrite(node plan.Node) (plan.Node, bool, error) {
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
	return plan.Pat(FromKind)
}

func (d DefaultFromAttributes) Rewrite(n plan.Node) (plan.Node, bool, error) {
	spec := n.ProcedureSpec().(*FromProcedureSpec)
	changed := false
	if spec.Org == nil && d.Org != nil {
		spec.Org = d.Org
		changed = true
	}
	if spec.Token == nil && d.Token != nil {
		spec.Token = d.Token
		changed = true
	}
	if spec.Host == nil && d.Host != nil {
		spec.Host = d.Host
		changed = true
	}
	return n, changed, nil
}
