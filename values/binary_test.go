package values_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

var (
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
		{lhs: int64(6), op: "+", rhs: intNullValue, want: nil},
		// uint + uint
		{lhs: uint64(6), op: "+", rhs: uint64(4), want: uint64(10)},
		{lhs: uint64(6), op: "+", rhs: uintNullValue, want: nil},
		// float + float
		{lhs: 4.5, op: "+", rhs: 8.2, want: 12.7},
		{lhs: 4.5, op: "+", rhs: floatNullValue, want: nil},
		// string + string
		{lhs: "a", op: "+", rhs: "b", want: "ab"},
		{lhs: "a", op: "+", rhs: stringNullValue, want: nil},
		// duration + duration
		{lhs: values.Duration(1), op: "+", rhs: values.Duration(2), want: values.Duration(3)},
		{lhs: values.Duration(1), op: "+", rhs: durationNullValue, want: nil},
		// null + null
		{lhs: nil, op: "+", rhs: nil, want: nil},
		// int - int
		{lhs: int64(6), op: "-", rhs: int64(4), want: int64(2)},
		{lhs: int64(6), op: "-", rhs: intNullValue, want: nil},
		// uint - uint
		{lhs: uint64(6), op: "-", rhs: uint64(4), want: uint64(2)},
		{lhs: uint64(6), op: "-", rhs: uintNullValue, want: nil},
		// float - float
		{lhs: 4.5, op: "-", rhs: 8.0, want: -3.5},
		{lhs: 4.5, op: "-", rhs: floatNullValue, want: nil},
		// duration - duration
		{lhs: values.Duration(5), op: "-", rhs: values.Duration(3), want: values.Duration(2)},
		{lhs: values.Duration(5), op: "-", rhs: durationNullValue, want: nil},
		// null - null
		{lhs: nil, op: "-", rhs: nil, want: nil},
		// int * int
		{lhs: int64(6), op: "*", rhs: int64(4), want: int64(24)},
		{lhs: int64(6), op: "*", rhs: intNullValue, want: nil},
		// uint * uint
		{lhs: uint64(6), op: "*", rhs: uint64(4), want: uint64(24)},
		{lhs: uint64(6), op: "*", rhs: uintNullValue, want: nil},
		// float * float
		{lhs: 4.5, op: "*", rhs: 8.2, want: 36.9},
		{lhs: 4.5, op: "*", rhs: floatNullValue, want: nil},
		// null * null
		{lhs: nil, op: "*", rhs: nil, want: nil},
		// int / int
		{lhs: int64(6), op: "/", rhs: int64(4), want: int64(1)},
		{lhs: int64(6), op: "/", rhs: intNullValue, want: nil},
		// uint / uint
		{lhs: uint64(6), op: "/", rhs: uint64(4), want: uint64(1)},
		{lhs: uint64(6), op: "/", rhs: uintNullValue, want: nil},
		// float / float
		{lhs: 5.0, op: "/", rhs: 2.0, want: 2.5},
		{lhs: 4.5, op: "/", rhs: floatNullValue, want: nil},
		// null / null
		{lhs: nil, op: "/", rhs: nil, want: nil},
		// int / zero
		{lhs: int64(8), op: "/", rhs: int64(0), want: nil, wantErr: fmt.Errorf("cannot divide by zero")},
		// int % int
		{lhs: int64(10), op: "%", rhs: int64(3), want: int64(1)},
		{lhs: int64(6), op: "%", rhs: intNullValue, want: nil},
		// uint * uint
		{lhs: uint64(6), op: "%", rhs: uint64(4), want: uint64(2)},
		{lhs: uint64(7), op: "%", rhs: uintNullValue, want: nil},
		// float * float
		{lhs: 3.8, op: "%", rhs: 8.2, want: 3.8},
		{lhs: 4.5, op: "%", rhs: floatNullValue, want: nil},
		// null * null
		{lhs: nil, op: "%", rhs: nil, want: nil},
		// int / zero
		{lhs: int64(2), op: "%", rhs: int64(0), want: nil, wantErr: fmt.Errorf("cannot mod zero")},
		// int % int
		{lhs: int64(2), op: "^", rhs: int64(4), want: float64(16)},
		{lhs: int64(6), op: "^", rhs: intNullValue, want: nil},
		// uint * uint
		{lhs: uint64(3), op: "^", rhs: uint64(2), want: float64(9)},
		{lhs: uint64(7), op: "^", rhs: uintNullValue, want: nil},
		// float * float
		{lhs: 3.8, op: "^", rhs: 2.0, want: 14.44},
		{lhs: 4.5, op: "^", rhs: floatNullValue, want: nil},
		// null * null
		{lhs: nil, op: "^", rhs: nil, want: nil},
		// int <= int
		{lhs: int64(6), op: "<=", rhs: int64(4), want: false},
		{lhs: int64(4), op: "<=", rhs: int64(4), want: true},
		{lhs: int64(4), op: "<=", rhs: int64(6), want: true},
		{lhs: int64(6), op: "<=", rhs: intNullValue, want: nil},
		// int <= uint
		{lhs: int64(6), op: "<=", rhs: uint64(4), want: false},
		{lhs: int64(4), op: "<=", rhs: uint64(4), want: true},
		{lhs: int64(4), op: "<=", rhs: uint64(6), want: true},
		{lhs: int64(6), op: "<=", rhs: uintNullValue, want: nil},
		// int <= float
		{lhs: int64(8), op: "<=", rhs: 6.7, want: false},
		{lhs: int64(6), op: "<=", rhs: 6.0, want: true},
		{lhs: int64(4), op: "<=", rhs: 6.7, want: true},
		{lhs: int64(8), op: "<=", rhs: floatNullValue, want: nil},
		// int <= null
		{lhs: int64(8), op: "<=", rhs: nil, want: nil},
		// uint <= int
		{lhs: uint64(6), op: "<=", rhs: int64(4), want: false},
		{lhs: uint64(4), op: "<=", rhs: int64(4), want: true},
		{lhs: uint64(4), op: "<=", rhs: int64(6), want: true},
		{lhs: uint64(6), op: "<=", rhs: intNullValue, want: nil},
		// uint <= uint
		{lhs: uint64(6), op: "<=", rhs: uint64(4), want: false},
		{lhs: uint64(4), op: "<=", rhs: uint64(4), want: true},
		{lhs: uint64(4), op: "<=", rhs: uint64(6), want: true},
		{lhs: uint64(6), op: "<=", rhs: uintNullValue, want: nil},
		// uint <= float
		{lhs: uint64(8), op: "<=", rhs: 6.7, want: false},
		{lhs: uint64(6), op: "<=", rhs: 6.0, want: true},
		{lhs: uint64(4), op: "<=", rhs: 6.7, want: true},
		{lhs: uint64(8), op: "<=", rhs: floatNullValue, want: nil},
		// uint <= null
		{lhs: uint64(8), op: "<=", rhs: nil, want: nil},
		// float <= int
		{lhs: 6.7, op: "<=", rhs: int64(4), want: false},
		{lhs: 6.0, op: "<=", rhs: int64(6), want: true},
		{lhs: 6.7, op: "<=", rhs: int64(8), want: true},
		{lhs: 6.7, op: "<=", rhs: intNullValue, want: nil},
		// float <= uint
		{lhs: 6.7, op: "<=", rhs: uint64(4), want: false},
		{lhs: 6.0, op: "<=", rhs: uint64(6), want: true},
		{lhs: 6.7, op: "<=", rhs: uint64(8), want: true},
		{lhs: 6.7, op: "<=", rhs: uintNullValue, want: nil},
		// float <= float
		{lhs: 8.2, op: "<=", rhs: 4.5, want: false},
		{lhs: 4.5, op: "<=", rhs: 4.5, want: true},
		{lhs: 4.5, op: "<=", rhs: 8.2, want: true},
		{lhs: 8.2, op: "<=", rhs: floatNullValue, want: nil},
		// float <= null
		{lhs: 6.7, op: "<=", rhs: nil, want: nil},
		// string <= string
		{lhs: "", op: "<=", rhs: "x", want: true},
		{lhs: "x", op: "<=", rhs: "", want: false},
		{lhs: "x", op: "<=", rhs: "x", want: true},
		{lhs: "x", op: "<=", rhs: "a", want: false},
		{lhs: "x", op: "<=", rhs: "abc", want: false},
		{lhs: "x", op: "<=", rhs: stringNullValue, want: nil},
		// string <= null
		{lhs: "x", op: "<=", rhs: nil, want: nil},
		// time <= time
		{lhs: values.Time(0), op: "<=", rhs: values.Time(1), want: true},
		{lhs: values.Time(0), op: "<=", rhs: values.Time(0), want: true},
		{lhs: values.Time(1), op: "<=", rhs: values.Time(0), want: false},
		{lhs: values.Time(0), op: "<=", rhs: nil, want: nil},
		// time <= null
		{lhs: values.Time(0), op: "<=", rhs: nil, want: nil},
		// null <= int
		{lhs: nil, op: "<=", rhs: int64(8), want: nil},
		// null <= uint
		{lhs: nil, op: "<=", rhs: uint64(8), want: nil},
		// null <= float
		{lhs: nil, op: "<=", rhs: 6.7, want: nil},
		// null <= string
		{lhs: nil, op: "<=", rhs: "x", want: nil},
		// null <= time
		{lhs: nil, op: "<=", rhs: values.Time(0), want: nil},
		// null <= null
		{lhs: nil, op: "<=", rhs: nil, want: nil},
		// int < int
		{lhs: int64(6), op: "<", rhs: int64(4), want: false},
		{lhs: int64(4), op: "<", rhs: int64(4), want: false},
		{lhs: int64(4), op: "<", rhs: int64(6), want: true},
		{lhs: int64(6), op: "<", rhs: intNullValue, want: nil},
		// int < uint
		{lhs: int64(6), op: "<", rhs: uint64(4), want: false},
		{lhs: int64(4), op: "<", rhs: uint64(4), want: false},
		{lhs: int64(4), op: "<", rhs: uint64(6), want: true},
		{lhs: int64(6), op: "<", rhs: uintNullValue, want: nil},
		// int < float
		{lhs: int64(8), op: "<", rhs: 6.7, want: false},
		{lhs: int64(6), op: "<", rhs: 6.0, want: false},
		{lhs: int64(4), op: "<", rhs: 6.7, want: true},
		{lhs: int64(8), op: "<", rhs: floatNullValue, want: nil},
		// int < null
		{lhs: int64(8), op: "<", rhs: nil, want: nil},
		// uint < int
		{lhs: uint64(6), op: "<", rhs: int64(4), want: false},
		{lhs: uint64(4), op: "<", rhs: int64(4), want: false},
		{lhs: uint64(4), op: "<", rhs: int64(6), want: true},
		{lhs: uint64(6), op: "<", rhs: intNullValue, want: nil},
		// uint < uint
		{lhs: uint64(6), op: "<", rhs: uint64(4), want: false},
		{lhs: uint64(4), op: "<", rhs: uint64(4), want: false},
		{lhs: uint64(4), op: "<", rhs: uint64(6), want: true},
		{lhs: uint64(6), op: "<", rhs: uintNullValue, want: nil},
		// uint < float
		{lhs: uint64(8), op: "<", rhs: 6.7, want: false},
		{lhs: uint64(6), op: "<", rhs: 6.0, want: false},
		{lhs: uint64(4), op: "<", rhs: 6.7, want: true},
		{lhs: uint64(8), op: "<", rhs: floatNullValue, want: nil},
		// uint < null
		{lhs: uint64(8), op: "<", rhs: nil, want: nil},
		// float < int
		{lhs: 6.7, op: "<", rhs: int64(4), want: false},
		{lhs: 6.0, op: "<", rhs: int64(6), want: false},
		{lhs: 6.7, op: "<", rhs: int64(8), want: true},
		{lhs: 6.7, op: "<", rhs: intNullValue, want: nil},
		// float < uint
		{lhs: 6.7, op: "<", rhs: uint64(4), want: false},
		{lhs: 6.0, op: "<", rhs: uint64(6), want: false},
		{lhs: 6.7, op: "<", rhs: uint64(8), want: true},
		{lhs: 6.7, op: "<", rhs: uintNullValue, want: nil},
		// float < float
		{lhs: 8.2, op: "<", rhs: 4.5, want: false},
		{lhs: 4.5, op: "<", rhs: 4.5, want: false},
		{lhs: 4.5, op: "<", rhs: 8.2, want: true},
		{lhs: 8.2, op: "<", rhs: floatNullValue, want: nil},
		// float < null
		{lhs: 8.2, op: "<", rhs: nil, want: nil},
		// string < string
		{lhs: "", op: "<", rhs: "x", want: true},
		{lhs: "x", op: "<", rhs: "", want: false},
		{lhs: "x", op: "<", rhs: "x", want: false},
		{lhs: "x", op: "<", rhs: "a", want: false},
		{lhs: "x", op: "<", rhs: "abc", want: false},
		{lhs: "x", op: "<", rhs: stringNullValue, want: nil},
		// string < null
		{lhs: "x", op: "<", rhs: nil, want: nil},
		// time < time
		{lhs: values.Time(0), op: "<", rhs: values.Time(1), want: true},
		{lhs: values.Time(0), op: "<", rhs: values.Time(0), want: false},
		{lhs: values.Time(1), op: "<", rhs: values.Time(0), want: false},
		{lhs: values.Time(0), op: "<", rhs: timeNullValue, want: nil},
		// time < null
		{lhs: values.Time(0), op: "<", rhs: nil, want: nil},
		// null < int
		{lhs: nil, op: "<", rhs: int64(8), want: nil},
		// null < uint
		{lhs: nil, op: "<", rhs: uint64(8), want: nil},
		// null < float
		{lhs: nil, op: "<", rhs: 6.7, want: nil},
		// null < string
		{lhs: nil, op: "<", rhs: "x", want: nil},
		// null < time
		{lhs: nil, op: "<", rhs: values.Time(0), want: nil},
		// null < null
		{lhs: nil, op: "<", rhs: nil, want: nil},
		// int >= int
		{lhs: int64(6), op: ">=", rhs: int64(4), want: true},
		{lhs: int64(4), op: ">=", rhs: int64(4), want: true},
		{lhs: int64(4), op: ">=", rhs: int64(6), want: false},
		{lhs: int64(6), op: ">=", rhs: intNullValue, want: nil},
		// int >= uint
		{lhs: int64(6), op: ">=", rhs: uint64(4), want: true},
		{lhs: int64(4), op: ">=", rhs: uint64(4), want: true},
		{lhs: int64(4), op: ">=", rhs: uint64(6), want: false},
		{lhs: int64(6), op: ">=", rhs: uintNullValue, want: nil},
		// int >= float
		{lhs: int64(8), op: ">=", rhs: 6.7, want: true},
		{lhs: int64(6), op: ">=", rhs: 6.0, want: true},
		{lhs: int64(4), op: ">=", rhs: 6.7, want: false},
		{lhs: int64(8), op: ">=", rhs: floatNullValue, want: nil},
		// int <= null
		{lhs: int64(8), op: ">=", rhs: nil, want: nil},
		// uint >= int
		{lhs: uint64(6), op: ">=", rhs: int64(4), want: true},
		{lhs: uint64(4), op: ">=", rhs: int64(4), want: true},
		{lhs: uint64(4), op: ">=", rhs: int64(6), want: false},
		{lhs: uint64(6), op: ">=", rhs: intNullValue, want: nil},
		// uint >= uint
		{lhs: uint64(6), op: ">=", rhs: uint64(4), want: true},
		{lhs: uint64(4), op: ">=", rhs: uint64(4), want: true},
		{lhs: uint64(4), op: ">=", rhs: uint64(6), want: false},
		{lhs: uint64(6), op: ">=", rhs: uintNullValue, want: nil},
		// uint >= float
		{lhs: uint64(8), op: ">=", rhs: 6.7, want: true},
		{lhs: uint64(6), op: ">=", rhs: 6.0, want: true},
		{lhs: uint64(4), op: ">=", rhs: 6.7, want: false},
		{lhs: uint64(8), op: ">=", rhs: floatNullValue, want: nil},
		// uint <= null
		{lhs: uint64(8), op: ">=", rhs: nil, want: nil},
		// float >= int
		{lhs: 6.7, op: ">=", rhs: int64(4), want: true},
		{lhs: 6.0, op: ">=", rhs: int64(6), want: true},
		{lhs: 6.7, op: ">=", rhs: int64(8), want: false},
		{lhs: 6.7, op: ">=", rhs: intNullValue, want: nil},
		// float >= uint
		{lhs: 6.7, op: ">=", rhs: uint64(4), want: true},
		{lhs: 6.0, op: ">=", rhs: uint64(6), want: true},
		{lhs: 6.7, op: ">=", rhs: uint64(8), want: false},
		{lhs: 6.7, op: ">=", rhs: uintNullValue, want: nil},
		// float >= float
		{lhs: 8.2, op: ">=", rhs: 4.5, want: true},
		{lhs: 4.5, op: ">=", rhs: 4.5, want: true},
		{lhs: 4.5, op: ">=", rhs: 8.2, want: false},
		{lhs: 8.2, op: ">=", rhs: floatNullValue, want: nil},
		// float <= null
		{lhs: 6.7, op: ">=", rhs: nil, want: nil},
		// string >= string
		{lhs: "", op: ">=", rhs: "x", want: false},
		{lhs: "x", op: ">=", rhs: "", want: true},
		{lhs: "x", op: ">=", rhs: "x", want: true},
		{lhs: "x", op: ">=", rhs: "a", want: true},
		{lhs: "x", op: ">=", rhs: "abc", want: true},
		{lhs: "x", op: ">=", rhs: stringNullValue, want: nil},
		// string <= null
		{lhs: "x", op: ">=", rhs: nil, want: nil},
		// time >= time
		{lhs: values.Time(0), op: ">=", rhs: values.Time(1), want: false},
		{lhs: values.Time(0), op: ">=", rhs: values.Time(0), want: true},
		{lhs: values.Time(1), op: ">=", rhs: values.Time(0), want: true},
		{lhs: values.Time(0), op: ">=", rhs: nil, want: nil},
		// time <= null
		{lhs: values.Time(0), op: ">=", rhs: nil, want: nil},
		// null <= int
		{lhs: nil, op: ">=", rhs: int64(8), want: nil},
		// null <= uint
		{lhs: nil, op: ">=", rhs: uint64(8), want: nil},
		// null <= float
		{lhs: nil, op: ">=", rhs: 6.7, want: nil},
		// null <= string
		{lhs: nil, op: ">=", rhs: "x", want: nil},
		// null <= time
		{lhs: nil, op: ">=", rhs: values.Time(0), want: nil},
		// null <= null
		{lhs: nil, op: ">=", rhs: nil, want: nil},
		// int > int
		{lhs: int64(6), op: ">", rhs: int64(4), want: true},
		{lhs: int64(4), op: ">", rhs: int64(4), want: false},
		{lhs: int64(4), op: ">", rhs: int64(6), want: false},
		{lhs: int64(6), op: ">", rhs: intNullValue, want: nil},
		// int > uint
		{lhs: int64(6), op: ">", rhs: uint64(4), want: true},
		{lhs: int64(4), op: ">", rhs: uint64(4), want: false},
		{lhs: int64(4), op: ">", rhs: uint64(6), want: false},
		{lhs: int64(6), op: ">", rhs: uintNullValue, want: nil},
		// int > float
		{lhs: int64(8), op: ">", rhs: 6.7, want: true},
		{lhs: int64(6), op: ">", rhs: 6.0, want: false},
		{lhs: int64(4), op: ">", rhs: 6.7, want: false},
		{lhs: int64(8), op: ">", rhs: floatNullValue, want: nil},
		// int < null
		{lhs: int64(8), op: ">", rhs: nil, want: nil},
		// uint > int
		{lhs: uint64(6), op: ">", rhs: int64(4), want: true},
		{lhs: uint64(4), op: ">", rhs: int64(4), want: false},
		{lhs: uint64(4), op: ">", rhs: int64(6), want: false},
		{lhs: uint64(6), op: ">", rhs: intNullValue, want: nil},
		// uint > uint
		{lhs: uint64(6), op: ">", rhs: uint64(4), want: true},
		{lhs: uint64(4), op: ">", rhs: uint64(4), want: false},
		{lhs: uint64(4), op: ">", rhs: uint64(6), want: false},
		{lhs: uint64(6), op: ">", rhs: uintNullValue, want: nil},
		// uint > float
		{lhs: uint64(8), op: ">", rhs: 6.7, want: true},
		{lhs: uint64(6), op: ">", rhs: 6.0, want: false},
		{lhs: uint64(4), op: ">", rhs: 6.7, want: false},
		{lhs: uint64(8), op: ">", rhs: floatNullValue, want: nil},
		// uint < null
		{lhs: uint64(8), op: ">", rhs: nil, want: nil},
		// float > int
		{lhs: 6.7, op: ">", rhs: int64(4), want: true},
		{lhs: 6.0, op: ">", rhs: int64(6), want: false},
		{lhs: 6.7, op: ">", rhs: int64(8), want: false},
		{lhs: 6.7, op: ">", rhs: intNullValue, want: nil},
		// float > uint
		{lhs: 6.7, op: ">", rhs: uint64(4), want: true},
		{lhs: 6.0, op: ">", rhs: uint64(6), want: false},
		{lhs: 6.7, op: ">", rhs: uint64(8), want: false},
		{lhs: 6.7, op: ">", rhs: uintNullValue, want: nil},
		// float > float
		{lhs: 8.2, op: ">", rhs: 4.5, want: true},
		{lhs: 4.5, op: ">", rhs: 8.2, want: false},
		{lhs: 4.5, op: ">", rhs: 4.5, want: false},
		{lhs: 8.2, op: ">", rhs: floatNullValue, want: nil},
		// string > string
		{lhs: "", op: ">", rhs: "x", want: false},
		{lhs: "x", op: ">", rhs: "", want: true},
		{lhs: "x", op: ">", rhs: "x", want: false},
		{lhs: "x", op: ">", rhs: "a", want: true},
		{lhs: "x", op: ">", rhs: "abc", want: true},
		{lhs: "x", op: ">", rhs: stringNullValue, want: nil},
		// time > time
		{lhs: values.Time(0), op: ">", rhs: values.Time(1), want: false},
		{lhs: values.Time(0), op: ">", rhs: values.Time(0), want: false},
		{lhs: values.Time(1), op: ">", rhs: values.Time(0), want: true},
		{lhs: values.Time(0), op: ">", rhs: timeNullValue, want: nil},
		// time < null
		{lhs: values.Time(0), op: ">", rhs: nil, want: nil},
		// null < int
		{lhs: nil, op: ">", rhs: int64(8), want: nil},
		// null < uint
		{lhs: nil, op: ">", rhs: uint64(8), want: nil},
		// null < float
		{lhs: nil, op: ">", rhs: 6.7, want: nil},
		// null < string
		{lhs: nil, op: ">", rhs: "x", want: nil},
		// null < time
		{lhs: nil, op: ">", rhs: values.Time(0), want: nil},
		// null < null
		{lhs: nil, op: ">", rhs: nil, want: nil},
		// bool == bool
		{lhs: true, op: "==", rhs: true, want: true},
		{lhs: true, op: "==", rhs: false, want: false},
		{lhs: false, op: "==", rhs: true, want: false},
		{lhs: false, op: "==", rhs: false, want: true},
		// bool == null
		{lhs: false, op: "==", rhs: nil, want: nil},
		// int == int
		{lhs: int64(4), op: "==", rhs: int64(4), want: true},
		{lhs: int64(6), op: "==", rhs: int64(4), want: false},
		{lhs: int64(4), op: "==", rhs: intNullValue, want: nil},
		// int == uint
		{lhs: int64(4), op: "==", rhs: uint64(4), want: true},
		{lhs: int64(6), op: "==", rhs: uint64(4), want: false},
		{lhs: int64(4), op: "==", rhs: uintNullValue, want: nil},
		// int == float
		{lhs: int64(4), op: "==", rhs: float64(4), want: true},
		{lhs: int64(6), op: "==", rhs: float64(4), want: false},
		{lhs: int64(4), op: "==", rhs: floatNullValue, want: nil},
		// int == null
		{lhs: int64(4), op: "==", rhs: nil, want: nil},
		// uint == int
		{lhs: uint64(4), op: "==", rhs: int64(4), want: true},
		{lhs: uint64(6), op: "==", rhs: int64(4), want: false},
		{lhs: uint64(4), op: "==", rhs: intNullValue, want: nil},
		// uint == uint
		{lhs: uint64(4), op: "==", rhs: uint64(4), want: true},
		{lhs: uint64(6), op: "==", rhs: uint64(4), want: false},
		{lhs: uint64(4), op: "==", rhs: uintNullValue, want: nil},
		// uint == float
		{lhs: uint64(4), op: "==", rhs: float64(4), want: true},
		{lhs: uint64(6), op: "==", rhs: float64(4), want: false},
		{lhs: uint64(4), op: "==", rhs: floatNullValue, want: nil},
		// uint == null
		{lhs: uint64(4), op: "==", rhs: nil, want: nil},
		// float == int
		{lhs: float64(4), op: "==", rhs: int64(4), want: true},
		{lhs: float64(6), op: "==", rhs: int64(4), want: false},
		{lhs: float64(4), op: "==", rhs: intNullValue, want: nil},
		// float == uint
		{lhs: float64(4), op: "==", rhs: uint64(4), want: true},
		{lhs: float64(6), op: "==", rhs: uint64(4), want: false},
		{lhs: float64(4), op: "==", rhs: uintNullValue, want: nil},
		// float == float
		{lhs: float64(4), op: "==", rhs: float64(4), want: true},
		{lhs: float64(6), op: "==", rhs: float64(4), want: false},
		{lhs: float64(4), op: "==", rhs: floatNullValue, want: nil},
		// float == null
		{lhs: float64(4), op: "==", rhs: nil, want: nil},
		// string == string
		{lhs: "a", op: "==", rhs: "a", want: true},
		{lhs: "a", op: "==", rhs: "b", want: false},
		{lhs: "a", op: "==", rhs: stringNullValue, want: nil},
		// string == null
		{lhs: "a", op: "==", rhs: nil, want: nil},
		// time == time
		{lhs: values.Time(0), op: "==", rhs: values.Time(1), want: false},
		{lhs: values.Time(0), op: "==", rhs: values.Time(0), want: true},
		{lhs: values.Time(1), op: "==", rhs: values.Time(0), want: false},
		{lhs: values.Time(0), op: "==", rhs: timeNullValue, want: nil},
		// time == null
		{lhs: values.Time(0), op: "==", rhs: nil, want: nil},
		// null == bool
		{lhs: nil, op: "==", rhs: true, want: nil},
		// null == int
		{lhs: nil, op: "==", rhs: int64(4), want: nil},
		// null == uint
		{lhs: nil, op: "==", rhs: uint64(4), want: nil},
		// null == float
		{lhs: nil, op: "==", rhs: float64(4), want: nil},
		// null == string
		{lhs: nil, op: "==", rhs: "a", want: nil},
		// null == time
		{lhs: nil, op: "==", rhs: values.Time(0), want: nil},
		// bool != bool
		{lhs: true, op: "!=", rhs: true, want: false},
		{lhs: true, op: "!=", rhs: false, want: true},
		{lhs: false, op: "!=", rhs: true, want: true},
		{lhs: false, op: "!=", rhs: false, want: false},
		// bool != null
		{lhs: false, op: "!=", rhs: nil, want: nil},
		// int != int
		{lhs: int64(4), op: "!=", rhs: int64(4), want: false},
		{lhs: int64(6), op: "!=", rhs: int64(4), want: true},
		{lhs: int64(4), op: "!=", rhs: intNullValue, want: nil},
		// int != uint
		{lhs: int64(4), op: "!=", rhs: uint64(4), want: false},
		{lhs: int64(6), op: "!=", rhs: uint64(4), want: true},
		{lhs: int64(4), op: "!=", rhs: uintNullValue, want: nil},
		// int != float
		{lhs: int64(4), op: "!=", rhs: float64(4), want: false},
		{lhs: int64(6), op: "!=", rhs: float64(4), want: true},
		{lhs: int64(4), op: "!=", rhs: uintNullValue, want: nil},
		// int != null
		{lhs: int64(4), op: "!=", rhs: nil, want: nil},
		// uint != int
		{lhs: uint64(4), op: "!=", rhs: int64(4), want: false},
		{lhs: uint64(6), op: "!=", rhs: int64(4), want: true},
		{lhs: uint64(4), op: "!=", rhs: intNullValue, want: nil},
		// uint != uint
		{lhs: uint64(4), op: "!=", rhs: uint64(4), want: false},
		{lhs: uint64(6), op: "!=", rhs: uint64(4), want: true},
		{lhs: uint64(4), op: "!=", rhs: uintNullValue, want: nil},
		// uint != float
		{lhs: uint64(4), op: "!=", rhs: float64(4), want: false},
		{lhs: uint64(6), op: "!=", rhs: float64(4), want: true},
		{lhs: uint64(4), op: "!=", rhs: floatNullValue, want: nil},
		// uint != null
		{lhs: uint64(4), op: "!=", rhs: nil, want: nil},
		// float != int
		{lhs: float64(4), op: "!=", rhs: int64(4), want: false},
		{lhs: float64(6), op: "!=", rhs: int64(4), want: true},
		{lhs: float64(4), op: "!=", rhs: intNullValue, want: nil},
		// float != uint
		{lhs: float64(4), op: "!=", rhs: uint64(4), want: false},
		{lhs: float64(6), op: "!=", rhs: uint64(4), want: true},
		{lhs: float64(4), op: "!=", rhs: uintNullValue, want: nil},
		// float != float
		{lhs: float64(4), op: "!=", rhs: float64(4), want: false},
		{lhs: float64(6), op: "!=", rhs: float64(4), want: true},
		{lhs: float64(4), op: "!=", rhs: floatNullValue, want: nil},
		// string != string
		{lhs: "a", op: "!=", rhs: "a", want: false},
		{lhs: "a", op: "!=", rhs: "b", want: true},
		{lhs: "a", op: "!=", rhs: stringNullValue, want: nil},
		// string == null
		{lhs: "a", op: "!=", rhs: nil, want: nil},
		// time != time
		{lhs: values.Time(0), op: "!=", rhs: values.Time(1), want: true},
		{lhs: values.Time(0), op: "!=", rhs: values.Time(0), want: false},
		{lhs: values.Time(1), op: "!=", rhs: values.Time(0), want: true},
		{lhs: values.Time(0), op: "!=", rhs: timeNullValue, want: nil},
		// time != null
		{lhs: values.Time(0), op: "!=", rhs: nil, want: nil},
		// null != bool
		{lhs: nil, op: "!=", rhs: true, want: nil},
		// null != int
		{lhs: nil, op: "!=", rhs: int64(4), want: nil},
		// null != uint
		{lhs: nil, op: "!=", rhs: uint64(4), want: nil},
		// null != float
		{lhs: nil, op: "!=", rhs: float64(4), want: nil},
		// null != string
		{lhs: nil, op: "!=", rhs: "a", want: nil},
		// null != time
		{lhs: nil, op: "!=", rhs: values.Time(0), want: nil},
		// string =~ regex
		{lhs: "abc", op: "=~", rhs: regexp.MustCompile(`.+`), want: true},
		{lhs: "abc", op: "=~", rhs: regexp.MustCompile(`b{2}`), want: false},
		{lhs: stringNullValue, op: "=~", rhs: regexp.MustCompile(`.*`), want: nil},
		// string !~ regex
		{lhs: "abc", op: "!~", rhs: regexp.MustCompile(`.+`), want: false},
		{lhs: "abc", op: "!~", rhs: regexp.MustCompile(`b{2}`), want: true},
		{lhs: stringNullValue, op: "!~", rhs: regexp.MustCompile(`.*`), want: nil},
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
			want := Value(tt.want)
			got, err := fn(left, right)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("unexpected lack of error, wanted: %s; got: %s", tt.wantErr, err)
				}
			} else if !ValueEqual(want, got) {
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
			return values.NewNull(semantic.Int)
		}
		return values.NewInt(*v)
	case *uint64:
		if v == nil {
			return values.NewNull(semantic.UInt)
		}
		return values.NewUInt(*v)
	case *float64:
		if v == nil {
			return values.NewNull(semantic.Float)
		}
		return values.NewFloat(*v)
	case *string:
		if v == nil {
			return values.NewNull(semantic.String)
		}
		return values.NewString(*v)
	case *bool:
		if v == nil {
			return values.NewNull(semantic.Bool)
		}
		return values.NewBool(*v)
	case *values.Time:
		if v == nil {
			return values.NewNull(semantic.Time)
		}
		return values.NewTime(*v)
	case *values.Duration:
		if v == nil {
			return values.NewNull(semantic.Duration)
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
