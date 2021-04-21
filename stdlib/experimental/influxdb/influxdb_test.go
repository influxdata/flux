package influxdb

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func Test_api(t *testing.T) {
	for _, test := range []struct {
		name string

		// passed to api func
		args values.Object

		// expected to be returned from api func call
		expectedErrorMessage string
		expectedStatusCode   int64
		expectedResponseBody []byte

		// status and responses to be returned from mock api host
		status   int
		response []byte

		// method, status, token and paths mock api host expects to receive
		expectedMethod       string
		expectedRequestBody  []byte
		expectedRequestPath  string
		expectedRequestToken string
	}{
		{
			name: "get",
			args: values.NewObjectWithValues(map[string]values.Value{
				"host":   values.NewString("placeholder"),
				"method": values.NewString("get"),
				"path":   values.NewString("/api/v2/foo"),
				"token":  values.NewString("passedtoken"),
			}),
			expectedStatusCode:   200,
			expectedRequestPath:  "/api/v2/foo",
			expectedMethod:       "get",
			expectedRequestToken: "passedtoken",
		},
		{
			name: "get with headers",
			args: values.NewObjectWithValues(map[string]values.Value{
				"host":   values.NewString("placeholder"),
				"method": values.NewString("get"),
				"path":   values.NewString("/api/v2/foo"),
				"token":  values.NewString("passedtoken"),
				"headers": newDictWithValues(map[string]string{
					"key": "value",
				}),
			}),
			expectedStatusCode:   200,
			expectedRequestPath:  "/api/v2/foo",
			expectedMethod:       "get",
			expectedRequestToken: "passedtoken",
		},
		{
			name: "get with query",
			args: values.NewObjectWithValues(map[string]values.Value{
				"host":   values.NewString("placeholder"),
				"method": values.NewString("get"),
				"path":   values.NewString("/api/v2/foo"),
				"token":  values.NewString("passedtoken"),
				"query": newDictWithValues(map[string]string{
					"key": "value",
				}),
			}),
			expectedStatusCode:   200,
			expectedRequestPath:  "/api/v2/foo",
			expectedMethod:       "get",
			expectedRequestToken: "passedtoken",
		},
		{
			name:     "get returning data",
			response: []byte(fakeData),
			args: values.NewObjectWithValues(map[string]values.Value{
				"host":   values.NewString("placeholder"),
				"method": values.NewString("get"),
				"path":   values.NewString("/api/v2/bar"),
				"token":  values.NewString("passedtoken"),
			}),
			expectedStatusCode:   200,
			expectedResponseBody: []byte(fakeData),
			expectedRequestPath:  "/api/v2/bar",
			expectedMethod:       "get",
			expectedRequestToken: "passedtoken",
		},
		{
			name:   "post created",
			status: 201,
			args: values.NewObjectWithValues(map[string]values.Value{
				"host":   values.NewString("placeholder"),
				"method": values.NewString("post"),
				"path":   values.NewString("/api/v2/baz"),
				"body":   values.NewBytes([]byte(`{"key":"value"}`)),
				"token":  values.NewString("passedtoken"),
			}),
			expectedRequestBody:  []byte(`{"key":"value"}`),
			expectedRequestPath:  "/api/v2/baz",
			expectedStatusCode:   201,
			expectedMethod:       "post",
			expectedRequestToken: "passedtoken",
		},
		{
			name:     "error",
			status:   500,
			response: []byte(`{"code":"internal","message":"internal error"}`),
			args: values.NewObjectWithValues(map[string]values.Value{
				"host":   values.NewString("placeholder"),
				"method": values.NewString("get"),
				"path":   values.NewString("/api/v2/bing"),
				"token":  values.NewString("passedtoken"),
			}),
			expectedStatusCode:   500,
			expectedResponseBody: []byte(`{"code":"internal","message":"internal error"}`),
			expectedRequestPath:  "/api/v2/bing",
			expectedMethod:       "get",
			expectedRequestToken: "passedtoken",
		},
		{
			name: "error missing args",
			args: values.NewObjectWithValues(map[string]values.Value{
				"host":   values.NewString("placeholder"),
				"method": values.NewString("get"),
			}),
			expectedErrorMessage: `missing required keyword argument "path"`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			newServer := func(status int, response []byte, expectedToken, expectedMethod, expectedPath string, expectedBody []byte) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if got, expected := strings.ToUpper(r.Method), strings.ToUpper(expectedMethod); got != expected {
						t.Errorf("unexpected request method: got %s, expected %s", got, expected)
					}

					if got := strings.TrimPrefix(r.Header.Get("Authorization"), "Token "); got != expectedToken {
						t.Errorf("unexpected request token: got %s, expected %s", got, expectedToken)
					}

					if r.URL.Path != expectedPath {
						t.Errorf("unexpected request path: got %s, expected %s", r.URL.Path, expectedPath)
					}

					if requestBody, _ := ioutil.ReadAll(r.Body); !bytes.Equal(requestBody, expectedBody) {
						t.Errorf("unexpected request body: got %s, expected %s", requestBody, expectedBody)
					}

					w.Header().Set("Date", "someday")
					if status != 0 {
						w.WriteHeader(status)
					}
					_, _ = w.Write([]byte(response))
				}))
			}

			apiServer := newServer(
				test.status,
				test.response,
				test.expectedRequestToken,
				test.expectedMethod,
				test.expectedRequestPath,
				test.expectedRequestBody,
			)
			test.args.Set("host", values.NewString(apiServer.URL))

			ctx := flux.NewDefaultDependencies().Inject(context.Background())

			result, err := api(ctx, test.args)

			if test.expectedErrorMessage != "" {
				if err == nil {
					t.Errorf("missing expected error: %s", test.expectedErrorMessage)
				} else if err.Error() != test.expectedErrorMessage {
					t.Errorf("unexpected error: expected %s, got %s", test.expectedErrorMessage, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if result == nil {
					t.Error("unexpected nil response body")
				} else if statusCode, _ := result.Object().Get("statusCode"); statusCode.Int() != test.expectedStatusCode {
					t.Errorf("unexpected status code: got %d, expected %d", statusCode.Int(), test.expectedStatusCode)
				} else if responseBody, _ := result.Object().Get("body"); !bytes.Equal(responseBody.Bytes(), test.expectedResponseBody) {
					t.Errorf("unexpected response body: got %s, expected %s", responseBody.Bytes(), test.expectedResponseBody)
				}
			}
		})
	}
}

const fakeData = ",result,table,_start,_stop,_time,_value,_field,_measurement,endpoint,org_id,status\n,_result,0,2021-03-13T19:29:20.9874663Z,2021-03-15T19:29:20.9874663Z,2021-03-14T04:15:52.3897524Z,15838,req_bytes,http_request,/api/v2/write,0000000000001002,204\n,_result,0,2021-03-13T19:29:20.9874663Z,2021-03-15T19:29:20.9874663Z,2021-03-14T04:16:02.3428779Z,7924,req_bytes,http_request,/api/v2/write,0000000000001002,204\n,_result,0,2021-03-13T19:29:20.9874663Z,2021-03-15T19:29:20.9874663Z,2021-03-14T04:16:12.437844Z,7924,req_bytes,http_request,/api/v2/write,0000000000001002,204"

func newDictWithValues(m map[string]string) values.Dictionary {
	dict := values.NewDict(semantic.NewDictType(semantic.BasicString, semantic.BasicString))
	for k, v := range m {
		dict.Insert(values.NewString(k), values.NewString(v))
	}
	return dict
}
