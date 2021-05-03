package servicenow_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
)

func TestServiceNow(t *testing.T) {
	ctx := dependenciestest.Default().Inject(context.Background())
	_, _, err := runtime.Eval(ctx, `
import "csv"
import "contrib/bonitoo-io/servicenow"

option url = "https://sandbox.service-now.com/api/global/em/jsonv2"
option username = "admin"
option password = "12345"

data = "
#group,false,false,false,false,false,false,false,false,false
#datatype,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,node,metric_type,resource,metric_name,message_key,description,severity
,,0,10.1.1.1,CPU,CPU-1,usage_idle,Alert-#1001,CPU-1 too busy,critical
"

process = servicenow.endpoint(url: url, username: username, password: password)(mapFn: (r) => ({
    node: r.node,
    metricType: r.metric_type,
    resource: r.resource,
    metricName: r.metric_name,
    messageKey: r.message_key,
    description: r.description,
    severity: r.severity,
    additionalInfo: {}
}))

csv.from(csv:data) |> process()
`)

	if err != nil {
		t.Error(err)
	}
}

func TestServiceNowPost(t *testing.T) {
	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		name           string
		URL            string
		event          Event
		additionalInfo string
	}{
		{
			name: "alert with defaults",
			URL:  s.URL,
			event: Event{
				Source:         "Flux",
				Description:    "some alert",
				Severity:       "minor",
				AdditionalInfo: "",
			},
			additionalInfo: "{}",
		},
		{
			name: "alert with all fields",
			URL:  s.URL,
			event: Event{
				Source:         "FluxCustom",
				Node:           "10.1.2.3",
				MetricType:     "CPU",
				Resource:       "CPU-2",
				MetricName:     "usage_user",
				MessageKey:     "Alert-#10001",
				Description:    "CPU-2 is rather busy",
				Severity:       "warning",
				AdditionalInfo: `{"metric-name":"usage_user","tid":13}`,
			},
			additionalInfo: `{ "metric-name": r.metric_name, "tid": 13 }`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s.Reset()

			fluxString := `import "csv"
import "contrib/bonitoo-io/servicenow"

url = "` + tc.URL + `"
username = "admin"
password = "12345"
source = "` + tc.event.Source + `"

data = "
#group,false,false,false,false,false,false,false,false,false
#datatype,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,node,metric_type,resource,metric_name,message_key,description,severity
,,0,10.1.1.1,CPU,CPU-1,usage_idle,Alert-#1001,CPU-1 too busy,major
,,0,` + strings.Join([]string{tc.event.Node, tc.event.MetricType, tc.event.Resource, tc.event.MetricName, tc.event.MessageKey, tc.event.Description, tc.event.Severity}, ",") + `
"

endpoint = servicenow.endpoint(url: url, username: username, password: password, source: source)(mapFn: (r) => ({
    node: r.node,
    metricType: r.metric_type,
    resource: r.resource,
    metricName: r.metric_name,
    messageKey: r.message_key,
    description: r.description,
    severity: r.severity,
    additionalInfo: ` + tc.additionalInfo + `
}))

csv.from(csv:data) |> endpoint()`

			prog, err := lang.Compile(fluxString, runtime.Default, time.Now())
			if err != nil {
				t.Fatal(err)
			}

			ctx := flux.NewDefaultDependencies().Inject(context.Background())
			query, err := prog.Start(ctx, &memory.Allocator{})
			if err != nil {
				t.Fatal(err)
			}

			var res flux.Result
			timer := time.NewTimer(1 * time.Second)
			select {
			case res = <-query.Results():
				timer.Stop()
			case <-timer.C:
				t.Fatal("query timeout")
			}

			var hasSent bool
			err = res.Tables().Do(func(table flux.Table) error {
				return table.Do(func(reader flux.ColReader) error {
					for i, meta := range reader.Cols() {
						if meta.Label == "_sent" {
							hasSent = true
							if v := reader.Strings(i).Value(0); string(v) != "true" {
								t.Fatalf("expecting _sent=true but got _sent=%v", string(v))
							}
							break
						}
					}
					return nil
				})
			})

			if err != nil {
				t.Fatal(err)
			}

			if !hasSent {
				t.Fatal("expected _sent column but didn't get one")
			}

			query.Done()
			if err := query.Err(); err != nil {
				t.Error(err)
			}

			reqs := s.Requests()
			if len(reqs) != 2 {
				t.Fatalf("expected 2 requests, received %d", len(reqs))
			}
			req := reqs[len(reqs)-1]
			if err = req.Events.Records[0].decodeSeverityToSource(); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.event, req.Events.Records[0]); diff != "" {
				t.Fatal(diff)
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
			URL:           r.URL.String(),
			Authorization: r.Header.Get("Authorization"),
		}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&sr.Events)
		if err != nil {
			t.Error(err)
		}
		s.mu.Lock()
		s.requests = append(s.requests, sr)
		s.mu.Unlock()
		w.WriteHeader(http.StatusOK)
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

func (s *Server) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests = []Request{}
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
	Events        Events
}

type Event struct {
	Source         string `json:"source"`
	Node           string `json:"node"`
	MetricType     string `json:"type"`
	Resource       string `json:"resource"`
	MetricName     string `json:"metric_name"`
	MessageKey     string `json:"message_key"`
	Description    string `json:"description"`
	Severity       string `json:"severity"`
	AdditionalInfo string `json:"additional_info"`
}

type Events struct {
	Records []Event `json:"records"`
}

func (a *Event) decodeSeverityToSource() error {
	switch a.Severity {
	case "0":
		a.Severity = "clear"
	case "1":
		a.Severity = "critical"
	case "2":
		a.Severity = "major"
	case "3":
		a.Severity = "minor"
	case "4":
		a.Severity = "warning"
	case "5":
		a.Severity = "info"
	default:
		return fmt.Errorf("unsupported severity: %s", a.Severity)
	}

	return nil
}
