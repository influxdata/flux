package array_test


import "testing"
import "array"
import "csv"

testcase from {
    want =
        csv.from(
            csv:
                "
#datatype,string,long,string,string,string,dateTime:RFC3339,boolean
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m0,f0,tagvalue,2018-12-19T22:13:30Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:13:40Z,true
,,0,m0,f0,tagvalue,2018-12-19T22:13:50Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:14:00Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:14:10Z,true
,,0,m0,f0,tagvalue,2018-12-19T22:14:20Z,true
",
        )
    got =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:13:30Z,
                    _value: false,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:13:40Z,
                    _value: true,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:13:50Z,
                    _value: false,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:14:00Z,
                    _value: false,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:14:10Z,
                    _value: true,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:14:20Z,
                    _value: true,
                },
            ],
        )

    testing.diff(want, got) |> yield()
}

testcase from_group {
    want =
        csv.from(
            csv:
                "
#datatype,string,long,string,string,string,dateTime:RFC3339,boolean
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m0,f0,tagvalue,2018-12-19T22:13:30Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:13:40Z,true
,,0,m0,f0,tagvalue,2018-12-19T22:13:50Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:14:00Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:14:10Z,true
,,0,m0,f0,tagvalue,2018-12-19T22:14:20Z,true
",
        )
    got =
        array.from(
            rows: [
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:13:30Z,
                    _value: false,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:13:40Z,
                    _value: true,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:13:50Z,
                    _value: false,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:14:00Z,
                    _value: false,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:14:10Z,
                    _value: true,
                },
                {
                    _measurement: "m0",
                    _field: "f0",
                    t0: "tagvalue",
                    _time: 2018-12-19T22:14:20Z,
                    _value: true,
                },
            ],
        )
            |> group(columns: ["_measurement", "_field", "t0"])

    testing.diff(want, got) |> yield()
}
testcase from_pipe {
    want =
        csv.from(
            csv:
                "
#datatype,string,long,string,string,string,dateTime:RFC3339,boolean
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m0,f0,tagvalue,2018-12-19T22:13:30Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:13:40Z,true
,,0,m0,f0,tagvalue,2018-12-19T22:13:50Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:14:00Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:14:10Z,true
,,0,m0,f0,tagvalue,2018-12-19T22:14:20Z,true
",
        )
    got =
        [
            {
                _measurement: "m0",
                _field: "f0",
                t0: "tagvalue",
                _time: 2018-12-19T22:13:30Z,
                _value: false,
            },
            {
                _measurement: "m0",
                _field: "f0",
                t0: "tagvalue",
                _time: 2018-12-19T22:13:40Z,
                _value: true,
            },
            {
                _measurement: "m0",
                _field: "f0",
                t0: "tagvalue",
                _time: 2018-12-19T22:13:50Z,
                _value: false,
            },
            {
                _measurement: "m0",
                _field: "f0",
                t0: "tagvalue",
                _time: 2018-12-19T22:14:00Z,
                _value: false,
            },
            {
                _measurement: "m0",
                _field: "f0",
                t0: "tagvalue",
                _time: 2018-12-19T22:14:10Z,
                _value: true,
            },
            {
                _measurement: "m0",
                _field: "f0",
                t0: "tagvalue",
                _time: 2018-12-19T22:14:20Z,
                _value: true,
            },
        ]
            |> array.from()

    testing.diff(want, got) |> yield()
}
