package influxdb

import (
	"context"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
)

type key int

const readerKey key = iota

// Dependency will inject the Provider into the dependency chain.
type Dependency struct {
	Provider Provider
}

// Inject will inject the Provider into the dependency chain.
func (d Dependency) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, readerKey, d.Provider)
}

// GetProvider will return the Provider for the current context.
// If no Provider has been injected into the dependencies,
// this will return a default provider.
func GetProvider(ctx context.Context) Provider {
	p := ctx.Value(readerKey)
	if p == nil {
		return HttpProvider{}
	}
	return p.(Provider)
}

// Config contains the common configuration for interacting with an influxdb instance.
type Config struct {
	Org    NameOrID
	Bucket NameOrID
	Host   string
	Token  string
}

// Predicate defines a predicate to filter storage with.
type Predicate struct {
	interpreter.ResolvedFunction

	// KeepEmpty determines if empty tables should be retained
	// if none of the rows pass the filter.
	KeepEmpty bool
}

// Copy produces a deep copy of the Predicate.
func (p *Predicate) Copy() Predicate {
	np := *p
	np.ResolvedFunction = p.ResolvedFunction.Copy()
	return np
}

// PredicateSet holds a set of predicates that will filter the results.
type PredicateSet []Predicate

// Copy produces a deep copy of the PredicateSet.
func (ps PredicateSet) Copy() PredicateSet {
	if ps == nil {
		return nil
	}

	nps := make([]Predicate, len(ps))
	for i := range ps {
		nps[i] = ps[i].Copy()
	}
	return nps
}

// Provider is an interface for creating a Reader that will read
// data from an influxdb instance.
//
// This provides different provider methods depending on the read
// method. The read methods can be expanded so implementors of this
// interface should embed the UnimplementedProvider to automatically
// implement new methods with a default unimplemented stub.
type Provider interface {
	// ReaderFor will construct a Reader using the given configuration parameters.
	// If the parameters are their zero values, appropriate defaults may be used
	// or an error may be returned if the implementation does not have a default.
	ReaderFor(ctx context.Context, conf Config, bounds flux.Bounds, predicateSet PredicateSet) (Reader, error)

	// SeriesCardinalityReaderFor will return a Reader
	// for the SeriesCardinality operation.
	SeriesCardinalityReaderFor(ctx context.Context, conf Config, bounds flux.Bounds, predicateSet PredicateSet) (Reader, error)
}

// Reader reads tables from an influxdb instance.
type Reader interface {
	// Read will produce flux.Table values using the memory.Allocator
	// and it will pass those tables to the given function.
	Read(ctx context.Context, f func(flux.Table) error, mem memory.Allocator) error
}

// UnimplementedProvider provides default implementations for a Provider.
// This implements all of the Provider methods by returning an error
// with the code codes.Unimplemented.
type UnimplementedProvider struct{}

var _ Provider = UnimplementedProvider{}

func (u UnimplementedProvider) ReaderFor(ctx context.Context, conf Config, bounds flux.Bounds, predicateSet PredicateSet) (Reader, error) {
	return nil, errors.New(codes.Unimplemented, "influxdb reader has not been implemented")
}

func (u UnimplementedProvider) SeriesCardinalityReaderFor(ctx context.Context, conf Config, bounds flux.Bounds, predicateSet PredicateSet) (Reader, error) {
	return nil, errors.New(codes.Unimplemented, "influxdb series cardinality reader has not been implemented")
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
