package line

import (
	"bufio"
	"io"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

// ResultDecoder decodes raw input strings from a reader into a flux.Result.
// It uses a separator to split the input into tokens and generate table rows.
// Tokens are kept as they are and put into a table with schema `_time`, `_value`.
// The `_value` column contains tokens.
// The `_time` column contains the timestamps for when each `_value` has been read.
// Strings in `_value` are obtained from the io.Reader passed to the Decode function.
// ResultDecoder outputs one table once the reader reaches EOF.
type ResultDecoder struct {
	reader *bufio.Reader
	stats  flux.Statistics
	config *ResultDecoderConfig
}

// NewResultDecoder creates a new result decoder from config.
func NewResultDecoder(config *ResultDecoderConfig) *ResultDecoder {
	return &ResultDecoder{config: config}
}

// TimeProvider gives the current time.
type TimeProvider interface {
	CurrentTime() values.Time
}

// ResultDecoderConfig is the configuration for a result decoder.
type ResultDecoderConfig struct {
	Separator    byte
	TimeProvider TimeProvider
}

func (rd *ResultDecoder) Do(f func(flux.Table) error) error {
	timeCol := flux.ColMeta{Label: "_time", Type: flux.TTime}
	valueCol := flux.ColMeta{Label: "_value", Type: flux.TString}
	key := execute.NewGroupKey(nil, nil)
	builder := execute.NewColListTableBuilder(key, &memory.Allocator{})
	timeIdx, err := builder.AddCol(timeCol)
	if err != nil {
		return err
	}
	valueIdx, err := builder.AddCol(valueCol)
	if err != nil {
		return err
	}

	var eof bool
	for !eof {
		s, err := rd.reader.ReadString(rd.config.Separator)
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		v := strings.Trim(s, string(rd.config.Separator))
		ts := rd.config.TimeProvider.CurrentTime()
		err = builder.AppendTime(timeIdx, ts)
		if err != nil {
			return err
		}
		err = builder.AppendString(valueIdx, v)
		if err != nil {
			return err
		}
	}

	tbl, err := builder.Table()
	if err != nil {
		return err
	}

	rd.stats = tbl.Statistics()

	return f(tbl)
}

func (*ResultDecoder) Name() string {
	return "_result"
}

func (rd *ResultDecoder) Tables() flux.TableIterator {
	return rd
}

func (rd *ResultDecoder) Statistics() flux.Statistics {
	return rd.stats
}

func (rd *ResultDecoder) Decode(r io.Reader) (flux.Result, error) {
	rd.reader = bufio.NewReader(r)
	return rd, nil
}
