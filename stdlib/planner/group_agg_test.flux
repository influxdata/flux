package planner_test

import "array"
import "testing"
import "csv"

// two fields, two tags keys, with three rows in each combo
input = "
#group,false,false,true,true,false,false,true,true
#datatype,string,long,string,string,dateTime:RFC3339,long,string,string
#default,_result,,,,,,,
,result,table,_field,_measurement,_time,_value,t0,t1
,,0,f0,m0,2021-07-06T23:06:30Z,3,t0v0,t1v0
,,0,f0,m0,2021-07-06T23:06:40Z,1,t0v0,t1v0
,,0,f0,m0,2021-07-06T23:06:50Z,0,t0v0,t1v0
,,1,f0,m0,2021-07-06T23:06:30Z,4,t0v0,t1v1
,,1,f0,m0,2021-07-06T23:06:40Z,3,t0v0,t1v1
,,1,f0,m0,2021-07-06T23:06:50Z,1,t0v0,t1v1
,,2,f0,m0,2021-07-06T23:06:30Z,1,t0v1,t1v0
,,2,f0,m0,2021-07-06T23:06:40Z,0,t0v1,t1v0
,,2,f0,m0,2021-07-06T23:06:50Z,4,t0v1,t1v0
,,3,f0,m0,2021-07-06T23:06:30Z,4,t0v1,t1v1
,,3,f0,m0,2021-07-06T23:06:40Z,0,t0v1,t1v1
,,3,f0,m0,2021-07-06T23:06:50Z,4,t0v1,t1v1

,,4,f1,m0,2021-07-06T23:06:30Z,0,t0v0,t1v0
,,4,f1,m0,2021-07-06T23:06:40Z,0,t0v0,t1v0
,,4,f1,m0,2021-07-06T23:06:50Z,0,t0v0,t1v0
,,5,f1,m0,2021-07-06T23:06:30Z,0,t0v0,t1v1
,,5,f1,m0,2021-07-06T23:06:40Z,4,t0v0,t1v1
,,5,f1,m0,2021-07-06T23:06:50Z,3,t0v0,t1v1
,,6,f1,m0,2021-07-06T23:06:30Z,3,t0v1,t1v0
,,6,f1,m0,2021-07-06T23:06:40Z,2,t0v1,t1v0
,,6,f1,m0,2021-07-06T23:06:50Z,1,t0v1,t1v0
,,7,f1,m0,2021-07-06T23:06:30Z,1,t0v1,t1v1
,,7,f1,m0,2021-07-06T23:06:40Z,0,t0v1,t1v1
,,7,f1,m0,2021-07-06T23:06:50Z,2,t0v1,t1v1
"

// Group + count test

// Group + count tests with no filter on field

testcase group_all_count {
    want = array.from(rows: [{"_value": 24}])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> group()
        |> count()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_one_tag_count {
    want = array.from(rows: [
        {"t0": "t0v0", "_value": 12},
        {"t0": "t0v1", "_value": 12},
    ]) |> group(columns: ["t0"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> group(columns: ["t0"])
        |> count()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_one_tag_and_field_count {
    want = array.from(rows: [
        {"t0": "t0v0", "_field": "f0", "_value": 6},
        {"t0": "t0v1", "_field": "f0", "_value": 6},
        {"t0": "t0v0", "_field": "f1", "_value": 6},
        {"t0": "t0v1", "_field": "f1", "_value": 6},
    ]) |> group(columns: ["t0", "_field"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> group(columns: ["t0", "_field"])
        |> count()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_all_tags_count {
    want = array.from(rows: [
        {"t0": "t0v0", "t1": "t1v0", "_value": 6},
        {"t0": "t0v1", "t1": "t1v0", "_value": 6},
        {"t0": "t0v0", "t1": "t1v1", "_value": 6},
        {"t0": "t0v1", "t1": "t1v1", "_value": 6},
    ]) |> group(columns: ["t0", "t1"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> group(columns: ["t0", "t1"])
        |> count()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_all_tags_and_field_count {
    want = array.from(rows: [
        {"_field": "f0", "t0": "t0v0", "t1": "t1v0", "_value": 3},
        {"_field": "f0", "t0": "t0v1", "t1": "t1v0", "_value": 3},
        {"_field": "f0", "t0": "t0v0", "t1": "t1v1", "_value": 3},
        {"_field": "f0", "t0": "t0v1", "t1": "t1v1", "_value": 3},
        {"_field": "f1", "t0": "t0v0", "t1": "t1v0", "_value": 3},
        {"_field": "f1", "t0": "t0v1", "t1": "t1v0", "_value": 3},
        {"_field": "f1", "t0": "t0v0", "t1": "t1v1", "_value": 3},
        {"_field": "f1", "t0": "t0v1", "t1": "t1v1", "_value": 3},
    ]) |> group(columns: ["t0", "t1", "_field"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> group(columns: ["t0", "t1", "_field"])
        |> count()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}


// Group + count tests with filter on field

testcase group_all_filter_field_count {
    want = array.from(rows: [{"_value": 12}])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> filter(fn: (r) => r._field == "f0")
        |> group()
        |> count()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_one_tag_filter_field_count {
    want = array.from(rows: [
        {"t0": "t0v0", "_value": 6},
        {"t0": "t0v1", "_value": 6},
    ]) |> group(columns: ["t0"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> filter(fn: (r) => r._field == "f0")
        |> group(columns: ["t0"])
        |> count()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_one_tag_and_field_filter_field_count {
    want = array.from(rows: [
        {"t0": "t0v0", "_field": "f0", "_value": 6},
        {"t0": "t0v1", "_field": "f0", "_value": 6},
    ]) |> group(columns: ["t0", "_field"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> filter(fn: (r) => r._field == "f0")
        |> group(columns: ["t0", "_field"])
        |> count()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_all_tags_filter_field_count {
    want = array.from(rows: [
        {"t0": "t0v0", "t1": "t1v0", "_value": 3},
        {"t0": "t0v1", "t1": "t1v0", "_value": 3},
        {"t0": "t0v0", "t1": "t1v1", "_value": 3},
        {"t0": "t0v1", "t1": "t1v1", "_value": 3},
    ]) |> group(columns: ["t0", "t1"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> filter(fn: (r) => r._field == "f0")
        |> group(columns: ["t0", "t1"])
        |> count()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_all_tags_and_field_filter_field_count {
    want = array.from(rows: [
        {"_field": "f0", "t0": "t0v0", "t1": "t1v0", "_value": 3},
        {"_field": "f0", "t0": "t0v1", "t1": "t1v0", "_value": 3},
        {"_field": "f0", "t0": "t0v0", "t1": "t1v1", "_value": 3},
        {"_field": "f0", "t0": "t0v1", "t1": "t1v1", "_value": 3},
    ]) |> group(columns: ["t0", "t1", "_field"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> filter(fn: (r) => r._field == "f0")
        |> group(columns: ["t0", "t1", "_field"])
        |> count()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

// Group + sum tests

// Group + sum tests with no filter on field

testcase group_all_sum {
    want = array.from(rows: [{"_value": 41}])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> group()
        |> sum()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_one_tag_sum {
    want = array.from(rows: [
        {"t0": "t0v0", "_value": 19},
        {"t0": "t0v1", "_value": 22},
    ]) |> group(columns: ["t0"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> group(columns: ["t0"])
        |> sum()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_one_tag_and_field_sum {
    want = array.from(rows: [
        {"t0": "t0v0", "_field": "f0", "_value": 12},
        {"t0": "t0v1", "_field": "f0", "_value": 13},
        {"t0": "t0v0", "_field": "f1", "_value": 7},
        {"t0": "t0v1", "_field": "f1", "_value": 9},
    ]) |> group(columns: ["t0", "_field"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> group(columns: ["t0", "_field"])
        |> sum()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_all_tags_sum {
    want = array.from(rows: [
        {"t0": "t0v0", "t1": "t1v0", "_value": 4},
        {"t0": "t0v1", "t1": "t1v0", "_value": 11},
        {"t0": "t0v0", "t1": "t1v1", "_value": 15},
        {"t0": "t0v1", "t1": "t1v1", "_value": 11},
    ]) |> group(columns: ["t0", "t1"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> group(columns: ["t0", "t1"])
        |> sum()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_all_tags_and_field_sum {
    want = array.from(rows: [
        {"_field": "f0", "t0": "t0v0", "t1": "t1v0", "_value": 4},
        {"_field": "f0", "t0": "t0v1", "t1": "t1v0", "_value": 5},
        {"_field": "f0", "t0": "t0v0", "t1": "t1v1", "_value": 8},
        {"_field": "f0", "t0": "t0v1", "t1": "t1v1", "_value": 8},
        {"_field": "f1", "t0": "t0v0", "t1": "t1v0", "_value": 0},
        {"_field": "f1", "t0": "t0v1", "t1": "t1v0", "_value": 6},
        {"_field": "f1", "t0": "t0v0", "t1": "t1v1", "_value": 7},
        {"_field": "f1", "t0": "t0v1", "t1": "t1v1", "_value": 3},
    ]) |> group(columns: ["t0", "t1", "_field"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> group(columns: ["t0", "t1", "_field"])
        |> sum()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}


// Group + sum tests with filter on field

testcase group_all_filter_field_sum {
    want = array.from(rows: [{"_value": 25}])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> filter(fn: (r) => r._field == "f0")
        |> group()
        |> sum()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_one_tag_filter_field_sum {
    want = array.from(rows: [
        {"t0": "t0v0", "_value": 12},
        {"t0": "t0v1", "_value": 13},
    ]) |> group(columns: ["t0"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> filter(fn: (r) => r._field == "f0")
        |> group(columns: ["t0"])
        |> sum()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_one_tag_and_field_filter_field_sum {
    want = array.from(rows: [
        {"t0": "t0v0", "_field": "f0", "_value": 12},
        {"t0": "t0v1", "_field": "f0", "_value": 13},
    ]) |> group(columns: ["t0", "_field"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> filter(fn: (r) => r._field == "f0")
        |> group(columns: ["t0", "_field"])
        |> sum()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_all_tags_filter_field_sum {
    want = array.from(rows: [
        {"t0": "t0v0", "t1": "t1v0", "_value": 4},
        {"t0": "t0v0", "t1": "t1v1", "_value": 8},
        {"t0": "t0v1", "t1": "t1v0", "_value": 5},
        {"t0": "t0v1", "t1": "t1v1", "_value": 8},
    ]) |> group(columns: ["t0", "t1"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> filter(fn: (r) => r._field == "f0")
        |> group(columns: ["t0", "t1"])
        |> sum()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}

testcase group_all_tags_and_field_filter_field_sum {
    want = array.from(rows: [
        {"_field": "f0", "t0": "t0v0", "t1": "t1v0", "_value": 4},
        {"_field": "f0", "t0": "t0v0", "t1": "t1v1", "_value": 8},
        {"_field": "f0", "t0": "t0v1", "t1": "t1v0", "_value": 5},
        {"_field": "f0", "t0": "t0v1", "t1": "t1v1", "_value": 8},
    ]) |> group(columns: ["t0", "t1", "_field"])
    got = testing.loadStorage(csv: input)
        |> range(start: -100y)
        |> filter(fn: (r) => r._field == "f0")
        |> group(columns: ["t0", "t1", "_field"])
        |> sum()
        |> drop(columns: ["_start", "_stop"])
    testing.diff(got, want)
}
