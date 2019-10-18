package socket

import (
	"testing"

	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/execute/executetest"
)

func TestFromSocketUrlValidation(t *testing.T) {
	testCases := executetest.SourceUrlValidationTestCases{
		{
			Name: "invalid scheme",
			Spec: &FromSocketProcedureSpec{
				URL:     "http://localhost:8090/abc/def",
				Decoder: "csv",
			},
			ErrMsg: "invalid scheme http",
		}, {
			Name: "ok",
			Spec: &FromSocketProcedureSpec{
				URL:     "tcp://localhost:12345/abc",
				Decoder: "csv",
			},
			ErrMsg: "connection refused",
		}, {
			Name: "validation failed",
			Spec: &FromSocketProcedureSpec{
				URL:     "tcp://127.0.0.1:12345/abc",
				Decoder: "csv",
			},
			V:      url.PrivateIPValidator{},
			ErrMsg: "it connects to a private IP",
		}, {
			Name: "no such host",
			Spec: &FromSocketProcedureSpec{
				URL:     "unix://notfound:12345/abc",
				Decoder: "csv",
			},
			V:      url.PrivateIPValidator{},
			ErrMsg: "no such host",
		},
	}
	testCases.Run(t, createFromSocketSource)
}
