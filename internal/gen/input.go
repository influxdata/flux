package gen

import (
	"container/heap"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	// DefaultNumPoints is the default number of points that should
	// be generated for each series.
	DefaultNumPoints = 6

	// DefaultPeriod is the default period between points in a series.
	DefaultPeriod = 10 * time.Second
)

// Tag includes the tag name and the cardinality for that tag in
// the schema.
type Tag struct {
	Name        string
	Cardinality int
}

// Schema describes the schema to be generated.
type Schema struct {
	// Start is the start time for generating data. This will default
	// so the current time would be the last point generated, but
	// truncated to the period.
	Start time.Time

	// Tags is a listing of tags and the generated cardinality for
	// that tag.
	Tags []Tag

	// NumPoints is the number of points that should be generated
	// for each series. This defaults to 6.
	NumPoints int

	// Nulls sets the percentage changes that a null value will
	// be used in the input. This should be a number between 0 and 1.
	Nulls float64

	// Period contains the distance between each point in a series.
	// This defaults to 10 seconds.
	Period time.Duration

	// GroupBy is a list of tags that, if they have the same value,
	// will have the same type even if the types ratio becomes
	// impossible to fulfill. This only does something if Types
	// has been set.
	GroupBy []string

	// Types includes a mapping of the column value type
	// to the ratio for how frequently it should show up
	// in the output. If this is left blank, all series
	// will be generated with a float value.
	Types map[flux.ColType]int

	// Seed is the (optional) seed to be used by the random
	// number generator. If this is null, the current time
	// will be used.
	Seed *int64

	// Alloc assigns an allocator to use when generating the
	// tables. If this is not set, an unlimited allocator is
	// used.
	Alloc *memory.Allocator
}

// Input constructs a TableIterator with randomly generated
// data according to the Schema.
func Input(schema Schema) (flux.TableIterator, error) {
	tags := schema.Tags

	var seed int64
	if schema.Seed != nil {
		seed = *schema.Seed
	} else {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))
	series := genSeriesKeys(tags, r)
	if len(series) == 0 {
		// If no tags were provided, then there is only one series
		// and it is the default one.
		series = []flux.GroupKey{
			execute.NewGroupKey(nil, nil),
		}
	}

	var ti typeInfo
	if len(schema.Types) > 0 {
		var total int
		for _, count := range schema.Types {
			total += count
		}
		for typ, count := range schema.Types {
			ti = append(ti, valueType{
				Type:   typ,
				Number: int(math.Round(float64(len(series)) * float64(count) / float64(total))),
			})
		}
	} else {
		ti = typeInfo{
			valueType{
				Type:   flux.TFloat,
				Number: len(series),
			},
		}
	}
	heap.Init(&ti)

	var groupTags []string
	if len(ti) > 1 {
		groupTags = schema.GroupBy
	}
	groups := seriesGroups(groupBy(series, groupTags))
	heap.Init(&groups)

	period := schema.Period
	if period == 0 {
		period = DefaultPeriod
	}

	numPoints := schema.NumPoints
	if numPoints == 0 {
		numPoints = DefaultNumPoints
	}

	alloc := schema.Alloc
	if alloc == nil {
		alloc = &memory.Allocator{}
	}
	g := &dataGenerator{
		Period:    values.Duration(period),
		NumPoints: numPoints,
		Nulls:     schema.Nulls,
		Allocator: alloc,
		Rand:      r,
		Groups:    groups,
		TypeInfo:  ti,
	}
	if !schema.Start.IsZero() {
		g.Start = values.ConvertTime(schema.Start)
	} else {
		ts := time.Now().Truncate(period).Add(-period * time.Duration(numPoints))
		g.Start = values.ConvertTime(ts)
	}
	return g, nil
}

// CsvInput generates a csv input based on the Schema.
func CsvInput(schema Schema) (string, error) {
	tables, err := Input(schema)
	if err != nil {
		return "", err
	}

	results := flux.NewSliceResultIterator([]flux.Result{
		&result{tables: tables},
	})

	var buf strings.Builder
	enc := csv.NewMultiResultEncoder(csv.DefaultEncoderConfig())
	if _, err := enc.Encode(&buf, results); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// seriesGroup is a group of series that should have the same type.
type seriesGroup struct {
	Series []flux.GroupKey
	Type   flux.ColType
}

type seriesGroups []seriesGroup

func (a *seriesGroups) Len() int {
	return len(*a)
}

func (a *seriesGroups) Less(i, j int) bool {
	return len((*a)[i].Series) > len((*a)[j].Series)
}

func (a *seriesGroups) Swap(i, j int) {
	(*a)[i], (*a)[j] = (*a)[j], (*a)[i]
}

func (a *seriesGroups) Push(x interface{}) {
	*a = append(*a, x.(seriesGroup))
}

func (a *seriesGroups) Pop() interface{} {
	sg := (*a)[len(*a)-1]
	*a = (*a)[:len(*a)-1]
	return sg
}

// valueType keeps a mapping of the number of series we wish to generate for each type.
type valueType struct {
	Type   flux.ColType
	Number int
}

type typeInfo []valueType

func (a *typeInfo) Len() int {
	return len(*a)
}

func (a *typeInfo) Less(i, j int) bool {
	return (*a)[i].Number > (*a)[j].Number
}

func (a *typeInfo) Swap(i, j int) {
	(*a)[i], (*a)[j] = (*a)[j], (*a)[i]
}

func (a *typeInfo) Push(x interface{}) {
	*a = append(*a, x.(valueType))
}

func (a *typeInfo) Pop() interface{} {
	vt := (*a)[len(*a)-1]
	*a = (*a)[:len(*a)-1]
	return vt
}

func genTagValue(r *rand.Rand, min, max int) string {
	var buf strings.Builder
	sz := r.Intn(max-min) + min
	for i := 0; i < sz; i++ {
		chars := 62
		if i == 0 {
			chars = 52
		}
		switch n := r.Intn(chars); {
		case n >= 0 && n < 26:
			buf.WriteByte('A' + byte(n))
		case n >= 26 && n < 52:
			buf.WriteByte('a' + byte(n-26))
		case n >= 52:
			buf.WriteByte('0' + byte(n-52))
		}
	}
	return buf.String()
}

func genTagValues(r *rand.Rand, cardinality, min, max int) []string {
	values := make([]string, 0, cardinality)
	for i := 0; i < cardinality; i++ {
		v := genTagValue(r, min, max)
		values = append(values, v)
	}
	return values
}

func appendTagKey(series []flux.GroupKey, k string, vs []string) []flux.GroupKey {
	if len(vs) == 0 {
		return series
	}

	if len(series) == 0 {
		series = []flux.GroupKey{nil}
	}

	newSeries := make([]flux.GroupKey, 0, len(series)*len(vs))
	for _, s := range series {
		for _, v := range vs {
			gkb := execute.NewGroupKeyBuilder(s)
			gkb.AddKeyValue(k, values.NewString(v))

			key, _ := gkb.Build()
			newSeries = append(newSeries, key)
		}
	}
	return newSeries
}

func genSeriesKeys(tags []Tag, r *rand.Rand) []flux.GroupKey {
	var keys []flux.GroupKey
	for _, tag := range tags {
		if tag.Cardinality == 0 {
			continue
		}
		keys = appendTagKey(keys, tag.Name, genTagValues(r, tag.Cardinality, 3, 8))
	}
	return keys
}

func groupBy(keys []flux.GroupKey, by []string) []seriesGroup {
	if len(by) == 0 {
		groups := make([]seriesGroup, 0, len(keys))
		for _, k := range keys {
			groups = append(groups, seriesGroup{
				Series: []flux.GroupKey{k},
			})
		}
		return groups
	}

	var groups []seriesGroup
	mapping := make(map[string]*seriesGroup)
	for _, k := range keys {
		parts := make([]string, 0, len(by))
		for _, s := range by {
			idx := execute.ColIdx(s, k.Cols())
			if idx == -1 {
				continue
			}
			parts = append(parts, k.ValueString(idx))
		}

		if len(parts) == 0 {
			groups = append(groups, seriesGroup{
				Series: []flux.GroupKey{k},
			})
			continue
		}

		groupKey := strings.Join(parts, ",")
		gr, ok := mapping[groupKey]
		if !ok {
			groups = append(groups, seriesGroup{
				Series: []flux.GroupKey{k},
			})
			mapping[groupKey] = &groups[len(groups)-1]
			continue
		}
		gr.Series = append(gr.Series, k)
	}
	return groups
}

type dataGenerator struct {
	Start     values.Time
	Period    values.Duration
	Jitter    values.Duration
	Nulls     float64
	NumPoints int
	Allocator *memory.Allocator

	Rand     *rand.Rand
	Groups   seriesGroups
	TypeInfo typeInfo
}

func (dg *dataGenerator) Do(f func(tbl flux.Table) error) error {
	for {
		if len(dg.Groups) == 0 {
			break
		}

		sg := heap.Pop(&dg.Groups).(seriesGroup)
		vt := heap.Pop(&dg.TypeInfo).(valueType)
		vt.Number -= len(sg.Series)
		sg.Type = vt.Type

		for _, s := range sg.Series {
			builder := execute.NewColListTableBuilder(s, dg.Allocator)
			startIdx, _ := builder.AddCol(flux.ColMeta{
				Label: execute.DefaultStartColLabel,
				Type:  flux.TTime,
			})
			stopIdx, _ := builder.AddCol(flux.ColMeta{
				Label: execute.DefaultStopColLabel,
				Type:  flux.TTime,
			})
			_ = execute.AddTableKeyCols(s, builder)
			start, stop := dg.Generate(builder, dg.Rand, sg.Type)
			for i := 0; i < dg.NumPoints; i++ {
				_ = builder.AppendTime(startIdx, start)
				_ = builder.AppendTime(stopIdx, stop)
				_ = execute.AppendKeyValues(s, builder)
			}

			table, err := builder.Table()
			if err != nil {
				builder.Release()
				return err
			}
			builder.Release()
			if err := f(table); err != nil {
				return err
			}
		}
		heap.Push(&dg.TypeInfo, vt)
	}
	return nil
}

func (dg *dataGenerator) Generate(tb execute.TableBuilder, r *rand.Rand, typ flux.ColType) (start, stop values.Time) {
	var next func() values.Value
	switch typ {
	case flux.TFloat:
		next = func() values.Value {
			if dg.Nulls > 0.0 && dg.Nulls > r.Float64() {
				return values.NewNull(semantic.Float)
			}
			v := rand.NormFloat64() * 50
			return values.NewFloat(v)
		}
	case flux.TInt:
		next = func() values.Value {
			if dg.Nulls > 0.0 && dg.Nulls > r.Float64() {
				return values.NewNull(semantic.Int)
			}
			v := rand.Intn(201) - 100
			return values.NewInt(int64(v))
		}
	case flux.TUInt:
		next = func() values.Value {
			if dg.Nulls > 0.0 && dg.Nulls > r.Float64() {
				return values.NewNull(semantic.UInt)
			}
			v := rand.Intn(101)
			return values.NewUInt(uint64(v))
		}
	case flux.TString:
		next = func() values.Value {
			if dg.Nulls > 0.0 && dg.Nulls > r.Float64() {
				return values.NewNull(semantic.String)
			}
			v := genTagValue(r, 3, 8)
			return values.NewString(v)
		}
	case flux.TBool:
		next = func() values.Value {
			if dg.Nulls > 0.0 && dg.Nulls > r.Float64() {
				return values.NewNull(semantic.Bool)
			}
			v := r.Intn(2) == 1
			return values.NewBool(v)
		}
	default:
		panic("implement me")
	}

	timeIdx, _ := tb.AddCol(flux.ColMeta{
		Label: execute.DefaultTimeColLabel,
		Type:  flux.TTime,
	})
	valueIdx, _ := tb.AddCol(flux.ColMeta{
		Label: execute.DefaultValueColLabel,
		Type:  typ,
	})

	start, stop = dg.Start, dg.Start
	for i := 0; i < dg.NumPoints; i++ {
		ts := dg.Start.Add(values.Duration(i) * dg.Period)
		if dg.Jitter != 0 {
			jitter := r.Intn(int(dg.Jitter)*2 + 1)
			ts = ts.Add(values.Duration(jitter))
		}
		_ = tb.AppendTime(timeIdx, ts)
		_ = tb.AppendValue(valueIdx, next())
		_ = tb.AppendValue(valueIdx, next())
		if ts > stop {
			stop = ts
		}
	}
	return start, stop
}

type result struct {
	tables flux.TableIterator
}

func (r *result) Name() string {
	return ""
}

func (r *result) Tables() flux.TableIterator {
	return r.tables
}
