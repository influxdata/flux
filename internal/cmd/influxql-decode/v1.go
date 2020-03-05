package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/influxdata/flux/influxql"
	"github.com/spf13/cobra"
)

func v1(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		f, err := os.Open(arg)
		if err != nil {
			return err
		}

		var resp influxql.Response
		if err := json.NewDecoder(f).Decode(&resp); err != nil {
			return err
		}

		b := new(bytes.Buffer)
		for _, res := range resp.Results {
			for _, s := range res.Series {
				seriesHeader := s.Name
				var tags []string
				for k, v := range s.Tags {
					tags = append(tags, fmt.Sprintf("%s=%s", k, v))
				}
				timeCol := -1
				for i, col := range s.Columns {
					if col == "time" {
						timeCol = i
						break
					}
				}
				if len(tags) > 0 {
					seriesHeader += "," + strings.Join(tags, ",")
				}
				seriesHeader += " "
				for _, row := range s.Values {
					b.WriteString(seriesHeader)
					for i, v := range row {
						if i == timeCol {
							continue
						}
						if _, ok := v.(string); ok {
							b.WriteString(fmt.Sprintf("%s=\"%s\"", s.Columns[i], v))
						} else {
							b.WriteString(fmt.Sprintf("%s=%v", s.Columns[i], v))
						}
						if i < len(s.Columns)-1 {
							b.WriteString(",")
						}
					}
					if timeCol >= 0 {
						ts, err := time.Parse(time.RFC3339Nano, row[timeCol].(string))
						if err != nil {
							return err
						}
						b.WriteString(fmt.Sprintf(" %d", ts.UnixNano()))
					}
					b.WriteString("\n")
				}
			}
		}
		fmt.Print(b.String())
	}
	return nil
}
