// Package influxdb implements the standard library functions
// for interacting with influxdb. It uses the influxdb
// dependency from the dependencies/influxdb package
// to implement the builtins.
package influxdb

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const (
	PackageName    = "influxdata/influxdb"
	FromKind       = "from"
	FromRemoteKind = "influxdata/influxdb.fromRemote"
)

type (
	// NameOrID signifies the name of an organization/bucket
	// or an ID for an organization/bucket.
	NameOrID = influxdb.NameOrID

	// Config contains the common configuration for interacting with an influxdb instance.
	Config = influxdb.Config

	// Predicate defines a predicate to filter storage with.
	Predicate = influxdb.Predicate

	// PredicateSet holds a set of predicates that will filter the results.
	PredicateSet = influxdb.PredicateSet
)

type FromOpSpec struct {
	Org    *NameOrID
	Bucket NameOrID
	Host   *string
	Token  *string
}

func init() {
	fromSignature := runtime.MustLookupBuiltinType("influxdata/influxdb", "from")

	runtime.RegisterPackageValue("influxdata/influxdb", FromKind, flux.MustValue(flux.FunctionValue(FromKind, createFromOpSpec, fromSignature)))
	flux.RegisterOpSpec(FromKind, newFromOp)
	plan.RegisterProcedureSpec(FromKind, newFromProcedure, FromKind)
	execute.RegisterSource(FromRemoteKind, createFromSource)
	plan.RegisterPhysicalRules(
		FromRemoteRule{},
		MergeRemoteRangeRule{},
		MergeRemoteFilterRule{},
	)
}

func createFromOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromOpSpec)

	if b, ok, err := GetNameOrID(args, "bucket", "bucketID"); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New(codes.Invalid, "must specify only one of bucket or bucketID")
	} else {
		spec.Bucket = b
	}

	if o, ok, err := GetNameOrID(args, "org", "orgID"); err != nil {
		return nil, err
	} else if ok {
		spec.Org = &o
	}

	if h, ok, err := args.GetString("host"); err != nil {
		return nil, err
	} else if ok {
		spec.Host = &h
	}

	if token, ok, err := args.GetString("token"); err != nil {
		return nil, err
	} else if ok {
		spec.Token = &token
	}
	return spec, nil
}

func GetNameOrID(args flux.Arguments, nameParam, idParam string) (NameOrID, bool, error) {
	name, nameOk, err := args.GetString(nameParam)
	if err != nil {
		return NameOrID{}, false, err
	}

	id, idOk, err := args.GetString(idParam)
	if err != nil {
		return NameOrID{}, false, err
	}

	if nameOk && idOk {
		return NameOrID{}, false, errors.Newf(codes.Invalid, "must specify one of %s or %s", nameParam, idParam)
	}
	return NameOrID{Name: name, ID: id}, nameOk || idOk, nil
}

func newFromOp() flux.OperationSpec {
	return new(FromOpSpec)
}

func (s *FromOpSpec) Kind() flux.OperationKind {
	return FromKind
}

var _ ProcedureSpec = (*FromProcedureSpec)(nil)

type FromProcedureSpec struct {
	plan.DefaultCost

	Org    *NameOrID
	Bucket NameOrID
	Host   *string
	Token  *string
}

func newFromProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &FromProcedureSpec{
		Org:    spec.Org,
		Bucket: spec.Bucket,
		Host:   spec.Host,
		Token:  spec.Token,
	}, nil
}

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return FromKind
}

func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromProcedureSpec)
	*ns = *s
	return ns
}

func (s *FromProcedureSpec) SetOrg(org *NameOrID)   { s.Org = org }
func (s *FromProcedureSpec) SetHost(host *string)   { s.Host = host }
func (s *FromProcedureSpec) SetToken(token *string) { s.Token = token }
func (s *FromProcedureSpec) GetOrg() *NameOrID      { return s.Org }
func (s *FromProcedureSpec) GetHost() *string       { return s.Host }
func (s *FromProcedureSpec) GetToken() *string      { return s.Token }

func (s *FromProcedureSpec) PostPhysicalValidate(id plan.NodeID) error {
	// This condition should never be met.
	// Customized planner rules within each binary should have
	// filled in either a default host or registered a from procedure
	// for when no host is specified.
	// We mark this as an internal error because it is a programming
	// error if this one ever gets hit.
	if s.Host == nil {
		return errors.New(codes.Internal, "from requires a remote host to be specified")
	}
	return nil
}

type FromRemoteProcedureSpec struct {
	plan.DefaultCost
	influxdb.Config
	Bounds       flux.Bounds
	PredicateSet influxdb.PredicateSet
}

func (s *FromRemoteProcedureSpec) Kind() plan.ProcedureKind {
	return FromRemoteKind
}

// TimeBounds implements plan.BoundsAwareProcedureSpec
func (s *FromRemoteProcedureSpec) TimeBounds(predecessorBounds *plan.Bounds) *plan.Bounds {
	bounds := &plan.Bounds{}

	// set the bounds to the range specified in from call, if there is one
	if !s.Bounds.IsEmpty() {
		bounds.Start = values.ConvertTime(s.Bounds.Start.Time(s.Bounds.Now))
		bounds.Stop = values.ConvertTime(s.Bounds.Stop.Time(s.Bounds.Now))
	}
	if predecessorBounds != nil {
		bounds = bounds.Intersect(predecessorBounds)
	}
	return bounds
}

func (s *FromRemoteProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromRemoteProcedureSpec)
	*ns = *s
	ns.PredicateSet = s.PredicateSet.Copy()
	return ns
}

func (s *FromRemoteProcedureSpec) PostPhysicalValidate(id plan.NodeID) error {
	if s.Bounds.IsEmpty() {
		var bucket string
		if s.Bucket.Name != "" {
			bucket = s.Bucket.Name
		} else {
			bucket = s.Bucket.ID
		}
		return errors.Newf(codes.Invalid, "cannot submit unbounded read to %q; try bounding 'from' with a call to 'range'", bucket)
	}
	return nil
}

func createFromSource(ps plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec := ps.(*FromRemoteProcedureSpec)
	if spec.Bounds.IsEmpty() {
		return nil, errors.Newf(codes.Invalid, "bounds must be set")
	}

	provider := influxdb.GetProvider(a.Context())
	reader, err := provider.ReaderFor(a.Context(), spec.Config, spec.Bounds, spec.PredicateSet)
	if err != nil {
		return nil, err
	}

	itr := &sourceIterator{
		reader: reader,
		mem:    a.Allocator(),
	}
	return execute.CreateSourceFromIterator(itr, id)
}
