package hex

builtin int : (v: string) => int
builtin string : (v: A) => string
builtin uint : (v: string) => uint
builtin bytes : (v: string) => bytes

toString = (tables=<-) => tables |> map(fn: (r) => ({r with _value: string(v: r._value)}))
toInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: int(v: r._value)}))
toUInt = (tables=<-) => tables |> map(fn: (r) => ({r with _value: uint(v: r._value)}))