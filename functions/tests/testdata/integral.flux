t_integral = (table=<-) =>
  table |> integral(unit: 10s)

testingTest(
    name: "integral",
    load: testLoadData,
    infile: "integral.in.csv",
    outfile: "integral.out.csv",
    test: t_integral,
)