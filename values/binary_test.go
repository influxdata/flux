package values_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var (
	boolNullValue     = (*bool)(nil)
	intNullValue      = (*int64)(nil)
	uintNullValue     = (*uint64)(nil)
	floatNullValue    = (*float64)(nil)
	stringNullValue   = (*string)(nil)
	timeNullValue     = (*values.Time)(nil)
	durationNullValue = (*values.Duration)(nil)
)

func TestBinaryOperator(t *testing.T) {
	for _, tt := range []struct {
		lhs, rhs interface{}
		op       string
		want     interface{}
		wantErr  error
	}{
		// int + int
		{lhs: int64(6), op: "+", rhs: int64(4), want: int64(10)},
		{lhs: int64(6), op: "+", rhs: intNullValue, want: intNullValue},
		// uint + uint
		{lhs: uint64(6), op: "+", rhs: uint64(4), want: uint64(10)},
		{lhs: uint64(6), op: "+", rhs: uintNullValue, want: uintNullValue},
		// float + float
		{lhs: 4.5, op: "+", rhs: 8.2, want: 12.7},
		{lhs: 4.5, op: "+", rhs: floatNullValue, want: floatNullValue},
		// string + string
		{lhs: "a", op: "+", rhs: "b", want: "ab"},
		{lhs: "a", op: "+", rhs: stringNullValue, want: stringNullValue},
		// duration + duration
		{lhs: values.ConvertDurationNsecs(1), op: "+", rhs: values.ConvertDurationNsecs(2), want: values.ConvertDurationNsecs(3)},
		{lhs: values.ConvertDurationNsecs(1), op: "+", rhs: durationNullValue, want: durationNullValue},
		// int - int
		{lhs: int64(6), op: "-", rhs: int64(4), want: int64(2)},
		{lhs: int64(6), op: "-", rhs: intNullValue, want: intNullValue},
		// uint - uint
		{lhs: uint64(6), op: "-", rhs: uint64(4), want: uint64(2)},
		{lhs: uint64(6), op: "-", rhs: uintNullValue, want: uintNullValue},
		// float - float
		{lhs: 4.5, op: "-", rhs: 8.0, want: -3.5},
		{lhs: 4.5, op: "-", rhs: floatNullValue, want: floatNullValue},
		// duration - duration
		{lhs: values.ConvertDurationNsecs(5), op: "-", rhs: values.ConvertDurationNsecs(3), want: values.ConvertDurationNsecs(2)},
		{lhs: values.ConvertDurationNsecs(5), op: "-", rhs: durationNullValue, want: durationNullValue},
		// int * int
		{lhs: int64(6), op: "*", rhs: int64(4), want: int64(24)},
		{lhs: int64(6), op: "*", rhs: intNullValue, want: intNullValue},
		// uint * uint
		{lhs: uint64(6), op: "*", rhs: uint64(4), want: uint64(24)},
		{lhs: uint64(6), op: "*", rhs: uintNullValue, want: uintNullValue},
		// float * float
		{lhs: 4.5, op: "*", rhs: 8.2, want: 36.9},
		{lhs: 4.5, op: "*", rhs: floatNullValue, want: floatNullValue},
		// int / int
		{lhs: int64(6), op: "/", rhs: int64(4), want: int64(1)},
		{lhs: int64(6), op: "/", rhs: intNullValue, want: intNullValue},
		// uint / uint
		{lhs: uint64(6), op: "/", rhs: uint64(4), want: uint64(1)},
		{lhs: uint64(6), op: "/", rhs: uintNullValue, want: uintNullValue},
		// float / float
		{lhs: 5.0, op: "/", rhs: 2.0, want: 2.5},
		{lhs: 4.5, op: "/", rhs: floatNullValue, want: floatNullValue},
		// int / zero
		{lhs: int64(8), op: "/", rhs: int64(0), want: nil, wantErr: errors.New(codes.FailedPrecondition, "cannot divide by zero")},
		// int % int
		{lhs: int64(10), op: "%", rhs: int64(3), want: int64(1)},
		{lhs: int64(6), op: "%", rhs: intNullValue, want: intNullValue},
		// uint * uint
		{lhs: uint64(6), op: "%", rhs: uint64(4), want: uint64(2)},
		{lhs: uint64(7), op: "%", rhs: uintNullValue, want: uintNullValue},
		// float * float
		{lhs: 3.8, op: "%", rhs: 8.2, want: 3.8},
		{lhs: 4.5, op: "%", rhs: floatNullValue, want: floatNullValue},
		// int / zero
		{lhs: int64(2), op: "%", rhs: int64(0), want: nil, wantErr: errors.Newf(codes.FailedPrecondition, "cannot mod zero")},
		// int % int
		{lhs: int64(2), op: "^", rhs: int64(4), want: float64(16)},
		{lhs: int64(6), op: "^", rhs: intNullValue, want: floatNullValue},
		// uint * uint
		{lhs: uint64(3), op: "^", rhs: uint64(2), want: float64(9)},
		{lhs: uint64(7), op: "^", rhs: uintNullValue, want: floatNullValue},
		// float * float
		{lhs: 3.8, op: "^", rhs: 2.0, want: 14.44},
		{lhs: 4.5, op: "^", rhs: floatNullValue, want: floatNullValue},
		// int <= int
		{lhs: int64(6), op: "<=", rhs: int64(4), want: false},
		{lhs: int64(4), op: "<=", rhs: int64(4), want: true},
		{lhs: int64(4), op: "<=", rhs: int64(6), want: true},
		{lhs: int64(6), op: "<=", rhs: intNullValue, want: boolNullValue},
		// int <= uint
		{lhs: int64(6), op: "<=", rhs: uint64(4), want: false},
		{lhs: int64(4), op: "<=", rhs: uint64(4), want: true},
		{lhs: int64(4), op: "<=", rhs: uint64(6), want: true},
		{lhs: int64(6), op: "<=", rhs: uintNullValue, want: boolNullValue},
		// int <= float
		{lhs: int64(8), op: "<=", rhs: 6.7, want: false},
		{lhs: int64(6), op: "<=", rhs: 6.0, want: true},
		{lhs: int64(4), op: "<=", rhs: 6.7, want: true},
		{lhs: int64(8), op: "<=", rhs: floatNullValue, want: boolNullValue},
		// int <= null
		{lhs: int64(8), op: "<=", rhs: intNullValue, want: boolNullValue},
		// uint <= int
		{lhs: uint64(6), op: "<=", rhs: int64(4), want: false},
		{lhs: uint64(4), op: "<=", rhs: int64(4), want: true},
		{lhs: uint64(4), op: "<=", rhs: int64(6), want: true},
		{lhs: uint64(6), op: "<=", rhs: intNullValue, want: boolNullValue},
		// uint <= uint
		{lhs: uint64(6), op: "<=", rhs: uint64(4), want: false},
		{lhs: uint64(4), op: "<=", rhs: uint64(4), want: true},
		{lhs: uint64(4), op: "<=", rhs: uint64(6), want: true},
		{lhs: uint64(6), op: "<=", rhs: uintNullValue, want: boolNullValue},
		// uint <= float
		{lhs: uint64(8), op: "<=", rhs: 6.7, want: false},
		{lhs: uint64(6), op: "<=", rhs: 6.0, want: true},
		{lhs: uint64(4), op: "<=", rhs: 6.7, want: true},
		{lhs: uint64(8), op: "<=", rhs: floatNullValue, want: boolNullValue},
		// uint <= null
		{lhs: uint64(8), op: "<=", rhs: uintNullValue, want: boolNullValue},
		// float <= int
		{lhs: 6.7, op: "<=", rhs: int64(4), want: false},
		{lhs: 6.0, op: "<=", rhs: int64(6), want: true},
		{lhs: 6.7, op: "<=", rhs: int64(8), want: true},
		{lhs: 6.7, op: "<=", rhs: intNullValue, want: boolNullValue},
		// float <= uint
		{lhs: 6.7, op: "<=", rhs: uint64(4), want: false},
		{lhs: 6.0, op: "<=", rhs: uint64(6), want: true},
		{lhs: 6.7, op: "<=", rhs: uint64(8), want: true},
		{lhs: 6.7, op: "<=", rhs: uintNullValue, want: boolNullValue},
		// float <= float
		{lhs: 8.2, op: "<=", rhs: 4.5, want: false},
		{lhs: 4.5, op: "<=", rhs: 4.5, want: true},
		{lhs: 4.5, op: "<=", rhs: 8.2, want: true},
		{lhs: 8.2, op: "<=", rhs: floatNullValue, want: boolNullValue},
		// float <= null
		{lhs: 6.7, op: "<=", rhs: floatNullValue, want: boolNullValue},
		// string <= string
		{lhs: "", op: "<=", rhs: "x", want: true},
		{lhs: "x", op: "<=", rhs: "", want: false},
		{lhs: "x", op: "<=", rhs: "x", want: true},
		{lhs: "x", op: "<=", rhs: "a", want: false},
		{lhs: "x", op: "<=", rhs: "abc", want: false},
		{lhs: "x", op: "<=", rhs: stringNullValue, want: boolNullValue},
		// string <= null
		{lhs: "x", op: "<=", rhs: stringNullValue, want: boolNullValue},
		// time <= time
		{lhs: values.Time(0), op: "<=", rhs: values.Time(1), want: true},
		{lhs: values.Time(0), op: "<=", rhs: values.Time(0), want: true},
		{lhs: values.Time(1), op: "<=", rhs: values.Time(0), want: false},
		{lhs: values.Time(0), op: "<=", rhs: timeNullValue, want: boolNullValue},
		// time <= null
		{lhs: values.Time(0), op: "<=", rhs: timeNullValue, want: boolNullValue},
		// null <= int
		{lhs: intNullValue, op: "<=", rhs: int64(8), want: boolNullValue},
		// null <= uint
		{lhs: uintNullValue, op: "<=", rhs: uint64(8), want: boolNullValue},
		// null <= float
		{lhs: floatNullValue, op: "<=", rhs: 6.7, want: boolNullValue},
		// null <= string
		{lhs: stringNullValue, op: "<=", rhs: "x", want: boolNullValue},
		// null <= time
		{lhs: timeNullValue, op: "<=", rhs: values.Time(0), want: boolNullValue},
		// int < int
		{lhs: int64(6), op: "<", rhs: int64(4), want: false},
		{lhs: int64(4), op: "<", rhs: int64(4), want: false},
		{lhs: int64(4), op: "<", rhs: int64(6), want: true},
		{lhs: int64(6), op: "<", rhs: intNullValue, want: boolNullValue},
		// int < uint
		{lhs: int64(6), op: "<", rhs: uint64(4), want: false},
		{lhs: int64(4), op: "<", rhs: uint64(4), want: false},
		{lhs: int64(4), op: "<", rhs: uint64(6), want: true},
		{lhs: int64(6), op: "<", rhs: uintNullValue, want: boolNullValue},
		// int < float
		{lhs: int64(8), op: "<", rhs: 6.7, want: false},
		{lhs: int64(6), op: "<", rhs: 6.0, want: false},
		{lhs: int64(4), op: "<", rhs: 6.7, want: true},
		{lhs: int64(8), op: "<", rhs: floatNullValue, want: boolNullValue},
		// int < null
		{lhs: int64(8), op: "<", rhs: intNullValue, want: boolNullValue},
		// uint < int
		{lhs: uint64(6), op: "<", rhs: int64(4), want: false},
		{lhs: uint64(4), op: "<", rhs: int64(4), want: false},
		{lhs: uint64(4), op: "<", rhs: int64(6), want: true},
		{lhs: uint64(6), op: "<", rhs: intNullValue, want: boolNullValue},
		// uint < uint
		{lhs: uint64(6), op: "<", rhs: uint64(4), want: false},
		{lhs: uint64(4), op: "<", rhs: uint64(4), want: false},
		{lhs: uint64(4), op: "<", rhs: uint64(6), want: true},
		{lhs: uint64(6), op: "<", rhs: uintNullValue, want: boolNullValue},
		// uint < float
		{lhs: uint64(8), op: "<", rhs: 6.7, want: false},
		{lhs: uint64(6), op: "<", rhs: 6.0, want: false},
		{lhs: uint64(4), op: "<", rhs: 6.7, want: true},
		{lhs: uint64(8), op: "<", rhs: floatNullValue, want: boolNullValue},
		// uint < null
		{lhs: uint64(8), op: "<", rhs: uintNullValue, want: boolNullValue},
		// float < int
		{lhs: 6.7, op: "<", rhs: int64(4), want: false},
		{lhs: 6.0, op: "<", rhs: int64(6), want: false},
		{lhs: 6.7, op: "<", rhs: int64(8), want: true},
		{lhs: 6.7, op: "<", rhs: intNullValue, want: boolNullValue},
		// float < uint
		{lhs: 6.7, op: "<", rhs: uint64(4), want: false},
		{lhs: 6.0, op: "<", rhs: uint64(6), want: false},
		{lhs: 6.7, op: "<", rhs: uint64(8), want: true},
		{lhs: 6.7, op: "<", rhs: uintNullValue, want: boolNullValue},
		// float < float
		{lhs: 8.2, op: "<", rhs: 4.5, want: false},
		{lhs: 4.5, op: "<", rhs: 4.5, want: false},
		{lhs: 4.5, op: "<", rhs: 8.2, want: true},
		{lhs: 8.2, op: "<", rhs: floatNullValue, want: boolNullValue},
		// float < null
		{lhs: 8.2, op: "<", rhs: floatNullValue, want: boolNullValue},
		// string < string
		{lhs: "", op: "<", rhs: "x", want: true},
		{lhs: "x", op: "<", rhs: "", want: false},
		{lhs: "x", op: "<", rhs: "x", want: false},
		{lhs: "x", op: "<", rhs: "a", want: false},
		{lhs: "x", op: "<", rhs: "abc", want: false},
		{lhs: "x", op: "<", rhs: stringNullValue, want: boolNullValue},
		// string < null
		{lhs: "x", op: "<", rhs: stringNullValue, want: boolNullValue},
		// time < time
		{lhs: values.Time(0), op: "<", rhs: values.Time(1), want: true},
		{lhs: values.Time(0), op: "<", rhs: values.Time(0), want: false},
		{lhs: values.Time(1), op: "<", rhs: values.Time(0), want: false},
		{lhs: values.Time(0), op: "<", rhs: timeNullValue, want: boolNullValue},
		// time < null
		{lhs: values.Time(0), op: "<", rhs: timeNullValue, want: boolNullValue},
		// null < int
		{lhs: intNullValue, op: "<", rhs: int64(8), want: boolNullValue},
		// null < uint
		{lhs: uintNullValue, op: "<", rhs: uint64(8), want: boolNullValue},
		// null < float
		{lhs: floatNullValue, op: "<", rhs: 6.7, want: boolNullValue},
		// null < string
		{lhs: stringNullValue, op: "<", rhs: "x", want: boolNullValue},
		// null < time
		{lhs: timeNullValue, op: "<", rhs: values.Time(0), want: boolNullValue},
		// int >= int
		{lhs: int64(6), op: ">=", rhs: int64(4), want: true},
		{lhs: int64(4), op: ">=", rhs: int64(4), want: true},
		{lhs: int64(4), op: ">=", rhs: int64(6), want: false},
		{lhs: int64(6), op: ">=", rhs: intNullValue, want: boolNullValue},
		// int >= uint
		{lhs: int64(6), op: ">=", rhs: uint64(4), want: true},
		{lhs: int64(4), op: ">=", rhs: uint64(4), want: true},
		{lhs: int64(4), op: ">=", rhs: uint64(6), want: false},
		{lhs: int64(6), op: ">=", rhs: uintNullValue, want: boolNullValue},
		// int >= float
		{lhs: int64(8), op: ">=", rhs: 6.7, want: true},
		{lhs: int64(6), op: ">=", rhs: 6.0, want: true},
		{lhs: int64(4), op: ">=", rhs: 6.7, want: false},
		{lhs: int64(8), op: ">=", rhs: floatNullValue, want: boolNullValue},
		// int <= null
		{lhs: int64(8), op: ">=", rhs: intNullValue, want: boolNullValue},
		// uint >= int
		{lhs: uint64(6), op: ">=", rhs: int64(4), want: true},
		{lhs: uint64(4), op: ">=", rhs: int64(4), want: true},
		{lhs: uint64(4), op: ">=", rhs: int64(6), want: false},
		{lhs: uint64(6), op: ">=", rhs: intNullValue, want: boolNullValue},
		// uint >= uint
		{lhs: uint64(6), op: ">=", rhs: uint64(4), want: true},
		{lhs: uint64(4), op: ">=", rhs: uint64(4), want: true},
		{lhs: uint64(4), op: ">=", rhs: uint64(6), want: false},
		{lhs: uint64(6), op: ">=", rhs: uintNullValue, want: boolNullValue},
		// uint >= float
		{lhs: uint64(8), op: ">=", rhs: 6.7, want: true},
		{lhs: uint64(6), op: ">=", rhs: 6.0, want: true},
		{lhs: uint64(4), op: ">=", rhs: 6.7, want: false},
		{lhs: uint64(8), op: ">=", rhs: floatNullValue, want: boolNullValue},
		// uint <= null
		{lhs: uint64(8), op: ">=", rhs: uintNullValue, want: boolNullValue},
		// float >= int
		{lhs: 6.7, op: ">=", rhs: int64(4), want: true},
		{lhs: 6.0, op: ">=", rhs: int64(6), want: true},
		{lhs: 6.7, op: ">=", rhs: int64(8), want: false},
		{lhs: 6.7, op: ">=", rhs: intNullValue, want: boolNullValue},
		// float >= uint
		{lhs: 6.7, op: ">=", rhs: uint64(4), want: true},
		{lhs: 6.0, op: ">=", rhs: uint64(6), want: true},
		{lhs: 6.7, op: ">=", rhs: uint64(8), want: false},
		{lhs: 6.7, op: ">=", rhs: uintNullValue, want: boolNullValue},
		// float >= float
		{lhs: 8.2, op: ">=", rhs: 4.5, want: true},
		{lhs: 4.5, op: ">=", rhs: 4.5, want: true},
		{lhs: 4.5, op: ">=", rhs: 8.2, want: false},
		{lhs: 8.2, op: ">=", rhs: floatNullValue, want: boolNullValue},
		// float <= null
		{lhs: 6.7, op: ">=", rhs: floatNullValue, want: boolNullValue},
		// string >= string
		{lhs: "", op: ">=", rhs: "x", want: false},
		{lhs: "x", op: ">=", rhs: "", want: true},
		{lhs: "x", op: ">=", rhs: "x", want: true},
		{lhs: "x", op: ">=", rhs: "a", want: true},
		{lhs: "x", op: ">=", rhs: "abc", want: true},
		{lhs: "x", op: ">=", rhs: stringNullValue, want: boolNullValue},
		// string <= null
		{lhs: "x", op: ">=", rhs: stringNullValue, want: boolNullValue},
		// time >= time
		{lhs: values.Time(0), op: ">=", rhs: values.Time(1), want: false},
		{lhs: values.Time(0), op: ">=", rhs: values.Time(0), want: true},
		{lhs: values.Time(1), op: ">=", rhs: values.Time(0), want: true},
		{lhs: values.Time(0), op: ">=", rhs: timeNullValue, want: boolNullValue},
		// time <= null
		{lhs: values.Time(0), op: ">=", rhs: timeNullValue, want: boolNullValue},
		// null <= int
		{lhs: intNullValue, op: ">=", rhs: int64(8), want: boolNullValue},
		// null <= uint
		{lhs: uintNullValue, op: ">=", rhs: uint64(8), want: boolNullValue},
		// null <= float
		{lhs: floatNullValue, op: ">=", rhs: 6.7, want: boolNullValue},
		// null <= string
		{lhs: stringNullValue, op: ">=", rhs: "x", want: boolNullValue},
		// null <= time
		{lhs: timeNullValue, op: ">=", rhs: values.Time(0), want: boolNullValue},
		// int > int
		{lhs: int64(6), op: ">", rhs: int64(4), want: true},
		{lhs: int64(4), op: ">", rhs: int64(4), want: false},
		{lhs: int64(4), op: ">", rhs: int64(6), want: false},
		{lhs: int64(6), op: ">", rhs: intNullValue, want: boolNullValue},
		// int > uint
		{lhs: int64(6), op: ">", rhs: uint64(4), want: true},
		{lhs: int64(4), op: ">", rhs: uint64(4), want: false},
		{lhs: int64(4), op: ">", rhs: uint64(6), want: false},
		{lhs: int64(6), op: ">", rhs: uintNullValue, want: boolNullValue},
		// int > float
		{lhs: int64(8), op: ">", rhs: 6.7, want: true},
		{lhs: int64(6), op: ">", rhs: 6.0, want: false},
		{lhs: int64(4), op: ">", rhs: 6.7, want: false},
		{lhs: int64(8), op: ">", rhs: floatNullValue, want: boolNullValue},
		// int < null
		{lhs: int64(8), op: ">", rhs: intNullValue, want: boolNullValue},
		// uint > int
		{lhs: uint64(6), op: ">", rhs: int64(4), want: true},
		{lhs: uint64(4), op: ">", rhs: int64(4), want: false},
		{lhs: uint64(4), op: ">", rhs: int64(6), want: false},
		{lhs: uint64(6), op: ">", rhs: intNullValue, want: boolNullValue},
		// uint > uint
		{lhs: uint64(6), op: ">", rhs: uint64(4), want: true},
		{lhs: uint64(4), op: ">", rhs: uint64(4), want: false},
		{lhs: uint64(4), op: ">", rhs: uint64(6), want: false},
		{lhs: uint64(6), op: ">", rhs: uintNullValue, want: boolNullValue},
		// uint > float
		{lhs: uint64(8), op: ">", rhs: 6.7, want: true},
		{lhs: uint64(6), op: ">", rhs: 6.0, want: false},
		{lhs: uint64(4), op: ">", rhs: 6.7, want: false},
		{lhs: uint64(8), op: ">", rhs: floatNullValue, want: boolNullValue},
		// uint < null
		{lhs: uint64(8), op: ">", rhs: uintNullValue, want: boolNullValue},
		// float > int
		{lhs: 6.7, op: ">", rhs: int64(4), want: true},
		{lhs: 6.0, op: ">", rhs: int64(6), want: false},
		{lhs: 6.7, op: ">", rhs: int64(8), want: false},
		{lhs: 6.7, op: ">", rhs: intNullValue, want: boolNullValue},
		// float > uint
		{lhs: 6.7, op: ">", rhs: uint64(4), want: true},
		{lhs: 6.0, op: ">", rhs: uint64(6), want: false},
		{lhs: 6.7, op: ">", rhs: uint64(8), want: false},
		{lhs: 6.7, op: ">", rhs: uintNullValue, want: boolNullValue},
		// float > float
		{lhs: 8.2, op: ">", rhs: 4.5, want: true},
		{lhs: 4.5, op: ">", rhs: 8.2, want: false},
		{lhs: 4.5, op: ">", rhs: 4.5, want: false},
		{lhs: 8.2, op: ">", rhs: floatNullValue, want: boolNullValue},
		// string > string
		{lhs: "", op: ">", rhs: "x", want: false},
		{lhs: "x", op: ">", rhs: "", want: true},
		{lhs: "x", op: ">", rhs: "x", want: false},
		{lhs: "x", op: ">", rhs: "a", want: true},
		{lhs: "x", op: ">", rhs: "abc", want: true},
		{lhs: "x", op: ">", rhs: stringNullValue, want: boolNullValue},
		// time > time
		{lhs: values.Time(0), op: ">", rhs: values.Time(1), want: false},
		{lhs: values.Time(0), op: ">", rhs: values.Time(0), want: false},
		{lhs: values.Time(1), op: ">", rhs: values.Time(0), want: true},
		{lhs: values.Time(0), op: ">", rhs: timeNullValue, want: boolNullValue},
		// time < null
		{lhs: values.Time(0), op: ">", rhs: timeNullValue, want: boolNullValue},
		// null < int
		{lhs: intNullValue, op: ">", rhs: int64(8), want: boolNullValue},
		// null < uint
		{lhs: uintNullValue, op: ">", rhs: uint64(8), want: boolNullValue},
		// null < float
		{lhs: floatNullValue, op: ">", rhs: 6.7, want: boolNullValue},
		// null < string
		{lhs: stringNullValue, op: ">", rhs: "x", want: boolNullValue},
		// null < time
		{lhs: timeNullValue, op: ">", rhs: values.Time(0), want: boolNullValue},
		// bool == bool
		{lhs: true, op: "==", rhs: true, want: true},
		{lhs: true, op: "==", rhs: false, want: false},
		{lhs: false, op: "==", rhs: true, want: false},
		{lhs: false, op: "==", rhs: false, want: true},
		// bool == null
		{lhs: false, op: "==", rhs: boolNullValue, want: boolNullValue},
		// int == int
		{lhs: int64(4), op: "==", rhs: int64(4), want: true},
		{lhs: int64(6), op: "==", rhs: int64(4), want: false},
		{lhs: int64(4), op: "==", rhs: intNullValue, want: boolNullValue},
		// int == uint
		{lhs: int64(4), op: "==", rhs: uint64(4), want: true},
		{lhs: int64(6), op: "==", rhs: uint64(4), want: false},
		{lhs: int64(4), op: "==", rhs: uintNullValue, want: boolNullValue},
		// int == float
		{lhs: int64(4), op: "==", rhs: float64(4), want: true},
		{lhs: int64(6), op: "==", rhs: float64(4), want: false},
		{lhs: int64(4), op: "==", rhs: floatNullValue, want: boolNullValue},
		// int == null
		{lhs: int64(4), op: "==", rhs: intNullValue, want: boolNullValue},
		// uint == int
		{lhs: uint64(4), op: "==", rhs: int64(4), want: true},
		{lhs: uint64(6), op: "==", rhs: int64(4), want: false},
		{lhs: uint64(4), op: "==", rhs: intNullValue, want: boolNullValue},
		// uint == uint
		{lhs: uint64(4), op: "==", rhs: uint64(4), want: true},
		{lhs: uint64(6), op: "==", rhs: uint64(4), want: false},
		{lhs: uint64(4), op: "==", rhs: uintNullValue, want: boolNullValue},
		// uint == float
		{lhs: uint64(4), op: "==", rhs: float64(4), want: true},
		{lhs: uint64(6), op: "==", rhs: float64(4), want: false},
		{lhs: uint64(4), op: "==", rhs: floatNullValue, want: boolNullValue},
		// uint == null
		{lhs: uint64(4), op: "==", rhs: uintNullValue, want: boolNullValue},
		// float == int
		{lhs: float64(4), op: "==", rhs: int64(4), want: true},
		{lhs: float64(6), op: "==", rhs: int64(4), want: false},
		{lhs: float64(4), op: "==", rhs: intNullValue, want: boolNullValue},
		// float == uint
		{lhs: float64(4), op: "==", rhs: uint64(4), want: true},
		{lhs: float64(6), op: "==", rhs: uint64(4), want: false},
		{lhs: float64(4), op: "==", rhs: uintNullValue, want: boolNullValue},
		// float == float
		{lhs: float64(4), op: "==", rhs: float64(4), want: true},
		{lhs: float64(6), op: "==", rhs: float64(4), want: false},
		{lhs: float64(4), op: "==", rhs: floatNullValue, want: boolNullValue},
		// float == null
		{lhs: float64(4), op: "==", rhs: floatNullValue, want: boolNullValue},
		// string == string
		{lhs: "a", op: "==", rhs: "a", want: true},
		{lhs: "a", op: "==", rhs: "b", want: false},
		{lhs: "a", op: "==", rhs: stringNullValue, want: boolNullValue},
		// string == null
		{lhs: "a", op: "==", rhs: stringNullValue, want: boolNullValue},
		// time == time
		{lhs: values.Time(0), op: "==", rhs: values.Time(1), want: false},
		{lhs: values.Time(0), op: "==", rhs: values.Time(0), want: true},
		{lhs: values.Time(1), op: "==", rhs: values.Time(0), want: false},
		{lhs: values.Time(0), op: "==", rhs: timeNullValue, want: boolNullValue},
		// time == null
		{lhs: values.Time(0), op: "==", rhs: timeNullValue, want: boolNullValue},
		// null == bool
		{lhs: boolNullValue, op: "==", rhs: true, want: boolNullValue},
		// null == int
		{lhs: intNullValue, op: "==", rhs: int64(4), want: boolNullValue},
		// null == uint
		{lhs: uintNullValue, op: "==", rhs: uint64(4), want: boolNullValue},
		// null == float
		{lhs: floatNullValue, op: "==", rhs: float64(4), want: boolNullValue},
		// null == string
		{lhs: stringNullValue, op: "==", rhs: "a", want: boolNullValue},
		// null == time
		{lhs: timeNullValue, op: "==", rhs: values.Time(0), want: boolNullValue},
		// bool != bool
		{lhs: true, op: "!=", rhs: true, want: false},
		{lhs: true, op: "!=", rhs: false, want: true},
		{lhs: false, op: "!=", rhs: true, want: true},
		{lhs: false, op: "!=", rhs: false, want: false},
		// bool != null
		{lhs: false, op: "!=", rhs: boolNullValue, want: boolNullValue},
		// int != int
		{lhs: int64(4), op: "!=", rhs: int64(4), want: false},
		{lhs: int64(6), op: "!=", rhs: int64(4), want: true},
		{lhs: int64(4), op: "!=", rhs: intNullValue, want: boolNullValue},
		// int != uint
		{lhs: int64(4), op: "!=", rhs: uint64(4), want: false},
		{lhs: int64(6), op: "!=", rhs: uint64(4), want: true},
		{lhs: int64(4), op: "!=", rhs: uintNullValue, want: boolNullValue},
		// int != float
		{lhs: int64(4), op: "!=", rhs: float64(4), want: false},
		{lhs: int64(6), op: "!=", rhs: float64(4), want: true},
		{lhs: int64(4), op: "!=", rhs: uintNullValue, want: boolNullValue},
		// int != null
		{lhs: int64(4), op: "!=", rhs: intNullValue, want: boolNullValue},
		// uint != int
		{lhs: uint64(4), op: "!=", rhs: int64(4), want: false},
		{lhs: uint64(6), op: "!=", rhs: int64(4), want: true},
		{lhs: uint64(4), op: "!=", rhs: intNullValue, want: boolNullValue},
		// uint != uint
		{lhs: uint64(4), op: "!=", rhs: uint64(4), want: false},
		{lhs: uint64(6), op: "!=", rhs: uint64(4), want: true},
		{lhs: uint64(4), op: "!=", rhs: uintNullValue, want: boolNullValue},
		// uint != float
		{lhs: uint64(4), op: "!=", rhs: float64(4), want: false},
		{lhs: uint64(6), op: "!=", rhs: float64(4), want: true},
		{lhs: uint64(4), op: "!=", rhs: floatNullValue, want: boolNullValue},
		// uint != null
		{lhs: uint64(4), op: "!=", rhs: uintNullValue, want: boolNullValue},
		// float != int
		{lhs: float64(4), op: "!=", rhs: int64(4), want: false},
		{lhs: float64(6), op: "!=", rhs: int64(4), want: true},
		{lhs: float64(4), op: "!=", rhs: intNullValue, want: boolNullValue},
		// float != uint
		{lhs: float64(4), op: "!=", rhs: uint64(4), want: false},
		{lhs: float64(6), op: "!=", rhs: uint64(4), want: true},
		{lhs: float64(4), op: "!=", rhs: uintNullValue, want: boolNullValue},
		// float != float
		{lhs: float64(4), op: "!=", rhs: float64(4), want: false},
		{lhs: float64(6), op: "!=", rhs: float64(4), want: true},
		{lhs: float64(4), op: "!=", rhs: floatNullValue, want: boolNullValue},
		// string != string
		{lhs: "a", op: "!=", rhs: "a", want: false},
		{lhs: "a", op: "!=", rhs: "b", want: true},
		{lhs: "a", op: "!=", rhs: stringNullValue, want: boolNullValue},
		// string == null
		{lhs: "a", op: "!=", rhs: stringNullValue, want: boolNullValue},
		// time != time
		{lhs: values.Time(0), op: "!=", rhs: values.Time(1), want: true},
		{lhs: values.Time(0), op: "!=", rhs: values.Time(0), want: false},
		{lhs: values.Time(1), op: "!=", rhs: values.Time(0), want: true},
		{lhs: values.Time(0), op: "!=", rhs: timeNullValue, want: boolNullValue},
		// time != null
		{lhs: values.Time(0), op: "!=", rhs: timeNullValue, want: boolNullValue},
		// null != bool
		{lhs: boolNullValue, op: "!=", rhs: true, want: boolNullValue},
		// null != int
		{lhs: intNullValue, op: "!=", rhs: int64(4), want: boolNullValue},
		// null != uint
		{lhs: uintNullValue, op: "!=", rhs: uint64(4), want: boolNullValue},
		// null != float
		{lhs: floatNullValue, op: "!=", rhs: float64(4), want: boolNullValue},
		// null != string
		{lhs: stringNullValue, op: "!=", rhs: "a", want: boolNullValue},
		// null != time
		{lhs: timeNullValue, op: "!=", rhs: values.Time(0), want: boolNullValue},
		// string =~ regex
		{lhs: "abc", op: "=~", rhs: regexp.MustCompile(`.+`), want: true},
		{lhs: "abc", op: "=~", rhs: regexp.MustCompile(`b{2}`), want: false},
		{lhs: stringNullValue, op: "=~", rhs: regexp.MustCompile(`.*`), want: boolNullValue},
		// string !~ regex
		{lhs: "abc", op: "!~", rhs: regexp.MustCompile(`.+`), want: false},
		{lhs: "abc", op: "!~", rhs: regexp.MustCompile(`b{2}`), want: true},
		{lhs: stringNullValue, op: "!~", rhs: regexp.MustCompile(`.*`), want: boolNullValue},
	} {
		t.Run(fmt.Sprintf("%v %s %v", tt.lhs, tt.op, tt.rhs), func(t *testing.T) {
			left, right := Value(tt.lhs), Value(tt.rhs)
			fn, err := values.LookupBinaryFunction(values.BinaryFuncSignature{
				Operator: ast.OperatorLookup(tt.op),
				Left:     left.Type().Nature(),
				Right:    right.Type().Nature(),
			})
			if err != nil {
				t.Fatal(err)
			}
			got, err := fn(left, right)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("unexpected lack of error, wanted: %s; got: %s", tt.wantErr, err)
				}
			} else if want := Value(tt.want); !ValueEqual(want, got) {
				t.Fatalf("unexpected value -want/+got\n\t- %s\n\t+ %s", want, got)
			}
		})
	}
}

// Value converts an interface into a value.
//
// If the interface is a pointer to a basic type that is null,
// it will create a null with the type of the pointer.
//
// Otherwise, it will use values.New to create the value.
func Value(v interface{}) values.Value {
	switch v := v.(type) {
	case *int64:
		if v == nil {
			return values.NewNull(semantic.BasicInt)
		}
		return values.NewInt(*v)
	case *uint64:
		if v == nil {
			return values.NewNull(semantic.BasicUint)
		}
		return values.NewUInt(*v)
	case *float64:
		if v == nil {
			return values.NewNull(semantic.BasicFloat)
		}
		return values.NewFloat(*v)
	case *string:
		if v == nil {
			return values.NewNull(semantic.BasicString)
		}
		return values.NewString(*v)
	case *bool:
		if v == nil {
			return values.NewNull(semantic.BasicBool)
		}
		return values.NewBool(*v)
	case *values.Time:
		if v == nil {
			return values.NewNull(semantic.BasicTime)
		}
		return values.NewTime(*v)
	case *values.Duration:
		if v == nil {
			return values.NewNull(semantic.BasicDuration)
		}
		return values.NewDuration(*v)
	}
	return values.New(v)
}

// ValueEqual compares two values and considers two null
// values to be equal to each other.
//
// The standard equality operator does not consider null
// to be equal to null.
func ValueEqual(l, r values.Value) bool {
	if l.IsNull() && r.IsNull() {
		return true
	}
	return l.Equal(r)
}
