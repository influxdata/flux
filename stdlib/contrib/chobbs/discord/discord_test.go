package discord_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
)

func TestDiscord(t *testing.T) {
	ctx := dependenciestest.Default().Inject(context.Background())
	_, scope, err := runtime.Eval(ctx, `
import "contrib/chobbs/discord"
send = discord.send(webhookToken:"ThisIsAFakeToken",webhookID:"123456789",username:"chobbs",content:"this is fake content!",avatar_url:"%s/somefakeurl.com/pic.png")
send == 204
`)

	if err != nil {
		t.Error("evaluation of discord.send failed: ", err)
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
	Username  string `json:"username"`
	Content   string `json:"content"`
	AvatarURL string `json:"avatar_url"`
}

func TestDiscordEndpoint(t *testing.T) {

	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		webhookID    string
		webhookToken string
		username     string
		content      string
		avatar_url   string
		extraArgs    string
	}{
		{
			webhookID:    "simple",
			webhookToken: "fakeToken",
			username:     "influxdb",
			content:      "whatever",
		},
		{
			webhookID:    "withAvatarUrl",
			webhookToken: "fakeToken",
			username:     "influxdb",
			content:      "whatever",
			avatar_url:   "myavaurl",
			extraArgs:    `,avatar_url:"myavaurl"`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.webhookID, func(t *testing.T) {

			fluxString := `import "csv"
import "contrib/chobbs/discord"
option discord.discordURL = "` + s.URL + `"

endpoint = discord.endpoint(webhookToken:webhookToken, webhookID:webhookID, username:username ` + tc.extraArgs + `)(mapFn: (r) => {
 return {content:r.qtext}
})

csv.from(csv:data) |> endpoint() `
			extern := `
webhookToken = "` + tc.webhookToken + `"
webhookID = "` + tc.webhookID + `"
username = "` + tc.username + `"
avatar_url = "` + tc.avatar_url + `"
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

			if req.URL != "/"+tc.webhookID+"/"+tc.webhookToken {
				t.Errorf("got URL: %s, expected %s", req.URL, "/"+tc.webhookID+"/"+tc.webhookToken)
			}
			if req.PostData.Content != tc.content {
				t.Errorf("got content: %s, expected %s", req.PostData.Content, tc.content)
			}
			if req.PostData.Username != tc.username {
				t.Errorf("got username: %s, expected %s", req.PostData.Username, tc.username)
			}
			if req.PostData.AvatarURL != tc.avatar_url {
				t.Errorf("got avatar_url: %s, expected %s", req.PostData.AvatarURL, tc.avatar_url)
			}
		})
	}

}
