package opsgenie_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/stdlib/contrib/sranka/opsgenie"
	"github.com/influxdata/flux/values"
)

func TestRespondersToJSON(t *testing.T) {
	testCases := []struct {
		name  string // test name
		input []string
		want  []map[string]string
		error string
	}{
		{
			name:  "user",
			input: []string{"user:a"},
			want:  []map[string]string{{"type": "user", "username": "a"}},
		},
		{
			name:  "team",
			input: []string{"team:x"},
			want:  []map[string]string{{"type": "team", "name": "x"}},
		},
		{
			name:  "escalation",
			input: []string{"escalation:x"},
			want:  []map[string]string{{"type": "escalation", "name": "x"}},
		},
		{
			name:  "schedule",
			input: []string{"schedule:y"},
			want:  []map[string]string{{"type": "schedule", "name": "y"}},
		},
		{
			name:  "error",
			input: []string{"test:y"},
			error: "unsupported",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fluxValues := make([]values.Value, len(tc.input))
			for i, v := range tc.input {
				fluxValues[i] = values.NewString(v)
			}
			args := interpreter.NewArguments(values.NewObjectWithValues(
				map[string]values.Value{
					"v": values.NewArrayWithBacking(semantic.NewArrayType(semantic.BasicString), fluxValues),
				}),
			)

			got, err := opsgenie.RespondersToJSON(args)
			if err != nil {
				if tc.error == "" || !strings.Contains(err.Error(), tc.error) {
					t.Fatal(err)
				}
				return
			}
			output := make([]map[string]string, 0, 5)
			if err := json.Unmarshal([]byte(got.Str()), &output); err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(tc.want, output) {
				t.Fatalf("unexpected details -want/+got\n\n%s\n\n", cmp.Diff(tc.want, output))
			}
		})
	}

	t.Run("missing required argument", func(t *testing.T) {
		args := interpreter.NewArguments(values.NewObjectWithValues(
			map[string]values.Value{}),
		)

		_, err := opsgenie.RespondersToJSON(args)
		if err == nil {
			t.Fatal("error expected, but none received")
		}
		if !strings.Contains(err.Error(), "missing required") {
			t.Fatal(err)
		}
	})
}
