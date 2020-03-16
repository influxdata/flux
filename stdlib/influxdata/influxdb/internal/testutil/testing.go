package testutil

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/csv"
	urldeps "github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
)

type (
	T = testing.T
	B = testing.B
)

type Want struct {
	Params url.Values
	Ast    *ast.Package
	Tables func() []*executetest.Table
}

func RunSourceTestHelper(t *testing.T, spec SourceProcedureSpec, want Want) {
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
			AST     *ast.Package `json:"ast"`
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

		if want, got := want.Ast, req.AST; !cmp.Equal(want, got) {
			t.Errorf("unexpected ast in request body -want/+got:\n%s", cmp.Diff(want, got))
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

	spec.SetHost(StringPtr(server.URL))

	deps := flux.NewDefaultDependencies()
	ctx := deps.Inject(context.Background())
	store := executetest.NewDataStore()
	s, err := CreateSource(ctx, spec)
	if err != nil {
		t.Fatal(err)
	}
	s.AddTransformation(store)
	s.Run(context.Background())

	if err := store.Err(); err != nil {
		t.Fatal(err)
	}

	got, err := executetest.TablesFromCache(store)
	if err != nil {
		t.Fatal(err)
	}
	executetest.NormalizeTables(got)

	tables := want.Tables()
	executetest.NormalizeTables(tables)

	if !cmp.Equal(tables, got) {
		t.Errorf("unexpected tables returned from server -want/+got:\n%s", cmp.Diff(tables, got))
	}
}

func RunSourceErrorTestHelper(t *testing.T, spec SourceProcedureSpec) {
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

			spec.SetHost(StringPtr(server.URL))

			deps := flux.NewDefaultDependencies()
			ctx := deps.Inject(context.Background())
			store := executetest.NewDataStore()
			s, err := CreateSource(ctx, spec)
			if err != nil {
				t.Fatal(err)
			}
			s.AddTransformation(store)
			s.Run(context.Background())

			got := store.Err()
			if got == nil {
				t.Fatal("expected error")
			}
			want := tt.want

			if !cmp.Equal(want, got) {
				t.Errorf("unexpected error:\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func RunSourceURLValidatorTestHelper(t *testing.T, spec SourceProcedureSpec) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("received unexpected request")
	}))
	defer server.Close()

	spec.SetHost(StringPtr(server.URL))

	deps := flux.NewDefaultDependencies()
	deps.Deps.URLValidator = urldeps.PrivateIPValidator{}
	ctx := deps.Inject(context.Background())
	if _, err := CreateSource(ctx, spec); err == nil {
		t.Fatal("expected error")
	}
}

func RunSourceHTTPClientTestHelper(t *testing.T, spec SourceProcedureSpec) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(""))
	}))
	defer server.Close()

	spec.SetHost(StringPtr(server.URL))

	counter := &requestCounter{}
	deps := flux.NewDefaultDependencies()
	deps.Deps.HTTPClient = &http.Client{
		Transport: counter,
	}
	ctx := deps.Inject(context.Background())
	store := executetest.NewDataStore()
	s, err := CreateSource(ctx, spec)
	if err != nil {
		t.Fatal(err)
	}
	s.AddTransformation(store)
	s.Run(context.Background())

	if err := store.Err(); err != nil {
		t.Fatal(err)
	}

	if counter.Count == 0 {
		t.Error("custom http client was not used")
	}
}

func StringPtr(v string) *string {
	return &v
}

func MustParseTime(v string) time.Time {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		panic(err)
	}
	return t
}

type requestCounter struct {
	Count int
}

func (r *requestCounter) RoundTrip(req *http.Request) (*http.Response, error) {
	r.Count++
	return http.DefaultTransport.RoundTrip(req)
}

func CreateSource(ctx context.Context, ps SourceProcedureSpec) (execute.Source, error) {
	id := executetest.RandomDatasetID()
	return influxdb.CreateSource(id, ps, mock.AdministrationWithContext(ctx))
}

type SourceProcedureSpec interface {
	influxdb.ProcedureSpec
	BuildQuery() *ast.File
}
