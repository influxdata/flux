package internal

// TODO: Remove this when https://github.com/influxdata/influxdb-client-go/pull/256 is merged

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	http2 "github.com/influxdata/influxdb-client-go/v2/api/http"
)

type RequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// service implements Service interface
type service struct {
	serverAPIURL  string
	serverURL     string
	authorization string
	client        RequestDoer
}

// NewService creates instance of http Service with given parameters
func NewService(serverURL, authorization string, roundTripper RequestDoer) http2.Service {
	apiURL, err := url.Parse(serverURL)
	serverAPIURL := serverURL
	if err == nil {
		apiURL, err = apiURL.Parse("api/v2/")
		if err == nil {
			serverAPIURL = apiURL.String()
		}
	}
	return &service{
		serverAPIURL:  serverAPIURL,
		serverURL:     serverURL,
		authorization: authorization,
		client:        roundTripper,
	}
}

func (s *service) ServerAPIURL() string {
	return s.serverAPIURL
}

func (s *service) ServerURL() string {
	return s.serverURL
}

func (s *service) SetAuthorization(authorization string) {
	s.authorization = authorization
}

func (s *service) Authorization() string {
	return s.authorization
}

func (s *service) DoPostRequest(ctx context.Context, url string, body io.Reader, requestCallback http2.RequestCallback, responseCallback http2.ResponseCallback) *http2.Error {
	return s.doHTTPRequestWithURL(ctx, http.MethodPost, url, body, requestCallback, responseCallback)
}

func (s *service) doHTTPRequestWithURL(ctx context.Context, method, url string, body io.Reader, requestCallback http2.RequestCallback, responseCallback http2.ResponseCallback) *http2.Error {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return http2.NewError(err)
	}
	return s.DoHTTPRequest(req, requestCallback, responseCallback)
}

func (s *service) DoHTTPRequest(req *http.Request, requestCallback http2.RequestCallback, responseCallback http2.ResponseCallback) *http2.Error {
	resp, err := s.DoHTTPRequestWithResponse(req, requestCallback)
	if err != nil {
		return http2.NewError(err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return s.parseHTTPError(resp)
	}
	if responseCallback != nil {
		err := responseCallback(resp)
		if err != nil {
			return http2.NewError(err)
		}
	}
	return nil
}

func (s *service) DoHTTPRequestWithResponse(req *http.Request, requestCallback http2.RequestCallback) (*http.Response, error) {
	if len(s.authorization) > 0 {
		req.Header.Set("Authorization", s.authorization)
	}
	if requestCallback != nil {
		requestCallback(req)
	}
	return s.client.Do(req)
}

func (s *service) parseHTTPError(r *http.Response) *http2.Error {
	// successful status code range
	if r.StatusCode >= 200 && r.StatusCode < 300 {
		return nil
	}
	defer func() {
		// discard body so connection can be reused
		_, _ = io.Copy(ioutil.Discard, r.Body)
		_ = r.Body.Close()
	}()

	perror := http2.NewError(nil)
	perror.StatusCode = r.StatusCode

	if v := r.Header.Get("Retry-After"); v != "" {
		r, err := strconv.ParseUint(v, 10, 32)
		if err == nil {
			perror.RetryAfter = uint(r)
		}
	}

	// json encoded error
	ctype, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ctype == "application/json" {
		perror.Err = json.NewDecoder(r.Body).Decode(perror)
	} else {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			perror.Err = err
			return perror
		}

		perror.Code = r.Status
		perror.Message = string(body)
	}

	if perror.Code == "" && perror.Message == "" {
		switch r.StatusCode {
		case http.StatusTooManyRequests:
			perror.Code = "too many requests"
			perror.Message = "exceeded rate limit"
		case http.StatusServiceUnavailable:
			perror.Code = "unavailable"
			perror.Message = "service temporarily unavailable"
		default:
			perror.Code = r.Status
			perror.Message = r.Header.Get("X-Influxdb-Error")
		}
	}

	return perror
}
