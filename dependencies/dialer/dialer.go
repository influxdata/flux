package dialer

import (
	"context"
	"net"
	"strings"
	"syscall"
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/internal/errors"
)

// Create a new *net.Dialer that uses the provided url.Validator to
// validate the destination OP address before connecting.
func New(urlValidator url.Validator) *net.Dialer {
	// ControlContext is called after DNS lookup, but before the network
	// connection is initiated.
	controlContext := func(ctx context.Context, network, address string, c syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}
		// Remove any zone from the host.
		host, _, _ = strings.Cut(host, "%")
		ip := net.ParseIP(host)
		if ip == nil {
			return errors.New(codes.Invalid, "no such host")
		}
		return urlValidator.ValidateIP(ip)
	}

	return &net.Dialer{
		Timeout:        30 * time.Second,
		KeepAlive:      30 * time.Second,
		ControlContext: controlContext,
	}
}
