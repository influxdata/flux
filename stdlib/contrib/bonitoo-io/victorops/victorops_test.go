package victorops_test

import (
	"context"
	"encoding/json"
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

func TestVictorOps(t *testing.T) {
	ctx := dependenciestest.Default().Inject(context.Background())
	_, _, err := runtime.Eval(ctx, `
import "csv"
import "contrib/bonitoo-io/victorops"

option url = "https://alert.victorops.com/integrations/generic/20131114/alert/apiKey/routingKey"

data = "
#group,false,false,false,false,false,false,false,false,false
#datatype,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,node,metric_type,resource,metric_name,event_id,description,severity
,,0,10.1.1.1,CPU,CPU-1,usage_idle,Alert-#1001,CPU-1 too busy,CRITICAL
"

process = victorops.endpoint(url: url)(mapFn: (r) => ({
    messageType: r.severity,
    entityID: r.event_id,
    entityDisplayName: "",
    stateMessage: r.description,
    timestamp: now()
}))

csv.from(csv:data) |> process()
`)

	if err != nil {
		t.Error(err)
	}
}

func TestVictorOpsPost(t *testing.T) {
	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		name        string
		URL         string
		extraParams string
		alert       Alert
		timestamp   string
	}{
		{
			name: "alert with defaults",
			URL:  s.URL,
			alert: Alert{
				MessageType:    "WARNING",
				EntityID:       "Alert-#2000",
				Message:        "CPU-2 too busy",
				MonitoringTool: "InfluxDB", // default value in endpoint()
				Timestamp:      1609459200,
			},
			timestamp: "2021-01-01T00:00:00Z",
		},
		{
			name:        "alert with all fields",
			URL:         s.URL,
			extraParams: `, monitoringTool: "InfluxDB NextGen"`,
			alert: Alert{
				MessageType:       "CRITICAL",
				EntityID:          "Alert-#2001",
				EntityDisplayName: "alert #2001",
				Message:           "CPU-2 too busy",
				MonitoringTool:    "InfluxDB NextGen",
				Timestamp:         1609459200,
			},
			timestamp: "2021-01-01T00:00:00Z",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s.Reset()

			fluxString := `import "csv"
import "contrib/bonitoo-io/victorops"

url = "` + tc.URL + `"

data = "
#group,false,false,false,false,false,false,false,false,false
#datatype,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,
,result,table,node,metric_type,resource,metric_name,event_id,description,severity
,,0,10.1.1.1,CPU,CPU-1,usage_idle,Alert-#1001,CPU-1 too busy,WARNING
,,0,` + strings.Join([]string{"node", "CPU", "CPU-2", "usage_user", tc.alert.EntityID, tc.alert.Message, tc.alert.MessageType}, ",") + `
"

endpoint = victorops.endpoint(url: url` + tc.extraParams + `)(mapFn: (r) => ({
    messageType: r.severity,
    entityID: r.event_id,
    entityDisplayName: "` + tc.alert.EntityDisplayName + `",
    stateMessage: r.description,
    timestamp: ` + tc.timestamp + `
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
			URL: r.URL.String(),
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
	URL   string
	Alert Alert
}

type Alert struct {
	MessageType       string `json:"message_type"`
	EntityID          string `json:"entity_id"`
	EntityDisplayName string `json:"entity_display_name"`
	Message           string `json:"state_message"`
	Timestamp         uint   `json:"state_start_time"`
	MonitoringTool    string `json:"monitoring_tool"`
}
