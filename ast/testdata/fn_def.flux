foo = () => from(bucket:"testdb")
bar = (x=<-) => x |> filter(fn: (r) => r.name =~ /.*0/)
baz = (y=<-) => y |> map(fn: (r) => {_time: r._time, io_time: r._value})

foo() |> bar() |> baz()