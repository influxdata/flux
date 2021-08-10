package hex

builtin int : (v: A) => int
builtin string : (v: A) => string
builtin uint : (v: A) => uint

toString = (tables=<-) => tables |> map(fn: (r) => ({r with _value: string(v: r._value)}))
toInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: int(v: r._value)}))
toUInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: uint(v: r._value)}))