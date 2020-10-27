package influxdb

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/influxdb"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const (
	CardinalityFuncName = "cardinality"
	CardinalityKind     = PackageName + "." + CardinalityFuncName
)

func init() {
	cardinalitySignature := runtime.MustLookupBuiltinType(PackageName, CardinalityFuncName)

	runtime.RegisterPackageValue(PackageName, CardinalityFuncName, flux.MustValue(flux.FunctionValue(CardinalityFuncName, createCardinalityOpSpec, cardinalitySignature)))
	plan.RegisterProcedureSpec(CardinalityKind, newCardinalityProcedure, CardinalityKind)
	execute.RegisterSource(CardinalityKind, createCardinalitySource)
}

func createCardinalityOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(CardinalityOpSpec)

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
		spec.Org = o
	}

	if h, ok, err := args.GetString("host"); err != nil {
		return nil, err
	} else if ok {
		spec.Host = h
	}

	if token, ok, err := args.GetString("token"); err != nil {
		return nil, err
	} else if ok {
		spec.Token = token
	}

	if start, err := args.GetRequiredTime("start"); err != nil {
		return nil, err
	} else {
		spec.Start = start
	}

	if stop, ok, err := args.GetTime("stop"); err != nil {
		return nil, err
	} else if ok {
		spec.Stop = stop
	} else {
		spec.Stop = flux.Now
	}

	if fn, ok, err := args.GetFunction("predicate"); err != nil {
		return nil, err
	} else if ok {
		predicate, err := interpreter.ResolveFunction(fn)
		if err != nil {
			return nil, err
		}
		spec.Predicate = influxdb.Predicate{
			ResolvedFunction: predicate,
		}
	}
	return spec, nil
}

type CardinalityOpSpec struct {
	influxdb.Config
	Start     flux.Time
	Stop      flux.Time
	Predicate influxdb.Predicate
}

func (s *CardinalityOpSpec) Kind() flux.OperationKind {
	return CardinalityKind
}

type CardinalityProcedureSpec struct {
	plan.DefaultCost
	influxdb.Config
	Bounds       flux.Bounds
	PredicateSet influxdb.PredicateSet
}

func newCardinalityProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*CardinalityOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	var predicateSet influxdb.PredicateSet
	if spec.Predicate.Fn != nil {
		predicateSet = influxdb.PredicateSet{spec.Predicate}
	}
	return &CardinalityProcedureSpec{
		Config: spec.Config,
		Bounds: flux.Bounds{
			Start: spec.Start,
			Stop:  spec.Stop,
			Now:   pa.Now(),
		},
		PredicateSet: predicateSet,
	}, nil
}

func (s *CardinalityProcedureSpec) Kind() plan.ProcedureKind {
	return CardinalityKind
}

func (s *CardinalityProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(CardinalityProcedureSpec)
	*ns = *s
	ns.PredicateSet = s.PredicateSet.Copy()
	return ns
}

// TimeBounds implements plan.BoundsAwareProcedureSpec
func (s *CardinalityProcedureSpec) TimeBounds(predecessorBounds *plan.Bounds) *plan.Bounds {
	bounds := &plan.Bounds{
		Start: values.ConvertTime(s.Bounds.Start.Time(s.Bounds.Now)),
		Stop:  values.ConvertTime(s.Bounds.Stop.Time(s.Bounds.Now)),
	}
	if predecessorBounds != nil {
		bounds = bounds.Intersect(predecessorBounds)
	}
	return bounds
}

func createCardinalitySource(ps plan.ProcedureSpec, id execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec := ps.(*CardinalityProcedureSpec)
	provider := influxdb.GetProvider(a.Context())

	reader, err := provider.SeriesCardinalityReaderFor(a.Context(), spec.Config, spec.Bounds, spec.PredicateSet)
	if err != nil {
		return nil, err
	}

	itr := &sourceIterator{
		reader: reader,
		mem:    a.Allocator(),
	}
	return execute.CreateSourceFromIterator(itr, id)
}
