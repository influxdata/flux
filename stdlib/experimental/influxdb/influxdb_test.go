package influxdb

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

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
		expectedErrorMessage          string
		expectedStatusCode            int64
		expectedResponseBody          []byte
		expectedResponseContentLength string
		expectedResponseHeaderCount   int

		// status and responses to be returned from mock api host
		status   int
		response []byte

		// method, status, token and paths mock api host expects to receive
		expectedMethod         string
		expectedRequestBody    []byte
		expectedRequestPath    string
		expectedRequestToken   string
		expectedRequestHeaders http.Header
		expectedRequestQuery   url.Values
	}{
		{
			name: "get",
			args: values.NewObjectWithValues(map[string]values.Value{
				"host":   values.NewString("placeholder"),
				"method": values.NewString("get"),
				"path":   values.NewString("/api/v2/foo"),
				"token":  values.NewString("passedtoken"),
			}),
			expectedStatusCode:            200,
			expectedRequestPath:           "/api/v2/foo",
			expectedMethod:                "get",
			expectedRequestToken:          "passedtoken",
			expectedResponseContentLength: "0",
			expectedResponseHeaderCount:   2,
		},
		{
			name: "get with headers",
			args: values.NewObjectWithValues(map[string]values.Value{
				"host":   values.NewString("placeholder"),
				"method": values.NewString("get"),
				"path":   values.NewString("/api/v2/foo"),
				"token":  values.NewString("passedtoken"),
				"headers": newDictWithValues(map[string]string{
					"Key": "Value",
				}),
			}),
			expectedStatusCode:   200,
			expectedRequestPath:  "/api/v2/foo",
			expectedMethod:       "get",
			expectedRequestToken: "passedtoken",
			expectedRequestHeaders: map[string][]string{
				"Accept-Encoding": {"gzip"},
				"Authorization":   {"Token passedtoken"},
				"User-Agent":      {"Go-http-client/1.1"},
				"Key":             {"Value"},
			},
			expectedResponseContentLength: "0",
			expectedResponseHeaderCount:   2,
		},
		{
			name: "get with query",
			args: values.NewObjectWithValues(map[string]values.Value{
				"host":   values.NewString("placeholder"),
				"method": values.NewString("get"),
				"path":   values.NewString("/api/v2/foo"),
				"token":  values.NewString("passedtoken"),
				"query": newDictWithValues(map[string]string{
					"Key": "Value",
				}),
			}),
			expectedStatusCode:   200,
			expectedRequestPath:  "/api/v2/foo",
			expectedMethod:       "get",
			expectedRequestToken: "passedtoken",
			expectedRequestQuery: map[string][]string{
				"Key": {"Value"},
			},
			expectedResponseContentLength: "0",
			expectedResponseHeaderCount:   2,
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
			expectedStatusCode:            200,
			expectedResponseBody:          []byte(fakeData),
			expectedRequestPath:           "/api/v2/bar",
			expectedMethod:                "get",
			expectedRequestToken:          "passedtoken",
			expectedResponseContentLength: "565",
			expectedResponseHeaderCount:   3,
		},
		{
			name:   "post created",
			status: 201,
			args: values.NewObjectWithValues(map[string]values.Value{
				"host":   values.NewString("placeholder"),
				"method": values.NewString("post"),
				"path":   values.NewString("/api/v2/baz"),
				"body":   values.NewBytes([]byte(`{"Key":"Value"}`)),
				"token":  values.NewString("passedtoken"),
			}),
			expectedRequestBody:  []byte(`{"Key":"Value"}`),
			expectedRequestPath:  "/api/v2/baz",
			expectedStatusCode:   201,
			expectedMethod:       "post",
			expectedRequestToken: "passedtoken",
			expectedRequestHeaders: http.Header{
				"Accept-Encoding": {"gzip"},
				"Authorization":   {"Token passedtoken"},
				"Content-Length":  {"15"},
				"User-Agent":      {"Go-http-client/1.1"},
			},
			expectedResponseContentLength: "0",
			expectedResponseHeaderCount:   2,
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
			expectedStatusCode:            500,
			expectedResponseBody:          []byte(`{"code":"internal","message":"internal error"}`),
			expectedRequestPath:           "/api/v2/bing",
			expectedMethod:                "get",
			expectedRequestToken:          "passedtoken",
			expectedResponseContentLength: "46",
			expectedResponseHeaderCount:   3,
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
			apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got, expected := strings.ToUpper(r.Method), strings.ToUpper(test.expectedMethod); got != expected {
					t.Errorf("unexpected request method: got %s, expected %s", got, expected)
				}

				if r.URL.Path != test.expectedRequestPath {
					t.Errorf("unexpected request path: got %s, expected %s", r.URL.Path, test.expectedRequestPath)
				}

				expectedHeaders := defaultExpectedRequestHeaders
				if test.expectedRequestHeaders != nil {
					expectedHeaders = test.expectedRequestHeaders
				}
				if diff := cmp.Diff(expectedHeaders, r.Header); diff != "" {
					t.Errorf("unexpected request headers: %s", diff)
				}

				expectedQuery := defaultExpectedRequestQuery
				if test.expectedRequestQuery != nil {
					expectedQuery = test.expectedRequestQuery
				}
				if diff := cmp.Diff(expectedQuery, r.URL.Query()); diff != "" {
					t.Errorf("unexpected request URL query: %s", diff)
				}

				if requestBody, _ := ioutil.ReadAll(r.Body); !bytes.Equal(requestBody, test.expectedRequestBody) {
					t.Errorf("unexpected request body: got %s, expected %s", requestBody, test.expectedRequestBody)
				}

				if test.status != 0 {
					w.WriteHeader(test.status)
				}
				_, _ = w.Write([]byte(test.response))
			}))
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

				// http.Server returns Content-Length and Date. The Date is difficult to test since it is
				// time dependent and contrary to documentation, it is impossible to suppress this behavior.
				// Therefore we only test for the Content-Length and total dictionary length here.
				responseHeadersObj, _ := result.Object().Get("headers")
				responseHeaders := responseHeadersObj.Dict()
				if got, expected := responseHeaders.Len(), test.expectedResponseHeaderCount; got != expected {
					t.Errorf("unexpected response headers: got %d, expected %d", got, expected)
				} else if contentLength := responseHeaders.Get(
					values.NewString("Content-Length"), values.NewString("")); contentLength.Str() != test.expectedResponseContentLength {
					t.Errorf("unexpected Content-Length header value: expected %s, got %s", contentLength.Str(), test.expectedResponseContentLength)
				}
			}
		})
	}
}

const fakeData = ",result,table,_start,_stop,_time,_value,_field,_measurement,endpoint,org_id,status\n,_result,0,2021-03-13T19:29:20.9874663Z,2021-03-15T19:29:20.9874663Z,2021-03-14T04:15:52.3897524Z,15838,req_bytes,http_request,/api/v2/write,0000000000001002,204\n,_result,0,2021-03-13T19:29:20.9874663Z,2021-03-15T19:29:20.9874663Z,2021-03-14T04:16:02.3428779Z,7924,req_bytes,http_request,/api/v2/write,0000000000001002,204\n,_result,0,2021-03-13T19:29:20.9874663Z,2021-03-15T19:29:20.9874663Z,2021-03-14T04:16:12.437844Z,7924,req_bytes,http_request,/api/v2/write,0000000000001002,204"

func newDictWithValues(m map[string]string) values.Dictionary {
	dict := values.NewDict(semantic.NewDictType(semantic.BasicString, semantic.BasicString))
	for k, v := range m {
		dict, _ = dict.Insert(values.NewString(k), values.NewString(v))
	}
	return dict
}

var defaultExpectedRequestHeaders = http.Header{
	"Accept-Encoding": {"gzip"},
	"Authorization":   {"Token passedtoken"},
	"User-Agent":      {"Go-http-client/1.1"},
}

var defaultExpectedRequestQuery = url.Values{}
