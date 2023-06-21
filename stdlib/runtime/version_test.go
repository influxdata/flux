//go:build go1.12
// +build go1.12

package runtime_test

import (
	"context"
	"runtime/debug"
	"testing"

	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/dependencies/dependenciestest"
	"github.com/InfluxCommunity/flux/dependency"
	"github.com/InfluxCommunity/flux/internal/errors"
	"github.com/InfluxCommunity/flux/stdlib/runtime"
	"github.com/InfluxCommunity/flux/values"
	"github.com/google/go-cmp/cmp"
)

func TestVersion(t *testing.T) {
	for _, tt := range []struct {
		name    string
		bi      *debug.BuildInfo
		want    values.Value
		wantErr error
	}{
		{
			name: "main module",
			bi: &debug.BuildInfo{
				Path: "github.com/InfluxCommunity/flux",
				Main: debug.Module{
					Path:    "github.com/InfluxCommunity/flux",
					Version: "v0.38.0",
				},
			},
			want: values.NewString("v0.38.0"),
		},
		{
			name: "replaced main module",
			bi: &debug.BuildInfo{
				Path: "github.com/InfluxCommunity/flux",
				Main: debug.Module{
					Path:    "github.com/InfluxCommunity/flux",
					Version: "v0.38.0",
					Replace: &debug.Module{
						Path:    "github.com/InfluxCommunity/flux",
						Version: "(devel)",
					},
				},
			},
			want: values.NewString("(devel)"),
		},
		{
			name: "dependency module",
			bi: &debug.BuildInfo{
				Path: "github.com/influxdata/influxdb",
				Main: debug.Module{
					Path:    "github.com/influxdata/influxdb",
					Version: "v2.0.0",
				},
				Deps: []*debug.Module{
					{
						Path:    "github.com/InfluxCommunity/flux",
						Version: "v0.38.0",
					},
				},
			},
			want: values.NewString("v0.38.0"),
		},
		{
			name: "replaced dependency module",
			bi: &debug.BuildInfo{
				Path: "github.com/influxdata/influxdb",
				Main: debug.Module{
					Path:    "github.com/influxdata/influxdb",
					Version: "v2.0.0",
				},
				Deps: []*debug.Module{
					{
						Path:    "github.com/InfluxCommunity/flux",
						Version: "v0.38.0",
						Replace: &debug.Module{
							Path: "github.com/InfluxCommunity/flux",
						},
					},
				},
			},
			want: values.NewString(""),
		},
		{
			name:    "build info not present",
			wantErr: errors.New(codes.NotFound, "build info is not present"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runtime.SetBuildInfo(tt.bi)

			ctx, deps := dependency.Inject(context.Background(), dependenciestest.Default())
			defer deps.Finish()
			got, err := runtime.Version(ctx, nil)
			if err != nil {
				if tt.wantErr != nil {
					if !cmp.Equal(tt.wantErr, err) {
						t.Fatalf("unexpected error -want/+got:\n%s", cmp.Diff(tt.wantErr, err))
					}
					return
				} else {
					t.Fatalf("unexpected error: %s", err)
				}
			}

			if !got.Equal(tt.want) {
				t.Fatalf("unexpected value -want/+got:\n\t- %v\n\t+ %v", tt.want, got)
			}
		})
	}
}
