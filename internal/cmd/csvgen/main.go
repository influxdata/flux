package main

import (
	"container/heap"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

var Flags struct {
	Start          string
	Keys           string
	TagCardinality string
	NumPoints      int
	Period         string
	GroupBy        string
	Types          string
}

// SeriesGroup is a group of series that should have the same type.
type SeriesGroup struct {
	Series []flux.GroupKey
	Type   flux.ColType
}

type SeriesGroups []SeriesGroup

func (a *SeriesGroups) Len() int {
	return len(*a)
}

func (a *SeriesGroups) Less(i, j int) bool {
	return len((*a)[i].Series) > len((*a)[j].Series)
}

func (a *SeriesGroups) Swap(i, j int) {
	(*a)[i], (*a)[j] = (*a)[j], (*a)[i]
}

func (a *SeriesGroups) Push(x interface{}) {
	*a = append(*a, x.(SeriesGroup))
}

func (a *SeriesGroups) Pop() interface{} {
	sg := (*a)[len(*a)-1]
	*a = (*a)[:len(*a)-1]
	return sg
}

// ValueType keeps a mapping of the number of series we wish to generate for each type.
type ValueType struct {
	Type   flux.ColType
	Number int
}

type TypeInfo []ValueType

func (a *TypeInfo) Len() int {
	return len(*a)
}

func (a *TypeInfo) Less(i, j int) bool {
	return (*a)[i].Number > (*a)[j].Number
}

func (a *TypeInfo) Swap(i, j int) {
	(*a)[i], (*a)[j] = (*a)[j], (*a)[i]
}

func (a *TypeInfo) Push(x interface{}) {
	*a = append(*a, x.(ValueType))
}

func (a *TypeInfo) Pop() interface{} {
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

func genSeriesKeys(tags map[string]int, r *rand.Rand) []flux.GroupKey {
	var keys []flux.GroupKey
	if c, ok := tags["_measurement"]; !ok || c > 0 {
		if !ok {
			c = 1
		}
		keys = appendTagKey(keys, "_measurement", genTagValues(r, c, 3, 8))
	}
	if c, ok := tags["_field"]; !ok || c > 0 {
		if !ok {
			c = 1
		}
		keys = appendTagKey(keys, "_field", genTagValues(r, c, 3, 8))
	}
	for k, c := range tags {
		if k == "_measurement" || k == "_field" {
			continue
		}
		keys = appendTagKey(keys, k, genTagValues(r, c, 3, 8))
	}
	return keys
}

func groupBy(keys []flux.GroupKey, by []string) []SeriesGroup {
	if len(by) == 0 {
		groups := make([]SeriesGroup, 0, len(keys))
		for _, k := range keys {
			groups = append(groups, SeriesGroup{
				Series: []flux.GroupKey{k},
			})
		}
		return groups
	}

	var groups []SeriesGroup
	mapping := make(map[string]*SeriesGroup)
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
			groups = append(groups, SeriesGroup{
				Series: []flux.GroupKey{k},
			})
			continue
		}

		groupKey := strings.Join(parts, ",")
		gr, ok := mapping[groupKey]
		if !ok {
			groups = append(groups, SeriesGroup{
				Series: []flux.GroupKey{k},
			})
			mapping[groupKey] = &groups[len(groups)-1]
			continue
		}
		gr.Series = append(gr.Series, k)
	}
	return groups
}

type DataGenerator struct {
	Start     values.Time
	Period    values.Duration
	Jitter    values.Duration
	NumPoints int
}

func (dg *DataGenerator) Generate(tb execute.TableBuilder, r *rand.Rand, typ flux.ColType) (start, stop values.Time) {
	var next func() values.Value
	switch typ {
	case flux.TFloat:
		next = func() values.Value {
			v := rand.NormFloat64() * 50
			return values.NewFloat(v)
		}
	case flux.TInt:
		next = func() values.Value {
			v := rand.Intn(201) - 100
			return values.NewInt(int64(v))
		}
	case flux.TUInt:
		next = func() values.Value {
			v := rand.Intn(101)
			return values.NewUInt(uint64(v))
		}
	case flux.TString:
		next = func() values.Value {
			v := genTagValue(r, 3, 8)
			return values.NewString(v)
		}
	case flux.TBool:
		next = func() values.Value {
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
		if ts > stop {
			stop = ts
		}
	}
	return start, stop
}

func getTags() (map[string]int, error) {
	if len(Flags.Keys) == 0 {
		return nil, nil
	}

	m := make(map[string]int)
	keys := strings.Split(Flags.Keys, ",")

	if len(Flags.TagCardinality) == 0 {
		for _, k := range keys {
			m[k] = 1
		}
		return m, nil
	}

	cardinality := strings.Split(Flags.TagCardinality, ",")
	if len(cardinality) != len(keys) {
		return nil, fmt.Errorf("cardinality must have the same number of entries as keys")
	}
	for i, k := range keys {
		c, err := strconv.Atoi(cardinality[i])
		if err != nil {
			return nil, fmt.Errorf("unable to parse %s as an integer: %s", cardinality[i], err)
		}
		m[k] = c
	}
	return m, nil
}

func main() {
	flag.StringVar(&Flags.Start, "s", "", "start time of the data")
	flag.StringVar(&Flags.Keys, "k", "", "comma-separated list of tags that should be included")
	flag.StringVar(&Flags.TagCardinality, "t", "", "comma-separated list of the cardinality for each tag specified")
	flag.IntVar(&Flags.NumPoints, "p", 6, "number of points to write for each series")
	flag.StringVar(&Flags.Period, "d", "10s", "the duration between each point")
	flag.StringVar(&Flags.GroupBy, "group-by", "", "ensure that series with these tags in common have the same types")
	flag.StringVar(&Flags.Types, "types", "", "list of which types should be used along with an optional frequency like float:2,int:1")
	flag.Parse()

	tags, err := getTags()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	series := genSeriesKeys(tags, r)

	var ti TypeInfo
	if len(Flags.Types) > 0 {
		var total int
		types := make(map[flux.ColType]int)
		for _, typstr := range strings.Split(Flags.Types, ",") {
			count := 1
			parts := strings.SplitN(typstr, ":", 2)
			if len(parts) == 2 {
				if c, err := strconv.Atoi(parts[1]); err != nil {
					panic(err)
				} else {
					count = c
				}
			}

			var typ flux.ColType
			switch parts[0] {
			case "float":
				typ = flux.TFloat
			case "int":
				typ = flux.TInt
			case "uint":
				typ = flux.TUInt
			case "string":
				typ = flux.TString
			case "bool":
				typ = flux.TBool
			}
			types[typ] = count
			total += count
		}

		for typ, count := range types {
			ti = append(ti, ValueType{
				Type:   typ,
				Number: int(math.Round(float64(len(series)) * float64(count) / float64(total))),
			})
		}
	} else {
		ti = TypeInfo{
			ValueType{
				Type:   flux.TFloat,
				Number: len(series),
			},
		}
	}
	heap.Init(&ti)

	var groupTags []string
	if len(ti) > 1 {
		groupTags = []string{"_measurement", "_field"}
		if len(Flags.GroupBy) > 0 {
			groupTags = append(groupTags, strings.Split(Flags.GroupBy, ",")...)
		}
	}
	groups := SeriesGroups(groupBy(series, groupTags))
	heap.Init(&groups)

	period, err := time.ParseDuration(Flags.Period)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	g := &DataGenerator{
		Start:     values.ConvertTime(time.Now().Truncate(time.Second)),
		Period:    values.Duration(period),
		NumPoints: Flags.NumPoints,
	}
	if len(Flags.Start) > 0 {
		ts, err := time.Parse(time.RFC3339Nano, Flags.Start)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		g.Start = values.ConvertTime(ts)
	}

	cache := execute.NewTableBuilderCache(&memory.Allocator{})
	cache.SetTriggerSpec(flux.DefaultTrigger)
	for {
		if len(groups) == 0 {
			break
		}

		sg := heap.Pop(&groups).(SeriesGroup)
		vt := heap.Pop(&ti).(ValueType)
		vt.Number -= len(sg.Series)
		sg.Type = vt.Type

		for _, s := range sg.Series {
			builder, _ := cache.TableBuilder(s)
			startIdx, _ := builder.AddCol(flux.ColMeta{
				Label: execute.DefaultStartColLabel,
				Type:  flux.TTime,
			})
			stopIdx, _ := builder.AddCol(flux.ColMeta{
				Label: execute.DefaultStopColLabel,
				Type:  flux.TTime,
			})
			_ = execute.AddTableKeyCols(s, builder)
			start, stop := g.Generate(builder, r, sg.Type)
			for i := 0; i < g.NumPoints; i++ {
				_ = builder.AppendTime(startIdx, start)
				_ = builder.AppendTime(stopIdx, stop)
				_ = execute.AppendKeyValues(s, builder)
			}
		}
		heap.Push(&ti, vt)
	}

	res := &Result{cache: cache}
	enc := csv.NewResultEncoder(csv.DefaultEncoderConfig())
	if _, err := enc.Encode(os.Stdout, res); err != nil {
		panic(err)
	}
}

type Result struct {
	cache execute.TableBuilderCache
}

func (r *Result) Do(f func(flux.Table) error) error {
	r.cache.ForEachBuilder(func(key flux.GroupKey, builder execute.TableBuilder) {
		table, err := builder.Table()
		if err != nil {
			return
		}
		_ = f(table)
	})
	return nil
}

func (r *Result) Name() string {
	return ""
}

func (r *Result) Tables() flux.TableIterator {
	return r
}

func (r *Result) Statistics() flux.Statistics {
	return flux.Statistics{}
}
