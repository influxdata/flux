package inputs

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/functions"
	"github.com/influxdata/flux/functions/inputs/storage"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/platform"
	"github.com/influxdata/platform/query"
	"github.com/pkg/errors"
)

const FromKind = "from"

type FromOpSpec struct {
	Bucket   string      `json:"bucket,omitempty"`
	BucketID platform.ID `json:"bucketID,omitempty"`
}

var fromSignature = semantic.FunctionSignature{
	Params: map[string]semantic.Type{
		"bucket":   semantic.String,
		"bucketID": semantic.String,
	},
	ReturnType: flux.TableObjectType,
}

func init() {
	flux.RegisterFunction(FromKind, createFromOpSpec, fromSignature)
	flux.RegisterOpSpec(FromKind, newFromOp)
	plan.RegisterProcedureSpec(FromKind, newFromProcedure, FromKind)
	plan.RegisterPhysicalRule(MergeFromRangeRule{})
	execute.RegisterSource(FromKind, createFromSource)
}

func createFromOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromOpSpec)

	if bucket, ok, err := args.GetString("bucket"); err != nil {
		return nil, err
	} else if ok {
		spec.Bucket = bucket
	}

	if bucketID, ok, err := args.GetString("bucketID"); err != nil {
		return nil, err
	} else if ok {
		err := spec.BucketID.DecodeFromString(bucketID)
		if err != nil {
			return nil, errors.Wrap(err, "invalid bucket ID")
		}
	}

	if spec.Bucket == "" && len(spec.BucketID) == 0 {
		return nil, errors.New("must specify one of bucket or bucketID")
	}
	if spec.Bucket != "" && len(spec.BucketID) != 0 {
		return nil, errors.New("must specify only one of bucket or bucketID")
	}
	return spec, nil
}

func newFromOp() flux.OperationSpec {
	return new(FromOpSpec)
}

func (s *FromOpSpec) Kind() flux.OperationKind {
	return FromKind
}

func (s *FromOpSpec) BucketsAccessed() (readBuckets, writeBuckets []platform.BucketFilter) {
	bf := platform.BucketFilter{}
	if s.Bucket != "" {
		bf.Name = &s.Bucket
	}

	if len(s.BucketID) > 0 {
		bf.ID = &s.BucketID
	}

	if bf.ID != nil || bf.Name != nil {
		readBuckets = append(readBuckets, bf)
	}
	return readBuckets, writeBuckets
}

type FromProcedureSpec struct {
	plan.DefaultCost
	Bucket   string
	BucketID platform.ID

	BoundsSet bool
	Bounds    flux.Bounds

	FilterSet bool
	Filter    *semantic.FunctionExpression

	DescendingSet bool
	Descending    bool

	LimitSet     bool
	PointsLimit  int64
	SeriesLimit  int64
	SeriesOffset int64

	WindowSet bool
	Window    plan.WindowSpec

	GroupingSet bool
	OrderByTime bool
	GroupMode   functions.GroupMode
	GroupKeys   []string

	AggregateSet    bool
	AggregateMethod string
}

// MergeFromRangeRule pushes a `range` into a `from`
type MergeFromRangeRule struct{}

// Name returns the name of the rule
func (rule MergeFromRangeRule) Name() string {
	return "MergeFromRangeRule"
}

// Pattern returns the pattern that matches `from -> range`
func (rule MergeFromRangeRule) Pattern() plan.Pattern {
	return plan.Pat(transformations.RangeKind, plan.Pat(FromKind))
}

// Rewrite attempts to rewrite a `from -> range` into a `FromRange`
func (rule MergeFromRangeRule) Rewrite(node plan.PlanNode) (plan.PlanNode, bool) {
	from := node.Predecessors()[0]
	fromSpec := from.ProcedureSpec().(*FromProcedureSpec)
	rangeSpec := node.ProcedureSpec().(*transformations.RangeProcedureSpec)
	fromRange := fromSpec.Copy().(*FromProcedureSpec)

	// Set new bounds to `range` bounds initially
	fromRange.Bounds = rangeSpec.Bounds

	var (
		now   = rangeSpec.Bounds.Now
		start = rangeSpec.Bounds.Start
		stop  = rangeSpec.Bounds.Stop
	)

	bounds := &plan.Bounds{
		Start: values.ConvertTime(start.Time(now)),
		Stop:  values.ConvertTime(stop.Time(now)),
	}

	// Intersect bounds if `from` already bounded
	if fromSpec.BoundsSet {
		now = fromSpec.Bounds.Now
		start = fromSpec.Bounds.Start
		stop = fromSpec.Bounds.Stop

		fromBounds := &plan.Bounds{
			Start: values.ConvertTime(start.Time(now)),
			Stop:  values.ConvertTime(stop.Time(now)),
		}

		bounds = bounds.Intersect(fromBounds)
		fromRange.Bounds = flux.Bounds{
			Start: flux.Time{Absolute: bounds.Start.Time()},
			Stop:  flux.Time{Absolute: bounds.Stop.Time()},
		}
	}

	fromRange.BoundsSet = true

	// Finally merge nodes into single operation
	merged, err := plan.MergePhysicalPlanNodes(node, from, fromRange)
	if err != nil {
		return node, false
	}
	return merged, true
}

func newFromProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FromProcedureSpec{
		Bucket:   spec.Bucket,
		BucketID: spec.BucketID,
	}, nil
}

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return FromKind
}

func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromProcedureSpec)

	ns.Bucket = s.Bucket
	if len(s.BucketID) > 0 {
		ns.BucketID = make(platform.ID, len(s.BucketID))
		copy(ns.BucketID, s.BucketID)
	}

	ns.BoundsSet = s.BoundsSet
	ns.Bounds = s.Bounds

	ns.FilterSet = s.FilterSet
	ns.Filter = s.Filter.Copy().(*semantic.FunctionExpression)

	ns.DescendingSet = s.DescendingSet
	ns.Descending = s.Descending

	ns.LimitSet = s.LimitSet
	ns.PointsLimit = s.PointsLimit
	ns.SeriesLimit = s.SeriesLimit
	ns.SeriesOffset = s.SeriesOffset

	ns.WindowSet = s.WindowSet
	ns.Window = s.Window

	ns.AggregateSet = s.AggregateSet
	ns.AggregateMethod = s.AggregateMethod

	return ns
}

// TimeBounds implements plan.BoundsAwareProcedureSpec
func (s *FromProcedureSpec) TimeBounds(predecessorBounds *plan.Bounds) *plan.Bounds {
	if s.BoundsSet {
		bounds := &plan.Bounds{
			Start: values.ConvertTime(s.Bounds.Start.Time(s.Bounds.Now)),
			Stop:  values.ConvertTime(s.Bounds.Stop.Time(s.Bounds.Now)),
		}
		return bounds
	}
	return nil
}

func createFromSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec := prSpec.(*FromProcedureSpec)
	var w execute.Window
	bounds := a.StreamContext().Bounds()
	if bounds == nil {
		return nil, errors.New("nil bounds passed to from")
	}

	if spec.WindowSet {
		w = execute.Window{
			Every:  execute.Duration(spec.Window.Every),
			Period: execute.Duration(spec.Window.Period),
			Round:  execute.Duration(spec.Window.Round),
			Start:  bounds.Start,
		}
	} else {
		duration := execute.Duration(bounds.Stop) - execute.Duration(bounds.Start)
		w = execute.Window{
			Every:  duration,
			Period: duration,
			Start:  bounds.Start,
		}
	}
	currentTime := w.Start + execute.Time(w.Period)

	deps := a.Dependencies()[FromKind].(storage.Dependencies)
	req := query.RequestFromContext(a.Context())
	if req == nil {
		return nil, errors.New("missing request on context")
	}
	orgID := req.OrganizationID

	var bucketID platform.ID
	// Determine bucketID
	switch {
	case spec.Bucket != "":
		b, ok := deps.BucketLookup.Lookup(orgID, spec.Bucket)
		if !ok {
			return nil, fmt.Errorf("could not find bucket %q", spec.Bucket)
		}
		bucketID = b
	case len(spec.BucketID) != 0:
		bucketID = spec.BucketID
	}

	return storage.NewSource(
		dsid,
		deps.Reader,
		storage.ReadSpec{
			OrganizationID:  orgID,
			BucketID:        bucketID,
			Predicate:       spec.Filter,
			PointsLimit:     spec.PointsLimit,
			SeriesLimit:     spec.SeriesLimit,
			SeriesOffset:    spec.SeriesOffset,
			Descending:      spec.Descending,
			OrderByTime:     spec.OrderByTime,
			GroupMode:       storage.GroupMode(spec.GroupMode),
			GroupKeys:       spec.GroupKeys,
			AggregateMethod: spec.AggregateMethod,
		},
		*bounds,
		w,
		currentTime,
	), nil
}

func InjectFromDependencies(depsMap execute.Dependencies, deps storage.Dependencies) error {
	if err := deps.Validate(); err != nil {
		return err
	}
	depsMap[FromKind] = deps
	return nil
}
