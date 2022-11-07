//go:build fipsonly

package sql

import (
	"database/sql"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
)

type snowflakeConfig struct {
	User     string
	Password string
	Protocol string
	Host     string
}

// How did these end up in the snowflake implementation?
const (
	layoutDate         = "2006-01-02"
	layoutTime         = "15:04:05"
	layoutTimeStampNtz = "2006-01-02T15:04:05.0000000000"
)

// errSnowflakeDisabled indicates Snowflake support has been disabled because of FIPS.
var errSnowflakeDisabled = errors.Wrap(ErrorDriverDisabled, codes.Unimplemented, "snowflake support disabled for FIPS")

func snowflakeParseDSN(dsn string) (cfg *snowflakeConfig, err error) {
	return nil, errSnowflakeDisabled
}

func NewSnowflakeRowReader(r *sql.Rows) (execute.RowReader, error) {
	return nil, errSnowflakeDisabled
}

// SnowflakeColumnTranslateFunc will go away once gosnowflake's OCSP implementation is updated.
func SnowflakeColumnTranslateFunc() translationFunc {
	return func(f flux.ColType, colname string) (string, error) {
		return "", errMssqlDisabled
	}
}
