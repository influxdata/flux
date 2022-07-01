package secret_test

import (
	"context"
	"testing"

	"github.com/influxdata/flux/mock"
)

func TestSecret_Service(t *testing.T) {
	ss := mock.SecretService{"key": "val"}
	val, err := ss.LoadSecret(context.Background(), "key")
	if err != nil {
		t.Fatal(err)
	}
	if val != "val" {
		t.Error("secret service returned wrong value")
	}

	if _, err = ss.LoadSecret(context.Background(), "k"); err == nil {
		t.Error("secret service should have errored on key lookup")
	}
}
