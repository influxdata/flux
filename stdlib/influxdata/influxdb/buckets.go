package influxdb

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const (
	BucketsKind       = "buckets"
	BucketsRemoteKind = "influxdata/influxdb.bucketsRemote"
)

type BucketsOpSpec struct {
	Org   *NameOrID
	Host  *string
	Token *string
}

func init() {
	bucketsSignature := runtime.MustLookupBuiltinType("influxdata/influxdb", "buckets")

	runtime.RegisterPackageValue("influxdata/influxdb", BucketsKind, flux.MustValue(flux.FunctionValue(BucketsKind, createBucketsOpSpec, bucketsSignature)))
	flux.RegisterOpSpec(BucketsKind, newBucketsOp)
	plan.RegisterProcedureSpec(BucketsKind, newBucketsProcedure, BucketsKind)
	execute.RegisterSource(BucketsRemoteKind, createBucketsSource)
	plan.RegisterPhysicalRules(BucketsRemoteRule{})
}

func createBucketsOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(BucketsOpSpec)

	if o, ok, err := GetNameOrID(args, "org", "orgID"); err != nil {
		return nil, err
	} else if ok {
		spec.Org = &o
	}

	if h, ok, err := args.GetString("host"); err != nil {
		return nil, err
	} else if ok {
		spec.Host = &h
	}

	if token, ok, err := args.GetString("token"); err != nil {
		return nil, err
	} else if ok {
		spec.Token = &token
	}
	return spec, nil
}

func newBucketsOp() flux.OperationSpec {
	return new(BucketsOpSpec)
}

func (s *BucketsOpSpec) Kind() flux.OperationKind {
	return BucketsKind
}

var _ ProcedureSpec = (*BucketsProcedureSpec)(nil)

type BucketsProcedureSpec struct {
	plan.DefaultCost

	Org   *NameOrID
	Host  *string
	Token *string
}

func newBucketsProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(*BucketsOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &BucketsProcedureSpec{
		Org:   s.Org,
		Host:  s.Host,
		Token: s.Token,
	}, nil
}

func (s *BucketsProcedureSpec) Kind() plan.ProcedureKind {
	return BucketsKind
}

func (s *BucketsProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(BucketsProcedureSpec)
	return ns
}

func (s *BucketsProcedureSpec) SetOrg(org *NameOrID)   { s.Org = org }
func (s *BucketsProcedureSpec) SetHost(host *string)   { s.Host = host }
func (s *BucketsProcedureSpec) SetToken(token *string) { s.Token = token }
func (s *BucketsProcedureSpec) GetOrg() *NameOrID      { return s.Org }
func (s *BucketsProcedureSpec) GetHost() *string       { return s.Host }
func (s *BucketsProcedureSpec) GetToken() *string      { return s.Token }

func (s *BucketsProcedureSpec) PostPhysicalValidate(id plan.NodeID) error {
	// This condition should never be met.
	// Customized planner rules within each binary should have
	// filled in either a default host or registered a from procedure
	// for when no host is specified.
	// We mark this as an internal error because it is a programming
	// error if this one ever gets hit.
	if s.Host == nil {
		return errors.New(codes.Internal, "buckets requires a remote host to be specified")
	}
	return nil
}

type BucketsRemoteProcedureSpec struct {
	plan.DefaultCost
	*BucketsProcedureSpec
}

func (s *BucketsRemoteProcedureSpec) Kind() plan.ProcedureKind {
	return BucketsRemoteKind
}

func (s *BucketsRemoteProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(BucketsRemoteProcedureSpec)
	*ns = *s
	ns.BucketsProcedureSpec = s.BucketsProcedureSpec.Copy().(*BucketsProcedureSpec)
	return ns
}

func (s *BucketsRemoteProcedureSpec) PostPhysicalValidate(id plan.NodeID) error {
	if s.Org == nil {
		return errors.New(codes.Invalid, "listing buckets from a remote host requires an organization to be set")
	}
	return nil
}

func createBucketsSource(ps plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec := ps.(*BucketsRemoteProcedureSpec)
	return CreateSource(id, spec, a)
}

func (s *BucketsRemoteProcedureSpec) BuildQuery() *ast.File {
	query := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "buckets"},
	}
	return &ast.File{
		Package: &ast.PackageClause{
			Name: &ast.Identifier{Name: "main"},
		},
		Name: "query.flux",
		Body: []ast.Statement{
			&ast.ExpressionStatement{Expression: query},
		},
	}
}
