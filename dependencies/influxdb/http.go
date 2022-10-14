package influxdb

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	stdhttp "net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/astutil"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies/http"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/influxdb-client-go/v2/api"
	apihttp "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/influxdata/line-protocol/v2/lineprotocol"
)

// HttpProvider is an implementation of the Provider that
// implements the read methods with HTTP calls to an influxdb query
// endpoint.
type HttpProvider struct {
	DefaultConfig Config
}

var _ Provider = HttpProvider{}

func (h HttpProvider) ReaderFor(ctx context.Context, conf Config, bounds flux.Bounds, predicateSet PredicateSet) (Reader, error) {
	c, err := h.clientFor(ctx, conf)
	if err != nil {
		return nil, err
	}
	return filteredHttpReader{
		HttpClient:   c,
		Bounds:       bounds,
		PredicateSet: predicateSet,
	}, nil
}

func (h HttpProvider) SeriesCardinalityReaderFor(ctx context.Context, conf Config, bounds flux.Bounds, predicateSet PredicateSet) (Reader, error) {
	// If any of the predicates use keep empty then they are not
	// valid for series cardinality reader.
	for _, p := range predicateSet {
		if p.KeepEmpty {
			return nil, errors.New(codes.Unimplemented, "keep empty filter option is not allowed for the series cardinality reader")
		}
	}

	// Retrieve the client and create the http reader.
	c, err := h.clientFor(ctx, conf)
	if err != nil {
		return nil, err
	}
	return seriesCardinalityHttpReader{
		HttpClient:   c,
		Bounds:       bounds,
		PredicateSet: predicateSet,
	}, nil
}

func (h HttpProvider) WriterFor(ctx context.Context, conf Config) (Writer, error) {
	httpClient, err := h.clientFor(ctx, conf)
	if err != nil {
		return nil, err
	}

	service := apihttp.NewService(httpClient.Config.Host, "Token "+httpClient.Config.Token, apihttp.DefaultOptions().SetHTTPDoer(httpClient.Client))

	writeOptions := write.DefaultOptions()
	writeOptions.SetMaxRetries(0)
	writer := api.NewWriteAPI(httpClient.Config.Org.IdOrName(), httpClient.Config.Bucket.IdOrName(), service, writeOptions)

	return newHttpWriter(writer)
}

func (h HttpProvider) clientFor(ctx context.Context, conf Config) (*HttpClient, error) {
	deps := flux.GetDependencies(ctx)
	var (
		httpc http.Client
		err   error
	)
	// If no host is specified, it means we're making a request
	// against an internal source. In that case, we want to obscure
	// errors from the http client.
	if conf.Host == "" {
		conf.Host = h.DefaultConfig.Host
		httpc, err = deps.PrivateHTTPClient()
	} else {
		httpc, err = deps.HTTPClient()
	}
	if err != nil {
		return nil, err
	}

	if conf.Org.IsZero() {
		conf.Org = h.DefaultConfig.Org
	}
	if conf.Bucket.IsZero() {
		conf.Bucket = h.DefaultConfig.Bucket
	}
	if err := h.validateHost(deps, conf.Host); err != nil {
		return nil, err
	}
	if conf.Token == "" {
		conf.Token = h.DefaultConfig.Token
	}
	return &HttpClient{
		Client: httpc,
		Config: conf,
	}, nil
}

func (h HttpProvider) validateHost(deps flux.Dependencies, host string) error {
	if host == "" {
		return errors.New(codes.Invalid, "influxdb provider requires a host to be specified")
	}

	validator, err := deps.URLValidator()
	if err != nil {
		return err
	}

	u, err := url.Parse(host)
	if err != nil {
		return err
	}
	return validator.Validate(u)
}

// HttpClient is an http client for reading from an influxdb instance.
type HttpClient struct {
	Client http.Client
	Config Config
}

// Query will create a new http.Request, send it to the server, then
// decode the request as a flux.TableIterator and invoke the function with
// each flux.Table.
func (h *HttpClient) Query(ctx context.Context, f func(table flux.Table) error, file *ast.File, now time.Time, mem memory.Allocator) error {
	req, err := h.newRequest(ctx, file, now)
	if err != nil {
		return err
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Newf(codes.Invalid, "error when reading response body: %s", err)
		}
		return h.parseError(data)
	}
	return h.processResult(resp.Body, f, mem)
}

// newFile constructs a new ast.File with the default values filled in.
func (h *HttpClient) newFile(imports map[string]*ast.ImportDeclaration) ast.File {
	file := ast.File{
		Package: &ast.PackageClause{
			Name: &ast.Identifier{Name: "main"},
		},
		Name: "query.flux",
	}
	file.Imports = make([]*ast.ImportDeclaration, 0, len(imports))
	for _, decl := range imports {
		file.Imports = append(file.Imports, decl)
	}
	sort.Slice(file.Imports, func(i, j int) bool {
		return file.Imports[i].Path.Value < file.Imports[j].Path.Value
	})
	return file
}

// appendFromArgs will append properties for the common from arguments
// in the HttpClient.
func (h *HttpClient) appendFromArgs(properties []*ast.Property) []*ast.Property {
	if properties == nil {
		properties = make([]*ast.Property, 0, 1)
	}

	var arg ast.Property
	if bucket := h.Config.Bucket; bucket.ID != "" {
		arg.Key = &ast.Identifier{Name: "bucketID"}
		arg.Value = &ast.StringLiteral{Value: bucket.ID}
	} else {
		arg.Key = &ast.Identifier{Name: "bucket"}
		arg.Value = &ast.StringLiteral{Value: bucket.Name}
	}
	return append(properties, &arg)
}

// appendRangeArgs will append properties for the common range arguments
// in the HttpClient.
func (h *HttpClient) appendRangeArgs(properties []*ast.Property, bounds flux.Bounds) []*ast.Property {
	if properties == nil {
		properties = make([]*ast.Property, 0, 2)
	}

	properties = append(properties, &ast.Property{
		Key:   &ast.Identifier{Name: "start"},
		Value: ast.DateTimeLiteralFromValue(bounds.Start.Time(bounds.Now)),
	})
	if !bounds.Stop.IsZero() {
		properties = append(properties, &ast.Property{
			Key:   &ast.Identifier{Name: "stop"},
			Value: ast.DateTimeLiteralFromValue(bounds.Stop.Time(bounds.Now)),
		})
	}
	return properties
}

// newRequest will create a new http.Request for the query endpoint.
// The body will be an encoded ast.File.
func (h *HttpClient) newRequest(ctx context.Context, file *ast.File, now time.Time) (*stdhttp.Request, error) {
	u, err := url.Parse(h.Config.Host)
	if err != nil {
		return nil, err
	}
	u.Path += "/api/v2/query"

	if org := h.Config.Org; org.IsValid() {
		u.RawQuery = func() string {
			params := make(url.Values)
			if org.ID != "" {
				params.Set("orgID", org.ID)
			} else {
				params.Set("org", org.Name)
			}
			return params.Encode()
		}()
	}

	body, err := h.newRequestBody(file, now)
	if err != nil {
		return nil, err
	}

	req, err := stdhttp.NewRequest("POST", u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if token := h.Config.Token; token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	return req.WithContext(ctx), nil
}

// newRequestBody will produce a new request body for the http client
// with a formatted ast.File.
func (h *HttpClient) newRequestBody(file *ast.File, now time.Time) ([]byte, error) {
	var req struct {
		Query   string `json:"query"`
		Dialect struct {
			Header         bool     `json:"header"`
			DateTimeFormat string   `json:"dateTimeFormat"`
			Annotations    []string `json:"annotations"`
		} `json:"dialect"`
		Now time.Time `json:"now"`
	}
	query, err := astutil.Format(file)
	if err != nil {
		return nil, err
	}
	req.Query = query
	req.Dialect.Header = true
	req.Dialect.DateTimeFormat = "RFC3339Nano"
	req.Dialect.Annotations = []string{"group", "datatype", "default"}
	req.Now = now
	return json.Marshal(req)
}

// processResult reads a single csv encoded result from the io.Reader.
// Produced tables are passed to the function. If there is more than one
// result, this method will discard any additional results.
func (h *HttpClient) processResult(r io.ReadCloser, f func(flux.Table) error, mem memory.Allocator) error {
	config := csv.ResultDecoderConfig{Allocator: mem}
	dec := csv.NewMultiResultDecoder(config)
	results, err := dec.Decode(r)
	if err != nil {
		return err
	}
	defer results.Release()

	if !results.More() {
		return nil
	}
	res := results.Next()
	if err := res.Tables().Do(f); err != nil {
		return err
	}
	results.Release()
	return results.Err()
}

// parseError will parse an influxdb error.
func (h *HttpClient) parseError(p []byte) error {
	var e interface{}
	if err := json.Unmarshal(p, &e); err != nil {
		return err
	}
	return handleError(e)
}

// functionToAST will convert a resolved function back to its
// ast representation. If the function references any imports,
// this will reimport the values into the new script.
func (h *HttpClient) functionToAST(fn Predicate, imports map[string]*ast.ImportDeclaration) ast.Expression {
	// Iterate through the scope and include any imports.
	fn.Scope.Range(func(k string, v values.Value) {
		pkg, ok := v.(values.Package)
		if !ok {
			return
		}

		pkgpath := pkg.Path()
		if pkgpath == "" {
			return
		}
		h.includeImport(imports, k, pkgpath)
	})
	return semantic.ToAST(fn.Fn).(ast.Expression)
}

// includeImport will include the given import in the list of import declarations.
// It does not resolve name or path conflicts.
func (h *HttpClient) includeImport(imports map[string]*ast.ImportDeclaration, name, path string) {
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

type filteredHttpReader struct {
	*HttpClient
	Bounds       flux.Bounds
	PredicateSet PredicateSet
}

func (h filteredHttpReader) Read(ctx context.Context, f func(flux.Table) error, mem memory.Allocator) error {
	imports := make(map[string]*ast.ImportDeclaration)
	query := &ast.PipeExpression{
		Argument: &ast.CallExpression{
			Callee: &ast.Identifier{Name: "from"},
			Arguments: []ast.Expression{
				&ast.ObjectExpression{
					Properties: h.appendFromArgs(nil),
				},
			},
		},
		Call: &ast.CallExpression{
			Callee: &ast.Identifier{Name: "range"},
			Arguments: []ast.Expression{
				&ast.ObjectExpression{
					Properties: h.appendRangeArgs(nil, h.Bounds),
				},
			},
		},
	}
	for _, predicate := range h.PredicateSet {
		params := []*ast.Property{{
			Key:   &ast.Identifier{Name: "fn"},
			Value: h.functionToAST(predicate, imports),
		}}
		if predicate.KeepEmpty {
			params = append(params, &ast.Property{
				Key:   &ast.Identifier{Name: "onEmpty"},
				Value: ast.StringLiteralFromValue("keep"),
			})
		}
		query = &ast.PipeExpression{
			Argument: query,
			Call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "filter"},
				Arguments: []ast.Expression{
					&ast.ObjectExpression{
						Properties: params,
					},
				},
			},
		}
	}

	file := h.newFile(imports)
	file.Body = []ast.Statement{
		&ast.ExpressionStatement{Expression: query},
	}
	return h.Query(ctx, f, &file, h.Bounds.Now, mem)
}

type seriesCardinalityHttpReader struct {
	*HttpClient
	Bounds       flux.Bounds
	PredicateSet PredicateSet
}

func (h seriesCardinalityHttpReader) Read(ctx context.Context, f func(flux.Table) error, mem memory.Allocator) error {
	properties := make([]*ast.Property, 0, 4)
	properties = h.appendFromArgs(properties)
	properties = h.appendRangeArgs(properties, h.Bounds)

	imports := make(map[string]*ast.ImportDeclaration)
	if len(h.PredicateSet) > 0 {
		predicate := h.functionToAST(h.PredicateSet[0], imports)
		for _, p := range h.PredicateSet[1:] {
			predicate = &ast.LogicalExpression{
				Operator: ast.AndOperator,
				Left:     predicate,
				Right:    h.functionToAST(p, imports),
			}
		}
		properties = append(properties, &ast.Property{
			Key:   &ast.Identifier{Name: "predicate"},
			Value: predicate,
		})
	}

	// Need to find an appropriate name for our required
	// import. Unlike the function, we can name this anything
	// we want. We prefer influxdb but let's try to disambiguate
	// it in case the person used this name for something else.
	const pkgpath = "influxdata/influxdb"
	name, num := "influxdb", 1
	for {
		if decl, ok := imports[name]; ok && decl.Path.Value == pkgpath {
			// Import already present and the correct path.
			// This name is fine to use.
			break
		} else if ok {
			// An import with this name exists, but it didn't
			// match the path we want. We need to use a different
			// name.
			name, num = "influxdb"+strconv.Itoa(num), num+1
			continue
		}
		// Add an import with the present name.
		h.includeImport(imports, name, pkgpath)
	}

	file := h.newFile(imports)
	file.Body = []ast.Statement{
		&ast.ExpressionStatement{
			Expression: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: name},
					Property: &ast.Identifier{Name: "cardinality"},
				},
				Arguments: []ast.Expression{
					&ast.ObjectExpression{Properties: properties},
				},
			},
		},
	}
	return h.Query(ctx, f, &file, h.Bounds.Now, mem)
}

type httpWriter struct {
	writer      *api.WriteAPIImpl
	errChan     <-chan error
	latestError chan error
}

func newHttpWriter(writer *api.WriteAPIImpl) (*httpWriter, error) {
	w := &httpWriter{
		writer:      writer,
		errChan:     writer.Errors(),
		latestError: make(chan error, 1),
	}
	go func() {
		for err := range w.errChan {
			if err != nil {
				select {
				case w.latestError <- handleError(err):
				default:
				}
			}
		}
		close(w.latestError)
	}()
	return w, nil
}

var _ Writer = &httpWriter{}

// Write sends points asynchronously to the underlying write api.
// Errors are returned only on a best-effort basis.
func (h *httpWriter) Write(metric ...Metric) error {
	var enc lineprotocol.Encoder
	for _, m := range metric {
		record, err := h.encodeMetric(&enc, m)
		if err != nil {
			h.writer.Flush()

			// Wrap the error with failed precondition because
			// the error was caused by user data.
			return errors.Wrap(err, codes.FailedPrecondition)
		}

		if len(record) > 0 {
			record = bytes.TrimRight(record, "\n")
			h.writer.WriteRecord(string(record))
		}
		enc.Reset()
	}
	select {
	case err := <-h.latestError:
		return err
	default:
	}
	return nil
}

func (h *httpWriter) encodeMetric(enc *lineprotocol.Encoder, m Metric) ([]byte, error) {
	enc.StartLine(m.Name())

	// The line protocol encoder checks for ordering so we need to ensure that
	// this is sorted. We did not have that requirement previously so we have
	// to accept that input may be unordered.
	//
	// We sort the original slice which seems like it's not too healthy for the
	// API, but this probably isn't a big deal and the extra allocations to
	// allocate a new slice for input that's probably already sorted would kill us.
	//
	// We could set the encoder to be lax, but that would also prevent it from
	// telling us about invalid escape characters that we really do care about.
	tags := m.TagList()
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	// Encode the tags.
	for _, tag := range tags {
		// Skip tags with an empty value.
		if tag.Value == "" {
			continue
		}
		enc.AddTag(tag.Key, tag.Value)
	}

	// Iterate through the fields and try to construct a value from them.
	// If the value is not able to be written, such as NaN or Inf in the case
	// of floats, we skip over the field and don't write anything.
	fields, hasField := m.FieldList(), false
	for _, field := range fields {
		if v, ok := lineprotocol.NewValue(field.Value); ok {
			enc.AddField(field.Key, v)
			hasField = true
		}
	}

	// Check if we encoded at least one field. If we did not attempt to
	// encode a field, just return any associated error if one occurred.
	if !hasField {
		return nil, enc.Err()
	}

	// End the line by writing the timestamp.
	enc.EndLine(m.Time())

	// Return any error that happened while encoding.
	return enc.Bytes(), enc.Err()
}

func (h *httpWriter) Close() error {
	h.writer.Flush()
	h.writer.Close()
	var err error
	// This ensures latestError is closed which ensures errChan is closed
	for e := range h.latestError {
		err = e
	}
	return err
}
