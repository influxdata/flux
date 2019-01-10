package universe_test

import (
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/values"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

const (
	N     = 1e6
	Mu    = 10
	Sigma = 3

	seed = 42
)

// NormalData is a slice of N random values that are normaly distributed with mean Mu and standard deviation Sigma.
var NormalData []float64

// NormalTable is a table of data whose value col is NormalData.
var NormalTable flux.Table

func init() {
	dist := distuv.Normal{
		Mu:    Mu,
		Sigma: Sigma,
		Src:   rand.New(rand.NewSource(seed)),
	}
	NormalData = make([]float64, N)
	for i := range NormalData {
		NormalData[i] = dist.Rand()
	}
	start := execute.Time(time.Date(2016, 10, 10, 0, 0, 0, 0, time.UTC).UnixNano())
	stop := execute.Time(time.Date(2017, 10, 10, 0, 0, 0, 0, time.UTC).UnixNano())
	t1Value := "a"
	key := execute.NewGroupKey(
		[]flux.ColMeta{
			{Label: execute.DefaultStartColLabel, Type: flux.TTime},
			{Label: execute.DefaultStopColLabel, Type: flux.TTime},
			{Label: "t1", Type: flux.TString},
		},
		[]values.Value{
			values.NewTime(start),
			values.NewTime(stop),
			values.NewString(t1Value),
		},
	)
	normalTableBuilder := execute.NewColListTableBuilder(key, executetest.UnlimitedAllocator)

	if _, err := normalTableBuilder.AddCol(flux.ColMeta{Label: execute.DefaultTimeColLabel, Type: flux.TTime}); err != nil {
		return
	}
	if _, err := normalTableBuilder.AddCol(flux.ColMeta{Label: execute.DefaultStartColLabel, Type: flux.TTime}); err != nil {
		return
	}
	if _, err := normalTableBuilder.AddCol(flux.ColMeta{Label: execute.DefaultStopColLabel, Type: flux.TTime}); err != nil {
		return
	}
	if _, err := normalTableBuilder.AddCol(flux.ColMeta{Label: execute.DefaultValueColLabel, Type: flux.TFloat}); err != nil {
		return
	}
	if _, err := normalTableBuilder.AddCol(flux.ColMeta{Label: "t1", Type: flux.TString}); err != nil {
		return
	}
	if _, err := normalTableBuilder.AddCol(flux.ColMeta{Label: "t2", Type: flux.TString}); err != nil {
		return
	}

	times := make([]int64, N)
	startTimes := make([]int64, N)
	stopTimes := make([]int64, N)
	normalValues := NormalData
	t1 := make([]string, N)
	t2 := make([]string, N)

	for i, v := range normalValues {
		startTimes[i] = int64(start)
		stopTimes[i] = int64(stop)
		t1[i] = t1Value
		// There are roughly 1 million, 31 second intervals in a year.
		times[i] = int64(start + execute.Time(time.Duration(i*31)*time.Second))
		// Pick t2 based off the value
		switch int(v) % 3 {
		case 0:
			t2[i] = "x"
		case 1:
			t2[i] = "y"
		case 2:
			t2[i] = "z"
		}
	}

	timesArrow := arrow.NewInt(times, nil)
	startTimesArrow := arrow.NewInt(startTimes, nil)
	stopTimesArrow := arrow.NewInt(stopTimes, nil)
	normalValuesArrow := arrow.NewFloat(normalValues, nil)
	t1Arrow := arrow.NewString(t1, nil)
	t2Arrow := arrow.NewString(t2, nil)
	defer func() {
		timesArrow.Release()
		startTimesArrow.Release()
		stopTimesArrow.Release()
		normalValuesArrow.Release()
		t1Arrow.Release()
		t2Arrow.Release()
	}()
	_ = normalTableBuilder.AppendTimes(0, timesArrow)
	_ = normalTableBuilder.AppendTimes(1, startTimesArrow)
	_ = normalTableBuilder.AppendTimes(2, stopTimesArrow)
	_ = normalTableBuilder.AppendFloats(3, normalValuesArrow)
	_ = normalTableBuilder.AppendStrings(4, t1Arrow)
	_ = normalTableBuilder.AppendStrings(5, t2Arrow)

	NormalTable, _ = normalTableBuilder.Table()
}
