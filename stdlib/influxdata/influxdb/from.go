// From is an operation that mocks the real implementation of InfluxDB's from.
// It is used in Flux to compile queries that resemble real queries issued against InfluxDB.
// Implementors of the real from are expected to replace its implementation via flux.ReplacePackageValue.
package influxdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
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

	if b, ok, err := getNameOrID(args, "bucket", "bucketID"); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New(codes.Invalid, "must specify only one of bucket or bucketID")
	} else {
		spec.Bucket = b
	}

	if o, ok, err := getNameOrID(args, "org", "orgID"); err != nil {
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

func getNameOrID(args flux.Arguments, nameParam, idParam string) (NameOrID, bool, error) {
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

type source struct {
	id      execute.DatasetID
	spec    *FromRemoteProcedureSpec
	deps    flux.Dependencies
	mem     *memory.Allocator
	ts      execute.TransformationSet
	imports map[string]*ast.ImportDeclaration
}

func createFromSource(ps plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec := ps.(*FromRemoteProcedureSpec)
	if spec.Range == nil {
		return nil, errors.Newf(codes.Invalid, "bounds must be set")
	}

	// These parameters are only required for the remote influxdb
	// source. If running flux within influxdb, these aren't
	// required.
	if spec.Org == nil {
		return nil, errors.Newf(codes.Invalid, "org must be set")
	}

	deps := flux.GetDependencies(a.Context())
	s := &source{
		id:   id,
		spec: spec,
		deps: deps,
		mem:  a.Allocator(),
	}

	if err := s.validateHost(*spec.Host); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *source) AddTransformation(t execute.Transformation) {
	s.ts = append(s.ts, t)
}

func (s *source) Run(ctx context.Context) {
	err := s.run(ctx)
	s.ts.Finish(s.id, err)
}

func (s *source) run(ctx context.Context) error {
	req, err := s.newRequest(ctx)
	if err != nil {
		return err
	}

	client, err := s.deps.HTTPClient()
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Newf(codes.Invalid, "error when reading response body: %s", err)
		}
		return s.parseError(data)
	}
	return s.processResults(resp.Body)
}

func (s *source) validateHost(host string) error {
	validator, err := s.deps.URLValidator()
	if err != nil {
		return err
	}

	u, err := url.Parse(host)
	if err != nil {
		return err
	}
	return validator.Validate(u)
}

func (s *source) newRequest(ctx context.Context) (*http.Request, error) {
	u, err := url.Parse(*s.spec.Host)
	if err != nil {
		return nil, err
	}
	u.Path += "/api/v2/query"
	u.RawQuery = func() string {
		params := make(url.Values)
		if s.spec.Org.ID != "" {
			params.Set("orgID", s.spec.Org.ID)
		} else {
			params.Set("org", s.spec.Org.Name)
		}
		return params.Encode()
	}()

	// Validate that the produced url is allowed.
	urlv, err := s.deps.URLValidator()
	if err != nil {
		return nil, err
	}

	if err := urlv.Validate(u); err != nil {
		return nil, err
	}

	body, err := s.newRequestBody()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if s.spec.Token != nil {
		req.Header.Set("Authorization", "Token "+*s.spec.Token)
	}
	req.Header.Set("Content-Type", "application/json")
	return req.WithContext(ctx), nil
}

func (s *source) newRequestBody() ([]byte, error) {
	var req struct {
		AST     *ast.Package `json:"ast"`
		Dialect struct {
			Header         bool     `json:"header"`
			DateTimeFormat string   `json:"dateTimeFormat"`
			Annotations    []string `json:"annotations"`
		} `json:"dialect"`
	}
	// Build the query. This needs to be done first to build
	// up the list of imports.
	query := s.buildQuery()
	req.AST = &ast.Package{
		Package: "main",
		Files: []*ast.File{{
			Package: &ast.PackageClause{
				Name: &ast.Identifier{Name: "main"},
			},
			Imports: s.getImports(),
			Name:    "query.flux",
			Body: []ast.Statement{
				&ast.ExpressionStatement{Expression: query},
			},
		}},
	}
	req.Dialect.Header = true
	req.Dialect.DateTimeFormat = "RFC3339Nano"
	req.Dialect.Annotations = []string{"group", "datatype", "default"}
	return json.Marshal(req)
}

func (s *source) buildQuery() ast.Expression {
	expr := &ast.PipeExpression{
		Argument: &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "from"},
			Arguments: []ast.Expression{s.fromArgs()},
		},
		Call: &ast.CallExpression{
			Callee:    &ast.Identifier{Name: "range"},
			Arguments: []ast.Expression{s.rangeArgs()},
		},
	}
	for _, ps := range s.spec.Transformations {
		expr = &ast.PipeExpression{
			Argument: expr,
			Call:     s.toAST(ps),
		}
	}
	return expr
}

func (s *source) fromArgs() *ast.ObjectExpression {
	var arg ast.Property
	if s.spec.Bucket.ID != "" {
		arg.Key = &ast.Identifier{Name: "bucketID"}
		arg.Value = &ast.StringLiteral{Value: s.spec.Bucket.ID}
	} else {
		arg.Key = &ast.Identifier{Name: "bucket"}
		arg.Value = &ast.StringLiteral{Value: s.spec.Bucket.Name}
	}
	return &ast.ObjectExpression{
		Properties: []*ast.Property{&arg},
	}
}

func (s *source) rangeArgs() *ast.ObjectExpression {
	toLiteral := func(t flux.Time) ast.Literal {
		if t.IsRelative {
			// TODO(jsternberg): This seems wrong. Relative should be a values.Duration
			// and not a time.Duration.
			d := flux.ConvertDuration(t.Relative)
			return &ast.DurationLiteral{Values: d.AsValues()}
		}
		return &ast.DateTimeLiteral{Value: t.Absolute}
	}

	args := make([]*ast.Property, 0, 2)
	args = append(args, &ast.Property{
		Key:   &ast.Identifier{Name: "start"},
		Value: toLiteral(s.spec.Range.Bounds.Start),
	})
	if stop := s.spec.Range.Bounds.Stop; !stop.IsZero() && !(stop.IsRelative && stop.Relative == 0) {
		args = append(args, &ast.Property{
			Key:   &ast.Identifier{Name: "stop"},
			Value: toLiteral(s.spec.Range.Bounds.Stop),
		})
	}
	return &ast.ObjectExpression{Properties: args}
}

func (s *source) processResults(r io.ReadCloser) error {
	defer func() { _ = r.Close() }()

	config := csv.ResultDecoderConfig{Allocator: s.mem}
	dec := csv.NewMultiResultDecoder(config)
	results, err := dec.Decode(r)
	if err != nil {
		return err
	}
	defer results.Release()

	for results.More() {
		res := results.Next()
		if err := res.Tables().Do(func(table flux.Table) error {
			return s.ts.Process(s.id, table)
		}); err != nil {
			return err
		}
	}
	results.Release()
	return results.Err()
}

func (s *source) parseError(p []byte) error {
	var e interface{}
	if err := json.Unmarshal(p, &e); err != nil {
		return err
	}
	return handleError(e)
}

func (s *source) toAST(spec plan.ProcedureSpec) *ast.CallExpression {
	switch spec := spec.(type) {
	case *universe.FilterProcedureSpec:
		return s.filterToAST(spec)
	default:
		panic(fmt.Sprintf("unable to convert procedure spec of type %T to ast", spec))
	}
}

func (s *source) filterToAST(spec *universe.FilterProcedureSpec) *ast.CallExpression {
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
		s.includeImport(k, pkgpath)
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

func (s *source) includeImport(name, path string) {
	// Look to see if we have already included an import
	// with this name.
	if _, ok := s.imports[name]; ok {
		return
	}

	if s.imports == nil {
		s.imports = make(map[string]*ast.ImportDeclaration)
	}
	decl := &ast.ImportDeclaration{
		Path: &ast.StringLiteral{Value: path},
		As:   &ast.Identifier{Name: name},
	}
	s.imports[name] = decl
}

func (s *source) getImports() []*ast.ImportDeclaration {
	if len(s.imports) == 0 {
		return nil
	}

	decls := make([]*ast.ImportDeclaration, 0, len(s.imports))
	for _, decl := range s.imports {
		decls = append(decls, decl)
	}
	return decls
}
