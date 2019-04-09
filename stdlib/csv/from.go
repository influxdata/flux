package csv

import (
	"fmt"
	"io/ioutil"
	"os"

	"context"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/pkg/errors"
)

const FromCSVKind = "fromCSV"

type FromCSVOpSpec struct {
	CSV  string `json:"csv"`
	File string `json:"file"`
}

func init() {
	fromCSVSignature := semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"csv":  semantic.String,
			"file": semantic.String,
		},
		Required: nil,
		Return:   flux.TableObjectType,
	}
	flux.RegisterPackageValue("csv", "from", flux.FunctionValue(FromCSVKind, createFromCSVOpSpec, fromCSVSignature))
	flux.RegisterOpSpec(FromCSVKind, newFromCSVOp)
	plan.RegisterProcedureSpec(FromCSVKind, newFromCSVProcedure, FromCSVKind)
	execute.RegisterSource(FromCSVKind, createFromCSVSource)
}

func createFromCSVOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromCSVOpSpec)

	if csv, ok, err := args.GetString("csv"); err != nil {
		return nil, err
	} else if ok {
		spec.CSV = csv
	}

	if file, ok, err := args.GetString("file"); err != nil {
		return nil, err
	} else if ok {
		spec.File = file
	}

	if spec.CSV == "" && spec.File == "" {
		return nil, errors.New("must provide csv raw text or filename")
	}

	if spec.CSV != "" && spec.File != "" {
		return nil, errors.New("must provide exactly one of the parameters csv or file")
	}

	if spec.File != "" {
		if _, err := os.Stat(spec.File); err != nil {
			return nil, errors.Wrap(err, "failed to stat csv file: ")
		}
	}

	return spec, nil
}

func newFromCSVOp() flux.OperationSpec {
	return new(FromCSVOpSpec)
}

func (s *FromCSVOpSpec) Kind() flux.OperationKind {
	return FromCSVKind
}

type FromCSVProcedureSpec struct {
	plan.DefaultCost
	CSV  string
	File string
}

func newFromCSVProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromCSVOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FromCSVProcedureSpec{
		CSV:  spec.CSV,
		File: spec.File,
	}, nil
}

func (s *FromCSVProcedureSpec) Kind() plan.ProcedureKind {
	return FromCSVKind
}

func (s *FromCSVProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromCSVProcedureSpec)
	ns.CSV = s.CSV
	ns.File = s.File
	return ns
}

func createFromCSVSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*FromCSVProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", prSpec)
	}

	csvText := spec.CSV
	// if spec.File non-empty then spec.CSV is empty
	if spec.File != "" {
		csvBytes, err := ioutil.ReadFile(spec.File)
		if err != nil {
			return nil, errors.Wrap(err, "csv.from() failed to read file")
		}
		csvText = string(csvBytes)
	}
	csvSource := CSVSource{id: dsid, tx: csvText}

	return &csvSource, nil
}

type CSVSource struct {
	id execute.DatasetID
	tx string
	ts []execute.Transformation
}

func (c *CSVSource) AddTransformation(t execute.Transformation) {
	c.ts = append(c.ts, t)
}

func (c *CSVSource) Run(ctx context.Context) {
	var err error
	var max execute.Time
	maxSet := false
	for _, t := range c.ts {
		// For each downstream transformation, instantiate a new result
		// decoder. This way a table instance goes to one and only one
		// transformation. Unlike other sources, tables from csv sources
		// are not read-only. They contain mutable state and therefore
		// cannot be shared among goroutines.
		decoder := csv.NewResultDecoder(csv.ResultDecoderConfig{})
		result, decodeErr := decoder.Decode(strings.NewReader(c.tx))
		if decodeErr != nil {
			err = decodeErr
			goto FINISH
		}
		err = result.Tables().Do(func(tbl flux.Table) error {
			err := t.Process(c.id, tbl)
			if err != nil {
				return err
			}
			if idx := execute.ColIdx(execute.DefaultStopColLabel, tbl.Key().Cols()); idx >= 0 {
				if stop := tbl.Key().ValueTime(idx); !maxSet || stop > max {
					max = stop
					maxSet = true
				}
			}
			return nil
		})
		if err != nil {
			goto FINISH
		}
	}

	if maxSet {
		for _, t := range c.ts {
			if err = t.UpdateWatermark(c.id, max); err != nil {
				goto FINISH
			}
		}
	}

FINISH:
	for _, t := range c.ts {
		err = errors.Wrap(err, "error in csv.from()")
		t.Finish(c.id, err)
	}
}
