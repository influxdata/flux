package feature_test

import (
	"context"
	"testing"

	"github.com/InfluxCommunity/flux/internal/pkg/feature"
	"github.com/google/go-cmp/cmp"
)

type flagger map[string]interface{}

func (f flagger) FlagValue(ctx context.Context, flag feature.Flag) interface{} {
	v, ok := f[flag.Key()]
	if !ok {
		return flag.Default()
	}
	return v
}

type metrics map[string]interface{}

func (m metrics) Inc(key string, value interface{}) {
	m[key] = value
}

func TestMetrics(t *testing.T) {
	for _, tt := range []struct {
		name    string
		flagger flagger
		flags   []feature.Flag
		want    metrics
	}{
		{
			name: "normal",
			flagger: flagger{
				"a": true,
			},
			flags: []feature.Flag{
				feature.MakeBoolFlag("A", "a", "", false),
				feature.MakeBoolFlag("B", "b", "", false),
			},
			want: metrics{
				"a": true,
				"b": false,
			},
		},
		{
			name: "mistyped",
			flagger: flagger{
				"a": "true",
			},
			flags: []feature.Flag{
				feature.MakeBoolFlag("A", "a", "", false),
			},
			want: metrics{
				"a": false,
			},
		},
	} {
		// Note: You cannot use t.Parallel() with this
		// because it modifies global state.
		t.Run(tt.name, func(t *testing.T) {
			ctx := feature.Inject(context.Background(), tt.flagger)

			got := metrics{}
			feature.SetMetrics(got)
			defer feature.SetMetrics(nil)

			for _, flag := range tt.flags {
				switch f := flag.(type) {
				case feature.BoolFlag:
					_ = f.Enabled(ctx)
				case feature.IntFlag:
					_ = f.Int(ctx)
				case feature.FloatFlag:
					_ = f.Float(ctx)
				case feature.StringFlag:
					_ = f.String(ctx)
				default:
					panic("unreachable")
				}
			}

			if !cmp.Equal(tt.want, got) {
				t.Errorf("unexpected metrics -want/+got\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}
