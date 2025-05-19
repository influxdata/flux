package iox

import (
	"context"

	"github.com/apache/arrow/go/v7/arrow/array"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
)

type key int

const clientKey key = iota

type Config = influxdb.Config

// Dependency holds the iox.Dependency to be injected.
type Dependency struct {
	Provider Provider
}

// Inject will inject the iox dependency into the dependency chain.
func (d Dependency) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, clientKey, d.Provider)
}

// Provider provides access to a Client with the given configuration.
type Provider interface {
	// ClientFor will return a client with the given configuration.
	ClientFor(ctx context.Context, conf Config) (Client, error)
}

// GetProvider retrieves the iox Provider.
func GetProvider(ctx context.Context) Provider {
	p := ctx.Value(clientKey)
	if p == nil {
		return ErrorProvider{}
	}
	return p.(Provider)
}

// ColumnType defines the column data types IOx can represent.
type ColumnType int32

const (
	// ColumnTypeUnknown is an invalid column type.
	ColumnTypeUnknown ColumnType = 0
	// ColumnType_I64 is an int64.
	ColumnType_I64 ColumnType = 1
	// ColumnType_U64 is an uint64.
	ColumnType_U64 ColumnType = 2
	// ColumnType_F64 is an float64.
	ColumnType_F64 ColumnType = 3
	// ColumnType_BOOL is a bool.
	ColumnType_BOOL ColumnType = 4
	// ColumnType_STRING is a string.
	ColumnType_STRING ColumnType = 5
	// ColumnType_TIME is a timestamp.
	ColumnType_TIME ColumnType = 6
	// ColumnType_TAG is a tag value.
	ColumnType_TAG ColumnType = 7
)

func (c ColumnType) String() string {
	switch c {
	case ColumnType_I64:
		return "int64"
	case ColumnType_U64:
		return "uint64"
	case ColumnType_F64:
		return "float64"
	case ColumnType_BOOL:
		return "bool"
	case ColumnType_STRING:
		return "string"
	case ColumnType_TIME:
		return "timestamp"
	case ColumnType_TAG:
		return "tag"
	default:
		return "unknown"
	}
}

// RecordReader is similar to the RecordReader interface provided by Arrow's array
// package, but includes a method for detecting errors that are sent mid-stream.
type RecordReader interface {
	array.RecordReader
	Err() error
}

// Client provides a way to query an iox instance.
type Client interface {
	// Query will initiate a query using the given query string, parameters, and memory allocator
	// against the iox instance. It returns an array.RecordReader from the arrow flight api.
	Query(ctx context.Context, query string, params []interface{}, mem memory.Allocator) (RecordReader, error)

	// GetSchema will retrieve a schema for the given table if this client supports that capability.
	// If this Client doesn't support this capability, it should return a flux error with the code
	// codes.Unimplemented.
	GetSchema(ctx context.Context, table string) (map[string]ColumnType, error)
}

// ErrorProvider is an implementation of the Provider that returns an error.
type ErrorProvider struct{}

func (u ErrorProvider) ClientFor(ctx context.Context, conf Config) (Client, error) {
	return nil, errors.New(codes.Invalid, "iox client has not been configured")
}
