package pagerduty_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/flux/dependencies"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/lang"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/values"

	_ "github.com/influxdata/flux/builtin"
)

func TestPagerduty(t *testing.T) {
	ctx, deps := context.Background(), dependenciestest.Default()
	_, _, err := flux.Eval(ctx, deps, `
import "csv"
import "pagerduty"
data = "
#datatype,string,string,string,string,string,string,string,string,string,string,string,string
#group,false,false,false,false,false,false,false,false,false,false,false,false
#default,_result,,,,,,,,,,,
,result,_routingKey,_client,_clientURL,_class,_group,_severity,_source,_summary,_timestamp
,,fakeRoutingKey,fakeClient,fakeClientURL,fakeClass,fakeGroup,fakeSeverity,fakeSource,fakeSummary,fakeTimestamp
"
process = pagerduty.endpoint(url:url, token:token)( mapFn: 
	(r) => {
		return {routingKey:r._routingKey,client:r._client,clientURL:r._clientURL,class:r._class,group:r._group,eventAction:r._eventAction,severity:r._severity,source:r._source,summary:r._summary,timestamp:r._timestamp}
	}
)
csv.from(csv:data) |> process()
`, func(s values.Scope) {
		s.Set("url", values.New("http://fakeurl.com/fakeyfake"))
		s.Set("token", values.New("faketoken"))
	})

	if err != nil {
		t.Error(err)
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
			URL:           r.URL.String(),
			Authorization: r.Header.Get("Authorization"),
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

type Request struct {
	URL           string
	Authorization string
	PostData      PostData
}

type Payload struct {
	Summary   string `json:"summary"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	Source    string `json:"source"`
	Class     string `json:"class"`
	Group     string `json:"group"`
	Component string `json:"component"`
}

type PostData struct {
	Token       string  `json:"token"`
	RoutingKey  string  `json:"routing_key"`
	Client      string  `json:"client"`
	ClientURL   string  `json:"client_url"`
	DedupKey    string  `json:"dedup_key"`
	EventAction string  `json:"event_action"`
	Payload     Payload `json:"payload"`
}

func TestPagerdutySendEvent(t *testing.T) {
	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		name          string
		otherGroupKey string
		pagerdutyURL  string
		token         string
		routingKey    string
		client        string
		clientURL     string
		class         string
		group         string
		severity      string
		component     string
		source        string
		summary       string
		timestamp     string
		eventAction   string
	}{
		{
			name:          "testCase1",
			otherGroupKey: "foo",
			pagerdutyURL:  s.URL,
			token:         "fakeToken1",
			routingKey:    "fakeRoutingKey",
			client:        "fakeClient1",
			clientURL:     "http://fakepagerduty.com",
			class:         "deploy",
			group:         "app-stack",
			severity:      "warning",
			component:     "influxdb",
			source:        "monitoringtool:vendor:region",
			summary:       "this is a testing summary",
			timestamp:     "2015-07-17T08:42:58.315+0000",
			eventAction:   "trigger",
		},
		{
			name:          "testCase2",
			otherGroupKey: "foo2",
			pagerdutyURL:  s.URL,
			token:         "fakeToken2",
			routingKey:    "fakeRoutingKey",
			client:        "fakeClient2",
			clientURL:     "http://fakepagerduty.com",
			class:         "deploy",
			group:         "app-stack",
			severity:      "ok",
			component:     "postgres",
			source:        "monitoringtool:vendor:region",
			summary:       "this is another testing summary",
			timestamp:     "2016-07-17T08:42:58.315+0000",
			eventAction:   "resolve",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxString := `import "csv"
import "pagerduty"

endpoint = pagerduty.endpoint(url:url, token:token)(mapFn: (r) => {
    return {
		routingKey:r.froutingKey,
		client:r.qclient,
		clientURL:r.qclientURL,
		class:r.wclass,
		group:r.wgroup,
		severity:r.wseverity,
		component:r.wcomponent,
		source:r.wsource,
		summary:r.wsummary,
		timestamp:r.wtimestamp,
	}
})

csv.from(csv:data) |> endpoint()
`

			prog, err := lang.Compile(fluxString, time.Now(), lang.WithExtern(&ast.File{Body: []ast.Statement{
				&ast.VariableAssignment{
					ID: &ast.Identifier{
						Name: "url",
					},
					Init: &ast.StringLiteral{
						Value: tc.pagerdutyURL,
					},
				},
				&ast.VariableAssignment{
					ID: &ast.Identifier{
						Name: "token",
					},
					Init: &ast.StringLiteral{
						Value: tc.token,
					},
				},
				&ast.VariableAssignment{
					ID: &ast.Identifier{
						Name: "data",
					},
					Init: &ast.StringLiteral{
						Value: `#datatype,string,string,string,string,string,string,string,string,string,string,string,string,string,string,long
#group,false,false,false,true,false,false,false,false,false,false,false,false,true,true,true
#default,_result,,,,,,,,,,,,,,
,result,,froutingKey,qclient,qclientURL,wclass,wgroup,wseverity,wcomponent,wsource,wsummary,wtimestamp,name,otherGroupKey,groupKey2
,,,` + strings.Join([]string{
							tc.routingKey,
							tc.client,
							tc.clientURL,
							tc.class,
							tc.group,
							tc.severity,
							tc.component,
							tc.source,
							tc.summary,
							tc.timestamp,
							tc.name,
							tc.otherGroupKey,
							"0"}, ","),
					},
				},
			}}))

			if err != nil {
				t.Error(err)
			}
			prog.SetExecutorDependencies(execute.Dependencies{dependencies.InterpreterDepsKey: dependencies.NewDefaults()})
			query, err := prog.Start(context.Background(), &memory.Allocator{})
			if err != nil {
				t.Fatal(err)
			}

			res := <-query.Results()
			defer func() {
				query.Done()
				if err := query.Err(); err != nil {
					t.Fatal("query error: ", err)
				}
			}()

			var Sent bool
			err = res.Tables().Do(func(table flux.Table) error {

				if table.Empty() {
					t.Errorf("results table is empty")
				}
				return table.Do(func(reader flux.ColReader) error {
					if reader == nil {
						return nil
					}
					for i, meta := range reader.Cols() {
						if meta.Label == "_sent" {
							Sent = true
							if reader.Strings(i).ValueString(0) != "true" {
								t.Errorf("expected _sent to be true but got %s", reader.Strings(i).ValueString(0))
							}
						}
					}
					return nil
				})
			})

			if err != nil {
				t.Error(err)
			}
			if !Sent {
				t.Error("expected a _sent column but didnt get one")
			}
			reqs := s.Requests()
			if len(reqs) < 1 {
				t.Fatal("received no requests")
			}
			req := reqs[len(reqs)-1]
			if req.Authorization != "Token token="+tc.token {
				t.Errorf("token incorrect got %s, expected %s", req.Authorization, "Token "+tc.token)
			}

			if req.PostData.Client != tc.client {
				t.Errorf("got client %s, expected %s", req.PostData.Client, tc.client)
			}

			if req.PostData.EventAction != tc.eventAction {
				t.Errorf("got event action %s, expected %s", req.PostData.EventAction, tc.eventAction)
			}

			if req.PostData.ClientURL != tc.clientURL {
				t.Errorf("got client URL %s, expected %s", req.PostData.ClientURL, tc.clientURL)
			}

			if req.PostData.Payload.Summary != tc.summary {
				t.Errorf("got summary %s, expected %s", req.PostData.Payload.Summary, tc.summary)
			}

			if req.PostData.Payload.Timestamp != tc.timestamp {
				t.Errorf("got timestamp %s, expected %s", req.PostData.Payload.Timestamp, tc.timestamp)
			}

			if req.PostData.Payload.Group != tc.group {
				t.Errorf("got group %s, expected %s", req.PostData.Payload.Group, tc.group)
			}

			if req.PostData.Payload.Class != tc.class {
				t.Errorf("got class %s, expected %s", req.PostData.Payload.Class, tc.class)
			}

			if req.PostData.Payload.Source != tc.source {
				t.Errorf("got source %s, expected %s", req.PostData.Payload.Source, tc.source)
			}

			if req.PostData.Payload.Severity != tc.severity {
				t.Errorf("got severity %s, expected %s", req.PostData.Payload.Severity, tc.severity)
			}

			if req.PostData.Payload.Component != tc.component {
				t.Errorf("got component %s, expected %s", req.PostData.Payload.Component, tc.component)
			}

		})
	}
}
