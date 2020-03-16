package v1

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
)

const (
	DatabasesKind       = "databases"
	DatabasesRemoteKind = "influxdata/influxdb/v1.databasesRemote"
)

type DatabasesOpSpec struct {
	Org   *influxdb.NameOrID
	Host  *string
	Token *string
}

func init() {
	databasesSignature := runtime.MustLookupBuiltinType("influxdata/influxdb/v1", "databases")

	runtime.RegisterPackageValue("influxdata/influxdb/v1", DatabasesKind, flux.MustValue(flux.FunctionValue(DatabasesKind, createDatabasesOpSpec, databasesSignature)))
	flux.RegisterOpSpec(DatabasesKind, newDatabasesOp)
	plan.RegisterProcedureSpec(DatabasesKind, newDatabasesProcedure, DatabasesKind)
	execute.RegisterSource(DatabasesRemoteKind, createDatabasesSource)
	plan.RegisterPhysicalRules(DatabasesRemoteRule{})
}

func createDatabasesOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(DatabasesOpSpec)

	if o, ok, err := influxdb.GetNameOrID(args, "org", "orgID"); err != nil {
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

func newDatabasesOp() flux.OperationSpec {
	return new(DatabasesOpSpec)
}

func (s *DatabasesOpSpec) Kind() flux.OperationKind {
	return DatabasesKind
}

type DatabasesProcedureSpec struct {
	plan.DefaultCost

	Org   *influxdb.NameOrID
	Host  *string
	Token *string
}

func newDatabasesProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*DatabasesOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &DatabasesProcedureSpec{
		Org:   spec.Org,
		Host:  spec.Host,
		Token: spec.Token,
	}, nil
}

func (s *DatabasesProcedureSpec) Kind() plan.ProcedureKind {
	return DatabasesKind
}

func (s *DatabasesProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(DatabasesProcedureSpec)
	*ns = *s
	return ns
}

func (s *DatabasesProcedureSpec) SetOrg(org *influxdb.NameOrID) { s.Org = org }
func (s *DatabasesProcedureSpec) SetHost(host *string)          { s.Host = host }
func (s *DatabasesProcedureSpec) SetToken(token *string)        { s.Token = token }
func (s *DatabasesProcedureSpec) GetOrg() *influxdb.NameOrID    { return s.Org }
func (s *DatabasesProcedureSpec) GetHost() *string              { return s.Host }
func (s *DatabasesProcedureSpec) GetToken() *string             { return s.Token }

func (s *DatabasesProcedureSpec) PostPhysicalValidate(id plan.NodeID) error {
	// This condition should never be met.
	// Customized planner rules within each binary should have
	// filled in either a default host or registered a from procedure
	// for when no host is specified.
	// We mark this as an internal error because it is a programming
	// error if this one ever gets hit.
	if s.Host == nil {
		return errors.New(codes.Internal, "databases requires a remote host to be specified")
	}
	return nil
}

type DatabasesRemoteProcedureSpec struct {
	plan.DefaultCost
	*DatabasesProcedureSpec
}

func (s *DatabasesRemoteProcedureSpec) Kind() plan.ProcedureKind {
	return DatabasesRemoteKind
}

func (s *DatabasesRemoteProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(DatabasesRemoteProcedureSpec)
	*ns = *s
	ns.DatabasesProcedureSpec = s.DatabasesProcedureSpec.Copy().(*DatabasesProcedureSpec)
	return ns
}

func (s *DatabasesRemoteProcedureSpec) PostPhysicalValidate(id plan.NodeID) error {
	if s.Org == nil {
		return errors.New(codes.Invalid, "reading from a remote host requires an organization to be set")
	}
	return nil
}

func createDatabasesSource(ps plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec := ps.(*DatabasesRemoteProcedureSpec)
	return influxdb.CreateSource(id, spec, a)
}

func (s *DatabasesRemoteProcedureSpec) BuildQuery() *ast.File {
	query := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "v1"},
			Property: &ast.Identifier{Name: "databases"},
		},
	}
	return &ast.File{
		Name: "query.flux",
		Package: &ast.PackageClause{
			Name: &ast.Identifier{Name: "main"},
		},
		Imports: []*ast.ImportDeclaration{{
			Path: &ast.StringLiteral{Value: "influxdata/influxdb/v1"},
		}},
		Body: []ast.Statement{
			&ast.ExpressionStatement{Expression: query},
		},
	}
}
