// Package influxdb exposes interfaces an implementations for accessing
// influxdb through the Dependency interface.
package influxdb

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	stderrors "errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	httpdep "github.com/influxdata/flux/dependencies/http"
	"github.com/influxdata/flux/internal/errors"
	lp "github.com/influxdata/line-protocol"
)

type key int

const pointsWriterKey key = iota

// PointsWriterDependency will inject the PointsWriterProvider
// into the dependency chain.
type PointsWriterDependency struct {
	Provider PointsWriterProvider
}

// Inject will inject the PointsWriterProvider into the dependency chain.
func (d PointsWriterDependency) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, pointsWriterKey, d.Provider)
}

// PointsWriterProvider is an interface for creating a PointsWriter
// that will write points to the given location.
type PointsWriterProvider interface {
	// WriterFor will construct a PointsWriter using the given
	// parameters. If the parameters are their zero values,
	// appropriate defaults may be used or an error may be
	// returned if the implementation does not have a default.
	WriterFor(ctx context.Context, org, bucket NameOrID, host, token string) (PointsWriter, error)
}

// PointsWriter is an interface for writing points to the influxdb
// storage engine. The point may be buffered and not immediately written.
type PointsWriter interface {
	// WritePoint will encode the point. This function may or may
	// not trigger a write.
	WritePoint(lp.Metric) error

	// Close will flush any pending writes and close any active
	// connections.
	Close() error
}

// GetPointsWriter will return the configured PointsWriterProvider.
// If one hasn't been configured, this function will attempt to
// create one using the configured HTTPClient.
func GetPointsWriter(ctx context.Context) PointsWriterProvider {
	pw := ctx.Value(pointsWriterKey)
	if pw == nil {
		return DefaultWriterProvider
	}
	return pw.(PointsWriterProvider)
}

// DefaultWriterProvider exposes the default PointsWriterProvider
// for systems that aren't configured with a specific implementation.
//
// The default will write points using the public API. It requires
// an organization, bucket, host, and token.
var DefaultWriterProvider PointsWriterProvider = httpPointsWriterProvider{}

type httpPointsWriterProvider struct{}

func (p httpPointsWriterProvider) WriterFor(ctx context.Context, org, bucket NameOrID, host, token string) (PointsWriter, error) {
	deps := flux.GetDependencies(ctx)
	httpc, err := deps.HTTPClient()
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrap(err, codes.Invalid, "url parse")
	}

	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	u.Path += "api/v2/write"

	params := url.Values{}
	if org.ID != "" {
		params.Set("orgID", org.ID)
	} else {
		params.Set("org", org.Name)
	}

	if bucket.ID != "" {
		params.Set("bucketID", bucket.ID)
	} else {
		params.Set("bucket", bucket.Name)
	}
	params.Set("precision", "ns")
	u.RawQuery = params.Encode()
	return &httpPointsWriter{
		c:     httpc,
		url:   u.String(),
		token: token,
	}, nil
}

type httpPointsWriter struct {
	c      httpdep.Client
	url    string
	token  string
	buffer bytes.Buffer
	w      *gzip.Writer
	enc    *lp.Encoder
	n      int
}

const maxPointsSize = 5000

func (h *httpPointsWriter) WritePoint(m lp.Metric) error {
	if h.n == 0 {
		h.w = gzip.NewWriter(&h.buffer)
		h.enc = lp.NewEncoder(h.w)
	}
	if _, err := h.enc.Encode(m); err != nil {
		return err
	}
	h.n++

	if h.n >= maxPointsSize {
		if err := h.writeBuffer(); err != nil {
			return err
		}
	}
	return nil
}

func (h *httpPointsWriter) Close() error {
	return h.writeBuffer()
}

func (h *httpPointsWriter) writeBuffer() error {
	if h.w == nil {
		return nil
	}
	if err := h.w.Close(); err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", h.url, &h.buffer)
	req.Header.Set("Authorization", "Token "+h.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "text/plan; charset=utf-8")

	resp, err := h.c.Do(req)
	if err != nil {
		return err
	}

	if err := h.checkError(resp); err != nil {
		return err
	}
	h.buffer.Reset()
	h.n = 0
	return nil
}

func (h *httpPointsWriter) checkError(resp *http.Response) error {
	switch resp.StatusCode / 100 {
	case 4, 5:
		// We will attempt to parse this error outside of this block.
	case 2:
		return nil
	default:
		// TODO(jsternberg): Figure out what to do here?
		return errors.Newf(codes.Invalid, "unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	if resp.StatusCode == http.StatusUnsupportedMediaType {
		return errors.Newf(codes.Invalid, "invalid media type: %q", resp.Header.Get("Content-Type"))
	}

	var perr struct {
		Msg  string `json:"message"`
		Code string `json:"code"`
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// Assume JSON if there is no content-type.
		contentType = "application/json"
	}
	mediatype, _, _ := mime.ParseMediaType(contentType)

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return errors.New(codes.Invalid, "failed to read error response")
	}

	switch mediatype {
	case "application/json":
		if err := json.Unmarshal(buf.Bytes(), &perr); err != nil {
			err := stderrors.New(firstLineAsError(buf))
			return errors.Wrapf(err, codes.Invalid, "attempted to unmarshal error as JSON but failed: %q", err)
		}
	default:
		perr.Msg = firstLineAsError(buf)
	}

	var code codes.Code
	switch resp.StatusCode {
	case http.StatusBadRequest:
		code = codes.Invalid
	case http.StatusUnauthorized:
		code = codes.Unauthenticated
	case http.StatusForbidden:
		code = codes.PermissionDenied
	case http.StatusRequestEntityTooLarge, http.StatusTooManyRequests:
		code = codes.ResourceExhausted
	case http.StatusServiceUnavailable:
		code = codes.Unavailable
	default:
		code = codes.Unknown
	}

	switch perr.Code {
	case "internal error":
		code = codes.Internal
	case "not found":
		code = codes.NotFound
	case "conflict":
		code = codes.FailedPrecondition
	case "invalid", "empty value", "unprocessable entity", "method not allowed":
		code = codes.Invalid
	case "unavailable":
		code = codes.Unavailable
	case "forbidden":
		code = codes.PermissionDenied
	case "too many requests":
		code = codes.ResourceExhausted
	case "unauthorized":
		code = codes.Unauthenticated
	}
	return errors.New(code, perr.Msg)
}

func firstLineAsError(buf bytes.Buffer) string {
	line, _ := buf.ReadString('\n')
	return strings.TrimSuffix(line, "\n")
}

// NameOrID signifies the name of an organization/bucket
// or an ID for an organization/bucket.
type NameOrID struct {
	ID   string
	Name string
}

// IsValid will return true if both the name and the id are not
// set at the same time.
func (n NameOrID) IsValid() bool {
	return (n.ID != "" && n.Name == "") || (n.ID == "" && n.Name != "")
}

// IsZero will return true if neither the id nor name are set.
func (n NameOrID) IsZero() bool {
	return n.ID == "" && n.Name == ""
}
