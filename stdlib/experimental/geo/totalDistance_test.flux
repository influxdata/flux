package geo_test


import "array"
import "experimental/geo"
import "testing"

testcase geo_total_distance_testcase {
    got =
        array.from(
            rows: [
                {id: "ABC1", _time: 2022-01-01T00:00:00Z, lat: 85.1, lon: 42.2},
                {id: "ABC1", _time: 2022-01-01T01:00:00Z, lat: 71.3, lon: 50.8},
                {id: "ABC1", _time: 2022-01-01T02:00:00Z, lat: 63.1, lon: 62.3},
                {id: "ABC1", _time: 2022-01-01T03:00:00Z, lat: 50.6, lon: 74.9},
                {id: "DEF2", _time: 2022-01-01T00:00:00Z, lat: -10.8, lon: -12.2},
                {id: "DEF2", _time: 2022-01-01T01:00:00Z, lat: -16.3, lon: -0.8},
                {id: "DEF2", _time: 2022-01-01T02:00:00Z, lat: -23.2, lon: 12.3},
                {id: "DEF2", _time: 2022-01-01T03:00:00Z, lat: -30.4, lon: 24.9},
            ],
        )
            |> group(columns: ["id"])
            |> geo.totalDistance()

    want =
        array.from(
            rows: [
                {id: "ABC1", _value: 4157.144498077607},
                {id: "DEF2", _value: 4428.129653320098},
            ],
        )
            |> group(columns: ["id"])

    testing.diff(got: got, want: want)
}

testcase geo_total_distance_custom_column_testcase {
    got =
        array.from(
            rows: [
                {id: "ABC1", _time: 2022-01-01T00:00:00Z, lat: 85.1, lon: 42.2},
                {id: "ABC1", _time: 2022-01-01T01:00:00Z, lat: 71.3, lon: 50.8},
                {id: "ABC1", _time: 2022-01-01T02:00:00Z, lat: 63.1, lon: 62.3},
                {id: "ABC1", _time: 2022-01-01T03:00:00Z, lat: 50.6, lon: 74.9},
                {id: "DEF2", _time: 2022-01-01T00:00:00Z, lat: -10.8, lon: -12.2},
                {id: "DEF2", _time: 2022-01-01T01:00:00Z, lat: -16.3, lon: -0.8},
                {id: "DEF2", _time: 2022-01-01T02:00:00Z, lat: -23.2, lon: 12.3},
                {id: "DEF2", _time: 2022-01-01T03:00:00Z, lat: -30.4, lon: 24.9},
            ],
        )
            |> group(columns: ["id"])
            |> geo.totalDistance(outputColumn: "foo")

    want =
        array.from(
            rows: [{id: "ABC1", foo: 4157.144498077607}, {id: "DEF2", foo: 4428.129653320098}],
        )
            |> group(columns: ["id"])

    testing.diff(got: got, want: want)
}
