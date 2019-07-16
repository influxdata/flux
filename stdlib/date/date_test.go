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
