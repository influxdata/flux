package line_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/line"
	"github.com/influxdata/flux/mock"
)

func TestResultDecoder(t *testing.T) {
	tcs := []struct {
		name      string
		separator byte
		input     string
		want      *executetest.Result
	}{
		{
			name:      "newline separator",
			separator: '\n',
			input: `these are
raw
strings for an
awesome
line decoder!

empty line above
`,
			want: &executetest.Result{
				Nm: "_result",
				Tbls: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(0), "these are"},
						{execute.Time(1), "raw"},
						{execute.Time(2), "strings for an"},
						{execute.Time(3), "awesome"},
						{execute.Time(4), "line decoder!"},
						{execute.Time(5), ""},
						{execute.Time(6), "empty line above"},
					},
				}},
			},
		},
		{
			name:      "space separator",
			separator: ' ',
			input: `these are
raw
strings for an
awesome
line decoder!

empty line above
`,
			want: &executetest.Result{
				Nm: "_result",
				Tbls: []*executetest.Table{{
					ColMeta: []flux.ColMeta{
						{Label: "_time", Type: flux.TTime},
						{Label: "_value", Type: flux.TString},
					},
					Data: [][]interface{}{
						{execute.Time(0), "these"},
						{execute.Time(1), "are\nraw\nstrings"},
						{execute.Time(2), "for"},
						{execute.Time(3), "an\nawesome\nline"},
						{execute.Time(4), "decoder!\n\nempty"},
						{execute.Time(5), "line"},
						// "above" is not added because the reader doesn't find the separator at the end.
					},
				}},
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			decoder := line.NewResultDecoder(&line.ResultDecoderConfig{
				Separator:    tc.separator,
				TimeProvider: &mock.AscendingTimeProvider{},
			})

			r, err := decoder.Decode(bytes.NewReader([]byte(tc.input)))
			if err != nil {
				t.Fatal(err)
			}

			got := &executetest.Result{
				Nm: r.Name(),
			}
			err = r.Tables().Do(func(table flux.Table) error {
				ct, err := executetest.ConvertTable(table)
				if err != nil {
					return err
				}

				got.Tbls = append(got.Tbls, ct)
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}

			executetest.NormalizeTables(got.Tbls)
			executetest.NormalizeTables(tc.want.Tbls)

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected tables -want/+got\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
