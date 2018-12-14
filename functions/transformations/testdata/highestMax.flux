t_highestMax = (table=<-) =>
  table
    |> highestMax(n: 3, groupColumns: ["_measurement", "host"])

testingTest(name: "highestMax", load: fromCSV, infile: "highestMax.in.csv", outfile: "highestMax.out.csv", test: t_highestMax)