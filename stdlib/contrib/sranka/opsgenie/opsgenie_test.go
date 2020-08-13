package opsgenie_test

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
	_ "github.com/influxdata/flux/builtin"
	_ "github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
)

func TestSendAlert(t *testing.T) {
	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		checkName   string
		message     string
		extraArgs   string
		alias       string
		description string
		responders  []map[string]string
		visibleTo   []map[string]string
		actions     []string
		tags        []string
		details     map[string]string
		entity      string
		priority    string
	}{
		{
			checkName: "simplest",
			message:   "my message",
		},
		{
			checkName: "withAlias",
			message:   "my message",
			alias:     "my alias",
			extraArgs: `, alias:"my alias"`,
		},
		{
			checkName: "messageTooLong",
			message:   strings.Repeat("0123456789", 20),
		},
		{
			checkName:   "allProps",
			message:     "msg",
			alias:       "mal",
			actions:     []string{"a1"},
			description: "md",
			details:     map[string]string{"mk": "mv"},
			entity:      "me",
			priority:    "P1",
			responders:  []map[string]string{{"name": "myt", "type": "team"}, {"username": "m@t", "type": "user"}},
			tags:        []string{"mt1"},
			visibleTo:   []map[string]string{{"name": "yt", "type": "team"}},
			extraArgs: `,alias:"mal",actions: ["a1"],description:"md",details:"{\"mk\":\"mv\"}",entity:"me",priority:"P1",tags:["mt1"],
										responders:["team:myt","user:m@t"],
										visibleTo:["team:yt"]`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.checkName, func(t *testing.T) {
			ctx := flux.NewDefaultDependencies().Inject(context.Background())
			fluxString := `
import "csv"
import "contrib/sranka/opsgenie"
url = "` + s.URL + `"
apiKey = "fakeApiKey"
message = "` + tc.message + `"
status = opsgenie.sendAlert(url:url,apiKey:apiKey,message:message` + tc.extraArgs + `)`
			// fmt.Println(fluxString)
			_, _, err := runtime.Eval(ctx, fluxString)
			if err != nil {
				t.Fatal(err)
			}
			req := s.Request()
			if req.URL != "/v2/alerts" {
				t.Errorf("got URL: %s, expected %s", req.URL, "/v2/alerts")
			}
			if req.PostData.Message != tc.message {
				if len(tc.message) < 130 || req.PostData.Message != tc.message[:130] {
					t.Errorf("got message: %s, expected %s", req.PostData.Message, tc.message)
				}
			}
			if tc.alias != "" && req.PostData.Alias != tc.alias {
				t.Errorf("got alias: %s, expected %s", req.PostData.Alias, tc.alias)
			}
			if tc.alias == "" && req.PostData.Alias != tc.message {
				t.Errorf("got alias: %s, expected %s", req.PostData.Alias, tc.message)
			}
			if req.PostData.Description != tc.description {
				t.Errorf("got description: %s, expected %s", req.PostData.Description, tc.description)
			}
			if len(req.PostData.Responders) == 0 && len(tc.responders) != 0 {
				t.Errorf("got responders: %v, expected %v", req.PostData.Responders, tc.responders)
			}
			if len(req.PostData.Responders) != 0 && !cmp.Equal(req.PostData.Responders, tc.responders) {
				t.Fatalf("unexpected responders -want/+got\n\n%s\n\n", cmp.Diff(req.PostData.Responders, tc.responders))
			}
			if len(req.PostData.VisibleTo) == 0 && len(tc.visibleTo) != 0 {
				t.Errorf("got visibleTo: %v, expected %v", req.PostData.VisibleTo, tc.visibleTo)
			}
			if len(req.PostData.VisibleTo) != 0 && !cmp.Equal(req.PostData.VisibleTo, tc.visibleTo) {
				t.Fatalf("unexpected visibleTo -want/+got\n\n%s\n\n", cmp.Diff(req.PostData.VisibleTo, tc.visibleTo))
			}
			if len(req.PostData.Actions) == 0 && len(tc.actions) != 0 {
				t.Errorf("got actions: %v, expected %v", req.PostData.Actions, tc.actions)
			}
			if len(req.PostData.Actions) != 0 && !cmp.Equal(req.PostData.Actions, tc.actions) {
				t.Fatalf("unexpected actions -want/+got\n\n%s\n\n", cmp.Diff(req.PostData.Actions, tc.actions))
			}
			if len(req.PostData.Tags) == 0 && len(tc.tags) != 0 {
				t.Errorf("got tags: %v, expected %v", req.PostData.Tags, tc.tags)
			}
			if len(req.PostData.Tags) != 0 && !cmp.Equal(req.PostData.Tags, tc.tags) {
				t.Fatalf("unexpected tags -want/+got\n\n%s\n\n", cmp.Diff(req.PostData.Tags, tc.tags))
			}
			if len(req.PostData.Details) == 0 && len(tc.details) != 0 {
				t.Errorf("got details: %v, expected %v", req.PostData.Details, tc.details)
			}
			if len(req.PostData.Details) != 0 && !cmp.Equal(req.PostData.Details, tc.details) {
				t.Fatalf("unexpected details -want/+got\n\n%s\n\n", cmp.Diff(req.PostData.Details, tc.details))
			}
			if req.PostData.Entity != tc.entity {
				t.Errorf("got entity: %s, expected %s", req.PostData.Entity, tc.entity)
			}
			if tc.priority != "" && req.PostData.Priority != tc.priority {
				t.Errorf("got priority: %s, expected %s", req.PostData.Priority, tc.priority)
			}
			if tc.priority == "" && req.PostData.Priority != "P3" {
				t.Errorf("got priority: %s, expected %s", req.PostData.Priority, "P3")
			}
		})
	}
}

func TestEndpoint(t *testing.T) {
	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		checkName   string
		message     string
		extraArgs   string
		alias       string
		description string
		responders  []map[string]string
		visibleTo   []map[string]string
		actions     []string
		tags        []string
		details     map[string]string
		entity      string
		priority    string
	}{
		{
			checkName:   "allProps",
			message:     "msg",
			alias:       "mal",
			actions:     []string{"a1"},
			description: "md",
			details:     map[string]string{"mk": "mv"},
			entity:      "me",
			priority:    "P1",
			responders:  []map[string]string{{"name": "myt", "type": "team"}, {"username": "m@t", "type": "user"}},
			tags:        []string{"mt1"},
			visibleTo:   []map[string]string{{"name": "yt", "type": "team"}},
			extraArgs: `,description:"md",priority:"P1",actions: ["a1"],details:"{\"mk\":\"mv\"}",tags:["mt1"],
										responders:["team:myt","user:m@t"],
										visibleTo:["team:yt"]`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.checkName, func(t *testing.T) {
			fluxString := `import "csv"
import "contrib/sranka/opsgenie"

endpoint = opsgenie.endpoint(url:url, apiKey:apiKey, entity: entity)(mapFn: (r) => {
 return {message:r.qmessage,alias:r.qalias` + tc.extraArgs + `}
})

csv.from(csv:data) |> endpoint() `
			extern := `
url = "` + s.URL + `"
apiKey = "` + tc.checkName + `"
entity = "` + tc.entity + `"
data = "
#datatype,string,string,string,string
#group,false,false,false,false
#default,_result,,,
,result,,qmessage,qalias
,,,` + strings.Join([]string{tc.message, tc.alias}, ",") + `"`

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
			req := s.Request()
			if req.URL != "/v2/alerts" {
				t.Errorf("got URL: %s, expected %s", req.URL, "/v2/alerts")
			}
			if req.PostData.Message != tc.message {
				if len(tc.message) < 130 || req.PostData.Message != tc.message[:130] {
					t.Errorf("got message: %s, expected %s", req.PostData.Message, tc.message)
				}
			}
			if tc.alias != "" && req.PostData.Alias != tc.alias {
				t.Errorf("got alias: %s, expected %s", req.PostData.Alias, tc.alias)
			}
			if tc.alias == "" && req.PostData.Alias != tc.message {
				t.Errorf("got alias: %s, expected %s", req.PostData.Alias, tc.message)
			}
			if req.PostData.Description != tc.description {
				t.Errorf("got description: %s, expected %s", req.PostData.Description, tc.description)
			}
			if len(req.PostData.Responders) == 0 && len(tc.responders) != 0 {
				t.Errorf("got responders: %v, expected %v", req.PostData.Responders, tc.responders)
			}
			if len(req.PostData.Responders) != 0 && !cmp.Equal(req.PostData.Responders, tc.responders) {
				t.Fatalf("unexpected responders -want/+got\n\n%s\n\n", cmp.Diff(req.PostData.Responders, tc.responders))
			}
			if len(req.PostData.VisibleTo) == 0 && len(tc.visibleTo) != 0 {
				t.Errorf("got visibleTo: %v, expected %v", req.PostData.VisibleTo, tc.visibleTo)
			}
			if len(req.PostData.VisibleTo) != 0 && !cmp.Equal(req.PostData.VisibleTo, tc.visibleTo) {
				t.Fatalf("unexpected visibleTo -want/+got\n\n%s\n\n", cmp.Diff(req.PostData.VisibleTo, tc.visibleTo))
			}
			if len(req.PostData.Actions) == 0 && len(tc.actions) != 0 {
				t.Errorf("got actions: %v, expected %v", req.PostData.Actions, tc.actions)
			}
			if len(req.PostData.Actions) != 0 && !cmp.Equal(req.PostData.Actions, tc.actions) {
				t.Fatalf("unexpected actions -want/+got\n\n%s\n\n", cmp.Diff(req.PostData.Actions, tc.actions))
			}
			if len(req.PostData.Tags) == 0 && len(tc.tags) != 0 {
				t.Errorf("got tags: %v, expected %v", req.PostData.Tags, tc.tags)
			}
			if len(req.PostData.Tags) != 0 && !cmp.Equal(req.PostData.Tags, tc.tags) {
				t.Fatalf("unexpected tags -want/+got\n\n%s\n\n", cmp.Diff(req.PostData.Tags, tc.tags))
			}
			if len(req.PostData.Details) == 0 && len(tc.details) != 0 {
				t.Errorf("got details: %v, expected %v", req.PostData.Details, tc.details)
			}
			if len(req.PostData.Details) != 0 && !cmp.Equal(req.PostData.Details, tc.details) {
				t.Fatalf("unexpected details -want/+got\n\n%s\n\n", cmp.Diff(req.PostData.Details, tc.details))
			}
			if req.PostData.Entity != tc.entity {
				t.Errorf("got entity: %s, expected %s", req.PostData.Entity, tc.entity)
			}
			if tc.priority != "" && req.PostData.Priority != tc.priority {
				t.Errorf("got priority: %s, expected %s", req.PostData.Priority, tc.priority)
			}
			if tc.priority == "" && req.PostData.Priority != "P3" {
				t.Errorf("got priority: %s, expected %s", req.PostData.Priority, "P3")
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
		w.WriteHeader(201)
	}))
	s.ts = ts
	s.URL = ts.URL + "/v2/alerts"
	return s
}
func (s *Server) Request() Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() { s.requests = nil }()
	if len(s.requests) == 0 {
		return Request{}
	} else {
		return s.requests[0]
	}
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
	Header   http.Header
	PostData PostData
}

type PostData struct {
	Message     string              `json:"message"`
	Alias       string              `json:"alias"`
	Description string              `json:"description"`
	Responders  []map[string]string `json:"responders"`
	VisibleTo   []map[string]string `json:"visibleTo"`
	Actions     []string            `json:"actions"`
	Tags        []string            `json:"tags"`
	Details     map[string]string   `json:"details"`
	Entity      string              `json:"entity"`
	Source      string              `json:"source"`
	Priority    string              `json:"priority"`
	Note        string              `json:"note"`
}
