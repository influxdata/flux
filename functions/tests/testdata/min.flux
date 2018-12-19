t_min = (table=<-) =>
  table
  |> min()
testingTest(name: "min", load: testLoadData, infile: "min.in.csv", outfile: "min.out.csv", test: t_min)