t_first = (table=<-) =>
  table
  |> first()
testingTest(name: "first", load: testLoadData, infile: "first.in.csv", outfile: "first.out.csv", test: t_first)