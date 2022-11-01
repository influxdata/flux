package csv

import (
	"fmt"
	"net/http"
	"time"

	"github.com/influxdata/flux"
)

const DialectType = "csv"

// AddDialectMappings adds the influxql specific dialect mappings.
func AddDialectMappings(mappings flux.DialectMappings) error {
	return mappings.Add(DialectType, func() flux.Dialect {
		return &Dialect{
			ResultEncoderConfig: DefaultEncoderConfig(),
		}
	})
}

// Dialect describes the output format of queries in CSV.
type Dialect struct {
	ResultEncoderConfig
}

func (d Dialect) SetHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Transfer-Encoding", "chunked")
	if d.ResultEncoderConfig.DownloadHeader {
		timestamp := time.Now().Format(time.RFC3339)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"influxdata_%s.csv\"; filename*=UTF-8''influxdata_%s.csv", timestamp, timestamp))
	}
}

func (d Dialect) Encoder() flux.MultiResultEncoder {
	return NewMultiResultEncoder(d.ResultEncoderConfig)
}
func (d Dialect) DialectType() flux.DialectType {
	return DialectType
}

func DefaultDialect() *Dialect {
	return &Dialect{
		ResultEncoderConfig: DefaultEncoderConfig(),
	}
}
