t_cov = () => {
    left = testLoadData(file: "cov.in.csv")
        |> range(start:2018-05-22T19:53:00Z, stop:2018-05-22T19:55:00Z)
        |> drop(columns: ["_start", "_stop"])
        |> filter(fn: (r) => r.user == "user1")
        |> group(columns: ["_measurement"])

    right = testLoadData(file: "cov.in.csv")
        |> range(start:2018-05-22T19:53:00Z, stop:2018-05-22T19:55:00Z)
        |> drop(columns: ["_start", "_stop"])
        |> filter(fn: (r) => r.user == "user2")
        |> group(columns: ["_measurement"])

    got = cov(x:left, y:right, on: ["_time", "_measurement"])
    want = testLoadData(file: "cov.out.csv")
    return assertEquals(name: "cov", want: want, got: got)
}

t_cov()
