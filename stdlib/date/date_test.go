package date

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/values"
)

func TestTimeFns_Time(t *testing.T) {
	testCases := []struct {
		name string
		time string
		want int64
	}{
		{
			name: "second",
			time: "2019-06-03T13:59:01.000000000Z",
			want: 1,
		},
		{
			name: "minute",
			time: "2019-06-03T13:59:01.000000000Z",
			want: 59,
		},
		{
			name: "hour",
			time: "2019-06-03T13:59:01.000000000Z",
			want: 13,
		},
		{
			name: "weekDay",
			time: "2019-06-03T13:59:01.000000000Z",
			want: 1,
		},
		{
			name: "monthDay",
			time: "2019-06-03T13:59:01.000000000Z",
			want: 3,
		},
		{
			name: "yearDay",
			time: "2019-06-03T13:59:01.000000000Z",
			want: 154,
		},
		{
			name: "month",
			time: "2019-06-03T13:59:01.000000000Z",
			want: 6,
		},
		{
			name: "year",
			time: "2019-06-03T13:59:01.000000000Z",
			want: 2019,
		},
		{
			name: "week",
			time: "2019-06-03T13:59:01.000000000Z",
			want: 23,
		},
		{
			name: "quarter",
			time: "2019-06-03T13:59:01.000000000Z",
			want: 2,
		},
		{
			name: "millisecond",
			time: "2019-06-03T13:59:01.123456789Z",
			want: 123,
		},
		{
			name: "microsecond",
			time: "2019-06-03T13:59:01.123456789Z",
			want: 123456,
		},
		{
			name: "nanosecond",
			time: "2019-06-03T13:59:01.123456789Z",
			want: 123456789,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxFn := SpecialFns[tc.name]
			time, err := values.ParseTime(tc.time)
			if err != nil {
				t.Fatal(err)
			}
			fluxArg := values.NewObjectWithValues(map[string]values.Value{
				"t": values.NewTime(time),
				"location": values.NewObjectWithValues(map[string]values.Value{
					"zone":   values.NewString("UTC"),
					"offset": values.NewDuration(values.Duration{}),
				}),
			})
			got, err := fluxFn.Call(dependenciestest.InjectAllDeps(context.Background()), fluxArg)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want != got.Int() {
				t.Errorf("input %v: expected %v, got %f", time, tc.want, got)
			}
		})
	}
}
func TestTimeFns_Duration(t *testing.T) {
	now, err := values.ParseTime("2020-01-01T17:14:39.123456789Z")
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		name string
		time string
		want int64
	}{
		{
			name: "second",
			time: "1s",
			want: 40,
		},
		{
			name: "second",
			time: "2mo",
			want: 39,
		},
		{
			name: "minute",
			time: "5m",
			want: 19,
		},
		{
			name: "minute",
			time: "2mo",
			want: 14,
		},
		{
			name: "hour",
			time: "-5h",
			want: 12,
		},
		{
			name: "hour",
			time: "2mo",
			want: 17,
		},
		{
			name: "weekDay",
			time: "1w",
			want: 3,
		},
		{
			name: "weekDay",
			time: "2mo",
			want: 0,
		},
		{
			name: "monthDay",
			time: "-1d",
			want: 31,
		},
		{
			name: "monthDay",
			time: "2mo",
			want: 1,
		},
		{
			name: "yearDay",
			time: "1d",
			want: 2,
		},
		{
			name: "yearDay",
			time: "2mo",
			want: 61,
		},
		{
			name: "month",
			time: "-1mo",
			want: 12,
		},
		{
			name: "month",
			time: "2mo",
			want: 3,
		},
		{
			name: "year",
			time: "-1y",
			want: 2019,
		},
		{
			name: "year",
			time: "10y",
			want: 2030,
		},
		{
			name: "week",
			time: "1w",
			want: 2,
		},
		{
			name: "week",
			time: "2mo",
			want: 9,
		},
		{
			name: "quarter",
			time: "3mo",
			want: 2,
		},
		{
			name: "quarter",
			time: "2mo",
			want: 1,
		},
		{
			name: "millisecond",
			time: "1ms",
			want: 124,
		},
		{
			name: "millisecond",
			time: "2mo",
			want: 123,
		},
		{
			name: "microsecond",
			time: "10us",
			want: 123466,
		},
		{
			name: "microsecond",
			time: "2mo",
			want: 123456,
		},
		{
			name: "nanosecond",
			time: "-10ns",
			want: 123456779,
		},
		{
			name: "nanosecond",
			time: "2mo",
			want: 123456789,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxFn := SpecialFns[tc.name]
			time, err := values.ParseDuration(tc.time)
			if err != nil {
				t.Fatal(err)
			}
			fluxArg := values.NewObjectWithValues(map[string]values.Value{
				"t": values.NewDuration(time),
				"location": values.NewObjectWithValues(map[string]values.Value{
					"zone":   values.NewString("UTC"),
					"offset": values.NewDuration(values.Duration{}),
				}),
			})

			//Setup deps with specific now time
			deps := dependenciestest.Default()
			execDeps := dependenciestest.ExecutionDefault()
			nowVar := now.Time()
			execDeps.Now = &nowVar
			ctx := deps.Inject(context.Background())
			ctx = execDeps.Inject(ctx)
			got, err := fluxFn.Call(ctx, fluxArg)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want != got.Int() {
				t.Errorf("input %v: expected %v, got %f", time, tc.want, got)
			}
		})
	}
}

func TestNilErrors(t *testing.T) {
	testCases := []string{
		"second",
		"minute",
		"hour",
		"weekDay",
		"monthDay",
		"yearDay",
		"month",
		"year",
		"week",
		"quarter",
		"millisecond",
		"microsecond",
		"nanosecond",
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			fluxFn := SpecialFns[tc]
			fluxArg := values.NewObjectWithValues(map[string]values.Value{"t": values.NewString("")})
			fluxArg.Set("t", nil)
			_, err := fluxFn.Call(dependenciestest.InjectAllDeps(context.Background()), fluxArg)
			if err == nil {
				t.Errorf("%s did not error", tc)
			}
			if err.Error() != "argument t was nil" {
				t.Errorf("expected: argument t was nil, got: %v", err.Error())
			}
		})
	}
}

func TestTruncate_Time(t *testing.T) {
	testCases := []struct {
		name string
		time string
		unit string
		want string
	}{
		{
			name: "second",
			time: "2019-06-03T13:59:01.000000000Z",
			unit: "1s",
			want: "2019-06-03T13:59:01.000000000Z",
		},
		{
			name: "minute",
			time: "2019-06-03T13:59:01.000000000Z",
			unit: "1m",
			want: "2019-06-03T13:59:00.000000000Z",
		},
		{
			name: "hour",
			time: "2019-06-03T13:59:01.000000000Z",
			unit: "1h",
			want: "2019-06-03T13:00:00.000000000Z",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxFn := SpecialFns["truncate"]
			time, err := values.ParseTime(tc.time)
			if err != nil {
				t.Fatal(err)
			}
			unit, err := values.ParseDuration(tc.unit)
			if err != nil {
				t.Fatal(err)
			}
			fluxArg := values.NewObjectWithValues(map[string]values.Value{
				"t":    values.NewTime(time),
				"unit": values.NewDuration(unit),
				"location": values.NewObjectWithValues(map[string]values.Value{
					"zone":   values.NewString("UTC"),
					"offset": values.NewDuration(values.Duration{}),
				}),
			})
			got, err := fluxFn.Call(dependenciestest.InjectAllDeps(context.Background()), fluxArg)
			if err != nil {
				t.Fatal(err)
			}

			wanted, err := values.ParseTime(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			if wanted != got.Time() {
				t.Errorf("input %v: expected %v, got %v", time, wanted, got.Time())
			}
		})
	}
}
func TestTruncate_Duration(t *testing.T) {
	now, err := values.ParseTime("2019-06-03T13:59:01.000000000Z")
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		name string
		time string
		unit string
		want string
	}{
		{
			name: "second",
			time: "-5mo",
			unit: "1s",
			want: "2019-01-03T13:59:01.000000000Z",
		},
		{
			name: "minute",
			time: "-5mo",
			unit: "1m",
			want: "2019-01-03T13:59:00.000000000Z",
		},
		{
			name: "hour",
			time: "-5mo",
			unit: "1h",
			want: "2019-01-03T13:00:00.000000000Z",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxFn := SpecialFns["truncate"]
			time, err := values.ParseDuration(tc.time)
			if err != nil {
				t.Fatal(err)
			}
			unit, err := values.ParseDuration(tc.unit)
			if err != nil {
				t.Fatal(err)
			}
			fluxArg := values.NewObjectWithValues(map[string]values.Value{
				"t":    values.NewDuration(time),
				"unit": values.NewDuration(unit),
				"location": values.NewObjectWithValues(map[string]values.Value{
					"zone":   values.NewString("UTC"),
					"offset": values.NewDuration(values.Duration{}),
				}),
			})

			//Setup deps with specific now time
			deps := dependenciestest.Default()
			execDeps := dependenciestest.ExecutionDefault()
			nowVar := now.Time()
			execDeps.Now = &nowVar
			ctx := deps.Inject(context.Background())
			ctx = execDeps.Inject(ctx)

			got, err := fluxFn.Call(ctx, fluxArg)
			if err != nil {
				t.Fatal(err)
			}

			wanted, err := values.ParseTime(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			if wanted != got.Time() {
				t.Errorf("input %v: expected %v, got %v", time, wanted, got.Time())
			}
		})
	}
}

func TestTruncateNilErrors(t *testing.T) {
	tc := struct {
		name string
		time string
		unit string
		want string
	}{
		name: "second",
		time: "2019-06-03T13:59:01.000000000Z",
		unit: "1s",
		want: "2019-06-03T13:59:01.000000000Z",
	}
	t.Run(tc.name, func(t *testing.T) {
		fluxFn := SpecialFns["truncate"]
		time, err := values.ParseTime(tc.time)
		if err != nil {
			t.Fatal(err)
		}
		unit, err := values.ParseDuration(tc.unit)
		if err != nil {
			t.Fatal(err)
		}
		fluxArg := values.NewObjectWithValues(map[string]values.Value{"t": values.NewTime(time), "unit": values.NewDuration(unit)})
		fluxArg.Set("t", nil)
		_, err = fluxFn.Call(dependenciestest.InjectAllDeps(context.Background()), fluxArg)
		if err == nil {
			t.Errorf("%s did not error", tc)
		}
		if err.Error() != "argument t was nil" {
			t.Errorf("expected: argument t was nil, got: %v", err.Error())
		}
	})
}

func TestTimeFnsWithLocationZone(t *testing.T) {
	const (
		US_Pacific     = "America/Los_Angeles"
		Australia_East = "Australia/Sydney"
	)
	testCases := []struct {
		name    string
		locname string
		time    string
		want    int64
	}{
		{
			name:    "minute",
			locname: Australia_East,
			time:    "2019-06-03T04:59:00.000000000Z",
			want:    59,
		},
		{
			name:    "hour",
			locname: US_Pacific,
			time:    "2019-06-03T08:00:00.000000000Z",
			want:    1,
		},
		{
			name:    "weekDay",
			locname: US_Pacific,
			time:    "2011-12-30T18:00:00.000000000Z",
			want:    5,
		},
		{
			name:    "monthDay",
			locname: US_Pacific,
			time:    "2019-06-03T13:59:01.000000000Z",
			want:    3,
		},
		{
			name:    "yearDay",
			locname: Australia_East,
			time:    "2019-06-03T23:00:00.000000000Z",
			want:    155,
		},
		{
			name:    "month",
			locname: Australia_East,
			time:    "2019-06-03T13:59:01.000000000Z",
			want:    6,
		},
		{
			name:    "year",
			locname: Australia_East,
			time:    "2011-12-31T23:59:00.000000000Z",
			want:    2012,
		},
		{
			name:    "week",
			locname: Australia_East,
			time:    "2011-12-31T23:59:00.000000000Z",
			want:    52,
		},
		{
			name:    "quarter",
			locname: Australia_East,
			time:    "2019-03-31T23:00:00.000000000Z",
			want:    2,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxFn := SpecialFns[tc.name]
			time, err := values.ParseTime(tc.time)
			if err != nil {
				t.Fatal(err)
			}
			fluxArg := values.NewObjectWithValues(map[string]values.Value{
				"t": values.NewTime(time),
				"location": values.NewObjectWithValues(map[string]values.Value{
					"zone":   values.NewString(tc.locname),
					"offset": values.NewDuration(values.Duration{}),
				}),
			})
			got, err := fluxFn.Call(dependenciestest.InjectAllDeps(context.Background()), fluxArg)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want != got.Int() {
				t.Errorf("input %v: expected %v, got %f", time, tc.want, got)
			}
		})
	}
}

func TestTimeFnsWithLocationOffset(t *testing.T) {
	testCases := []struct {
		name   string
		offset values.Duration
		time   string
		want   int64
	}{
		{
			name:   "minute",
			offset: values.ConvertDurationNsecs(time.Minute),
			time:   "2019-06-03T13:59:01.000000000Z",
			want:   0,
		},
		{
			name:   "hour",
			offset: values.ConvertDurationNsecs(time.Hour),
			time:   "2019-06-03T13:59:01.000000000Z",
			want:   14,
		},
		{
			name:   "weekDay",
			offset: values.ConvertDurationNsecs(24 * time.Hour),
			time:   "2011-12-30T18:00:00.000000000Z",
			want:   6,
		},
		{
			name:   "monthDay",
			offset: values.ConvertDurationNsecs(24 * time.Hour),
			time:   "2019-06-03T13:59:01.000000000Z",
			want:   4,
		},
		{
			name:   "yearDay",
			offset: values.ConvertDurationNsecs(24 * time.Hour),
			time:   "2019-06-03T13:59:01.000000000Z",
			want:   155,
		},
		{
			name:   "month",
			offset: values.ConvertDurationNsecs(48 * time.Hour),
			time:   "2019-06-30T13:59:01.000000000Z",
			want:   7,
		},
		{
			name:   "year",
			offset: values.ConvertDurationNsecs(-24 * time.Hour),
			time:   "2019-01-01T13:59:01.000000000Z",
			want:   2018,
		},
		{
			name:   "week",
			offset: values.ConvertDurationNsecs(48 * time.Hour),
			time:   "2011-12-31T23:59:00.000000000Z",
			want:   1,
		},
		{
			name:   "quarter",
			offset: values.ConvertDurationNsecs(-5 * time.Minute),
			time:   "2019-04-01T00:01:00.000000000Z",
			want:   1,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxFn := SpecialFns[tc.name]
			time, err := values.ParseTime(tc.time)
			if err != nil {
				t.Fatal(err)
			}
			fluxArg := values.NewObjectWithValues(map[string]values.Value{
				"t": values.NewTime(time),
				"location": values.NewObjectWithValues(map[string]values.Value{
					"zone":   values.NewString("UTC"),
					"offset": values.NewDuration(tc.offset),
				}),
			})
			got, err := fluxFn.Call(dependenciestest.InjectAllDeps(context.Background()), fluxArg)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want != got.Int() {
				t.Errorf("input %v: expected %v, got %f", time, tc.want, got)
			}
		})
	}
}

func TestTimeFnsWithLocationZoneOffset(t *testing.T) {
	const (
		US_Pacific     = "America/Los_Angeles"
		Australia_East = "Australia/Sydney"
	)
	testCases := []struct {
		name    string
		locname string
		offset  values.Duration
		time    string
		want    int64
	}{
		{
			name:    "minute",
			locname: Australia_East,
			offset:  values.ConvertDurationNsecs(time.Minute),
			time:    "2019-06-03T04:59:00.000000000Z",
			want:    0,
		},
		{
			name:    "hour",
			locname: US_Pacific,
			offset:  values.ConvertDurationNsecs(time.Hour),
			time:    "2019-06-03T08:00:00.000000000Z",
			want:    2,
		},
		{
			name:    "weekDay",
			locname: US_Pacific,
			offset:  values.ConvertDurationNsecs(24 * time.Hour),
			time:    "2011-12-30T18:00:00.000000000Z",
			want:    6,
		},
		{
			name:    "monthDay",
			locname: US_Pacific,
			offset:  values.ConvertDurationNsecs(24 * time.Hour),
			time:    "2019-06-03T13:59:01.000000000Z",
			want:    4,
		},
		{
			name:    "yearDay",
			locname: Australia_East,
			offset:  values.ConvertDurationNsecs(24 * time.Hour),
			time:    "2019-06-03T23:00:00.000000000Z",
			want:    156,
		},
		{
			name:    "month",
			locname: Australia_East,
			offset:  values.ConvertDurationNsecs(48 * time.Hour),
			time:    "2019-06-03T13:59:01.000000000Z",
			want:    6,
		},
		{
			name:    "year",
			locname: Australia_East,
			offset:  values.ConvertDurationNsecs(-24 * time.Hour),
			time:    "2011-12-31T23:59:00.000000000Z",
			want:    2011,
		},
		{
			name:    "week",
			locname: Australia_East,
			offset:  values.ConvertDurationNsecs(48 * time.Hour),
			time:    "2011-12-31T23:59:00.000000000Z",
			want:    1,
		},
		{
			name:    "quarter",
			locname: Australia_East,
			offset:  values.ConvertDurationNsecs(-5 * time.Minute),
			time:    "2019-03-31T23:00:00.000000000Z",
			want:    2,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxFn := SpecialFns[tc.name]
			time, err := values.ParseTime(tc.time)
			if err != nil {
				t.Fatal(err)
			}
			fluxArg := values.NewObjectWithValues(map[string]values.Value{
				"t": values.NewTime(time),
				"location": values.NewObjectWithValues(map[string]values.Value{
					"zone":   values.NewString(tc.locname),
					"offset": values.NewDuration(tc.offset),
				}),
			})
			got, err := fluxFn.Call(dependenciestest.InjectAllDeps(context.Background()), fluxArg)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want != got.Int() {
				t.Errorf("input %v: expected %v, got %f", time, tc.want, got)
			}
		})
	}
}
