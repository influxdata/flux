package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/spf13/cobra"
)

// decodeCSV will read a csv file with results into a mapping of table iterators.
func decodeCSV(name string) (map[string]flux.TableIterator, error) {
	r, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	dec := csv.NewMultiResultDecoder(csv.ResultDecoderConfig{})
	results, err := dec.Decode(r)
	if err != nil {
		return nil, err
	}
	defer results.Release()

	tables := make(map[string]flux.TableIterator)
	for results.More() {
		res := results.Next()

		var iter table.Iterator
		if err := res.Tables().Do(func(t flux.Table) error {
			cpy, err := execute.CopyTable(t)
			if err != nil {
				return err
			}
			iter = append(iter, cpy)
			return nil
		}); err != nil {
			return nil, err
		}
		tables[res.Name()] = iter
	}
	results.Release()

	if err := results.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

// prefix will prefix each line in the string with another string.
func prefix(s, prefix string) string {
	if prefix == "" {
		return s
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		sb.WriteString(prefix)
		sb.WriteString(": ")
		sb.WriteString(scanner.Text())
		sb.WriteString("\n")
	}
	return sb.String()
}

// diff will print out a diff between each table iterator.
func diff(want, got map[string]flux.TableIterator) {
	names := make([]string, 0, len(want))
	for name := range want {
		names = append(names, name)
	}
	sort.Strings(names)
	for name := range got {
		if i := sort.SearchStrings(names[:len(want)], name); i < len(want) && names[i] == name {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		diffS := table.Diff(want[name], got[name])
		if len(names) > 1 {
			diffS = prefix(diffS, name)
		}
		fmt.Println(diffS)
	}
}

func runE(cmd *cobra.Command, args []string) error {
	want, err := decodeCSV(args[0])
	if err != nil {
		return err
	}

	got, err := decodeCSV(args[1])
	if err != nil {
		return err
	}
	diff(want, got)
	return nil
}

func main() {
	cmd := &cobra.Command{
		Use:  "tablediff <file1> <file2>",
		RunE: runE,
		Args: cobra.ExactArgs(2),
	}
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
