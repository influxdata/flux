package testutil

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/csv"
	influxdeps "github.com/influxdata/flux/dependencies/influxdb"
	urldeps "github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/plan/plantest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"go.uber.org/zap/zaptest"
)

type (
	T = testing.T
	B = testing.B
)

type Want struct {
	Params url.Values
	Query  string
	Tables func() []*executetest.Table
	Err    error
}

func RunSourceTestHelper(t *testing.T, spec plan.PhysicalProcedureSpec, want Want) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if want, got := "/api/v2/query", r.URL.Path; want != got {
			t.Errorf("unexpected query path -want/+got:\n- %q\n+ %q", want, got)
		}
		if want, got := want.Params, r.URL.Query(); !cmp.Equal(want, got) {
			t.Errorf("unexpected query params -want/+got:\n%s", cmp.Diff(want, got))
		}
		if want, got := "application/json", r.Header.Get("Content-Type"); want != got {
			t.Errorf("unexpected query content type -want/+got:\n- %q\n+ %q", want, got)
			return
		}

		var req struct {
			Query   string `json:"query"`
			Dialect struct {
				Header         bool     `json:"header"`
				DateTimeFormat string   `json:"dateTimeFormat"`
				Annotations    []string `json:"annotations"`
			} `json:"dialect"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("client did not send json: %s", err)
			return
		}

		if want, got := want.Query, req.Query; !cmp.Equal(want, got) {
			t.Errorf("unexpected query in request body -want/+got:\n%s", cmp.Diff(want, got))
		}

		w.Header().Add("Content-Type", "text/csv")
		results := flux.NewSliceResultIterator([]flux.Result{
			&executetest.Result{
				Nm:   "_result",
				Tbls: want.Tables(),
			},
		})
		enc := csv.NewMultiResultEncoder(csv.ResultEncoderConfig{
			Annotations: req.Dialect.Annotations,
			NoHeader:    !req.Dialect.Header,
			Delimiter:   ',',
		})
		if _, err := enc.Encode(w, results); err != nil {
			t.Errorf("error encoding results: %s", err)
		}
	}))
	defer server.Close()

	if ps, ok := spec.(influxdb.ProcedureSpec); ok {
		ps.SetHost(&server.URL)
	}

	provider := influxdeps.Dependency{
		Provider: influxdeps.HttpProvider{
			DefaultConfig: influxdeps.Config{
				Host: server.URL,
			},
		},
	}

	deps := flux.NewDefaultDependencies()
	ctx := deps.Inject(context.Background())
	ctx = provider.Inject(ctx)
	ExecuteSourceTestHelper(t, ctx, spec, want)
}

func RunSourceErrorTestHelper(t *testing.T, spec plan.PhysicalProcedureSpec) {
	t.Helper()

	for _, tt := range []struct {
		name string
		fn   func(w http.ResponseWriter)
		want error
	}{
		{
			name: "internal error",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = io.WriteString(w, `{"code":"internal error","message":"An internal error has occurred"}`)
			},
			want: errors.New(codes.Internal, "An internal error has occurred"),
		},
		{
			name: "not found",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = io.WriteString(w, `{"code":"not found","message":"bucket not found"}`)
			},
			want: errors.New(codes.NotFound, "bucket not found"),
		},
		{
			name: "invalid",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"code":"invalid","message":"query was invalid"}`)
			},
			want: errors.New(codes.Invalid, "query was invalid"),
		},
		{
			name: "unavailable",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = io.WriteString(w, `{"code":"unavailable","message":"service unavailable"}`)
			},
			want: errors.New(codes.Unavailable, "service unavailable"),
		},
		{
			name: "forbidden",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = io.WriteString(w, `{"code":"forbidden","message":"user does not have access to bucket"}`)
			},
			want: errors.New(codes.PermissionDenied, "user does not have access to bucket"),
		},
		{
			name: "unauthorized",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = io.WriteString(w, `{"code":"unauthorized","message":"credentials required"}`)
			},
			want: errors.New(codes.Unauthenticated, "credentials required"),
		},
		{
			name: "nested influxdb error",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"code":"invalid","message":"query was invalid","error":{"code":"not found","message":"resource not found"}}`)
			},
			want: errors.Wrap(
				errors.New(codes.NotFound, "resource not found"),
				codes.Invalid,
				"query was invalid",
			),
		},
		{
			name: "nested internal error",
			fn: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"code":"invalid","message":"query was invalid","error":"internal error"}`)
			},
			want: errors.Wrap(
				errors.New(codes.Unknown, "internal error"),
				codes.Invalid,
				"query was invalid",
			),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.fn(w)
			}))
			defer server.Close()

			want := Want{
				Err: tt.want,
			}

			if ps, ok := spec.(influxdb.ProcedureSpec); ok {
				ps.SetHost(&server.URL)
			}

			provider := influxdeps.Dependency{
				Provider: influxdeps.HttpProvider{
					DefaultConfig: influxdeps.Config{
						Host: server.URL,
					},
				},
			}

			deps := flux.NewDefaultDependencies()
			ctx := deps.Inject(context.Background())
			ctx = provider.Inject(ctx)
			ExecuteSourceTestHelper(t, ctx, spec, want)
		})
	}
}

func RunSourceURLValidatorTestHelper(t *testing.T, spec plan.PhysicalProcedureSpec) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("received unexpected request")
	}))
	defer server.Close()

	want := Want{
		Err: &flux.Error{
			Msg: "failed to initialize execute state",
			Err: &flux.Error{
				Code: codes.Invalid,
				Msg:  "url is not valid, it connects to a private IP",
			},
		},
	}

	if ps, ok := spec.(influxdb.ProcedureSpec); ok {
		ps.SetHost(&server.URL)
	}

	provider := influxdeps.Dependency{
		Provider: influxdeps.HttpProvider{
			DefaultConfig: influxdeps.Config{
				Host: server.URL,
			},
		},
	}

	deps := flux.NewDefaultDependencies()
	deps.Deps.URLValidator = urldeps.PrivateIPValidator{}
	ctx := deps.Inject(context.Background())
	ctx = provider.Inject(ctx)
	ExecuteSourceTestHelper(t, ctx, spec, want)
}

func RunSourceHTTPClientTestHelper(t *testing.T, spec plan.PhysicalProcedureSpec) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(""))
	}))
	defer server.Close()

	want := Want{
		Tables: func() []*executetest.Table {
			return nil
		},
	}

	if ps, ok := spec.(influxdb.ProcedureSpec); ok {
		ps.SetHost(&server.URL)
	}

	provider := influxdeps.Dependency{
		Provider: influxdeps.HttpProvider{
			DefaultConfig: influxdeps.Config{
				Host: server.URL,
			},
		},
	}

	counter := &requestCounter{}
	deps := flux.NewDefaultDependencies()
	deps.Deps.HTTPClient = &http.Client{
		Transport: counter,
	}
	ctx := deps.Inject(context.Background())
	ctx = provider.Inject(ctx)
	ExecuteSourceTestHelper(t, ctx, spec, want)

	if counter.Count == 0 {
		t.Error("custom http client was not used")
	}
}

func ExecuteSourceTestHelper(t *testing.T, ctx context.Context, spec plan.PhysicalProcedureSpec, want Want) {
	t.Helper()

	logger := zaptest.NewLogger(t)

	ps := plantest.CreatePlanSpec(&plantest.PlanSpec{
		Nodes: []plan.Node{
			plan.CreatePhysicalNode(plan.NodeID(spec.Kind()), spec),
			plan.CreatePhysicalNode("yield", &universe.YieldProcedureSpec{
				Name: "_result",
			}),
		},
		Edges: [][2]int{
			{0, 1},
		},
		Resources: flux.ResourceManagement{
			ConcurrencyQuota: 1,
			MemoryBytesQuota: math.MaxInt64,
		},
	})

	mem := &memory.Allocator{}
	executor := execute.NewExecutor(logger)
	results, _, err := executor.Execute(ctx, ps, mem)
	if err != nil {
		if diff := cmp.Diff(want.Err, err); diff != "" {
			t.Errorf("unexpected error -want/+got:\n%s", diff)
		}
		return
	}

	if len(results) != 1 {
		t.Fatalf("expected exactly one result, got %d", len(results))
	}

	var res flux.Result
	for _, r := range results {
		res = r
	}

	if res == nil {
		t.Fatal("expected non-null result")
	}

	tables := want.Tables
	if want.Err != nil {
		tables = func() []*executetest.Table {
			return []*executetest.Table{{Err: want.Err}}
		}
	}

	var wantT table.Iterator
	for _, tbl := range tables() {
		wantT = append(wantT, tbl)
	}
	gotT := res.Tables()

	if diff := table.Diff(wantT, gotT); diff != "" {
		t.Errorf("unexpected output -want/+got:\n%s", diff)
	}
}

type requestCounter struct {
	Count int
}

func (r *requestCounter) RoundTrip(req *http.Request) (*http.Response, error) {
	r.Count++
	return http.DefaultTransport.RoundTrip(req)
}
