// Package socket implements a source that gets input from a socket connection and produces tables given a decoder.
// This is a good candidate for streaming use cases. For now, it produces a single table for everything
// that it receives from the start to the end of the connection.
package socket

import (
	"context"
	"io"
	"net"
	neturl "net/url"
	"strings"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/line"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const FromSocketKind = "fromSocket"

type FromSocketOpSpec struct {
	URL     string `json:"url"`
	Decoder string `json:"decoder"`
}

func init() {
	fromSocketSignature := runtime.MustLookupBuiltinType("socket", "from")

	runtime.RegisterPackageValue("socket", "from", flux.MustValue(flux.FunctionValue(FromSocketKind, createFromSocketOpSpec, fromSocketSignature)))
	flux.RegisterOpSpec(FromSocketKind, newFromSocketOp)
	plan.RegisterProcedureSpec(FromSocketKind, newFromSocketProcedure, FromSocketKind)
	execute.RegisterSource(FromSocketKind, createFromSocketSource)
}

// nowTimeProvider provides wall clock time.
type nowTimeProvider struct{}

func (a *nowTimeProvider) CurrentTime() values.Time {
	return values.ConvertTime(time.Now())
}

var (
	decoders = []string{"csv", "line"}
	schemes  = []string{"tcp", "unix"}
)

func contains(ss []string, s string) bool {
	for _, st := range ss {
		if st == s {
			return true
		}
	}
	return false
}

func createFromSocketOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromSocketOpSpec)

	if url, err := args.GetRequiredString("url"); err != nil {
		return nil, err
	} else {
		spec.URL = url
	}

	if d, ok, err := args.GetString("decoder"); err != nil {
		return nil, err
	} else if ok {
		spec.Decoder = d
	} else {
		spec.Decoder = decoders[0]
	}

	if !contains(decoders, spec.Decoder) {
		return nil, errors.Newf(codes.Invalid, "invalid decoder %s, must be one of %v", spec.Decoder, decoders)
	}

	return spec, nil
}

func newFromSocketOp() flux.OperationSpec {
	return new(FromSocketOpSpec)
}

func (s *FromSocketOpSpec) Kind() flux.OperationKind {
	return FromSocketKind
}

type FromSocketProcedureSpec struct {
	plan.DefaultCost
	URL     string
	Decoder string
}

func newFromSocketProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromSocketOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &FromSocketProcedureSpec{
		URL:     spec.URL,
		Decoder: spec.Decoder,
	}, nil
}

func (s *FromSocketProcedureSpec) Kind() plan.ProcedureKind {
	return FromSocketKind
}

func (s *FromSocketProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromSocketProcedureSpec)
	ns.URL = s.URL
	ns.Decoder = s.Decoder
	return ns
}

func createFromSocketSource(s plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := s.(*FromSocketProcedureSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", s)
	}

	// known issue with url.Parse for detecting the presence of a scheme: https://github.com/golang/go/issues/19779
	var scheme, address string
	if !strings.Contains(spec.URL, "://") {
		// no scheme specified, use default scheme and use the entire url as address
		spec.URL = schemes[0] + "://" + spec.URL
	}
	url, err := neturl.Parse(spec.URL)
	if err != nil {
		return nil, errors.Newf(codes.Invalid, "invalid url: %v", err)
	}
	deps := flux.GetDependencies(a.Context())
	validator, err := deps.URLValidator()
	if err != nil {
		return nil, err
	}
	if err := validator.Validate(url); err != nil {
		return nil, errors.Newf(codes.Invalid, "url did not pass validation: %v", err)
	}
	scheme = url.Scheme
	address = url.Host
	if !contains(schemes, scheme) {
		return nil, errors.Newf(codes.Invalid, "invalid scheme %s, must be one of %v", scheme, schemes)
	}

	conn, err := net.Dial(scheme, address)
	if err != nil {
		return nil, errors.Wrap(err, codes.Inherit, "error in creating socket source")
	}

	return NewSocketSource(spec, conn, &nowTimeProvider{}, dsid)
}

func NewSocketSource(spec *FromSocketProcedureSpec, rc io.ReadCloser, tp line.TimeProvider, dsid execute.DatasetID) (execute.Source, error) {
	var decoder flux.ResultDecoder
	switch spec.Decoder {
	case "csv":
		decoder = csv.NewResultDecoder(csv.ResultDecoderConfig{})
	case "line":
		decoder = line.NewResultDecoder(&line.ResultDecoderConfig{
			Separator:    '\n',
			TimeProvider: tp,
		})
	}

	if decoder == nil {
		return nil, errors.Newf(codes.Invalid, "unknown decoder type: %v", spec.Decoder)
	}

	return &socketSource{
		d:       dsid,
		rc:      rc,
		decoder: decoder,
	}, nil
}

type socketSource struct {
	execute.ExecutionNode
	d       execute.DatasetID
	rc      io.ReadCloser
	decoder flux.ResultDecoder
	ts      []execute.Transformation
}

func (ss *socketSource) AddTransformation(t execute.Transformation) {
	ss.ts = append(ss.ts, t)
}

func (ss *socketSource) Run(ctx context.Context) {
	defer ss.rc.Close()
	result, err := ss.decoder.Decode(ss.rc)
	if err != nil {
		err = errors.Wrap(err, codes.Inherit, "decode error")
	} else {
		err = result.Tables().Do(func(tbl flux.Table) error {
			for _, t := range ss.ts {
				if err := t.Process(ss.d, tbl); err != nil {
					return err
				}
			}
			return nil
		})
	}

	for _, t := range ss.ts {
		t.Finish(ss.d, err)
	}
}
