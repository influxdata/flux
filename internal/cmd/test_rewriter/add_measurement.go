package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/internal/errors"
	"github.com/spf13/cobra"
)

var addMeasurementCmd = &cobra.Command{
	Use:   "add-measurement [test files...]",
	Short: "Update test inData and outData to have a _measurement column",
	RunE:  addMeasurementE,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(addMeasurementCmd)
	addMeasurementCmd.Flags().StringVar(&flagMeasurementName, "measurement-name", "m", "value to populate new column")
}

var (
	flagMeasurementName string
)

func addMeasurementE(cmd *cobra.Command, args []string) error {
	return doSubCommand(addMeasurementColumn, args)
}

func addMeasurementColumn(fileName string) error {
	astPkg, err := getFileAST(fileName)
	if err != nil {
		return err
	}

	df := &dataFinder{
		dataStmts: make(map[string]*ast.VariableAssignment),
	}
	ast.Walk(df, astPkg)

	if _, ok := df.dataStmts["inData"]; !ok {
		fmt.Printf("  No inData; skipping\n")
		return nil
	}

	for _, a := range df.dataStmts {
		lack, err := csvLacksMeasurementColumn(a)
		if err != nil {
			return err
		}
		if !lack {
			fmt.Printf("  %v has _measurement, skipping\n", a.ID.Name)
			return nil
		}
	}

	var q string
	{
		var sb strings.Builder
		sb.WriteString(`
import "csv"
import "experimental"
`)
		for _, a := range df.dataStmts {
			sb.WriteString(ast.Format(a) + "\n")
			sb.WriteString(fmt.Sprintf(`
csv.from(csv: %v)
  |> experimental.set(o: {_measurement: "%v"})
  |> experimental.group(mode: "extend", columns: ["_measurement"])
  |> yield(name: "%v")
`, a.ID.Name, flagMeasurementName, a.ID.Name))
		}
		q = sb.String()
	}

	ri, err := runQuery(q)
	if err != nil {
		return err
	}
	defer ri.Release()

	for ri.More() {
		r := ri.Next()

		var bb bytes.Buffer
		enc := csv.NewResultEncoder(csv.DefaultEncoderConfig())
		if _, err := enc.Encode(&bb, r); err != nil {
			return err
		}
		if err := replaceStringLitRHS(df.dataStmts[r.Name()], "\n"+bb.String()); err != nil {
			return err
		}
	}
	if ri.Err() != nil {
		return ri.Err()
	}

	if err := rewriteFile(fileName, astPkg); err != nil {
		return nil
	}
	fmt.Printf("  Rewrote %s with measurement columns added.\n", fileName)
	return nil
}

func csvLacksMeasurementColumn(a *ast.VariableAssignment) (bool, error) {
	bb := bytes.NewBuffer([]byte(a.Init.(*ast.StringLiteral).Value))
	dec := csv.NewResultDecoder(csv.ResultDecoderConfig{})
	r, err := dec.Decode(bb)
	if err != nil {
		return false, err
	}
	lacks := true
	if err := r.Tables().Do(func(t flux.Table) error {
		for _, c := range t.Cols() {
			if c.Label == "_measurement" {
				lacks = false
				break
			}
		}
		return nil
	}); err != nil {
		return false, err
	}
	return lacks, nil
}

func replaceStringLitRHS(va *ast.VariableAssignment, v string) error {
	sl, ok := va.Init.(*ast.StringLiteral)
	if !ok {
		return errors.New(codes.Invalid, "funky assignment")
	}
	sl.Value = v
	sl.Loc = nil
	return nil
}

type dataFinder struct {
	dataStmts map[string]*ast.VariableAssignment
}

func (d *dataFinder) Visit(node ast.Node) ast.Visitor {
	return d
}

func (d *dataFinder) Done(node ast.Node) {
	switch n := node.(type) {
	case *ast.VariableAssignment:
		if n.ID.Name == "inData" || n.ID.Name == "outData" {
			d.dataStmts[n.ID.Name] = n
		}
	}
}
