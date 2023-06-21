package sql

import (
	"testing"

	"github.com/InfluxCommunity/flux/dependencies/url"
)

func TestLocalhostIsInvalid(t *testing.T) {
	validator := url.PrivateIPValidator{}
	if validateDataSource(validator, "mock", "postgres://localhost/database") == nil {
		t.Error("localhost is a private ip; expected validator to fail")
	}
}

func TestLocalhostIsValidForBigQuery(t *testing.T) {
	validator := url.PrivateIPValidator{}
	if validateDataSource(validator, "bigquery", "bigquery://localhost/") != nil {
		t.Error("bigquery DSNs contain no host info; expected validator to pass")
	}
}
