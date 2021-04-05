package alerta_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestAlerta(t *testing.T) {
	ctx := dependenciestest.Default().Inject(context.Background())
	_, _, err := runtime.Eval(ctx, `
import "csv"
import "contrib/bonitoo-io/alerta"

option url = "https://alerta.io:8080/alert"
option apiKey = "some key"

data = "
#group,false,false,false,false,false,false,false,false,false
#datatype,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,node,metric_type,resource,metric_name,alert_id,description,severity
,,0,10.1.1.1,CPU,CPU-1,usage_idle,Alert-#1001,CPU-1 too busy,major
"

process = alerta.endpoint(url: url, apiKey: apiKey)(mapFn: (r) => ({
    resource: r.resource,
    event: r.alert_id,
    severity: r.severity,
    service: [r.node],
    group: "",
    value: r.description,
    text: "",
    tags: ["dc1"],
    attributes: {metric_name:r.metric_name},
    origin: "InfluxDB",
    type: "external",
    timestamp: now()
}))

csv.from(csv:data) |> process()
`)

	if err != nil {
		t.Error(err)
	}
}

func TestAlertaPost(t *testing.T) {
	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		name   string
		URL    string
		env    string
		origin string
		alert  Alert
		fn     string
		extras bool
	}{
		{
			name: "alert with defaults",
			URL:  s.URL,
			alert: Alert{
				Resource: "CPU-1",
				Event:    "Alert-#1001",
				Severity: "major",
				Service:  []string{},
				Tags:     []string{},
				Attributes: map[string]interface{}{
					"metric": "usage_user",
				},
				Type:      "external",
				Timestamp: "2021-04-01T01:02:03.456Z", // precision is cut for Alerta to 3 decimal digits
			},
			fn: "alerta.endpoint(url: url, apiKey: apiKey)",
		},
		{
			name:   "alert with all fields",
			URL:    s.URL,
			env:    "Production",
			origin: "Telegraf",
			alert: Alert{
				Resource: "CPU-2",
				Event:    "Alert-#1002",
				Severity: "minor",
				Service:  []string{"10.1.1.1"},
				Tags:     []string{"dc1"},
				Attributes: map[string]interface{}{
					"metric": "usage_user",
				},
				Value:     "CPU-2 too busy",
				Type:      "external",
				Origin:    "Telegraf",
				Timestamp: "2021-04-01T01:02:03.456Z", // precision is cut for Alerta to 3 decimal digits
			},
			fn:     "alerta.endpoint(url: url, apiKey: apiKey, environment: environment, origin: origin)",
			extras: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s.Reset()

			alert := tc.alert
			fluxString := `import "csv"
import "contrib/bonitoo-io/alerta"

url = "` + tc.URL + `"
apiKey = "some key"
environment = "` + tc.env + `"
origin = "` + tc.origin + `"
extras = ` + strconv.FormatBool(tc.extras) + `

data = "
#group,false,false,false,false,false,false,false,false,false
#datatype,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,node,metric_type,resource,metric_name,alert_id,description,severity
,,0,10.1.1.1,CPU,CPU-1,usage_idle,Alert-#1001,CPU-1 too busy,major
,,0,10.1.1.1,CPU,` + strings.Join([]string{alert.Resource, "usage_user", alert.Event, alert.Value, alert.Severity}, ",") + `
"

endpoint = ` + tc.fn + `(mapFn: (r) => ({
    resource: r.resource,
    event: r.alert_id,
    severity: r.severity,
    service: if extras then [r.node] else [],
    group: "",
    value: r.description,
    text: "",
    tags: if extras then ["dc1"] else [],
    attributes: {metric:r.metric_name},
    origin: "InfluxDB",
    type: "external",
    timestamp: 2021-04-01T01:02:03.456789000Z
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
			if diff := cmp.Diff(tc.alert, req.Alert); diff != "" {
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
		err := dec.Decode(&sr.Alert)
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
	Alert         Alert
}

type Alert struct {
	Resource   string                 `json:"resource"`
	Event      string                 `json:"event"`
	Severity   string                 `json:"severity"`
	Service    []string               `json:"service"`
	Group      string                 `json:"group"`
	Value      string                 `json:"value"`
	Tags       []string               `json:"tags"`
	Attributes map[string]interface{} `json:"attributes"`
	Origin     string                 `json:"origin"`
	Type       string                 `json:"type"`
	Timestamp  string                 `json:"createTime"`
}
