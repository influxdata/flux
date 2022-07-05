package influxdb_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/dependency"
	influxdb2 "github.com/influxdata/flux/stdlib/influxdata/influxdb"
)

type RoundTrip struct {
	RequestValidator      func(_ *http.Request) error
	RequestValidatorError error
	HandlerFn             func(req *http.Request) (*http.Response, error)

	Bodies bytes.Buffer
}

func (f *RoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	if f.RequestValidatorError == nil {
		f.RequestValidatorError = f.RequestValidator(req)
	}
	_, err := io.Copy(&f.Bodies, req.Body)
	if err != nil {
		panic(fmt.Sprintf("Error while copying request body: %s", err))
	}

	if f.HandlerFn != nil {
		return f.HandlerFn(req)
	}

	return &http.Response{
		StatusCode: 200,
		Status:     "Body generated by test client",

		// Send response to be tested
		Body: ioutil.NopCloser(new(bytes.Buffer)),

		// Must be set to non-nil value or it panics
		Header: make(http.Header),
	}, nil
}

func cpuMetric(usage float64, ns int) influxdb.Metric {
	tm := time.Date(2017, 11, 17, 0, 0, 0, ns, time.UTC)
	return &influxdb2.RowMetric{
		NameStr: "cpu",
		Tags: []*influxdb.Tag{
			{Key: "host", Value: "localhost"},
			{Key: "id", Value: "cpua"},
		},
		Fields: []*influxdb.Field{
			{Key: "usage_user", Value: usage},
			{Key: "log", Value: "message"},
		},
		TS: tm,
	}
}

func diskMetric(usage float64, ns int) influxdb.Metric {
	tm := time.Date(2017, 11, 17, 0, 0, 0, ns, time.UTC)
	return &influxdb2.RowMetric{
		NameStr: "disk",
		Tags: []*influxdb.Tag{
			{Key: "id", Value: "/dev/sdb"},
		},
		Fields: []*influxdb.Field{
			{Key: "usage_disk", Value: usage},
			{Key: "log", Value: "disk message"},
		},
		TS: tm,
	}
}

func TestHttpWriter_Write(t *testing.T) {
	tests := []struct {
		name     string
		metric   [][]influxdb.Metric
		wantBody string
		wantErr  bool
	}{
		{
			name: "basic",
			metric: [][]influxdb.Metric{
				{
					cpuMetric(95, 1),
					cpuMetric(96, 2),
					cpuMetric(97, 3),
					cpuMetric(95, 4),
					// should be skipped
					cpuMetric(math.Inf(1), 5),
					cpuMetric(math.Inf(-1), 6),
				},
				{
					diskMetric(45, 1),
					diskMetric(46, 2),
					diskMetric(47, 3),
					diskMetric(45, 4),
					// should be skipped
					diskMetric(math.NaN(), 5),
				},
				{
					&influxdb2.RowMetric{
						NameStr: "skipped",
						Fields: []*influxdb.Field{
							{Key: "value", Value: math.NaN()},
						},
					},
				},
			},
			wantBody: `cpu,host=localhost,id=cpua usage_user=95,log="message" 1510876800000000001
cpu,host=localhost,id=cpua usage_user=96,log="message" 1510876800000000002
cpu,host=localhost,id=cpua usage_user=97,log="message" 1510876800000000003
cpu,host=localhost,id=cpua usage_user=95,log="message" 1510876800000000004
cpu,host=localhost,id=cpua log="message" 1510876800000000005
cpu,host=localhost,id=cpua log="message" 1510876800000000006
disk,id=/dev/sdb usage_disk=45,log="disk message" 1510876800000000001
disk,id=/dev/sdb usage_disk=46,log="disk message" 1510876800000000002
disk,id=/dev/sdb usage_disk=47,log="disk message" 1510876800000000003
disk,id=/dev/sdb usage_disk=45,log="disk message" 1510876800000000004
disk,id=/dev/sdb log="disk message" 1510876800000000005
`,
		},
		{
			name: "invalid empty field key",
			metric: [][]influxdb.Metric{
				{
					&influxdb2.RowMetric{
						NameStr: "cpu",
						Fields: []*influxdb.Field{
							{
								// Empty field key is invalid.
								Key:   "",
								Value: int64(1),
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid escape field key",
			metric: [][]influxdb.Metric{
				{
					&influxdb2.RowMetric{
						NameStr: "cpu",
						Fields: []*influxdb.Field{
							{
								// Field key with escape at the end is invalid.
								Key:   "invalid\\",
								Value: int64(1),
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := influxdb.HttpProvider{
				DefaultConfig: influxdb.Config{
					Host:  "http://myhost.com:8085",
					Token: "mytoken",
				},
			}
			deps := dependenciestest.Default()
			roundTripper := &RoundTrip{
				RequestValidator: func(req *http.Request) error {
					url := req.URL
					values := url.Query()
					if val, exp := req.Header.Get("Authorization"), "Token mytoken"; val != exp {
						return fmt.Errorf("token does not match, expected %s, got %s", exp, val)
					}
					if val, exp := values.Get("bucket"), "mybucket"; val != exp {
						return fmt.Errorf("bucket does not match, expected %s, got %s", exp, val)
					}
					if val, exp := values.Get("org"), "myorg"; val != exp {
						return fmt.Errorf("org does not match, expected %s, got %s", exp, val)
					}
					if val, exp := url.Host, "myhost.com:8085"; val != exp {
						return fmt.Errorf("host does not match, expected %s, got %s", exp, val)
					}
					if val, exp := url.Path, "/api/v2/write"; val != exp {
						return fmt.Errorf("path does not match, expected %s, got %s", exp, val)
					}
					return nil
				},
			}
			deps.Deps.Deps.HTTPClient = &http.Client{
				Transport: roundTripper,
			}
			ctx, span := dependency.Inject(context.Background(), deps)
			defer span.Finish()
			writer, err := h.WriterFor(ctx, influxdb.Config{
				Org:    influxdb.NameOrID{Name: "myorg"},
				Bucket: influxdb.NameOrID{Name: "mybucket"},
			})
			if err != nil {
				t.Errorf("WriterFor() error = %v", err)
			}
			for i := range tt.metric {
				if err := writer.Write(tt.metric[i]...); err != nil {
					if tt.wantErr {
						return
					}
					t.Errorf("Write() error = %v", err)
				}
			}
			writer.Close()
			if roundTripper.RequestValidatorError != nil {
				t.Errorf("Query validation error = %v", roundTripper.RequestValidatorError)
			}
			if roundTripper.Bodies.String() != tt.wantBody {
				t.Error(cmp.Diff(tt.wantBody, roundTripper.Bodies.String()))
			}
			if tt.wantErr {
				t.Error("expected error but none occurred")
			}
		})
	}
}

func TestHttpWriter_Write_Error(t *testing.T) {
	h := influxdb.HttpProvider{
		DefaultConfig: influxdb.Config{
			Host:  "http://myhost.com:8085",
			Token: "mytoken",
		},
	}
	deps := dependenciestest.Default()
	roundTripper := &RoundTrip{
		RequestValidator: func(req *http.Request) error {
			return nil
		},
		HandlerFn: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Status:     http.StatusText(http.StatusTooManyRequests),

				// Send response to be tested
				Body: ioutil.NopCloser(strings.NewReader(`{"code":"too many requests","message":"write limit reached"}`)),

				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
			}, nil
		},
	}
	deps.Deps.Deps.HTTPClient = &http.Client{
		Transport: roundTripper,
	}
	ctx, span := dependency.Inject(context.Background(), deps)
	defer span.Finish()
	writer, err := h.WriterFor(ctx, influxdb.Config{
		Org:    influxdb.NameOrID{Name: "myorg"},
		Bucket: influxdb.NameOrID{Name: "mybucket"},
	})
	if err != nil {
		t.Errorf("WriterFor() error = %v", err)
	}

	// We're going to write a metric. It's not really guaranteed
	// when we will receive the error and we probably won't receive
	// an error when only writing one metric, but we'll check anyway.
	//
	// An error on Write or Close is what we are looking for.
	metrics := []influxdb.Metric{
		cpuMetric(95, 1),
	}
	err = writer.Write(metrics...)
	if closeErr := writer.Close(); err == nil {
		err = closeErr
	}

	if err == nil {
		t.Error("expected error, got nil")
	} else if ferr, ok := err.(*flux.Error); !ok {
		t.Errorf("expected flux error, but got a non-flux error: %v", err)
	} else {
		if want, got := codes.ResourceExhausted, ferr.Code; want != got {
			t.Errorf("unexpected code -want/+got:\n\t- %s\n\t+ %s", want, got)
		}
		if want, got := "write limit reached", ferr.Msg; want != got {
			t.Errorf("unexpected message -want/+got:\n\t- %s\n\t+ %s", want, got)
		}
	}
}
