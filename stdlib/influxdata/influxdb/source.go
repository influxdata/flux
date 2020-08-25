package influxdb

import (
	"bytes"
	"context"
	"encoding/json"
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
)

type ProcedureSpec interface {
	GetOrg() *NameOrID
	GetHost() *string
	GetToken() *string
	SetOrg(org *NameOrID)
	SetHost(host *string)
	SetToken(token *string)
}

type RemoteProcedureSpec interface {
	GetOrg() *NameOrID
	GetHost() *string
	GetToken() *string

	BuildQuery() *ast.File
}

type source struct {
	id   execute.DatasetID
	spec RemoteProcedureSpec
	deps flux.Dependencies
	mem  *memory.Allocator
	ts   execute.TransformationSet
}

func CreateSource(id execute.DatasetID, spec RemoteProcedureSpec, a execute.Administration) (execute.Source, error) {
	deps := flux.GetDependencies(a.Context())
	s := &source{
		id:   id,
		spec: spec,
		deps: deps,
		mem:  a.Allocator(),
	}

	if err := s.validateHost(*spec.GetHost()); err != nil {
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
	u, err := url.Parse(*s.spec.GetHost())
	if err != nil {
		return nil, err
	}
	u.Path += "/api/v2/query"
	if org := s.spec.GetOrg(); org != nil {
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
	if token := s.spec.GetToken(); token != nil {
		req.Header.Set("Authorization", "Token "+*token)
	}
	req.Header.Set("Content-Type", "application/json")
	return req.WithContext(ctx), nil
}

func (s *source) newRequestBody() ([]byte, error) {
	var req struct {
		Query   string `json:"query"`
		Dialect struct {
			Header         bool     `json:"header"`
			DateTimeFormat string   `json:"dateTimeFormat"`
			Annotations    []string `json:"annotations"`
		} `json:"dialect"`
	}
	// Build the query. This needs to be done first to build
	// up the list of imports.
	req.Query = ast.Format(s.spec.BuildQuery())
	req.Dialect.Header = true
	req.Dialect.DateTimeFormat = "RFC3339Nano"
	req.Dialect.Annotations = []string{"group", "datatype", "default"}
	return json.Marshal(req)
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
