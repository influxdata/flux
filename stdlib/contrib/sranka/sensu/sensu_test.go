package sensu_test

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
	_ "github.com/influxdata/flux/builtin"
	_ "github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
)

func TestSensu(t *testing.T) {
	s := NewServer(t)
	defer s.Close()
	ctx := dependenciestest.Default().Inject(context.Background())
	_, _, err := runtime.Eval(ctx, `
import "csv"
import "contrib/sranka/sensu"

url = "`+s.URL+`"
apiKey = "fakeApiKey"
checkName = "compilationCheckOnly"
text = "abc"
status = sensu.event(url:url,apiKey:apiKey,checkName:checkName,text:text)`)

	if err != nil {
		t.Error(err)
	}
}

func TestSensuEndpoint(t *testing.T) {
	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		checkName  string // also name of the test
		text       string
		handlers   []string
		status     int
		state      string
		namespace  string
		entityName string

		endpointExtraArgs string
		sensuEntityName   string
		sensuCheckName    string
	}{
		{
			checkName: "simple",
			text:      "abc",
		},
		{
			checkName: "okStatus",
			text:      "abc",
			status:    0,
			state:     "passing",
		},
		{
			checkName: "failedStatus",
			text:      "abc",
			status:    1,
			state:     "failing",
		},
		{
			checkName:         "withHandlers",
			text:              "bcd",
			handlers:          []string{"myHandler"},
			endpointExtraArgs: `handlers: ["myHandler"],`,
		},
		{
			checkName:         "withNamespace",
			text:              "wns",
			namespace:         "myNs",
			endpointExtraArgs: `namespace: "myNs",`,
		},
		{
			checkName:         "withEntityName",
			text:              "wns",
			entityName:        "myEntity",
			endpointExtraArgs: `entityName: "myEntity",`,
		},
		{
			checkName:         "converted entity name",
			text:              "stu",
			entityName:        "spaces to underscore:",
			endpointExtraArgs: `entityName: "spaces to underscore:",`,
			sensuCheckName:    "converted_entity_name",
			sensuEntityName:   "spaces_to_underscore_",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.checkName, func(t *testing.T) {
			s.Reset()
			fluxString := `import "csv"
import "contrib/sranka/sensu"

url = "` + s.URL + `"
apiKey = "fakeKey"
handlers = ["` + strings.Join(tc.handlers, `","`) + `"]
namespace = "` + tc.namespace + `"
entityName = "` + tc.entityName + `"

endpoint = sensu.endpoint(` + tc.endpointExtraArgs + `url:url,apiKey:apiKey)(mapFn: (r) => {
	return {checkName:r.qcheck,text:r.qtext, status:r.qstatus}
 })
 
 data = "
#datatype,string,string,string,string,long
#group,false,false,false,false,false
#default,_result,,,,
,result,,qcheck,qtext,qstatus
,,,` + strings.Join([]string{tc.checkName, tc.text, strconv.Itoa(tc.status)}, ",") + `"

csv.from(csv:data) |> endpoint()
 `
			prog, err := lang.Compile(fluxString, runtime.Default, time.Now())
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

			namespace := tc.namespace
			if namespace == "" {
				namespace = "default"
			}
			if req.URL != "/api/core/v2/namespaces/"+namespace+"/events" {
				t.Errorf("got URL: %s, expected %s", req.URL, "/api/core/v2/namespaces/"+namespace+"/events")
			}
			if req.PostData.Entity.EntityClass != "proxy" {
				t.Errorf("got entity_class: %s, expected %s", req.PostData.Entity.EntityClass, "proxy")
			}
			entityName := tc.entityName
			if entityName == "" {
				entityName = "influxdb"
			}
			if tc.sensuEntityName != "" {
				entityName = tc.sensuEntityName
			}
			if req.PostData.Entity.Metadata.Name != entityName {
				t.Errorf("got entityName: %s, expected %s", req.PostData.Entity.Metadata.Name, entityName)
			}
			if req.PostData.Check.Output != tc.text {
				t.Errorf("got text: %s, expected %s", req.PostData.Check.Output, tc.text)
			}
			if req.PostData.Check.Status != tc.status {
				t.Errorf("got status: %d, expected %d", req.PostData.Check.Status, tc.status)
			}
			state := tc.state
			if state == "" {
				if tc.status == 0 {
					state = "passing"
				} else {
					state = "failing"
				}
			}
			if req.PostData.Check.State != state {
				t.Errorf("got state: %s, expected %s", req.PostData.Check.State, state)
			}
			handlers := tc.handlers
			if len(handlers) == 0 {
				handlers = []string{}
			}
			if !cmp.Equal(handlers, req.PostData.Check.Handlers) {
				t.Fatalf("unexpected handlers -want/+got\n\n%s\n\n", cmp.Diff(handlers, req.PostData.Check.Handlers))
			}
			checkName := tc.checkName
			if tc.sensuCheckName != "" {
				checkName = tc.sensuCheckName
			}
			if req.PostData.Check.Metadata.Name != checkName {
				t.Errorf("got checkName: %s, expected %s", req.PostData.Check.Metadata.Name, checkName)
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
			URL: r.URL.String(), // r.URL.String(),
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
	URL      string
	PostData PostData
}

type PostData struct {
	Entity struct {
		EntityClass string `json:"entity_class"`
		Metadata    struct {
			Name string `json:"name"`
		} `json:"metadata"`
	} `json:"entity"`
	Check struct {
		Output   string   `json:"output"`
		State    string   `json:"state"`
		Status   int      `json:"status"`
		Handlers []string `json:"handlers"`
		Interval int      `json:"interval"`
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
	} `json:"check"`
}
