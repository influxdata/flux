package sql

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	flux "github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// additional and seperate tests that can be run without needing functions to be Exported in sql, just to be testable
func TestCorrectBatchSize(t *testing.T) {
	// given the combination of row width and supplied batchSize argument from user, verify that it is modified as required
	userBatchSize := 1000
	rowWidth := 10
	correctedSize := correctBatchSize(userBatchSize, rowWidth)
	if !cmp.Equal(99, correctedSize) {
		t.Log(cmp.Diff(90, correctedSize))
		t.Fail()
	}

	// verify that the batchSoze is not lower than the width of a single row - if it ever is, we have a big problem
	userBatchSize = 1
	correctedSize = correctBatchSize(userBatchSize, rowWidth)
	if !cmp.Equal(10, correctedSize) {
		t.Log(cmp.Diff(10, correctedSize))
		t.Fail()
	}

	userBatchSize = -1
	correctedSize = correctBatchSize(userBatchSize, rowWidth)
	if !cmp.Equal(10, correctedSize) {
		t.Log(cmp.Diff(10, correctedSize))
		t.Fail()
	}
}

func TestTranslation(t *testing.T) {
	// verify that this works as it should - used here to avoid exporting this functionality
	columnLabel := "apples"
	// verify invalid return error
	_, err := getTranslationFunc("bananas")
	if !cmp.Equal(errors.New(codes.Internal, "invalid driverName: bananas").Error(), err.Error()) {
		t.Log(cmp.Diff(errors.New(codes.Internal, "invalid driverName: bananas").Error(), err.Error()))
		t.Fail()
	}

	// verify that valid returns expected hapiness for SQLITE
	sqlT, err := getTranslationFunc("sqlite3")
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	// float
	v, err := sqlT()(flux.TFloat, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" FLOAT", v) {
		t.Log(cmp.Diff(columnLabel+" FLOAT", v))
		t.Fail()
	}
	// int
	v, err = sqlT()(flux.TInt, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" INT", v) {
		t.Log(cmp.Diff(columnLabel+" INT", v))
		t.Fail()
	}
	// uint
	v, err = sqlT()(flux.TUInt, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" INT", v) {
		t.Log(cmp.Diff(columnLabel+" INT", v))
		t.Fail()
	}
	// string
	v, err = sqlT()(flux.TString, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" TEXT", v) {
		t.Log(cmp.Diff(columnLabel+" TEXT", v))
		t.Fail()
	}
	// time
	v, err = sqlT()(flux.TTime, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" DATETIME", v) {
		t.Log(cmp.Diff(columnLabel+" DATETIME", v))
		t.Fail()
	}
	// as SQLITE has NO BOOLEAN column type, we need to return an error rather than doing implicit conversions
	v, err = sqlT()(flux.TBool, columnLabel)
	if cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal("DB does not support column type bool", err.Error()) {
		t.Log(cmp.Diff("DB does not support column type bool", err.Error()))
		t.Fail()
	}

	// verify that valid returns expected hapiness for postgres
	sqlT, err = getTranslationFunc("postgres")
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}

	// float
	v, err = sqlT()(flux.TFloat, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" FLOAT", v) {
		t.Log(cmp.Diff(columnLabel+" FLOAT", v))
		t.Fail()
	}
	// int
	v, err = sqlT()(flux.TInt, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" BIGINT", v) {
		t.Log(cmp.Diff(columnLabel+" BIGINT", v))
		t.Fail()
	}
	// uint
	v, err = sqlT()(flux.TUInt, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" BIGINT", v) {
		t.Log(cmp.Diff(columnLabel+" BIGINT", v))
		t.Fail()
	}
	// string
	v, err = sqlT()(flux.TString, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" text", v) {
		t.Log(cmp.Diff(columnLabel+" text", v))
		t.Fail()
	}
	// time
	v, err = sqlT()(flux.TTime, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" TIMESTAMP", v) {
		t.Log(cmp.Diff(columnLabel+" TIMESTAMP", v))
		t.Fail()
	}
	// bool
	v, err = sqlT()(flux.TBool, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" BOOL", v) {
		t.Log(cmp.Diff(columnLabel+" BOOL", v))
		t.Fail()
	}

	// verify that valid returns expected hapiness for MySQL
	sqlT, err = getTranslationFunc("mysql")
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}

	// float
	v, err = sqlT()(flux.TFloat, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" FLOAT", v) {
		t.Log(cmp.Diff(columnLabel+" FLOAT", v))
		t.Fail()
	}
	// int
	v, err = sqlT()(flux.TInt, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" BIGINT", v) {
		t.Log(cmp.Diff(columnLabel+" BIGINT", v))
		t.Fail()
	}
	// uint
	v, err = sqlT()(flux.TUInt, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" BIGINT", v) {
		t.Log(cmp.Diff(columnLabel+" BIGINT", v))
		t.Fail()
	}
	// string
	v, err = sqlT()(flux.TString, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" TEXT(16383)", v) {
		t.Log(cmp.Diff(columnLabel+" TEXT(16383)", v))
		t.Fail()
	}
	// time
	v, err = sqlT()(flux.TTime, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" DATETIME", v) {
		t.Log(cmp.Diff(columnLabel+" DATETIME", v))
		t.Fail()
	}
	// bool
	v, err = sqlT()(flux.TBool, columnLabel)
	if !cmp.Equal(nil, err) {
		t.Log(cmp.Diff(nil, err))
		t.Fail()
	}
	if !cmp.Equal(columnLabel+" BOOL", v) {
		t.Log(cmp.Diff(columnLabel+" BOOL", v))
		t.Fail()
	}

}
