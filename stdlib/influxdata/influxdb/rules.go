package influxdb

import (
	"context"

	"github.com/influxdata/flux/dependencies/influxdb"
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

func (p FromRemoteRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	spec := node.ProcedureSpec().(*FromProcedureSpec)
	if spec.Host == nil {
		return node, false, nil
	}

	config := influxdb.Config{
		Bucket: spec.Bucket,
	}
	if spec.Org != nil {
		config.Org = *spec.Org
	}
	if spec.Host != nil {
		config.Host = *spec.Host
	}
	if spec.Token != nil {
		config.Token = *spec.Token
	}

	return plan.CreateUniquePhysicalNode(ctx, "fromRemote", &FromRemoteProcedureSpec{
		Config: config,
	}), true, nil
}

type MergeRemoteRangeRule struct{}

func (p MergeRemoteRangeRule) Name() string {
	return "influxdata/influxdb.MergeRemoteRangeRule"
}

func (p MergeRemoteRangeRule) Pattern() plan.Pattern {
	return plan.Pat(universe.RangeKind, plan.Pat(FromRemoteKind))
}

func (p MergeRemoteRangeRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	fromNode := node.Predecessors()[0]
	fromSpec := fromNode.ProcedureSpec().(*FromRemoteProcedureSpec)
	if !fromSpec.Bounds.IsEmpty() {
		return node, false, nil
	}

	rangeSpec := node.ProcedureSpec().(*universe.RangeProcedureSpec)
	newFromSpec := fromSpec.Copy().(*FromRemoteProcedureSpec)
	newFromSpec.Bounds = rangeSpec.Bounds
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

func (p MergeRemoteFilterRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	fromNode := node.Predecessors()[0]
	fromSpec := fromNode.ProcedureSpec().(*FromRemoteProcedureSpec)
	if fromSpec.Bounds.IsEmpty() {
		return node, false, nil
	}
	filterSpec := node.ProcedureSpec().(*universe.FilterProcedureSpec)

	// Attempt to construct the new from procedure spec and see
	// it we can create a reader using it.
	fromSpec = fromSpec.Copy().(*FromRemoteProcedureSpec)
	fromSpec.PredicateSet = append(fromSpec.PredicateSet, influxdb.Predicate{
		ResolvedFunction: filterSpec.Fn,
		KeepEmpty:        filterSpec.KeepEmptyTables,
	})

	provider := influxdb.GetProvider(ctx)
	if _, err := provider.ReaderFor(ctx, fromSpec.Config, fromSpec.Bounds, fromSpec.PredicateSet); err != nil {
		// TODO(jsternberg): It might be possible to push part of
		// a predicate and this is done in influxdb. Update this section
		// to also try and split the predicate into multiple sets
		// so we can partially push down a filter.
		return node, false, nil
	}

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

func (p BucketsRemoteRule) Rewrite(ctx context.Context, node plan.Node) (plan.Node, bool, error) {
	spec := node.ProcedureSpec().(*BucketsProcedureSpec)
	if spec.Host == nil {
		return node, false, nil
	}

	return plan.CreateUniquePhysicalNode(ctx, "bucketsRemote", &BucketsRemoteProcedureSpec{
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

func (d DefaultFromAttributes) Rewrite(ctx context.Context, n plan.Node) (plan.Node, bool, error) {
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
