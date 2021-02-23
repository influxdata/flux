package bigpanda_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
)

func TestBigPanda(t *testing.T) {
	s := NewServer(t)
	defer s.Close()
	ctx := dependenciestest.Default().Inject(context.Background())
	_, scope, err := runtime.Eval(ctx, `
import "csv"
import "contrib/rhajek/bigpanda"

option url = "`+s.URL+`"
option appKey = "myappkey1"
option token = "bigpandatoken1"

data = "
#datatype,string,string,string,string
#group,false,false,false,false
#default,_result,,,
,result,host,description,check
,,kozel.local,this is a lot of text yay,cpu_check
"

process = bigpanda.endpoint(url:url, appKey:appKey, token:token)( mapFn:
	(r) => {
		return {host:r.host, status:r.status, description:r.description, check:r.check}
	}
)

csv.from(csv:data) |> process()

`)

	if err != nil {
		t.Error(err)
	}
	_ = scope
}

type Request struct {
	URL      string
	PostData PostData
}

type PostData struct {
	AppKey      string `json:"app_key"`
	Status      string `json:"status"`
	Host        string `json:"host"`
	Description string `json:"description"`
	Check       string `json:"check"`
}

func TestBigPandaPost(t *testing.T) {

	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		name        string
		host        string
		level       string
		check       string
		description string
		status      string
	}{
		{
			name:        "simple.crit",
			host:        "kozel.local",
			level:       "crit",
			check:       "aCheck",
			description: "aDescription",
			status:      "critical",
		},
		{
			name:        "simple.info",
			host:        "kozel.local",
			level:       "info",
			check:       "aCheck",
			description: "aDescription",
			status:      "ok",
		},
		{
			name:        "simple.warn",
			host:        "kozel.local",
			level:       "warn",
			check:       "aCheck",
			description: "aDescription",
			status:      "warning",
		},
		{
			name:        "simple.unknown",
			host:        "kozel.local",
			level:       "unknown",
			check:       "aCheck",
			description: "aDescription",
			status:      "critical",
		},
		{
			name:        "simple.ok",
			host:        "kozel.local",
			level:       "ok",
			check:       "aCheck",
			description: "aDescription",
			status:      "ok",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxString := `
import "csv"
import "contrib/rhajek/bigpanda"

endpoint = bigpanda.endpoint(url: url, appKey: appKey, token: "token123")(mapFn: (r) => {
	return {r with status: bigpanda.statusFromLevel(level: r.level)}
})

csv.from(csv:data) |> endpoint() 
`

			extern := `
url = "` + s.URL + `"
appKey = "myappkey1"
data = "
#datatype,string,string,string,string,string,string
#group,false,false,false,false,false,false
#default,_result,,,,,
,result,,host,level,description,check
,,,` + strings.Join([]string{tc.host, tc.level, tc.description, tc.check}, ",") + `"`

			extHdl, err := runtime.Default.Parse(extern)
			if err != nil {
				t.Fatal(err)
			}
			prog, err := lang.Compile(fluxString, runtime.Default, time.Now(), lang.WithExtern(extHdl))
			if err != nil {
				t.Fatal(err)
			}
			ctx := flux.NewDefaultDependencies().Inject(context.Background())
			query, err := prog.Start(ctx, &memory.Allocator{})

			if err != nil {
				t.Fatal(err)
			}
			res := <-query.Results()
			_ = res

			var HasSent bool
			err = res.Tables().Do(func(table flux.Table) error {
				return table.Do(func(reader flux.ColReader) error {
					if reader == nil {
						return nil
					}
					for i, meta := range reader.Cols() {
						if meta.Label == "_sent" {
							HasSent = true
							if v := reader.Strings(i).Value(0); string(v) != "true" {
								t.Fatalf("expecting _sent=true but got _sent=%v", string(v))
							}
						}
					}
					return nil
				})
			})
			if !HasSent {
				t.Fatal("expected a _sent column but didn't get one")
			}
			if err != nil {
				t.Fatal(err)
			}

			query.Done()
			if err := query.Err(); err != nil {
				t.Error(err)
			}
			reqs := s.Requests()

			if len(reqs) < 1 {
				t.Fatal("received no requests")
			}
			req := reqs[len(reqs)-1]

			if req.PostData.Host != tc.host {
				t.Errorf("got host: %s, expected %s", req.PostData.Host, tc.host)
			}
			if req.PostData.Status != tc.status {
				t.Errorf("got status: %s, expected %s", req.PostData.Status, tc.status)
			}
			if req.PostData.Check != tc.check {
				t.Errorf("got check: %s, expected %s", req.PostData.Check, tc.check)
			}
			if req.PostData.Description != tc.description {
				t.Errorf("got description: %s, expected %s", req.PostData.Description, tc.check)
			}
		})
	}

}

type Server struct {
	mu       sync.Mutex
	ts       *httptest.Server
	URL      string
	requests []Request
	closed   bool
}

func NewServer(t *testing.T) *Server {
	s := new(Server)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sr := Request{
			URL: r.URL.String(),
		}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&sr.PostData)
		if err != nil {
			t.Error(err)
		}
		s.mu.Lock()
		s.requests = append(s.requests, sr)
		s.mu.Unlock()
	}))
	s.ts = ts
	s.URL = ts.URL
	return s
}
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.requests
}
func (s *Server) Close() {
	if s.closed {
		return
	}
	s.closed = true
	s.ts.Close()
}
