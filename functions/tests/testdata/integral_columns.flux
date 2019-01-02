t_integral_columns = (table=<-) =>
  table |> integral(columns: ["v1", "v2"], unit: 10s)

testingTest(
    name: "integral_columns",
    load: testLoadData,
    infile: "integral_columns.in.csv",
    outfile: "integral_columns.out.csv",
    test: t_integral_columns,
)