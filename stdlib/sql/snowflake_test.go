package sql

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/gosnowflake"
)

func TestSnowflake_FileTransfer(t *testing.T) {
	cfg, err := gosnowflake.ParseDSN("user:password@my_organization-my_account/mydb")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Transporter = dependenciestest.RoundTripFunc(func(req *http.Request) *http.Response {
		resp := &http.Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
		}
		resp.Body = io.NopCloser(strings.NewReader(`{"success":true}`))
		return resp
	})
	connector := gosnowflake.NewConnector(gosnowflake.SnowflakeDriver{}, *cfg)
	db := sql.OpenDB(connector)
	defer db.Close()

	rows, err := db.Query("PUT file:///etc/passwd @~")
	if err == nil {
		t.Error("expected error")
		_ = rows.Close()
	} else {
		var snowflakeErr *gosnowflake.SnowflakeError
		if !errors.As(err, &snowflakeErr) {
			t.Errorf("unexpected error type: %T", err)
		} else {
			if got, want := snowflakeErr.Message, "file transfer not allowed"; got != want {
				t.Errorf("unexpected error message: %q != %q", want, got)
			}
			if got, want := snowflakeErr.Number, gosnowflake.ErrNotImplemented; got != want {
				t.Errorf("unexpected error number: %d != %d", want, got)
			}
		}
	}
}
