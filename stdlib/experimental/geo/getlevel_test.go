package geo_test

import (
	"context"
	"strings"
	"testing"

	_ "github.com/InfluxCommunity/flux/fluxinit/static"
	"github.com/InfluxCommunity/flux/querytest"
	"github.com/InfluxCommunity/flux/stdlib/experimental/geo"
	"github.com/InfluxCommunity/flux/values"
)

func TestGetLevel_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "experimental/geo" geo.getLevel()`,
			WantErr: true, // missing required parameter(s)
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestGetLevel_Process(t *testing.T) {
	testCases := []struct {
		name         string
		token        string
		want         int64
		wantErr      bool
		errSubstring string
	}{
		{
			name:  "level 9",
			token: "89c28c",
			want:  9,
		},
		{
			name:         "invalid token",
			token:        "89c28",
			wantErr:      true,
			errSubstring: "invalid token specified",
		},
		{
			name:         "complete invalid token",
			token:        "%*^&(*^*%&$",
			wantErr:      true,
			errSubstring: "invalid token specified",
		},
	}

	for _, tc := range testCases {
		tc := tc
		getLevel := geo.Functions["getLevel"]
		result, err := getLevel.Call(context.Background(),
			values.NewObjectWithValues(map[string]values.Value{
				"token": values.NewString(tc.token),
			}),
		)
		if err != nil {
			if !tc.wantErr {
				t.Error(err.Error())
			}
			if tc.errSubstring != "" && !strings.Contains(err.Error(), tc.errSubstring) {
				t.Errorf("[%s] expected error with '%s', got '%v'", tc.name, tc.errSubstring, err)
			}
		} else if tc.want != result.Int() {
			t.Errorf("[%s] expected %v (%T), got %v (%T)", tc.name, tc.want, tc.want, result, result)
		}
	}
}
