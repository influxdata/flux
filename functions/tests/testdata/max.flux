t_max = (table=<-) =>
  table
  |> max()
testingTest(name: "max", load: testLoadData, infile: "max.in.csv", outfile: "max.out.csv", test: t_max)