//go:build fipsonly

package sql

// This file contains code to effectively disable mssql support in flux.
// This is simply a stopgap until flux can be modified to avoid making
// connections that will use NTLM. Once that happens, most of the code in
// this file will go away and be replaced by a function to check if
// NTLM is required.

import (
	"database/sql"
	neturl "net/url"
	"strings"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/execute"
	"github.com/InfluxCommunity/flux/internal/errors"
)

// mssqlConfig is copied from mssql.go. This copy will go away when the NTLM check is added.
type mssqlConfig struct {
	Scheme   string
	Host     string
	User     string
	Password string
	Database string
	// Azure auth
	//AzureAuth string
	//*AzureConfig
}

// errMssqlDisabled indicates MSSQL support has been disabled because of FIPS.
var errMssqlDisabled = errors.Wrap(ErrorDriverDisabled, codes.Unimplemented, "mssql driver disabled for FIPS")

// mssqlParseDSN will go away once NTLM check is added.
func mssqlParseDSN(dsn string) (cfg *mssqlConfig, err error) {
	return nil, errMssqlDisabled
}

// NewMssqlRowReader will go away once NTLM check is added.
func NewMssqlRowReader(r *sql.Rows) (execute.RowReader, error) {
	return nil, errMssqlDisabled
}

// MssqlColumnTranslateFunc will go away once NTLM check is added.
func MssqlColumnTranslateFunc() translationFunc {
	return func(f flux.ColType, colname string) (string, error) {
		return "", errMssqlDisabled
	}
}

// isMssqlDriver is copied from mssql.go. This copy will go away once NTLM check is added.
func isMssqlDriver(driverName string) bool {
	return driverName == "mssql" || driverName == "sqlserver"
}

// These constants are copied from mssql.go and will go away when NTLM check is added.
const (
	// SQL Server cannot insert into table with IDENTITY set unless explicitly enabled.
	// https://docs.microsoft.com/en-us/sql/t-sql/statements/set-identity-insert-transact-sql
	mssqlIdentityInsertEnabled = "identity insert=on"
)

// mssqlCheckParameter is copied from mssql.go. This copy will go away once NTLM check is added.
func mssqlCheckParameter(dsn string, option string) bool {
	raw, err := neturl.QueryUnescape(dsn)
	if err != nil {
		raw = dsn
	}
	return strings.Contains(strings.ToLower(raw), option)
}
