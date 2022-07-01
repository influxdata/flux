package bigtable

import (
	"context"
	"net"
	"time"

	"cloud.google.com/go/bigtable"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

type key int

const providerKey key = iota

// Inject will inject this Provider into the dependency chain.
func Inject(ctx context.Context, provider Provider) context.Context {
	return context.WithValue(ctx, providerKey, provider)
}

// Dependency will inject the Provider into the dependency chain.
type Dependency struct {
	Provider Provider
}

// Inject will inject the Dialer into the dependency chain.
func (d Dependency) Inject(ctx context.Context) context.Context {
	return Inject(ctx, d.Provider)
}

// GetProvider will return the Provider for the current context.
// If no Provider has been injected into the dependencies,
// this will return a default provider.
func GetProvider(ctx context.Context) Provider {
	p := ctx.Value(providerKey)
	if p == nil {
		return ErrorProvider{}
	}
	return p.(Provider)
}

// Provider provides a method to create a new bigtable.Client.
type Provider interface {
	NewClient(ctx context.Context, project, instance string, opts ...option.ClientOption) (*bigtable.Client, error)
}

// DefaultProvider is the default provider that uses the default bigtable client.
type DefaultProvider struct{}

func (d DefaultProvider) NewClient(ctx context.Context, project, instance string, opts ...option.ClientOption) (*bigtable.Client, error) {
	opts = append([]option.ClientOption{
		option.WithGRPCDialOption(grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
			dialer, err := flux.GetDialer(ctx)
			if err != nil {
				return nil, err
			}
			return dialer.DialContext(ctx, "tcp", address)
		})),
	}, opts...)
	return bigtable.NewClient(ctx, project, instance, opts...)
}

// DefaultProvider is the default provider that uses the default bigtable client.
type ErrorProvider struct{}

func (ErrorProvider) NewClient(ctx context.Context, project, instance string, opts ...option.ClientOption) (*bigtable.Client, error) {
	return nil, errors.New(codes.Invalid, "Provider.NewClient called on an error dependency")
}

// Forwarding types and functions for convenience.

type (
	Client     = bigtable.Client
	Table      = bigtable.Table
	RowSet     = bigtable.RowSet
	RowRange   = bigtable.RowRange
	Row        = bigtable.Row
	Filter     = bigtable.Filter
	ReadOption = bigtable.ReadOption
	ReadItem   = bigtable.ReadItem
)

func RowFilter(f Filter) ReadOption {
	return bigtable.RowFilter(f)
}

func InfiniteRange(start string) RowRange {
	return bigtable.InfiniteRange(start)
}

func PassAllFilter() Filter {
	return bigtable.PassAllFilter()
}

func SingleRow(row string) RowSet {
	return bigtable.SingleRow(row)
}

func ChainFilters(sub ...Filter) Filter {
	return bigtable.ChainFilters(sub...)
}

func FamilyFilter(pattern string) Filter {
	return bigtable.FamilyFilter(pattern)
}

func NewRange(begin, end string) RowRange {
	return bigtable.NewRange(begin, end)
}

func TimestampRangeFilter(startTime, endTime time.Time) Filter {
	return bigtable.TimestampRangeFilter(startTime, endTime)
}

func PrefixRange(prefix string) RowRange {
	return bigtable.PrefixRange(prefix)
}

func LimitRows(limit int64) ReadOption {
	return bigtable.LimitRows(limit)
}
