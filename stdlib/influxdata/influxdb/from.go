// From is an operation that mocks the real implementation of InfluxDB's from.
// It is used in Flux to compile queries that resemble real queries issued against InfluxDB.
// Implementors of the real from are expected to replace its implementation via flux.ReplacePackageValue.
package influxdb

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
)

const (
	FromKind       = "from"
	FromRemoteKind = "influxdata/influxdb.fromRemote"
)

// NameOrID signifies the name of an organization/bucket
// or an ID for an organization/bucket.
type NameOrID struct {
	ID   string
	Name string
}

type FromOpSpec struct {
	Org    *NameOrID
	Bucket NameOrID
	Host   *string
	Token  *string
}

func init() {
	fromSignature := runtime.MustLookupBuiltinType("influxdata/influxdb", "from")

	runtime.RegisterPackageValue("influxdata/influxdb", FromKind, flux.MustValue(flux.FunctionValue(FromKind, createFromOpSpec, fromSignature)))
	flux.RegisterOpSpec(FromKind, newFromOp)
	plan.RegisterProcedureSpec(FromKind, newFromProcedure, FromKind)
	execute.RegisterSource(FromRemoteKind, createFromSource)
	plan.RegisterPhysicalRules(
		FromRemoteRule{},
		MergeRemoteRangeRule{},
		MergeRemoteFilterRule{},
	)
}

func createFromOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromOpSpec)

	if b, ok, err := GetNameOrID(args, "bucket", "bucketID"); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New(codes.Invalid, "must specify only one of bucket or bucketID")
	} else {
		spec.Bucket = b
	}

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

func GetNameOrID(args flux.Arguments, nameParam, idParam string) (NameOrID, bool, error) {
	name, nameOk, err := args.GetString(nameParam)
	if err != nil {
		return NameOrID{}, false, err
	}

	id, idOk, err := args.GetString(idParam)
	if err != nil {
		return NameOrID{}, false, err
	}

	if nameOk && idOk {
		return NameOrID{}, false, errors.Newf(codes.Invalid, "must specify one of %s or %s", nameParam, idParam)
	}
	return NameOrID{Name: name, ID: id}, nameOk || idOk, nil
}

func newFromOp() flux.OperationSpec {
	return new(FromOpSpec)
}

func (s *FromOpSpec) Kind() flux.OperationKind {
	return FromKind
}

var _ ProcedureSpec = (*FromProcedureSpec)(nil)

type FromProcedureSpec struct {
	plan.DefaultCost

	Org    *NameOrID
	Bucket NameOrID
	Host   *string
	Token  *string
}

func newFromProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &FromProcedureSpec{
		Org:    spec.Org,
		Bucket: spec.Bucket,
		Host:   spec.Host,
		Token:  spec.Token,
	}, nil
}

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return FromKind
}

func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromProcedureSpec)
	*ns = *s
	return ns
}

func (s *FromProcedureSpec) SetOrg(org *NameOrID)   { s.Org = org }
func (s *FromProcedureSpec) SetHost(host *string)   { s.Host = host }
func (s *FromProcedureSpec) SetToken(token *string) { s.Token = token }
func (s *FromProcedureSpec) GetOrg() *NameOrID      { return s.Org }
func (s *FromProcedureSpec) GetHost() *string       { return s.Host }
func (s *FromProcedureSpec) GetToken() *string      { return s.Token }

func (s *FromProcedureSpec) PostPhysicalValidate(id plan.NodeID) error {
	// This condition should never be met.
	// Customized planner rules within each binary should have
	// filled in either a default host or registered a from procedure
	// for when no host is specified.
	// We mark this as an internal error because it is a programming
	// error if this one ever gets hit.
	if s.Host == nil {
		return errors.New(codes.Internal, "from requires a remote host to be specified")
	}
	return nil
}

type FromRemoteProcedureSpec struct {
	plan.DefaultCost

	*FromProcedureSpec
	Range           *universe.RangeProcedureSpec
	Transformations []plan.ProcedureSpec
}

func (s *FromRemoteProcedureSpec) Kind() plan.ProcedureKind {
	return FromRemoteKind
}

func (s *FromRemoteProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromRemoteProcedureSpec)
	*ns = *s
	ns.FromProcedureSpec = s.FromProcedureSpec.Copy().(*FromProcedureSpec)
	if s.Range != nil {
		ns.Range = s.Range.Copy().(*universe.RangeProcedureSpec)
	}
	if len(s.Transformations) > 0 {
		// Add an extra slot for a transformation in anticipation
		// of one being appended.
		ns.Transformations = make([]plan.ProcedureSpec, len(s.Transformations), len(s.Transformations)+1)
		copy(ns.Transformations, s.Transformations)
	}
	return ns
}

func (s *FromRemoteProcedureSpec) PostPhysicalValidate(id plan.NodeID) error {
	if s.Org == nil {
		return errors.New(codes.Invalid, "reading from a remote host requires an organization to be set")
	} else if s.Range == nil {
		var bucket string
		if s.Bucket.Name != "" {
			bucket = s.Bucket.Name
		} else {
			bucket = s.Bucket.ID
		}
		return errors.Newf(codes.Invalid, "cannot submit unbounded read to %q; try bounding 'from' with a call to 'range'", bucket)
	}
	return nil
}

func createFromSource(ps plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec := ps.(*FromRemoteProcedureSpec)
	if spec.Range == nil {
		return nil, errors.Newf(codes.Invalid, "bounds must be set")
	}
	return CreateSource(id, spec, a)
}

func (s *FromRemoteProcedureSpec) BuildQuery() *ast.File {
	imports := make(map[string]*ast.ImportDeclaration)
	query := &ast.PipeExpression{
		Argument: &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "from"},
			Arguments: []ast.Expression{s.fromArgs()},
		},
		Call: &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "range"},
			Arguments: []ast.Expression{s.rangeArgs()},
		},
	}
	for _, ps := range s.Transformations {
		query = &ast.PipeExpression{
			Argument: query,
			Call:     s.toAST(ps, imports),
		}
	}
	file := &ast.File{
		Package: &ast.PackageClause{
			Name: &ast.Identifier{Name: "main"},
		},
		Name: "query.flux",
		Body: []ast.Statement{
			&ast.ExpressionStatement{Expression: query},
		},
	}

	if len(imports) > 0 {
		file.Imports = make([]*ast.ImportDeclaration, 0, len(imports))
		for _, decl := range imports {
			file.Imports = append(file.Imports, decl)
		}
	}
	return file
}

func (s *FromRemoteProcedureSpec) fromArgs() *ast.ObjectExpression {
	var arg ast.Property
	if s.Bucket.ID != "" {
		arg.Key = &ast.Identifier{Name: "bucketID"}
		arg.Value = &ast.StringLiteral{Value: s.Bucket.ID}
	} else {
		arg.Key = &ast.Identifier{Name: "bucket"}
		arg.Value = &ast.StringLiteral{Value: s.Bucket.Name}
	}
	return &ast.ObjectExpression{
		Properties: []*ast.Property{&arg},
	}
}

func (s *FromRemoteProcedureSpec) rangeArgs() *ast.ObjectExpression {
	toLiteral := func(t flux.Time) ast.Expression {
		if t.IsRelative {
			// TODO(jsternberg): This seems wrong. Relative should be a values.Duration
			// and not a time.Duration.
			d := flux.ConvertDuration(t.Relative)
			var expr ast.Expression = &ast.DurationLiteral{
				Values: d.AsValues(),
			}
			if d.IsNegative() {
				expr = &ast.UnaryExpression{
					Operator: ast.SubtractionOperator,
					Argument: expr,
				}
			}
			return expr
		}
		return &ast.DateTimeLiteral{Value: t.Absolute}
	}

	args := make([]*ast.Property, 0, 2)
	args = append(args, &ast.Property{
		Key:   &ast.Identifier{Name: "start"},
		Value: toLiteral(s.Range.Bounds.Start),
	})
	if stop := s.Range.Bounds.Stop; !stop.IsZero() && !(stop.IsRelative && stop.Relative == 0) {
		args = append(args, &ast.Property{
			Key:   &ast.Identifier{Name: "stop"},
			Value: toLiteral(s.Range.Bounds.Stop),
		})
	}
	return &ast.ObjectExpression{Properties: args}
}

func (s *FromRemoteProcedureSpec) toAST(spec plan.ProcedureSpec, imports map[string]*ast.ImportDeclaration) *ast.CallExpression {
	switch spec := spec.(type) {
	case *universe.FilterProcedureSpec:
		return s.filterToAST(spec, imports)
	default:
		panic(fmt.Sprintf("unable to convert procedure spec of type %T to ast", spec))
	}
}

func (s *FromRemoteProcedureSpec) filterToAST(spec *universe.FilterProcedureSpec, imports map[string]*ast.ImportDeclaration) *ast.CallExpression {
	// Iterate through the scope and include any imports.
	spec.Fn.Scope.Range(func(k string, v values.Value) {
		pkg, ok := v.(values.Package)
		if !ok {
			return
		}

		pkgpath := pkg.Path()
		if pkgpath == "" {
			return
		}
		s.includeImport(imports, k, pkgpath)
	})

	fn := semantic.ToAST(spec.Fn.Fn).(ast.Expression)
	properties := []*ast.Property{{
		Key:   &ast.Identifier{Name: "fn"},
		Value: fn,
	}}
	if spec.KeepEmptyTables {
		properties = append(properties, &ast.Property{
			Key:   &ast.Identifier{Name: "onEmpty"},
			Value: &ast.StringLiteral{Value: "keep"},
		})
	}
	return &ast.CallExpression{
		Callee: &ast.Identifier{Name: "filter"},
		Arguments: []ast.Expression{
			&ast.ObjectExpression{Properties: properties},
		},
	}
}

func (s *FromRemoteProcedureSpec) includeImport(imports map[string]*ast.ImportDeclaration, name, path string) {
	// Look to see if we have already included an import
	// with this name.
	if _, ok := imports[name]; ok {
		return
	}

	decl := &ast.ImportDeclaration{
		Path: &ast.StringLiteral{Value: path},
		As:   &ast.Identifier{Name: name},
	}
	imports[name] = decl
}
