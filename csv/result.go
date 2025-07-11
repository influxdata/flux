// Package csv contains the csv result encoders and decoders.
package csv

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/iocounter"
	"github.com/influxdata/flux/values"
)

const (
	defaultMaxBufferCount = 1000

	annotationIdx = 0
	resultIdx     = 1
	tableIdx      = 2

	defaultRecordStartIdx = 3

	datatypeAnnotation = "datatype"
	groupAnnotation    = "group"
	defaultAnnotation  = "default"

	resultLabel = "result"
	tableLabel  = "table"

	commentPrefix = "#"

	stringDatatype = "string"
	timeDatatype   = "dateTime"
	floatDatatype  = "double"
	boolDatatype   = "boolean"
	intDatatype    = "long"
	uintDatatype   = "unsignedLong"

	timeDataTypeWithFmt = "dateTime:RFC3339"

	nullValue = ""
)

// ResultDecoder decodes a csv representation of a result.
type ResultDecoder struct {
	c ResultDecoderConfig
}

// NewResultDecoder creates a new ResultDecoder.
func NewResultDecoder(c ResultDecoderConfig) *ResultDecoder {
	if c.MaxBufferCount == 0 {
		c.MaxBufferCount = defaultMaxBufferCount
	}
	return &ResultDecoder{
		c: c,
	}
}

// ResultDecoderConfig are options that can be specified on the ResultDecoder.
type ResultDecoderConfig struct {
	// NoAnnotations indicates that the CSV data will not have annotations rows.
	// Without annotations the decoder will assume every column is of type string and
	// that all data is in a single table in a single result.
	NoAnnotations bool
	// NoHeader indicates that the CSV data will not have a header row.
	NoHeader bool
	// MaxBufferCount is the maximum number of rows that will be buffered when decoding.
	// If 0, then a value of 1000 will be used.
	MaxBufferCount int
	// Allocator is the memory allocator that will be used during decoding.
	// The default is to use an unlimited allocator when this is not set.
	Allocator memory.Allocator
	// Context is the context for this ResultDecoder.
	// When the context is canceled, the decoder will also be canceled.
	// This defaults to context.Background.
	Context context.Context
}

func (d *ResultDecoder) Decode(r io.Reader) (flux.Result, error) {
	return newResultDecoder(newCSVReader(r), d.c, nil)
}

// MultiResultDecoder reads multiple results from a single csv file.
// Results are delimited by an empty line.
type MultiResultDecoder struct {
	c ResultDecoderConfig
}

// NewMultiResultDecoder creates a new MultiResultDecoder.
func NewMultiResultDecoder(c ResultDecoderConfig) *MultiResultDecoder {
	if c.MaxBufferCount == 0 {
		c.MaxBufferCount = defaultMaxBufferCount
	}
	return &MultiResultDecoder{
		c: c,
	}
}

func (d *MultiResultDecoder) Decode(r io.ReadCloser) (flux.ResultIterator, error) {
	return &resultIterator{
		c:  d.c,
		r:  r,
		cr: newCSVReader(r),
	}, nil
}

// resultIterator iterates through the results encoded in r.
type resultIterator struct {
	c    ResultDecoderConfig
	r    io.ReadCloser
	cr   *bufferedCSVReader
	next *resultDecoder
	err  error

	canceled bool
}

func (r *resultIterator) More() bool {
	if r.next == nil || !r.next.eof {
		var extraMeta *tableMetadata
		if r.next != nil {
			extraMeta = r.next.extraMeta
		}
		r.next, r.err = newResultDecoder(r.cr, r.c, extraMeta)
		if r.err == nil {
			return true
		}
		if r.err == io.EOF {
			// Do not report EOF errors
			r.err = nil
		}
	}

	// Release the resources for this query.
	r.Release()
	return false
}

func (r *resultIterator) Next() flux.Result {
	return r.next
}

func (r *resultIterator) Release() {
	if r.canceled {
		return
	}

	if err := r.r.Close(); err != nil && r.err == nil {
		r.err = err
	}
	r.canceled = true
}

func (r *resultIterator) Err() error {
	return r.err
}

func (r *resultIterator) Statistics() flux.Statistics {
	return flux.Statistics{}
}

type resultDecoder struct {
	id string
	c  ResultDecoderConfig

	cr *bufferedCSVReader

	extraMeta *tableMetadata

	eof bool
}

func newResultDecoder(cr *bufferedCSVReader, c ResultDecoderConfig, extraMeta *tableMetadata) (*resultDecoder, error) {
	d := &resultDecoder{
		c:         c,
		cr:        cr,
		extraMeta: extraMeta,
	}
	// We need to know the result ID before we return
	if extraMeta == nil {
		tm, err := readMetadata(d.cr, c)
		if err != nil {
			if err == io.EOF {
				return nil, err
			} else if sfe, ok := err.(*serializedFluxError); ok {
				return nil, sfe.err
			}
			return nil, errors.Wrap(err, codes.Inherit, "failed to read metadata")
		}
		d.extraMeta = &tm
	}
	d.id = d.extraMeta.ResultID
	return d, nil
}

func newCSVReader(r io.Reader) *bufferedCSVReader {
	csvr := csv.NewReader(r)
	csvr.ReuseRecord = true
	// Do not check record size
	csvr.FieldsPerRecord = -1
	csvr.LazyQuotes = true
	return &bufferedCSVReader{
		r:    csvr,
		line: nil,
	}
}

func (r *resultDecoder) Name() string {
	return r.id
}

func (r *resultDecoder) Tables() flux.TableIterator {
	return r
}

func (r *resultDecoder) Abort(error) {
	panic("not implemented")
}

func (r *resultDecoder) Do(f func(flux.Table) error) error {
	ctx := r.c.Context
	if ctx == nil {
		ctx = context.Background()
	}

	var meta tableMetadata
	newMeta := true
	for !r.eof {
		if newMeta {
			if r.extraMeta != nil {
				meta = *r.extraMeta
				r.extraMeta = nil
			} else {
				tm, err := readMetadata(r.cr, r.c)
				if err != nil {
					if err == io.EOF {
						goto EOF
					}
					if sfe, ok := err.(*serializedFluxError); ok {
						return sfe.err
					}
					return errors.Wrap(err, codes.Inherit, "failed to read metadata")
				}
				meta = tm
			}

			if meta.ResultID != r.id {
				r.extraMeta = &meta
				return nil
			}
		}

		// Create a new table
		tbl, err := newTable(r.cr, r.c, meta)
		if err != nil {
			return err
		}
		// Call f on the table, f can return before the table has been fully consumed.
		if err := f(tbl); err != nil {
			return err
		}
		// Block until the table has been fully consumed or the context is canceled
		select {
		case <-tbl.done:
		case <-ctx.Done():
			return ctx.Err()
		}

		// Now that the table has been consumed we can check for more tables
		// Check next line to see if we have a new meta block
		line, err := r.cr.Read()
		if err != nil {
			if err == io.EOF {
				goto EOF
			}
			return err
		}
		// cannot be a new meta block if there are no annotations
		newMeta = !r.c.NoAnnotations && len(line) > annotationIdx && line[annotationIdx] != ""
		err = r.cr.Unread(line)
		if err != nil {
			return err
		}
		// track whether we hit the EOF
		r.eof = tbl.eof
	}
EOF:
	r.eof = true
	return nil
}

type tableMetadata struct {
	ResultID       string
	TableID        string
	Cols           []colMeta
	Groups         []bool
	Defaults       []values.Value
	NumFields      int
	RecordStartIdx int
}

// serializedFluxError represents an error that occurred during
// Flux execution that has been serialized to CSV.
type serializedFluxError struct {
	err error
}

func (sfe *serializedFluxError) Error() string {
	return sfe.err.Error()
}

// readMetadata reads the table annotations and header.
// If there is no more data, returns (tableMetadata{}, io.EOF).
// In case of an actual error:
//   - if it's error that was serialized to CSV, it will be wrapped in serializedFluxError.
//   - otherwise, it's a serialization error, it will be returned as-is.
func readMetadata(r *bufferedCSVReader, c ResultDecoderConfig) (tableMetadata, error) {
	n := -1
	var recordStartIdx int
	var resultID, tableID string
	var datatypes, groups, defaults []string
	if c.NoAnnotations {
		// No annotations means that we are going to treat all rows as part of the same result with exactly one table
		resultID = "_result"
		tableID = "0"
		recordStartIdx = 0
		line, err := r.Read()
		if err != nil {
			return tableMetadata{}, err
		}
		n = len(line)
		// We treat all columns as strings
		datatypes = make([]string, n)
		groups = make([]string, n)
		defaults = make([]string, n)
		for i := range datatypes {
			datatypes[i] = "string"
			groups[i] = "false"
			defaults[i] = ""
		}
		// put this line back now that we know its length
		err = r.Unread(line)
		if err != nil {
			return tableMetadata{}, err
		}
	} else {
		recordStartIdx = defaultRecordStartIdx
		for datatypes == nil || groups == nil || defaults == nil {
			line, err := r.Read()
			if err != nil {
				if err == io.EOF {
					if datatypes == nil && groups == nil && defaults == nil {
						return tableMetadata{}, err
					}
					switch {
					case datatypes == nil:
						return tableMetadata{}, errors.New(codes.FailedPrecondition, "missing expected annotation datatype. "+
							"Consider using the mode: \"raw\" for csv that is not expected to have annotations.")
					case groups == nil:
						return tableMetadata{}, fmt.Errorf("missing expected annotation group")
					case defaults == nil:
						return tableMetadata{}, fmt.Errorf("missing expected annotation default")
					}
				}
				return tableMetadata{}, err
			}
			if n == -1 {
				n = len(line)
			}
			if n != len(line) {
				return tableMetadata{}, errors.Wrap(csv.ErrFieldCount, codes.Invalid, "failed to read annotations")
			}
			switch annotation := strings.TrimPrefix(line[annotationIdx], commentPrefix); annotation {
			case datatypeAnnotation:
				if defaultRecordStartIdx > len(line) {
					return tableMetadata{}, errors.Wrap(csv.ErrFieldCount, codes.Invalid, "failed to read \"datatype\" annotation")
				}
				datatypes = copyLine(line[defaultRecordStartIdx:])
			case groupAnnotation:
				if defaultRecordStartIdx > len(line) {
					return tableMetadata{}, errors.Wrap(csv.ErrFieldCount, codes.Invalid, "failed to read \"group\" annotation")
				}
				groups = copyLine(line[defaultRecordStartIdx:])
			case defaultAnnotation:
				if defaultRecordStartIdx > len(line) {
					return tableMetadata{}, errors.Wrap(csv.ErrFieldCount, codes.Invalid, "failed to read \"default\" annotation")
				}
				resultID = line[resultIdx]
				tableID = line[tableIdx]
				if _, err := strconv.ParseInt(tableID, 10, 64); tableID != "" && err != nil {
					return tableMetadata{}, fmt.Errorf("default Table ID is not an integer")
				}
				defaults = copyLine(line[defaultRecordStartIdx:])
			default:
				if !strings.HasPrefix(line[annotationIdx], commentPrefix) {
					switch {
					case datatypes == nil:
						return tableMetadata{}, errors.New(codes.FailedPrecondition, "missing expected annotation datatype. "+
							"consider using the mode: \"raw\" for csv that is not expected to have annotations.")
					case groups == nil:
						return tableMetadata{}, fmt.Errorf("missing expected annotation group")
					case defaults == nil:
						return tableMetadata{}, fmt.Errorf("missing expected annotation default")
					}
				}
				// Ignore unsupported/unknown annotations.
			}
		}
	}

	// Determine column labels
	var labels []string
	if c.NoHeader {
		labels = make([]string, len(datatypes))
		for i := range labels {
			labels[i] = fmt.Sprintf("col%d", i)
		}
	} else {
		// Read header row
		line, err := r.Read()
		if err != nil {
			if err == io.EOF {
				return tableMetadata{}, errors.New(codes.Invalid, "missing expected header row")
			}
			return tableMetadata{}, err
		}
		if n != len(line) {
			return tableMetadata{}, errors.Wrap(csv.ErrFieldCount, codes.Invalid, "failed to read header row")
		}

		if len(line) > 1 && line[1] == "error" {
			// Read the first row and return the error.
			line, err := r.Read()
			if err != nil || n != len(line) {
				if err == io.EOF {
					return tableMetadata{}, errors.Wrap(io.ErrUnexpectedEOF, codes.Invalid)
				} else if err == nil && n != len(line) {
					return tableMetadata{}, errors.Wrap(csv.ErrFieldCount, codes.Invalid)
				}
				return tableMetadata{}, errors.Wrap(err, codes.Inherit, "failed to read error value")
			}
			// TODO: We should determine the correct error code here:
			//   https://github.com/influxdata/flux/issues/1916
			return tableMetadata{}, &serializedFluxError{err: errors.New(codes.Internal, line[1])}
		}

		labels = line[recordStartIdx:]
	}

	cols := make([]colMeta, len(labels))
	defaultValues := make([]values.Value, len(labels))
	groupValues := make([]bool, len(labels))

	for j, label := range labels {
		t, desc, err := decodeType(datatypes[j])
		if err != nil {
			return tableMetadata{}, errors.Wrapf(err, codes.Invalid, "column %q has invalid datatype", label)
		}
		cols[j].ColMeta.Label = label
		cols[j].ColMeta.Type = t
		if t == flux.TTime {
			switch desc {
			case "RFC3339":
				cols[j].fmt = time.RFC3339
			case "RFC3339Nano":
				cols[j].fmt = time.RFC3339Nano
			default:
				cols[j].fmt = desc
			}
		}
		if defaults[j] == nullValue {
			defaultValues[j] = values.NewNull(flux.SemanticType(cols[j].ColMeta.Type))
		} else if defaults[j] == "" {
			// for now, the null value is always represented with "", so this is
			// unreachable.
			// When we support the #null annotation we'll want to distinguish
			// between "" (for strings) and null here.
			panic("unreachable")
		} else {
			v, err := decodeValue(defaults[j], cols[j])
			if err != nil {
				return tableMetadata{}, errors.Wrapf(err, codes.Invalid, "column %q has invalid default value", label)
			}
			defaultValues[j] = v
		}
		groupValues[j] = groups[j] == "true"
	}

	return tableMetadata{
		ResultID:       resultID,
		TableID:        tableID,
		Cols:           cols,
		Groups:         groupValues,
		Defaults:       defaultValues,
		NumFields:      n,
		RecordStartIdx: recordStartIdx,
	}, nil
}

type tableDecoder struct {
	r *bufferedCSVReader
	c ResultDecoderConfig

	meta tableMetadata

	used  int32
	empty bool

	initialized bool
	id          string

	key     flux.GroupKey
	colMeta []flux.ColMeta
	cols    []array.Builder
	nrows   int

	done chan struct{}

	eof bool
}

func newTable(
	r *bufferedCSVReader,
	c ResultDecoderConfig,
	meta tableMetadata,
) (*tableDecoder, error) {
	b := &tableDecoder{
		r:    r,
		c:    c,
		meta: meta,
		// assume its empty until we append a record
		empty: true,
		done:  make(chan struct{}),
	}
	more, err := b.advance()
	if !more {
		close(b.done)
	}
	if err != nil {
		return nil, err
	}
	if !b.initialized {
		return b, b.init(nil)
	}
	return b, nil
}

func (d *tableDecoder) Do(f func(flux.ColReader) error) error {
	if !atomic.CompareAndSwapInt32(&d.used, 0, 1) {
		return errors.New(codes.Internal, "table already read")
	}

	// Ensure that all internal memory is released when we exit.
	defer d.release()

	// Send off first batch from first advance call.
	if err := d.Emit(f); err != nil {
		return err
	}

	select {
	case <-d.done:
		return nil
	default:
	}

	more := true
	defer close(d.done)
	for more {
		var err error
		more, err = d.advance()
		if err != nil {
			return err
		}
		if err := d.Emit(f); err != nil {
			return err
		}
	}
	return nil
}

func (d *tableDecoder) Done() {
	_ = d.Do(func(flux.ColReader) error { return nil })
}

// advance reads the csv data until the end of the table or bufSize rows have been read.
// Advance returns whether there is more data and any error.
func (d *tableDecoder) advance() (bool, error) {
	var line, record []string
	var err error
	for !d.initialized || d.nrows < d.c.MaxBufferCount {
		line, err = d.r.Read()
		if err != nil {
			if err == io.EOF {
				d.eof = true
				return false, nil
			}
			return false, err
		}
		// whatever this line is, it's not part of this table so goto DONE
		if len(line) != d.meta.NumFields {
			// If this line is another annotation then the current table has no more data.
			// Otherwise if this line is not another annotation or if we are not expecting annotations
			// we have an ErrFieldCount, meaning that a row we expected to be a data row has the wrong number of fields.
			if d.c.NoAnnotations || (len(line) > annotationIdx && line[annotationIdx] == "") {
				return false, csv.ErrFieldCount
			}
			goto DONE
		}

		// Check for new annotation
		if !d.c.NoAnnotations && line[annotationIdx] != "" {
			goto DONE
		}

		if !d.initialized {
			if err := d.init(line); err != nil {
				return false, err
			}
			d.initialized = true
		}

		// check if we have tableID that is now different
		if !d.c.NoAnnotations && line[tableIdx] != "" && line[tableIdx] != d.id {
			goto DONE
		}

		record = line[d.meta.RecordStartIdx:]
		err = d.appendRecord(record)
		if err != nil {
			return false, err
		}
	}
	return true, nil

DONE:
	// table is done

	// unread the last line
	err = d.r.Unread(line)
	if err != nil {
		return false, err
	}
	if !d.initialized {
		// if we found a new annotation without any data rows, then the table is empty and we
		// init using the meta.Default column values.
		if d.empty {
			if err := d.init(nil); err != nil {
				return false, err
			}
		} else {
			return false, errors.New(codes.Internal, "table was not initialized, missing group key data")
		}
	}
	return false, nil
}

func (d *tableDecoder) init(line []string) error {
	if d.c.NoAnnotations {
		d.id = "0"
	} else {
		if len(line) != 0 {
			d.id = line[tableIdx]
		} else if d.meta.TableID != "" {
			d.id = d.meta.TableID
		} else {
			return errors.New(codes.Invalid, "missing table ID")
		}
	}
	var record []string
	if len(line) != 0 {
		record = line[d.meta.RecordStartIdx:]
	}
	keyCols := make([]flux.ColMeta, 0, len(d.meta.Cols))
	keyValues := make([]values.Value, 0, len(d.meta.Cols))
	for j, c := range d.meta.Cols {
		if d.meta.Groups[j] {
			var value values.Value
			if record != nil && record[j] != "" {
				// TODO: consider treatment of nullValue here
				v, err := decodeValue(record[j], c)
				if err != nil {
					return err
				}
				value = v
			} else {
				value = d.meta.Defaults[j]
			}
			keyCols = append(keyCols, c.ColMeta)
			keyValues = append(keyValues, value)
		}
	}

	d.key = execute.NewGroupKey(keyCols, keyValues)
	alloc := memory.DefaultAllocator
	if d.c.Allocator != nil {
		alloc = d.c.Allocator
	}
	if len(d.meta.Cols) > 0 {
		d.colMeta = make([]flux.ColMeta, len(d.meta.Cols))
		d.cols = make([]array.Builder, len(d.meta.Cols))
		for i, c := range d.meta.Cols {
			d.colMeta[i] = c.ColMeta
			d.cols[i] = arrow.NewBuilder(c.Type, alloc)
		}
	}

	return nil
}

func (d *tableDecoder) appendRecord(record []string) error {
	d.empty = false
	for j, c := range d.meta.Cols {
		if record[j] == "" {
			v := d.meta.Defaults[j]
			if err := arrow.AppendValue(d.cols[j], v); err != nil {
				return err
			}
			continue
		}
		if err := decodeValueInto(c, record[j], d.cols[j]); err != nil {
			return err
		}
	}
	d.nrows++
	return nil
}

func (d *tableDecoder) Empty() bool {
	return d.empty
}

func (d *tableDecoder) Key() flux.GroupKey {
	return d.key
}

func (d *tableDecoder) Cols() []flux.ColMeta {
	return d.colMeta
}

func (d *tableDecoder) Emit(f func(flux.ColReader) error) error {
	cr := arrow.TableBuffer{
		GroupKey: d.key,
		Columns:  d.colMeta,
		Values:   make([]array.Array, len(d.cols)),
	}
	for i, c := range d.cols {
		// Creating a new array resets the builder so
		// we do not have to release the memory or
		// reinitialize the builder.
		cr.Values[i] = c.NewArray()
	}
	d.nrows = 0

	defer cr.Release()
	return f(&cr)
}

func (d *tableDecoder) release() {
	for _, c := range d.cols {
		c.Release()
	}
	d.cols = nil
}

type colMeta struct {
	flux.ColMeta
	fmt string
}

type ResultEncoder struct {
	c       ResultEncoderConfig
	written bool
}

// ResultEncoderConfig are options that can be specified on the ResultEncoder.
type ResultEncoderConfig struct {
	// Annotations is a list of annotations to include.
	Annotations []string

	// NoHeader indicates whether a header row should be added.
	NoHeader bool

	// If present, add the download header matching this AttachmentFilename.
	AttachmentFilename string

	// Delimiter is the character to delimite columns.
	// It must not be \r, \n, or the Unicode replacement character (0xFFFD).
	Delimiter rune
}

func (c ResultEncoderConfig) MarshalJSON() ([]byte, error) {
	request := struct {
		Header      bool     `json:"header,omitempty"`
		Delimiter   string   `json:"delimiter"`
		Annotations []string `json:"annotations,omitempty"`
	}{
		Delimiter:   string(c.Delimiter),
		Annotations: c.Annotations,
		Header:      !c.NoHeader,
	}

	return json.Marshal(request)
}

func (c *ResultEncoderConfig) UnmarshalJSON(b []byte) error {
	request := &struct {
		Header      *bool    `json:"header,omitempty"`
		Delimiter   string   `json:"delimiter"`
		Annotations []string `json:"annotations,omitempty"`
	}{}

	if err := json.Unmarshal(b, request); err != nil {
		return err
	}

	if request.Delimiter == "" {
		request.Delimiter = ","
	}
	c.Delimiter, _ = utf8.DecodeRuneInString(request.Delimiter)

	c.NoHeader = false
	if request.Header != nil {
		c.NoHeader = !*request.Header
	}

	c.Annotations = request.Annotations

	return nil
}

func DefaultEncoderConfig() ResultEncoderConfig {
	return ResultEncoderConfig{
		Annotations: []string{datatypeAnnotation, groupAnnotation, defaultAnnotation},
	}
}

// NewResultEncoder creates a new encoder with the provided configuration.
func NewResultEncoder(c ResultEncoderConfig) *ResultEncoder {
	return &ResultEncoder{
		c: c,
	}
}

func (e *ResultEncoder) csvWriter(w io.Writer) *csv.Writer {
	writer := csv.NewWriter(w)
	if e.c.Delimiter != 0 {
		writer.Comma = e.c.Delimiter
	}
	writer.UseCRLF = true
	return writer
}

type csvEncoderError struct {
	err error
}

func (e *csvEncoderError) Error() string {
	return fmt.Sprintf("csv encoder error: %s", e.err.Error())
}

func (e *csvEncoderError) IsEncoderError() bool {
	return true
}

func (e *csvEncoderError) Unwrap() error {
	return e.err
}

func wrapEncodingError(err error) error {
	if err == nil {
		return err
	}
	return &csvEncoderError{err: err}
}

func (e *ResultEncoder) Encode(w io.Writer, result flux.Result) (int64, error) {
	tableID := 0
	tableIDStr := "0"
	metaCols := []colMeta{
		{ColMeta: flux.ColMeta{Label: "", Type: flux.TInvalid}},
		{ColMeta: flux.ColMeta{Label: resultLabel, Type: flux.TString}},
		{ColMeta: flux.ColMeta{Label: tableLabel, Type: flux.TInt}},
	}
	writeCounter := &iocounter.Writer{Writer: w}
	writer := e.csvWriter(writeCounter)

	var lastCols []colMeta
	var lastGroupCols []flux.ColMeta
	var lastEmpty bool

	resultName := result.Name()
	err := result.Tables().Do(func(tbl flux.Table) error {
		e.written = true
		// Update cols with table cols
		cols := metaCols
		for _, c := range tbl.Cols() {
			cm := colMeta{ColMeta: c}
			if c.Type == flux.TTime {
				cm.fmt = time.RFC3339Nano
			}
			cols = append(cols, cm)
		}
		// pre-allocate row slice
		row := make([]string, len(cols))

		if lastEmpty || tbl.Empty() || schemaChanged(cols, lastCols, tbl.Key().Cols(), lastGroupCols) {
			if len(lastCols) > 0 {
				// Write out empty line if not first table
				writer.Write(nil)
			}

			if err := writeSchema(writer, &e.c, row, cols, tbl.Empty(), tbl.Key(), resultName, tableIDStr); err != nil {
				return wrapEncodingError(err)
			}
		}

		if execute.ContainsStr(e.c.Annotations, defaultAnnotation) {
			for j := range cols {
				switch j {
				case annotationIdx:
					row[j] = ""
				case resultIdx:
					row[j] = ""
				case tableIdx:
					row[j] = tableIDStr
				default:
					row[j] = ""
				}
			}
		} else {
			for j := range cols {
				switch j {
				case annotationIdx:
					row[j] = ""
				case resultIdx:
					row[j] = resultName
				case tableIdx:
					row[j] = tableIDStr
				default:
					row[j] = ""
				}
			}
		}

		if err := tbl.Do(func(cr flux.ColReader) error {
			record := row[defaultRecordStartIdx:]
			l := cr.Len()
			for i := 0; i < l; i++ {
				for j, c := range cols[defaultRecordStartIdx:] {
					v, err := encodeValueFrom(i, j, c, cr)
					if err != nil {
						return wrapEncodingError(err)
					}
					record[j] = v
				}
				writer.Write(row)
			}
			writer.Flush()
			return wrapEncodingError(writer.Error())
		}); err != nil {
			return err
		}

		tableID++
		tableIDStr = strconv.Itoa(tableID)
		lastCols = cols
		lastGroupCols = tbl.Key().Cols()
		lastEmpty = tbl.Empty()
		writer.Flush()
		return wrapEncodingError(writer.Error())
	})
	return writeCounter.Count(), err
}

func (e *ResultEncoder) EncodeError(w io.Writer, err error) error {
	writer := e.csvWriter(w)
	if e.written {
		// Write out empty line
		writer.Write(nil)
	}

	for _, anno := range e.c.Annotations {
		switch anno {
		case datatypeAnnotation:
			writer.Write([]string{commentPrefix + datatypeAnnotation, "string", "string"})
		case groupAnnotation:
			writer.Write([]string{commentPrefix + groupAnnotation, "true", "true"})
		case defaultAnnotation:
			writer.Write([]string{commentPrefix + defaultAnnotation, "", ""})
		}
	}
	writer.Write([]string{"", "error", "reference"})
	// TODO: Add referenced code
	writer.Write([]string{"", err.Error(), ""})
	writer.Flush()
	return writer.Error()
}

func writeSchema(writer *csv.Writer, c *ResultEncoderConfig, row []string, cols []colMeta, useKeyDefaults bool, key flux.GroupKey, resultName, tableID string) error {
	defaults := make([]string, len(row))
	for j, c := range cols {
		switch j {
		case annotationIdx:
		case resultIdx:
			defaults[j] = resultName
		case tableIdx:
			if useKeyDefaults {
				defaults[j] = tableID
			} else {
				defaults[j] = ""
			}
		default:
			if useKeyDefaults {
				kj := execute.ColIdx(c.Label, key.Cols())
				if kj >= 0 {
					v, err := encodeValue(key.Value(kj), c)
					if err != nil {
						return err
					}
					defaults[j] = v
				} else {
					defaults[j] = nullValue
				}
			} else {
				defaults[j] = nullValue
			}
		}
	}
	if err := writeAnnotations(writer, c.Annotations, row, defaults, cols, key); err != nil {
		return err
	}

	if !c.NoHeader {
		// Write labels header
		for j, c := range cols {
			row[j] = c.Label
		}
		writer.Write(row)
	}
	return writer.Error()
}

func writeAnnotations(writer *csv.Writer, annotations []string, row, defaults []string, cols []colMeta, key flux.GroupKey) error {
	for _, annotation := range annotations {
		switch annotation {
		case datatypeAnnotation:
			if err := writeDatatypes(writer, row, cols); err != nil {
				return err
			}
		case groupAnnotation:
			if err := writeGroups(writer, row, cols, key); err != nil {
				return err
			}
		case defaultAnnotation:
			if err := writeDefaults(writer, row, defaults); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported annotation %q", annotation)
		}
	}
	return writer.Error()
}

func writeDatatypes(writer *csv.Writer, row []string, cols []colMeta) error {
	for j, c := range cols {
		if j == annotationIdx {
			row[j] = commentPrefix + datatypeAnnotation
			continue
		}
		switch c.Type {
		case flux.TBool:
			row[j] = boolDatatype
		case flux.TInt:
			row[j] = intDatatype
		case flux.TUInt:
			row[j] = uintDatatype
		case flux.TFloat:
			row[j] = floatDatatype
		case flux.TString:
			row[j] = stringDatatype
		case flux.TTime:
			row[j] = timeDataTypeWithFmt
		default:
			return fmt.Errorf("unknown column type %v", c.Type)
		}
	}
	return writer.Write(row)
}

func writeGroups(writer *csv.Writer, row []string, cols []colMeta, key flux.GroupKey) error {
	for j, c := range cols {
		if j == annotationIdx {
			row[j] = commentPrefix + groupAnnotation
			continue
		}
		if j == resultIdx || j == tableIdx {
			row[j] = "false"
			continue
		}
		row[j] = strconv.FormatBool(key.HasCol(c.Label))
	}
	return writer.Write(row)
}

func writeDefaults(writer *csv.Writer, row, defaults []string) error {
	for j := range defaults {
		switch j {
		case annotationIdx:
			row[j] = commentPrefix + defaultAnnotation
		default:
			row[j] = defaults[j]
		}
	}
	return writer.Write(row)
}

func decodeValue(value string, c colMeta) (values.Value, error) {
	if value == nullValue {
		return values.NewNull(flux.SemanticType(c.Type)), nil
	}

	var val values.Value
	switch c.Type {
	case flux.TBool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		val = values.NewBool(v)
	case flux.TInt:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		val = values.NewInt(v)
	case flux.TUInt:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, err
		}
		val = values.NewUInt(v)
	case flux.TFloat:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		val = values.NewFloat(v)
	case flux.TString:
		val = values.NewString(value)
	case flux.TTime:
		v, err := decodeTime(value, c.fmt)
		if err != nil {
			return nil, err
		}
		val = values.NewTime(v)
	default:
		return nil, fmt.Errorf("unsupported type %v", c.Type)
	}
	return val, nil
}

func decodeValueInto(c colMeta, value string, b array.Builder) error {
	switch c.Type {
	case flux.TBool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		return arrow.AppendBool(b, v)
	case flux.TInt:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		return arrow.AppendInt(b, v)
	case flux.TUInt:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		return arrow.AppendUint(b, v)
	case flux.TFloat:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		return arrow.AppendFloat(b, v)
	case flux.TString:
		return arrow.AppendString(b, value)
	case flux.TTime:
		t, err := decodeTime(value, c.fmt)
		if err != nil {
			return err
		}
		return arrow.AppendTime(b, t)
	default:
		return fmt.Errorf("unsupported type %v", c.Type)
	}
}

func encodeValue(value values.Value, c colMeta) (string, error) {
	if value.IsNull() {
		return nullValue, nil
	}

	switch c.Type {
	case flux.TBool:
		return strconv.FormatBool(value.Bool()), nil
	case flux.TInt:
		return strconv.FormatInt(value.Int(), 10), nil
	case flux.TUInt:
		return strconv.FormatUint(value.UInt(), 10), nil
	case flux.TFloat:
		return strconv.FormatFloat(value.Float(), 'f', -1, 64), nil
	case flux.TString:
		return value.Str(), nil
	case flux.TTime:
		return encodeTime(value.Time(), c.fmt), nil
	default:
		return "", fmt.Errorf("unknown type %v", c.Type)
	}
}

func encodeValueFrom(i, j int, c colMeta, cr flux.ColReader) (string, error) {
	var v = nullValue
	switch c.Type {
	case flux.TBool:
		if cr.Bools(j).IsValid(i) {
			v = strconv.FormatBool(cr.Bools(j).Value(i))
		}
	case flux.TInt:
		if cr.Ints(j).IsValid(i) {
			v = strconv.FormatInt(cr.Ints(j).Value(i), 10)
		}
	case flux.TUInt:
		if cr.UInts(j).IsValid(i) {
			v = strconv.FormatUint(cr.UInts(j).Value(i), 10)
		}
	case flux.TFloat:
		if cr.Floats(j).IsValid(i) {
			v = strconv.FormatFloat(cr.Floats(j).Value(i), 'f', -1, 64)
		}
	case flux.TString:
		if cr.Strings(j).IsValid(i) {
			v = cr.Strings(j).Value(i)
		}
	case flux.TTime:
		if cr.Times(j).IsValid(i) {
			v = encodeTime(execute.Time(cr.Times(j).Value(i)), c.fmt)
		}
	default:
		return "", fmt.Errorf("unknown type %v", c.Type)
	}

	return v, nil
}

func decodeTime(t string, fmt string) (execute.Time, error) {
	v, err := time.Parse(fmt, t)
	if err != nil {
		return 0, err
	}
	return values.ConvertTime(v), nil
}

func encodeTime(t execute.Time, fmt string) string {
	return t.Time().Format(fmt)
}

func copyLine(line []string) []string {
	cpy := make([]string, len(line))
	copy(cpy, line)
	return cpy
}

// decodeType returns the flux.ColType and any additional format description.
func decodeType(datatype string) (t flux.ColType, desc string, err error) {
	split := strings.SplitN(datatype, ":", 2)
	if len(split) > 1 {
		desc = split[1]
	}
	typ := split[0]
	switch typ {
	case boolDatatype:
		t = flux.TBool
	case intDatatype:
		t = flux.TInt
	case uintDatatype:
		t = flux.TUInt
	case floatDatatype:
		t = flux.TFloat
	case stringDatatype:
		t = flux.TString
	case timeDatatype:
		t = flux.TTime
	default:
		err = fmt.Errorf("unsupported data type %q", typ)
	}
	return
}

func schemaChanged(cols, lastCols []colMeta, groupCols, lastGroupCols []flux.ColMeta) bool {
	if lastGroupCols == nil {
		return true
	}
	if len(groupCols) != len(lastGroupCols) {
		return true
	}
	for j := range groupCols {
		if groupCols[j] != lastGroupCols[j] {
			return true
		}
	}

	if len(cols) != len(lastCols) {
		return true
	}
	for j := range cols {
		if cols[j] != lastCols[j] {
			return true
		}
	}
	return false
}

func NewMultiResultEncoder(c ResultEncoderConfig) flux.MultiResultEncoder {
	return &flux.DelimitedMultiResultEncoder{
		Delimiter: []byte("\r\n"),
		Encoder:   NewResultEncoder(c),
	}
}

// bufferedCSVReader allows for unreading a single line of the csv data
type bufferedCSVReader struct {
	r    *csv.Reader
	line []string
}

// Read returns the next line in the csv stream
func (b *bufferedCSVReader) Read() ([]string, error) {
	if len(b.line) > 0 {
		line := b.line
		b.line = nil
		return line, nil
	}
	return b.r.Read()
}

// Unread places the provided line back on the buffer.
// It is invalid to call unread multiple times without calling Read inbetween.
func (b *bufferedCSVReader) Unread(line []string) error {
	if len(b.line) > 0 {
		return errors.New(codes.Internal, "unread was called without reading first")
	}
	b.line = line
	return nil
}
