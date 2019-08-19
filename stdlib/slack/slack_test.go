package slack_test

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

	"github.com/influxdata/flux/memory"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/lang"

	"github.com/influxdata/flux/values"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/csv"

	//_ "github.com/influxdata/flux/stdlib" // Import the stdlib

	_ "github.com/influxdata/flux/builtin"
)

func TestSlack(t *testing.T) {

	_, scope, err := flux.Eval(`
import "csv"
import "slack"


data = "
#datatype,string,string,string,string,string,string,string
#group,false,false,false,false,false,false,false
#default,_result,,,
,result,qusername,qchannel,qworkspace,qtext,qiconEmoji,qiconEmoji,qcolor
,,fakeUser0,fakeChannel,workspace,this is a lot of text yay,\"#FF0000\"
"

process = slack.endpoint(url:url, token:token)( mapFn: 
	(r) => {
		return {username:r.qusername,channel:r.qchannel,workspace:r.qorkspace,text:r.qtext,iconEmoji:r.qiconEmoji,color:r.color}
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
	_ = scope
}

type Server struct {
	mu       sync.Mutex
	ts       *httptest.Server
	URL      string
	token    string
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
	URL           string
	Authorization string
	PostData      PostData
}

type PostData struct {
	Channel     string       `json:"channel"`
	Workspace   string       `json:"workspace"`
	Icon        string       `json:"icon_emoji"`
	Username    string       `json:"username"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Color    string   `json:"color"`
	Text     string   `json:"text"`
	MrkdwnIn []string `json:"mrkdwn_in"`
}

func TestSlackPost(t *testing.T) {

	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		name      string
		color     string
		text      string
		channel   string
		URL       string
		token     string
		username  string
		workspace string
		icon      string
	}{
		{
			name:     "....",
			color:    "#ababab",
			text:     "aaaaaaa",
			channel:  "fooo",
			URL:      s.URL,
			token:    "faketoken",
			username: "username",
			icon:     ":thumb:",
		},
	}
	//exe := execute.NewExecutor(nil, zap.NewNop())

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			//_, scope, err := flux.Eval(`
			fluxString := `import "csv"
import "slack"

endpoint = slack.endpoint(url:url, token:token)(mapFn: (r) => {
	return {username:r.fusername,channel:r.qchannel,workspace:r.qworkspace,text:r.qtext,iconEmoji:r.qiconEmoji,color:r.wcolor}
})

csv.from(csv:data) |> endpoint()
` /*, func(s interpreter.Scope) {
				s.Set("url", values.New("http://fakeurl.com/fakeyfake"))
				s.Set("token", values.New("faketoken"))

			})*/

			// http.Post(s.URL, "text/json", bytes.NewBufferString(`{"token":"hello"}`))

			prog, err := lang.Compile(fluxString, time.Now(), lang.WithExtern(&ast.File{Body: []ast.Statement{
				&ast.VariableAssignment{
					ID: &ast.Identifier{
						Name: "url",
					},
					Init: &ast.StringLiteral{
						Value: tc.URL,
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
						Value: `#datatype,string,string,string,string,string,string,string,string
#group,false,false,false,false,false,false,false,false
#default,_result,,,,,,,
,result,,fusername,qchannel,qworkspace,qtext,qiconEmoji,wcolor
,,,` + strings.Join([]string{tc.username, tc.channel, tc.workspace, tc.text, tc.icon, tc.color}, ","),
					},
				},
			}}))
			if err != nil {
				t.Fatal(err)
			}
			query, err := prog.Start(context.Background(), &memory.Allocator{})

			if err != nil {
				t.Fatal(err)
			}
			res := <-query.Results()
			_ = res
			fmt.Println(res.Name())
			err = res.Tables().Do(func(table flux.Table) error {
				return table.Do(func(reader flux.ColReader) error {
					//fmt.Println(reader.Strings(0))
					return nil
				})
			})
			if err != nil {
				t.Fatal(err)
			}

			query.Done()
			if err := query.Err(); err != nil {
				t.Error(err)
			}
			reqs := s.Requests()

			if len(reqs) < 1 {
				t.Fatal("recieved no requests")
			}
			req := reqs[len(reqs)-1]

			if s.URL != req.URL {
				t.Errorf("calling out to wrong slack url expected %s, got %s", s.URL, req.URL)
			}
			if req.Authorization != tc.token {
				t.Errorf("token incorrect got %s, expected %s", req.Authorization, tc.token)
			}
			if len(req.PostData.Attachments) != 1 {
				t.Fatalf("expected 1 attachment got %d", len(req.PostData.Attachments))
			}
			if req.PostData.Text != tc.text {
				t.Errorf(" got %s text, expected text of %s", tc.text, req.PostData.Text)
			}
			if req.PostData.Channel != tc.channel {
				t.Errorf("got channel: %s, expected %s", tc.channel, req.PostData.Channel)
			}
			if len(req.PostData.Attachments[0].MrkdwnIn) != 0 && req.PostData.Attachments[0].MrkdwnIn[0] != "text" {
				t.Errorf("mrkdwn_in field incorrect, should be lenth 1 with a string text in a json array")
			}
			if req.PostData.Attachments[0].Color != tc.color {
				t.Errorf("got color %s, expected %s", req.PostData.Attachments[0].Color, tc.color)
			}
			if req.PostData.Username != tc.username {
				t.Errorf("got username %s, expected %s", req.PostData.Username, tc.username)
			}
			if req.PostData.Workspace != tc.workspace {
				t.Errorf("got workspace %s, expected %s", req.PostData.Workspace, tc.workspace)
			}
			if req.PostData.Icon != tc.icon {
				t.Errorf("got icon-emoji %s, expected %s", req.PostData.Icon, tc.icon)
			}
		})
	}

}
