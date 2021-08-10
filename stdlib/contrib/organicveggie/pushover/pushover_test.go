package pushover_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	runtime "github.com/influxdata/flux/runtime"
)

func TestPushover(t *testing.T) {
	ctx := dependenciestest.Default().Inject(context.Background())

	_, _, err := runtime.Eval(ctx, `
import "contrib/organicveggie/pushover"
send = pushover.send(apiToken:"TestToken",userKey:"TestUserKey",content:"test content",)
send == 204
`)

	if err != nil {
		t.Error("evaluation of pushover.send failed: ", err)
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
	s.URL = ts.URL + "/"
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
	URL      string
	PostData PostData
}

type PostData struct {
	Token    string `json:"token"`
	UserKey  string `json:"user"`
	Content  string `json:"message"`
	Priority int    `json:"priority"`
	Device   string `json:"device"`
}

func TestPushoverEndpoint(t *testing.T) {
	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		name     string
		token    string
		user     string
		content  string
		priority int
		device   string
	}{
		{
			name:    "BasicSuccess",
			token:   "FakeToken",
			user:    "MyUser",
			content: "This is a test message.",
		},
		{
			name:     "CustomPrioritySuccess",
			token:    "PriorityToken",
			user:     "PriorityUser",
			content:  "This is a priority message.",
			priority: 2,
		},
		{
			name:    "DeviceSuccess",
			token:   "DeviceToken",
			user:    "DeviceUser",
			content: "This is a device message.",
			device:  "TestDevice",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fluxString := `import "csv"
			import "contrib/organicveggie/pushover"
			option pushover.pushoverURL = "` + s.URL + `"
			
			endpoint = pushover.endpoint(apiToken:token, userKey:user, priority:priority, device:device)(mapFn: (r) => {
			 return {content:r.qtext}
			})
			
			csv.from(csv:data) |> endpoint() `

			extern := `
token = "` + tc.token + `"
user = "` + tc.user + `"
priority = ` + strconv.Itoa(tc.priority) + `
device = "` + tc.device + `"
data = "
#datatype,string,string,string
#group,false,false,false
#default,_result,,
,result,,qtext
,,,` + tc.content + `"`

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

			var hasSent bool
			err = res.Tables().Do(func(table flux.Table) error {
				return table.Do(func(reader flux.ColReader) error {
					if reader == nil {
						return nil
					}
					for i, meta := range reader.Cols() {
						if meta.Label == "_sent" {
							hasSent = true
							if v := reader.Strings(i).Value(0); string(v) != "true" {
								t.Fatalf("expecting _sent=true but got _sent=%v", string(v))
							}
						}
					}
					return nil
				})
			})
			if !hasSent {
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
			if req.PostData.Content != tc.content {
				t.Errorf("got content: %s, expected %s", req.PostData.Content, tc.content)
			}
			if req.PostData.Token != tc.token {
				t.Errorf("got token: %s, expected %s", req.PostData.Token, tc.token)
			}
			if req.PostData.UserKey != tc.user {
				t.Errorf("got user: %s, expected %s", req.PostData.UserKey, tc.user)
			}
		})
	}
}
