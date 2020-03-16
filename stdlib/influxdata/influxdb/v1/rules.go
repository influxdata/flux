package v1

import "github.com/influxdata/flux/plan"

type DatabasesRemoteRule struct{}

func (p DatabasesRemoteRule) Name() string {
	return "influxdata/influxdb.DatabasesRemoteRule"
}

func (p DatabasesRemoteRule) Pattern() plan.Pattern {
	return plan.Pat(DatabasesKind)
}

func (p DatabasesRemoteRule) Rewrite(node plan.Node) (plan.Node, bool, error) {
	spec := node.ProcedureSpec().(*DatabasesProcedureSpec)
	if spec.Host == nil {
		return node, false, nil
	}

	return plan.CreatePhysicalNode("databasesRemote", &DatabasesRemoteProcedureSpec{
		DatabasesProcedureSpec: spec,
	}), true, nil
}
