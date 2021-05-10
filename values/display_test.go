package values_test

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func TestDisplay(t *testing.T) {
	testCases := []struct {
		value   values.Value
		display string
	}{
		{
			value:   values.NewNull(semantic.BasicInt),
			display: "<null>",
		},
		{
			value:   values.NewString("hi"),
			display: "hi",
		},
		{
			value:   values.NewBytes([]byte{10, 11, 12}),
			display: "[10 11 12]",
		},
		{
			value:   values.NewInt(1),
			display: "1",
		},
		{
			value:   values.NewUInt(1),
			display: "1",
		},
		{
			value:   values.NewFloat(1.1),
			display: "1.1",
		},
		{
			value:   values.NewBool(true),
			display: "true",
		},
		{
			value:   values.NewTime(values.ConvertTime(time.Date(2020, 4, 12, 1, 2, 3, 4, time.UTC))),
			display: "2020-04-12T01:02:03.000000004Z",
		},
		{
			value:   values.NewDuration(values.ConvertDurationNsecs(time.Minute + time.Second)),
			display: "1m1s",
		},
		{
			value:   values.NewRegexp(regexp.MustCompile(".*")),
			display: ".*",
		},
		{
			value: values.NewArrayWithBacking(
				semantic.NewArrayType(semantic.BasicInt),
				[]values.Value{
					values.NewInt(1),
					values.NewInt(2),
					values.NewInt(3),
				},
			),
			display: "[1, 2, 3]",
		},
		{
			value: values.NewArrayWithBacking(
				semantic.NewArrayType(semantic.BasicInt),
				[]values.Value{
					values.NewInt(1),
					values.NewInt(2),
					values.NewInt(3),
					values.NewInt(4),
				},
			),
			display: "[\n    1, \n    2, \n    3, \n    4\n]",
		},
		{
			value: values.NewObjectWithValues(
				map[string]values.Value{
					"a": values.NewInt(1),
					"b": values.NewInt(1),
					"c": values.NewInt(1),
				},
			),
			display: "{a: 1, b: 1, c: 1}",
		},
		{
			value: values.NewObjectWithValues(
				map[string]values.Value{
					"a": values.NewInt(1),
					"b": values.NewInt(1),
					"c": values.NewInt(1),
					"d": values.NewInt(1),
				},
			),
			display: "{\n    a: 1, \n    b: 1, \n    c: 1, \n    d: 1\n}",
		},
		{
			value:   values.NewEmptyDict(semantic.NewDictType(semantic.BasicInt, semantic.BasicInt)),
			display: "[:]",
		},
		{
			value: func() values.Value {
				b := values.NewDictBuilder(semantic.NewDictType(semantic.BasicInt, semantic.BasicInt))
				b.Insert(values.NewInt(1), values.NewInt(0))
				b.Insert(values.NewInt(2), values.NewInt(0))
				b.Insert(values.NewInt(3), values.NewInt(0))
				return b.Dict()
			}(),
			display: "[1: 0, 2: 0, 3: 0]",
		},
		{
			value: func() values.Value {
				b := values.NewDictBuilder(semantic.NewDictType(semantic.BasicInt, semantic.BasicInt))
				b.Insert(values.NewInt(1), values.NewInt(0))
				b.Insert(values.NewInt(2), values.NewInt(0))
				b.Insert(values.NewInt(3), values.NewInt(0))
				b.Insert(values.NewInt(4), values.NewInt(0))
				return b.Dict()
			}(),
			display: "[\n    1: 0, \n    2: 0, \n    3: 0, \n    4: 0\n]",
		},
		{
			value: values.NewFunction(
				"foo",
				semantic.NewFunctionType(semantic.BasicInt, nil),
				func(ctx context.Context, args values.Object) (values.Value, error) {
					return values.NewInt(1), nil
				},
				false,
			),
			display: "() => int",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.display, func(t *testing.T) {
			b := strings.Builder{}
			err := values.Display(&b, tc.value)
			if err != nil {
				t.Fatal(err)
			}
			got := b.String()
			if tc.display != got {
				t.Errorf("unexpected display strings diff: %s", cmp.Diff(tc.display, got))
			}
		})
	}
}
