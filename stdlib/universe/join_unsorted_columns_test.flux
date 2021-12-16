package universe_test


import "array"
import "testing"

lhs =
    array.from(
        rows: [
            {interface: "port1", router: "router1", _value: 123},
            {interface: "port1", router: "router2", _value: 855},
            {interface: "port2", router: "router1", _value: 1235},
            {interface: "port2", router: "router2", _value: 2432},
        ],
    )

rhs =
    array.from(
        rows: [
            {router: "router1", interface: "port1", data: "acme corp"},
            {router: "router1", interface: "port2", data: "foo corp"},
            {router: "router2", interface: "port2", data: "foo corp"},
            {router: "router2", interface: "port1", data: "bar corp"},
        ],
    )

testcase join_many_groups_to_one {
    want =
        array.from(
            rows: [
                {interface: "port1", router: "router1", _value: 123, data: "acme corp"},
                {interface: "port1", router: "router2", _value: 855, data: "bar corp"},
                {interface: "port2", router: "router1", _value: 1235, data: "foo corp"},
                {interface: "port2", router: "router2", _value: 2432, data: "foo corp"},
            ],
        )
            |> group(columns: ["interface", "router"])

    left = lhs |> group(columns: ["interface", "router"])
    right = rhs |> group(columns: [])

    got = join(tables: {left: left, right: right}, on: ["router", "interface"])

    testing.diff(want: want, got: got)
}

testcase join_one_to_one {
    want =
        array.from(
            rows: [
                {interface: "port1", router: "router1", _value: 123, data: "acme corp"},
                {interface: "port1", router: "router2", _value: 855, data: "bar corp"},
                {interface: "port2", router: "router1", _value: 1235, data: "foo corp"},
                {interface: "port2", router: "router2", _value: 2432, data: "foo corp"},
            ],
        )
            |> group(columns: [])

    left = lhs |> group(columns: [])
    right = rhs |> group(columns: [])

    got = join(tables: {left: left, right: right}, on: ["router", "interface"])

    testing.diff(want: want, got: got)
}
