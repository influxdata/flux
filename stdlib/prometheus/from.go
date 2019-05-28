package prometheus

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/prompb"
)

const FromKind = "prometheus.from"
const defaultStep = 10 * time.Second

//FromOpSpec From Prometheus request data struct
type FromOpSpec struct {
	URL       string                 `json:"url,omitempty"`
	User      string                 `json:"user,omitempty"`
	Password  string                 `json:"password,omitempty"`
	Matcher   []*prompb.LabelMatcher `json:"matcher,omitempty"`
	Query     string                 `json:"query,omitempty"`
	Step      time.Duration          `json:"step,omitempty"`
	HasQuery  bool
	HasAuth   bool
	BoundsSet bool
}

func init() {
	fromSignature := semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"url":      semantic.String,
			"query":    semantic.String,
			"name":     semantic.String,
			"user":     semantic.String,
			"password": semantic.String,
			"step":     semantic.Duration,
		},
		Required: nil,
		Return:   flux.TableObjectType,
	}

	flux.RegisterPackageValue("prometheus", "from", flux.FunctionValue(FromKind, createFromOpSpec, fromSignature))
	flux.RegisterOpSpec(FromKind, newFromOp)
	plan.RegisterProcedureSpec(FromKind, newFromProcedure, FromKind)
	execute.RegisterSource(FromKind, createFromSource)
	plan.RegisterPhysicalRules(MergeFromRangeRule{}, MergeFromFilterRule{})
}

// MergeFromRangeRule pushes a `range` into a `from`
type MergeFromRangeRule struct{}

// Name returns the name of the rule
func (rule MergeFromRangeRule) Name() string {
	return "MergeFromPromRangeRule"
}

// Pattern returns the pattern that matches `from -> range`
func (rule MergeFromRangeRule) Pattern() plan.Pattern {
	return plan.Pat(universe.RangeKind, plan.Pat(FromKind))
}

// Rewrite attempts to rewrite a `from -> range` into a `FromRange`
func (rule MergeFromRangeRule) Rewrite(node plan.Node) (plan.Node, bool, error) {
	from := node.Predecessors()[0]
	fromSpec := from.ProcedureSpec().(*FromProcedureSpec)
	rangeSpec := node.ProcedureSpec().(*universe.RangeProcedureSpec)
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
	merged, err := plan.MergeToLogicalNode(node, from, fromRange)
	if err != nil {
		return nil, false, err
	}

	return merged, true, nil
}

type MergeFromFilterRule struct {
}

func (MergeFromFilterRule) Name() string {
	return "MergeFromPromFilterRule"
}

func (MergeFromFilterRule) Pattern() plan.Pattern {
	return plan.Pat(universe.FilterKind, plan.Pat(FromKind))
}

func (MergeFromFilterRule) Rewrite(filterNode plan.Node) (plan.Node, bool, error) {
	filterSpec := filterNode.ProcedureSpec().(*universe.FilterProcedureSpec)
	fromNode := filterNode.Predecessors()[0]
	newFromSpec := fromNode.ProcedureSpec().Copy().(*FromProcedureSpec)

	if newFromSpec.FilterSet {
		return filterNode, true, nil
	} else {
		newFromSpec.FilterSet = true
		err := newFromSpec.SetMatcherFromFilter(filterSpec.Fn)
		if err != nil {
			return nil, false, err
		}
	}
	err := fromNode.ReplaceSpec(newFromSpec)
	if err != nil {
		return nil, false, err
	}

	return fromNode, true, nil
}
func createFromOpSpec(args flux.Arguments, administration *flux.Administration) (flux.OperationSpec, error) {
	spec := &FromOpSpec{}

	spec.Matcher = make([]*prompb.LabelMatcher, 0)

	if matcher, ok, err := args.GetString("name"); err != nil {
		return nil, err
	} else if ok {
		nameMatcher := &prompb.LabelMatcher{Type: prompb.LabelMatcher_EQ, Name: "__name__", Value: matcher}
		spec.Matcher = append(spec.Matcher, nameMatcher)
	}

	if url, ok, err := args.GetString("url"); err != nil {
		return nil, err
	} else if ok {
		if url == "" {
			return nil, fmt.Errorf("Invalid PromQl url in %q", FromKind)
		}
		spec.URL = url
	}

	if query, ok, err := args.GetString("query"); err != nil {
		return nil, err
	} else if ok {
		if query == "" {
			return nil, fmt.Errorf("Invalid PromQl query in %q", FromKind)
		}
		spec.HasQuery = true
		spec.Query = query
	}
	if user, ok, err := args.GetString("user"); err != nil {
		return nil, err
	} else if ok {
		spec.User = user
		spec.HasAuth = true
	}
	if password, ok, err := args.GetString("password"); err != nil {
		return nil, err
	} else if ok {
		spec.Password = password
		spec.HasAuth = true
	}
	if step, has, err := args.GetDuration("step"); err != nil {
		return nil, err
	} else if has {
		if !spec.HasQuery {
			return nil, fmt.Errorf("Step parameter only apply on a PromQL query in %q", FromKind)
		}
		spec.Step = time.Duration(step)
	} else {
		spec.Step = defaultStep
	}
	return spec, nil
}

func newFromOp() flux.OperationSpec {
	return new(FromOpSpec)
}

func (s *FromOpSpec) Kind() flux.OperationKind {
	return FromKind
}

type FromProcedureSpec struct {
	plan.DefaultCost
	URL       string
	Query     string
	User      string
	Password  string
	Matcher   []*prompb.LabelMatcher
	Step      time.Duration
	hasQuery  bool
	hasAuth   bool
	Bounds    flux.Bounds
	BoundsSet bool
	FilterSet bool
}

func newFromProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &FromProcedureSpec{
		URL:      spec.URL,
		Query:    spec.Query,
		User:     spec.User,
		Password: spec.Password,
		Matcher:  spec.Matcher,
		Step:     spec.Step,
		hasQuery: spec.HasQuery,
		hasAuth:  spec.HasAuth,
	}, nil
}

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

func (s *FromProcedureSpec) Kind() plan.ProcedureKind {
	return FromKind
}
func (s *FromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromProcedureSpec)
	ns.URL = s.URL
	ns.User = s.User
	ns.Password = s.Password
	ns.BoundsSet = s.BoundsSet
	ns.Bounds = s.Bounds
	ns.Query = s.Query
	ns.Matcher = make([]*prompb.LabelMatcher, len(s.Matcher))
	copy(ns.Matcher, s.Matcher)
	ns.Step = s.Step
	ns.hasQuery = s.hasQuery
	ns.hasAuth = s.hasAuth
	return ns
}

func (s *FromProcedureSpec) SetMatcherFromFilter(fn *semantic.FunctionExpression) error {
	m, err := s.toMatcher(fn.Block.Body.(semantic.Expression))
	if err != nil {
		return err
	}
	if m != nil {
		s.Matcher = append(s.Matcher, m...)
	}
	return nil
}

func (s *FromProcedureSpec) toMatcher(n semantic.Expression) ([]*prompb.LabelMatcher, error) {
	switch n := n.(type) {
	case *semantic.LogicalExpression:
		left, err := s.toMatcher(n.Left)
		if err != nil {
			return nil, errors.Wrap(err, "left hand side")
		}
		right, err := s.toMatcher(n.Right)
		if err != nil {
			return nil, errors.Wrap(err, "right hand side")
		}
		switch n.Operator {
		case ast.AndOperator:
			return append(left, right...), nil
		case ast.OrOperator:
			return nil, errors.New("or operator not supported in from")
		default:
			return nil, fmt.Errorf("unknown logical operator %v", n.Operator)
		}
	case *semantic.BinaryExpression:
		left, err := s.toLiteralMatcher(n.Left)
		if err != nil {
			return nil, errors.Wrap(err, "left hand side")
		}
		right, err := s.toLiteralMatcher(n.Right)
		if err != nil {
			return nil, errors.Wrap(err, "right hand side")
		}
		res := make([]*prompb.LabelMatcher, 0)
		switch n.Operator {
		case ast.EqualOperator:
			lm := &prompb.LabelMatcher{Type: prompb.LabelMatcher_EQ, Name: left, Value: right}
			res = append(res, lm)
			return res, nil
		case ast.NotEqualOperator:
			lm := &prompb.LabelMatcher{Type: prompb.LabelMatcher_NEQ, Name: left, Value: right}
			res = append(res, lm)
			return res, nil
		case ast.RegexpMatchOperator:
			lm := &prompb.LabelMatcher{Type: prompb.LabelMatcher_RE, Name: left, Value: right}
			res = append(res, lm)
			return res, nil
		case ast.NotRegexpMatchOperator:
			lm := &prompb.LabelMatcher{Type: prompb.LabelMatcher_NRE, Name: left, Value: right}
			res = append(res, lm)
			return res, nil
		case ast.StartsWithOperator:
			right = fmt.Sprintf("\"^%s.*\"", right)
			lm := &prompb.LabelMatcher{Type: prompb.LabelMatcher_RE, Name: left, Value: right}
			res = append(res, lm)
			return res, nil
		case ast.LessThanOperator:
			return nil, errors.New("< not supported")
		case ast.LessThanEqualOperator:
			return nil, errors.New("<= not supported")
		case ast.GreaterThanOperator:
			return nil, errors.New("> not supported")
		case ast.GreaterThanEqualOperator:
			return nil, errors.New(">= not supported")
		case ast.InOperator:
			lm := &prompb.LabelMatcher{Type: prompb.LabelMatcher_RE, Name: left, Value: right}
			res = append(res, lm)
			return res, nil
		default:
			return nil, fmt.Errorf("unknown operator %v", n.Operator)
		}
	default:
		return nil, fmt.Errorf("unsupported semantic expression type %T", n)
	}
}
func (s *FromProcedureSpec) toLiteralMatcher(n semantic.Expression) (string, error) {
	switch n := n.(type) {
	case *semantic.StringLiteral:
		return n.Value, nil
	case *semantic.IntegerLiteral:
		return fmt.Sprintf("%d", n.Value), nil
	case *semantic.BooleanLiteral:
		if n.Value {
			return "true", nil
		}
		return "false", nil
	case *semantic.FloatLiteral:
		return fmt.Sprintf("%f", n.Value), nil
	case *semantic.RegexpLiteral:
		return n.Value.String(), nil
	case *semantic.MemberExpression:
		if n.Property == "_value" {
			return "", errors.New("unable to push value filtering down to Prometheus")
		}
		return n.Property, nil
	case *semantic.ArrayExpression:
		vals := make([]string, 0, len(n.Elements))
		for _, e := range n.Elements {
			vals = append(vals, e.(*semantic.StringLiteral).Value)
		}
		return strings.Join(vals, "|"), nil
	case *semantic.DurationLiteral:
		return "", errors.New("duration literals not supported in storage predicates")
	case *semantic.DateTimeLiteral:
		return "", errors.New("time literals not supported in storage predicates")
	default:
		return "", fmt.Errorf("unsupported semantic expression type %T", n)
	}
}

// Prometheus.from Source
func createFromSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*FromProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", prSpec)
	}
	PromIterator := PromIterator{id: dsid, spec: spec, administration: a, index: -1}
	return execute.CreateSourceFromDecoder(&PromIterator, dsid, a)
}

type PromIterator struct {
	id               execute.DatasetID
	data             flux.Result
	ts               []execute.Transformation
	administration   execute.Administration
	spec             *FromProcedureSpec
	prom             *PromClient
	result           []*QueryRangeResponseResult
	resultRemoteRead []*prompb.TimeSeries
	start            time.Time
	end              time.Time
	index            int
}

func (pi *PromIterator) Connect() error {
	if pi.spec.hasAuth {
		prom, err := NewAuthPromClient(pi.spec.URL, pi.spec.User, pi.spec.Password)
		if err != nil {
			return err
		}
		pi.prom = prom
		return nil
	}
	prom, err := NewPromClient(pi.spec.URL)
	if err != nil {
		return err
	}
	pi.prom = prom
	return nil
}
func (pi *PromIterator) Fetch() (bool, error) {
	if pi.index == -1 {
		now := time.Now()
		pi.start = pi.spec.Bounds.Start.Time(now)
		pi.end = pi.spec.Bounds.Stop.Time(now)
		// Case of PromQL query
		if pi.spec.hasQuery {
			pi.resultRemoteRead = make([]*prompb.TimeSeries, 0)
			query := pi.spec.Query
			promRes, err := pi.prom.QueryRange(query, pi.start, pi.end, pi.spec.Step)
			if err != nil {
				return false, err
			}
			pi.result = promRes.Data.Result
			pi.index = 0
			if len(pi.result) == 0 {
				return false, nil
			}
			return true, nil
		}
		// Case of loading Prom matcher
		pi.result = make([]*QueryRangeResponseResult, 0)
		promQuery := &prompb.Query{StartTimestampMs: pi.start.UnixNano() / 1000000, EndTimestampMs: pi.end.UnixNano() / 1000000, Matchers: pi.spec.Matcher}

		resp, err := pi.prom.QueryRemoteRead(promQuery)
		if err != nil {
			return false, err
		}
		timeseries := make([]*prompb.TimeSeries, 0)
		for _, list := range resp.Results {
			timeseries = append(timeseries, list.Timeseries...)
		}
		pi.resultRemoteRead = timeseries
		pi.index = 0
		if len(pi.resultRemoteRead) == 0 {
			return false, nil
		}
		return true, nil
	}
	pi.index = pi.index + 1
	if pi.spec.hasQuery && pi.index >= len(pi.result) {
		return false, nil
	}
	if !pi.spec.hasQuery && pi.index >= len(pi.resultRemoteRead) {
		return false, nil
	}
	return true, nil
}

func (pi *PromIterator) Close() error {
	return nil
}

func (pi *PromIterator) Decode() (flux.Table, error) {
	//maxLimit := int64(64)
	if pi.spec.hasQuery {
		if len(pi.result) > 0 {
			return pi.ParseResult(pi.result[pi.index])
		}
		groupKey := execute.NewGroupKey(nil, nil)
		builder := execute.NewColListTableBuilder(groupKey, &memory.Allocator{})
		return builder.Table()
	}
	if len(pi.resultRemoteRead) > 0 {
		return pi.ParseRemoteRead(pi.resultRemoteRead[pi.index])
	}
	groupKey := execute.NewGroupKey(nil, nil)
	builder := execute.NewColListTableBuilder(groupKey, &memory.Allocator{})
	return builder.Table()
}

// ParseResult convert Provider result to influx format
func (pi *PromIterator) ParseResult(series *QueryRangeResponseResult) (flux.Table, error) {
	keyCols := make([]flux.ColMeta, 0, len(series.Metric)+2)
	keyValues := make([]values.Value, 0, len(series.Metric)+2)
	names := make([]string, 0, len(series.Metric))
	for n := range series.Metric {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		value := series.Metric[name]
		keyCols = append(keyCols, flux.ColMeta{Label: name, Type: flux.TString})
		keyValues = append(keyValues, values.NewString(value))
	}
	keyCols = append(keyCols, flux.ColMeta{Label: "_start", Type: flux.TTime})
	keyCols = append(keyCols, flux.ColMeta{Label: "_stop", Type: flux.TTime})
	keyValues = append(keyValues, values.NewTime(values.ConvertTime(pi.start)))
	keyValues = append(keyValues, values.NewTime(values.ConvertTime(pi.end)))
	key := execute.NewGroupKey(keyCols, keyValues)
	builder := execute.NewColListTableBuilder(key, &memory.Allocator{})
	for _, c := range keyCols {
		builder.AddCol(c)
	}
	valueIdx := len(keyCols)
	timeIdx := valueIdx + 1
	builder.AddCol(flux.ColMeta{Label: "_value", Type: flux.TFloat})
	builder.AddCol(flux.ColMeta{Label: "_time", Type: flux.TTime})
	for _, v := range series.Values {
		val, err := v.Value()
		if err != nil {
			continue
		}
		// Add all labels in each table line
		l := len(keyValues) - 2
		for i, v := range keyValues[:l] {
			builder.AppendString(i, v.Str())
		}
		// Add stat and end in each table line
		builder.AppendTime(l, values.ConvertTime(pi.start))
		builder.AppendTime(l+1, values.ConvertTime(pi.end))
		// Add current value and time in each table line
		builder.AppendFloat(valueIdx, val)
		builder.AppendTime(timeIdx, values.ConvertTime(v.Time()))
	}
	return builder.Table()
}

// ParseRemoteRead convert Provider result to influx format
func (pi *PromIterator) ParseRemoteRead(series *prompb.TimeSeries) (flux.Table, error) {
	keyCols := make([]flux.ColMeta, 0, len(series.Labels)+2)
	keyValues := make([]values.Value, 0, len(series.Labels)+2)
	for _, label := range series.Labels {
		keyCols = append(keyCols, flux.ColMeta{Label: label.Name, Type: flux.TString})
		keyValues = append(keyValues, values.NewString(label.Value))
	}
	keyCols = append(keyCols, flux.ColMeta{Label: "_start", Type: flux.TTime})
	keyCols = append(keyCols, flux.ColMeta{Label: "_stop", Type: flux.TTime})
	keyValues = append(keyValues, values.NewTime(values.ConvertTime(pi.start)))
	keyValues = append(keyValues, values.NewTime(values.ConvertTime(pi.end)))
	key := execute.NewGroupKey(keyCols, keyValues)
	builder := execute.NewColListTableBuilder(key, &memory.Allocator{})
	for _, c := range keyCols {
		builder.AddCol(c)
	}
	valueIdx := len(keyCols)
	timeIdx := valueIdx + 1
	builder.AddCol(flux.ColMeta{Label: "_value", Type: flux.TFloat})
	builder.AddCol(flux.ColMeta{Label: "_time", Type: flux.TTime})
	for _, v := range series.Samples {
		val := v.Value

		// Convert timestamp to Flux time unit
		tick := v.Timestamp * 1000000
		// Add all labels in each table line
		l := len(keyValues) - 2
		for i, v := range keyValues[:l] {
			builder.AppendString(i, v.Str())
		}
		// Add stat and end in each table line
		builder.AppendTime(l, values.ConvertTime(pi.start))
		builder.AppendTime(l+1, values.ConvertTime(pi.end))
		// Add current value and time in each table line
		builder.AppendFloat(valueIdx, val)
		builder.AppendTime(timeIdx, values.Time(tick))
	}
	return builder.Table()
}
