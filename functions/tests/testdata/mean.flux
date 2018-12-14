t_mean = (table=<-) =>
  table
    |> range(start:2018-05-22T19:53:26Z)
    |> group(columns:["_measurement", "_start"])
    |> mean()
    |> map(fn: (r) => ({mean: r._value}))
    |> duplicate(column:"_start", as: "_time")
    |> yield(name: "0")

testingTest(name: "mean", load: testLoadData, infile: "mean.in.csv", outfile: "mean.out.csv", test: t_mean)