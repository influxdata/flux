package date

import (
	"testing"

	"github.com/influxdata/flux/values"
)

func TestTimeFns(t *testing.T) {
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
			fluxArg := values.NewObjectWithValues(map[string]values.Value{"t": values.NewTime(time)})
			got, err := fluxFn.Call(fluxArg)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want != got.Int() {
				t.Errorf("input %v: expected %v, got %f", time, tc.want, got)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
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
			fluxArg := values.NewObjectWithValues(map[string]values.Value{"t": values.NewTime(time), "unit": values.NewDuration(values.Duration(unit))})
			got, err := fluxFn.Call(fluxArg)
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
