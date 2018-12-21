t_join = () => {
    left = testLoadData(file: "join.in.csv")
        |> range(start:2018-05-22T19:53:00Z, stop:2018-05-22T19:55:00Z)
        |> drop(columns: ["_start", "_stop"])
        |> filter(fn: (r) => r.user == "user1")
        |> group(columns: ["user"])

    right = testLoadData(file: "join.in.csv")
        |> range(start:2018-05-22T19:53:00Z, stop:2018-05-22T19:55:00Z)
        |> drop(columns: ["_start", "_stop"])
        |> filter(fn: (r) => r.user == "user2")
        |> group(columns: ["_measurement"])

    got = join(tables: {left:left, right:right}, on: ["_time", "_measurement"])
    want = testLoadData(file: "join.out.csv")
    return assertEquals(name: "join", want: want, got: got)
}

t_join()
