t_yield = (table=<-) =>
  table
    |> sort()
    |> limit(n: 3)
    |> yield(name: "1: lowest 3")
    |> sample(n: 2, pos: 1)
    |> yield(name: "2: 2nd row")

indata = fromCSV(file: "yield.in.csv")
indata |> t_yield()

//testingTest(name: "yield", load: fromCSV, infile: "yield.in.csv", outfile: "yield.out.csv", test: t_yield)