package prometheus

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
)

type PromClient struct {
	Server  *url.URL
	Auth    string
	hasAuth bool
}

func NewPromClient(addr string) (*PromClient, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	return &PromClient{
		Server: u,
	}, nil
}
func NewAuthPromClient(addr, user, password string) (*PromClient, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	return &PromClient{
		Server:  u,
		Auth:    basicAuth(user, password),
		hasAuth: true,
	}, nil
}
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

type QueryRangeResponse struct {
	Status string                  `json:"status"`
	Data   *QueryRangeResponseData `json:"data"`
}
type QueryRangeResponseData struct {
	Result []*QueryRangeResponseResult `json:"result"`
}
type QueryRangeResponseResult struct {
	Metric map[string]string          `json:"metric"`
	Values []*QueryRangeResponseValue `json:"values"`
}
type QueryRangeResponseValue []interface{}

func (v *QueryRangeResponseValue) Time() time.Time {
	t := (*v)[0].(float64)
	return time.Unix(int64(t), 0)
}
func (v *QueryRangeResponseValue) Value() (float64, error) {
	s := (*v)[1].(string)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}
func (c *PromClient) QueryRange(query string, start time.Time, end time.Time, step time.Duration) (*QueryRangeResponse, error) {
	u, err := url.Parse(fmt.Sprintf("./api/v1/query_range?query=%s&start=%s&end=%s&step=%s",
		url.QueryEscape(query),
		url.QueryEscape(fmt.Sprintf("%d", start.Unix())),
		url.QueryEscape(fmt.Sprintf("%d", end.Unix())),
		url.QueryEscape(fmt.Sprintf("%ds", int(step.Seconds()))),
	))
	if err != nil {
		return nil, err
	}
	u = c.Server.ResolveReference(u)
	req, err := http.NewRequest("GET", u.String(), nil)
	if c.hasAuth {
		req.Header.Add("Authorization", "Basic "+c.Auth)
	}
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		if err == io.EOF {
			return &QueryRangeResponse{}, nil
		}
		return nil, err
	}
	if 400 <= res.StatusCode {
		return nil, fmt.Errorf("error response: %s", string(body))
	}
	resp := &QueryRangeResponse{}
	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *PromClient) QueryRemoteRead(query *prompb.Query) (*prompb.ReadResponse, error) {
	u, err := url.Parse(fmt.Sprintf("./remote_read"))
	req := &prompb.ReadRequest{
		Queries: []*prompb.Query{
			query,
		},
	}
	u = c.Server.ResolveReference(u)
	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal read request: %v", err)
	}
	compressed := snappy.Encode(nil, data)
	httpReq, err := http.NewRequest("POST", u.String(), bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %v", err)
	}
	httpReq.Header.Add("Content-Encoding", "snappy")
	httpReq.Header.Add("Accept-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("X-Prometheus-Remote-Read-Version", "0.1.0")
	if c.hasAuth {
		httpReq.Header.Add("Authorization", "Basic "+c.Auth)
	}
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("server returned HTTP status %s", httpResp.Status)
	}
	compressed, err = ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}
	uncompressed, err := snappy.Decode(nil, compressed)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}
	var resp prompb.ReadResponse
	err = proto.Unmarshal(uncompressed, &resp)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response body: %v", err)
	}
	return &resp, nil
}
