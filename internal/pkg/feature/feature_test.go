package feature_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux/internal/pkg/feature"
)

func Test_feature(t *testing.T) {
	cases := []struct {
		name     string
		flag     feature.Flag
		values   map[string]interface{}
		expected interface{}
	}{
		{
			name: "bool happy path",
			flag: newFlag("test", false),
			values: map[string]interface{}{
				"test": true,
			},
			expected: true,
		},
		{
			name: "int happy path",
			flag: newFlag("test", 0),
			values: map[string]interface{}{
				"test": int32(42),
			},
			expected: int32(42),
		},
		{
			name: "float happy path",
			flag: newFlag("test", 0.0),
			values: map[string]interface{}{
				"test": 42.42,
			},
			expected: 42.42,
		},
		{
			name: "string happy path",
			flag: newFlag("test", ""),
			values: map[string]interface{}{
				"test": "restaurantattheendoftheuniverse",
			},
			expected: "restaurantattheendoftheuniverse",
		},
		{
			name:     "bool missing use default",
			flag:     newFlag("test", false),
			expected: false,
		},
		{
			name:     "bool missing use default true",
			flag:     newFlag("test", true),
			expected: true,
		},
		{
			name:     "int missing use default",
			flag:     newFlag("test", 65),
			expected: int32(65),
		},
		{
			name:     "float missing use default",
			flag:     newFlag("test", 65.65),
			expected: 65.65,
		},
		{
			name:     "string missing use default",
			flag:     newFlag("test", "mydefault"),
			expected: "mydefault",
		},

		{
			name: "bool invalid use default",
			flag: newFlag("test", true),
			values: map[string]interface{}{
				"test": "notabool",
			},
			expected: true,
		},
		{
			name: "int invalid use default",
			flag: newFlag("test", 42),
			values: map[string]interface{}{
				"test": 99.99,
			},
			expected: int32(42),
		},
		{
			name: "float invalid use default",
			flag: newFlag("test", 42.42),
			values: map[string]interface{}{
				"test": 99,
			},
			expected: 42.42,
		},
		{
			name: "string invalid use default",
			flag: newFlag("test", "restaurantattheendoftheuniverse"),
			values: map[string]interface{}{
				"test": true,
			},
			expected: "restaurantattheendoftheuniverse",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			flagger := testFlagsFlagger{
				m: test.values,
			}
			ctx := feature.Inject(context.Background(), flagger)

			var actual interface{}
			switch flag := test.flag.(type) {
			case feature.BoolFlag:
				actual = flag.Enabled(ctx)
			case feature.FloatFlag:
				actual = flag.Float(ctx)
			case feature.IntFlag:
				actual = flag.Int(ctx)
			case feature.StringFlag:
				actual = flag.String(ctx)
			default:
				t.Errorf("unknown flag type %T (%#v)", flag, flag)
			}

			if actual != test.expected {
				t.Errorf("unexpected flag value: got %v, want %v", actual, test.expected)
			}
		})
	}
}

type testFlagsFlagger struct {
	m   map[string]interface{}
}

func (f testFlagsFlagger) FlagValue(ctx context.Context, flag feature.Flag) interface{} {
	v, ok := f.m[flag.Key()]
	if !ok {
		v = flag.Default()
	}
	return v
}

func newFlag(key string, defaultValue interface{}) feature.Flag {
	return feature.MakeFlag(key, key, "", defaultValue)
}
