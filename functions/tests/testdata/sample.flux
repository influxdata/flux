t_sample = (table=<-) =>
  table
  |> sample(n: 3, pos: 1)
testingTest(name: "sample", load: testLoadData, infile: "sample.in.csv", outfile: "sample.out.csv", test: t_sample)