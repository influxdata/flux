package sql

import (
	"strings"
	"testing"
)

func TestHdb_IfNoExist(t *testing.T) {
	// test table with fqn
	q := hdbAddIfNotExist("stores.orders_copy", "CREATE TABLE stores.orders_copy (ORDER_ID BIGINT)")
	if !strings.HasPrefix(q, "DO") || !strings.Contains(q, "(:SCHEMA_NAME)") || !strings.Contains(q, "(:TABLE_NAME)") {
		t.Fail()
	}
	// test table in default schema
	q = hdbAddIfNotExist("orders_copy", "CREATE TABLE orders_copy (ORDER_ID BIGINT)")
	if !strings.HasPrefix(q, "DO") || strings.Contains(q, "(:SCHEMA_NAME)") || !strings.Contains(q, "(:TABLE_NAME)") {
		t.Fail()
	}
}
