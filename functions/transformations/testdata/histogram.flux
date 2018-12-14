t_histogram = (table=<-) =>
  table
    |> histogram(bins:[0.0,1.0,2.0])

testingTest(name: "histogram", load: fromCSV, infile: "histogram.in.csv", outfile: "histogram.out.csv", test: t_histogram)