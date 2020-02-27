package geo_test

import (
	"context"
	"strings"
	"testing"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/experimental/geo"
	"github.com/influxdata/flux/values"
)

func TestGetParent_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "no args",
			Raw:     `import "experimental/geo" geo.getParent()`,
			WantErr: true, // missing required parameter(s)
		},
		{
			Name:    "too few args",
			Raw:     `import "experimental/geo" geo.getParent(token: "89c284")`,
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

func TestGetParent_Process(t *testing.T) {
	testCases := []struct {
		name         string
		token        string
		level        int64
		want         string
		wantErr      bool
		errSubstring string
	}{
		{
			name:  "parent level 7",
			token: "89c284", // level 9
			level: 7,
			want:  "89c2c",
		},
		{
			name:         "wrong level",
			token:        "89c284", // level 9
			level:        11,       // cannot request higher level than source token
			wantErr:      true,
			errSubstring: "requested level greater then current level",
		},
		{
			name:         "invalid level",
			token:        "89c284", // level 9
			level:        0,
			wantErr:      true,
			errSubstring: "level value must be [1, 30]",
		},
		{
			name:         "invalid token",
			token:        "%*^&(*^*%&$",
			level:        7,
			wantErr:      true,
			errSubstring: "invalid token specified",
		},
	}

	for _, tc := range testCases {
		tc := tc
		getGrid := geo.Functions["getParent"]
		result, err := getGrid.Call(context.Background(),
			values.NewObjectWithValues(map[string]values.Value{
				"token": values.NewString(tc.token),
				"level": values.NewInt(tc.level),
			}),
		)
		if err != nil {
			if !tc.wantErr {
				t.Error(err.Error())
			}
			if tc.errSubstring != "" && !strings.Contains(err.Error(), tc.errSubstring) {
				t.Errorf("[%s] expected error with '%s', got '%v'", tc.name, tc.errSubstring, err)
			}
		} else if tc.want != result.Str() {
			t.Errorf("[%s] expected %v (%T), got %v (%T)", tc.name, tc.want, tc.want, result, result)
		}
	}
}
