package webexteams_test

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

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependency"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
)

func TestMessage(t *testing.T) {
	s := NewServer(t)
	defer s.Close()
	ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
	defer deps.Finish()

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

func TestEndpoint(t *testing.T) {
	s := NewServer(t)
	defer s.Close()

	url := s.URL // use "https://webexapis.com" for real e2e and verify manually
	roomId := "Y2lzY29zcGFyazovL3VybjpURUFNOmV1LWNlbnRyYWwtMV9rL1JPT00vZWZiYjU0NzAtZDk3My0xMWViLTk4NGYtMGI5OGY0MTJiMzZm"
	token := "YjcwN2ZjYTgtMDMzYi00NTE5LWJjMjMtN2U4Y2E4MWI0NTk3Y2FjZmMyZWQtNzg3_PE93_ed3fff69-a996-4e21-b5af-0dc3ad437459"

	testCases := []struct {
		name     string // also name of the test
		text     string
		markdown string
		roomId   string
	}{
		{
			name:   "text_to_room",
			text:   "abc",
			roomId: roomId,
		},
		{
			name:     "markdown_to_room",
			markdown: "**abc**",
			roomId:   roomId,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s.Reset()
			fluxString := `import "csv"
import "contrib/sranka/webexteams"

url = "` + url + `"
token = "` + token + `"

endpoint = webexteams.endpoint(url:url,token:token)(mapFn: (r) => {
	return {roomId:r.roomId,text:r.text,markdown:r.markdown}
 })

 data = "
#datatype,string,string,string,string,string
#group,false,false,false,false,false
#default,_result,,,,
,result,,roomId,text,markdown
,,,` + strings.Join([]string{tc.roomId, tc.text, tc.markdown}, ",") + `"

csv.from(csv:data) |> endpoint()
 `
			ctx := flux.NewDefaultDependencies().Inject(context.Background())

			prog, err := lang.Compile(ctx, fluxString, runtime.Default, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("*** ", tc.name)
			query, err := prog.Start(ctx, &memory.ResourceAllocator{})

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
					fmt.Println("*** TABLE ***")
					for i, meta := range reader.Cols() {
						fmt.Println(meta.Label)
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

			if req.URL != "/v1/messages" {
				t.Errorf("got URL: %s, expected %s", req.URL, "/v1/messages")
			}
			if !strings.HasPrefix(req.ContentType, "application/json") {
				t.Errorf("got content-type: %s, expected application/json", req.ContentType)
			}
			if req.Authorization != ("Bearer " + token) {
				t.Errorf("got authorization: %s, expected %s", req.Authorization, ("Bearer " + token))
			}
			if req.PostData.RoomID != tc.roomId {
				t.Errorf("got roomId: %s, expected %s", req.PostData.RoomID, tc.roomId)
			}
			if req.PostData.Text != tc.text {
				t.Errorf("got text: %s, expected %s", req.PostData.Text, tc.text)
			}
			if req.PostData.Markdown != tc.markdown {
				t.Errorf("got markdown: %s, expected %s", req.PostData.Markdown, tc.markdown)
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
			URL:           r.URL.String(), // r.URL.String(),
			ContentType:   r.Header.Get("content-type"),
			Authorization: r.Header.Get("authorization"),
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
	URL           string
	ContentType   string
	Authorization string
	PostData      PostData
}

type PostData struct {
	RoomID   string `json:"roomId"`
	PersonID string `json:"personId"`
	Text     string `json:"text"`
	Markdown string `json:"markdown"`
}
