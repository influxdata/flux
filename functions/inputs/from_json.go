package inputs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/influxql"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/pkg/errors"
)

const FromJSONKind = "fromJSON"
const bufferSize = 8192

func init() {
	fromJSONSignature := semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"json": semantic.String,
			"file": semantic.String,
		},
		Required: nil,
		Return:   flux.TableObjectType,
	}
	flux.RegisterFunction(FromJSONKind, createFromJSONOpSpec, fromJSONSignature)
	flux.RegisterOpSpec(FromJSONKind, newFromJSONOp)
	plan.RegisterProcedureSpec(FromJSONKind, newFromJSONProcedure, FromJSONKind)
	execute.RegisterSource(FromJSONKind, createFromJSONSource)
}

func createFromJSONOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	var spec = new(FromJSONOpSpec)

	if json, ok, err := args.GetString("json"); err != nil {
		return nil, err
	} else if ok {
		spec.JSON = json
	}

	if file, ok, err := args.GetString("file"); err != nil {
		return nil, err
	} else if ok {
		spec.File = file
	}

	if spec.JSON == "" && spec.File == "" {
		return nil, errors.New("must provide json raw text or filename")
	}

	if spec.JSON != "" && spec.File != "" {
		return nil, errors.New("must provide exactly one of the parameters json or file")
	}

	if spec.File != "" {
		if _, err := os.Stat(spec.File); err != nil {
			return nil, errors.Wrapf(err, "failed to stat json file: %s", spec.File)
		}
	}

	return spec, nil
}

// FromJSONOpSpec defines the `fromJSON` function signature
type FromJSONOpSpec struct {
	JSON string `json:"json"`
	File string `json:"file"`
}

func newFromJSONOp() flux.OperationSpec {
	return new(FromJSONOpSpec)
}

func (s *FromJSONOpSpec) Kind() flux.OperationKind {
	return FromJSONKind
}

// FromJSONProcedureSpec describes the `fromJSON` prodecure
type FromJSONProcedureSpec struct {
	plan.DefaultCost
	JSON string
	File string
}

func newFromJSONProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromJSONOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &FromJSONProcedureSpec{
		JSON: spec.JSON,
		File: spec.File,
	}, nil
}

func (s *FromJSONProcedureSpec) Kind() plan.ProcedureKind {
	return FromJSONKind
}

func (s *FromJSONProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromJSONProcedureSpec)
	ns.JSON = s.JSON
	ns.File = s.File
	return ns
}

func createFromJSONSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*FromJSONProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", prSpec)
	}

	var jsonReader io.Reader

	if spec.File != "" {
		f, err := os.Open(spec.File)
		if err != nil {
			return nil, err
		}
		jsonReader = bufio.NewReaderSize(f, bufferSize)
	} else {
		jsonReader = strings.NewReader(spec.JSON)
	}

	decoder := influxql.NewResultDecoder(a.Allocator())
	results, err := decoder.Decode(ioutil.NopCloser(jsonReader))
	if err != nil {
		return nil, err
	}

	return &JSONSource{id: dsid, results: results}, nil
}

type JSONSource struct {
	results flux.ResultIterator
	id      execute.DatasetID
	ts      []execute.Transformation
}

func (c *JSONSource) AddTransformation(t execute.Transformation) {
	c.ts = append(c.ts, t)
}

func (c *JSONSource) Run(ctx context.Context) {
	var err error
	var max execute.Time
	var maxSet bool

	err = c.results.Next().Tables().Do(func(tbl flux.Table) error {
		for _, t := range c.ts {
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
		}
		return nil
	})

	if err != nil {
		goto FINISH
	}

	if maxSet {
		for _, t := range c.ts {
			err = t.UpdateWatermark(c.id, max)
			if err != nil {
				goto FINISH
			}
		}
	}

	if c.results.More() {
		// It doesn't make sense to read multiple results
		err = errors.Wrap(err, "'fromJSON' supports only single results")
	}

FINISH:
	for _, t := range c.ts {
		t.Finish(c.id, err)
	}
}
