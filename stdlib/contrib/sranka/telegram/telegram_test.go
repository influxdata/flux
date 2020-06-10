package telegram_test

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

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/builtin"
	_ "github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
)

func TestTelegram(t *testing.T) {
	s := NewServer(t)
	defer s.Close()
	ctx := dependenciestest.Default().Inject(context.Background())
	_, scope, err := runtime.Eval(ctx, `
import "csv"
import "contrib/sranka/telegram"

option url = "`+s.URL+`"
option token = "faketoken"

data = "
#datatype,string,string,string
#group,false,false,false
#default,_result,,
,result,qchannel,qtext
,,fakeChannel,this is a lot of text yay
"

process = telegram.endpoint(url:url, token:token)( mapFn:
	(r) => {
		return {channel:r.qchannel,text:r.qtext, silent:true}
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
	s.URL = ts.URL + "/bot"
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
	Channel               string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
	DisableNotification   bool   `json:"disable_notification"`
}

func TestTelegramPost(t *testing.T) {

	s := NewServer(t)
	defer s.Close()

	testCases := []struct {
		token                 string
		name                  string
		text                  string
		channel               string
		parseMode             string
		silent                bool
		endpointExtraArgs     string
		disableWebPagePreview bool
	}{
		{
			name:                  "simple",
			token:                 "123",
			text:                  "aaaaaaab",
			channel:               "whatever",
			silent:                true,
			endpointExtraArgs:     "",
			parseMode:             "MarkdownV2",
			disableWebPagePreview: false,
		},
		{
			name:                  "nonDefaultOptionals",
			token:                 "123",
			text:                  "aaaaaaab",
			channel:               "whatever",
			silent:                false,
			endpointExtraArgs:     `disableWebPagePreview: true, parseMode: "HTML", `,
			parseMode:             "HTML",
			disableWebPagePreview: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			fluxString := `import "csv"
import "contrib/sranka/telegram"

endpoint = telegram.endpoint(` + tc.endpointExtraArgs + `url:url, token:token)(mapFn: (r) => {
 return {channel:r.qchannel,text:r.qtext, silent:` + strconv.FormatBool(tc.silent) + `}
})

csv.from(csv:data) |> endpoint() `
			extern := `
url = "` + s.URL + `"
token = "` + tc.token + `"
data = "
#datatype,string,string,string,string
#group,false,false,false,false
#default,_result,,,
,result,,qchannel,qtext
,,,` + strings.Join([]string{tc.channel, tc.text}, ",") + `"`

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

			if req.URL != "/bot"+tc.token+"/sendMessage" {
				t.Errorf("got URL: %s, expected %s", req.URL, "/bot"+tc.token+"/sendMessage")
			}
			if req.PostData.Channel != tc.channel {
				t.Errorf("got channel: %s, expected %s", req.PostData.Channel, tc.channel)
			}
			if req.PostData.ParseMode != tc.parseMode {
				t.Errorf("got parseMode: %s, expected %s", req.PostData.ParseMode, tc.parseMode)
			}
			if req.PostData.DisableNotification != tc.silent {
				t.Errorf("got disableNotification: %v, expected %v", req.PostData.DisableNotification, tc.silent)
			}
			if req.PostData.DisableWebPagePreview != tc.disableWebPagePreview {
				t.Errorf("got disableWebPagePreview: %v, expected %v", req.PostData.DisableWebPagePreview, tc.disableWebPagePreview)
			}
			if req.PostData.Text != tc.text {
				t.Errorf("got text: %s, expected %s", req.PostData.Text, tc.text)
			}
		})
	}

}
