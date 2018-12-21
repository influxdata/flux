t_covariance = (tables=<-) =>
  tables
    |> range(start: 2018-05-22T19:53:26Z)
    |> covariance(columns: ["x", "y"])

testingTest(name: "t_covariance", load: testLoadData, infile: "covariance.in.csv", outfile: "covariance.out.csv", test: t_covariance)
