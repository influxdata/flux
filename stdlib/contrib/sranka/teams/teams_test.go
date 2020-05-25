package teams_test

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
	_ "github.com/influxdata/flux/builtin"
	_ "github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
)

func TestTeams(t *testing.T) {
	s := NewServer(t)
	defer s.Close()
	ctx := dependenciestest.Default().Inject(context.Background())
	_, scope, err := runtime.Eval(ctx, `
import "csv"
import "contrib/sranka/teams"

option url = "`+s.URL+`"

data = "
#datatype,string,string,string,string
#group,false,false,false,false
#default,_result,,,
,result,qtitle,qtext,qsummary
,,fakeChannel,this is a lot of text yay, summarized
"

process = teams.endpoint(url:url)( mapFn:
	(r) => {
		return {title:r.qtitle,text:r.qtext,summary:r.qsummary}
	}
)

csv.from(csv:data) |> process()

`)

	if err != nil {
		t.Error(err)
	}
	_ = scope
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

type Request struct {
	URL      string
	PostData PostData
}

type PostData struct {
	Type    string `json:"@type"`
	Context string `json:"@context"`
	Title   string `json:"title"`
	Text    string `json:"text"`
	Summary string `json:"summary"`
}

func TestTeamsPost(t *testing.T) {

	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		name          string
		title         string
		text          string
		summary       string
		expectSummary string
	}{
		{
			name:          "simple",
			title:         "my title",
			text:          "aaaaaaab",
			summary:       "aaaSummary",
			expectSummary: "aaaSummary",
		},
		{
			name:          "summaryFromText",
			title:         "my title",
			text:          "aaaaaaab",
			summary:       "",
			expectSummary: "aaaaaaab",
		},
		{
			name:          "truncatedSummary",
			title:         "my title",
			text:          "aaaaaaab",
			summary:       "my 3456789-...20...--...30...--...40...--...50...--...60...--...70...--...80...-",
			expectSummary: "my 3456789-...20...--...30...--...40...--...50...--...60...--...70...-...",
		},
		{
			name:          "truncatedSummaryFromText",
			title:         "my title",
			text:          "my 3456789-...20...--...30...--...40...--...50...--...60...--...70...--...80...-",
			summary:       "",
			expectSummary: "my 3456789-...20...--...30...--...40...--...50...--...60...--...70...-...",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			fluxString := `import "csv"
import "contrib/sranka/teams"

endpoint = teams.endpoint(url:url)(mapFn: (r) => {
 return {title: r.qtitle, text: r.qtext, summary: r.qsummary }
})

csv.from(csv:data) |> endpoint() `
			extern := `
url = "` + s.URL + `"
data = "
#datatype,string,string,string,string,string
#group,false,false,false,false,false
#default,_result,,,,
,result,,qtitle,qtext,qsummary
,,,` + strings.Join([]string{tc.title, tc.text, tc.summary}, ",") + `"`

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

			if req.PostData.Title != tc.title {
				t.Errorf("got title: %s, expected %s", req.PostData.Title, tc.title)
			}
			if req.PostData.Text != tc.text {
				t.Errorf("got text: %s, expected %s", req.PostData.Text, tc.text)
			}
			if req.PostData.Summary != tc.expectSummary {
				t.Errorf("got summary: %s, expected %s", req.PostData.Summary, tc.expectSummary)
			}
		})
	}

}
