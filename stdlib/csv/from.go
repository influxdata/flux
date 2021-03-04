package csv

import (
	"context"
	"io"
	"io/ioutil"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/dependencies/filesystem"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const FromCSVKind = "fromCSV"

type FromCSVOpSpec struct {
	CSV  string `json:"csv"`
	File string `json:"file"`
	Mode string `json:"mode"`
}

const (
	annotationMode = "annotations"
	rawMode        = "raw"
)

func init() {
	fromCSVSignature := runtime.MustLookupBuiltinType("csv", "from")
	runtime.RegisterPackageValue("csv", "from", flux.MustValue(flux.FunctionValue(FromCSVKind, createFromCSVOpSpec, fromCSVSignature)))
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
		return nil, errors.New(codes.Invalid, "must provide csv raw text or filename")
	}

	if spec.CSV != "" && spec.File != "" {
		return nil, errors.New(codes.Invalid, "must provide exactly one of the parameters csv or file")
	}

	if mode, ok, err := args.GetString("mode"); err != nil {
		return nil, err
	} else if ok {
		spec.Mode = mode
	} else {
		// default to annotation mode
		spec.Mode = annotationMode
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
	Mode string
}

func newFromCSVProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromCSVOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &FromCSVProcedureSpec{
		CSV:  spec.CSV,
		File: spec.File,
		Mode: spec.Mode,
	}, nil
}

func (s *FromCSVProcedureSpec) Kind() plan.ProcedureKind {
	return FromCSVKind
}

func (s *FromCSVProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromCSVProcedureSpec)
	ns.CSV = s.CSV
	ns.File = s.File
	ns.Mode = s.Mode
	return ns
}

func createFromCSVSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*FromCSVProcedureSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", prSpec)
	}
	return CreateSource(spec, dsid, a)
}

func CreateSource(spec *FromCSVProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	var getDataStream func() (io.ReadCloser, error)
	if spec.File != "" {
		getDataStream = func() (io.ReadCloser, error) {
			f, err := filesystem.OpenFile(a.Context(), spec.File)
			if err != nil {
				return nil, errors.Wrap(err, codes.Inherit, "csv.from() failed to read file")
			}
			return f, nil
		}
	} else { // if spec.File is empty then spec.CSV is not empty
		getDataStream = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(strings.NewReader(spec.CSV)), nil
		}
	}
	csvSource := CSVSource{
		id:            dsid,
		getDataStream: getDataStream,
		alloc:         a.Allocator(),
		mode:          spec.Mode,
	}

	return &csvSource, nil
}

type CSVSource struct {
	execute.ExecutionNode
	id            execute.DatasetID
	getDataStream func() (io.ReadCloser, error)
	ts            []execute.Transformation
	alloc         *memory.Allocator
	mode          string
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
		config := csv.ResultDecoderConfig{
			Allocator: c.alloc,
			Context:   ctx,
		}
		switch c.mode {
		case rawMode:
			config.NoAnnotations = true
		default:
		}
		decoder := csv.NewMultiResultDecoder(config)
		var data io.ReadCloser
		data, err = c.getDataStream()
		if err != nil {
			goto FINISH
		}
		// Many applications will add a UTF BOM (byte order mark) to the beginning of csv files
		// We expect UTF8 encoded data so the byte order does not matter.
		// Therefore we skip the BOM if it exists.
		// See http://www.unicode.org/faq/utf_bom.html#BOM
		rc := newSkipBOMReader(data)
		results, decodeErr := decoder.Decode(rc)
		defer results.Release()
		if decodeErr != nil {
			err = decodeErr
			goto FINISH
		}

		if !results.More() {
			err = results.Err()
			goto FINISH
		}
		result := results.Next()

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
		if results.More() {
			err = errors.New(
				codes.FailedPrecondition,
				"csv.from() can only parse 1 result",
			)
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
	if err != nil {
		err = errors.Wrap(err, codes.Inherit, "error in csv.from()")
	}

	for _, t := range c.ts {
		t.Finish(c.id, err)
	}
}

// skipBOMReader wraps an io.ReadCloser and skips the BOM,
// if it exists at the beginning of the stream.
type skipBOMReader struct {
	io.ReadCloser
	buf     []byte
	checked bool
}

func newSkipBOMReader(r io.ReadCloser) io.ReadCloser {
	return &skipBOMReader{
		ReadCloser: r,
		// BOM is three bytes long
		buf:     []byte{0, 0, 0},
		checked: false,
	}
}

func (r *skipBOMReader) Read(p []byte) (n int, err error) {
	if !r.checked {
		r.checked = true
		// Read at least 3 bytes if possible
		// We do not use io.ReadFull or io.ReadAtLeast here as they
		// change the semantics of how errors are reported..
		total := 0
		for total < len(r.buf) {
			n, err := r.ReadCloser.Read(r.buf[total:])
			total += n
			if err != nil {
				m := copy(p, r.buf[:total])
				return m, err
			}
		}
		if r.buf[0] == 0xef && r.buf[1] == 0xbb && r.buf[2] == 0xbf {
			// Found the BOM, null the buffer so that we don't return those bytes
			r.buf = nil
		}
	}
	if len(r.buf) > 0 {
		n := copy(p, r.buf)
		r.buf = r.buf[n:]
		return n, nil
	}
	return r.ReadCloser.Read(p)
}
