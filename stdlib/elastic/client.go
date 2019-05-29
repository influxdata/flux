package elastic

import (
	"bytes"
	"encoding/json"
	"github.com/influxdata/flux/values"
	"net/http"
	"net/url"
)

const DefaultURL = "http://127.0.0.1:9200"
const DefaultTimeout = 30

type Client struct {
	URL      *url.URL
	User     string
	Password string
}

func (c *Client) hasAuth() bool {
	return c.User != ""
}

func NewClient(serverAddr, user, password string) (*Client, error) {
	if serverAddr == "" {
		serverAddr = DefaultURL;
	}
	serverUrl, err := url.Parse(serverAddr)
	if err != nil {
		return nil, err
	}
	return &Client{
		URL:      serverUrl,
		User:     user,
		Password: password,
	}, nil
}

func (c *Client) Query(query values.Object) (*map[string]interface{}, error) {
	u, err := url.Parse("./_search")
	if err != nil {
		return nil, err
	}
	u = c.URL.ResolveReference(u)

	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", u.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if c.hasAuth() {
		req.SetBasicAuth(c.User, c.Password)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req);
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}
	return result, nil

}

func (c *Client) Ping() error {
	//	TODO
	return nil
}
