package webexteams_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	_ "github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/runtime"
)

func TestMessage(t *testing.T) {
	s := NewServer(t)
	defer s.Close()
	ctx := dependenciestest.Default().Inject(context.Background())
	_, _, err := runtime.Eval(ctx, `
import "csv"
import "contrib/sranka/webexteams"

url = "`+s.URL+`"
token = "fakeApiKey"
roomId = "myRoomId"
text = "abc"
status = webexteams.message(url:url,token:token,roomId:roomId,text: "Hi there",markdown: "")`)

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
