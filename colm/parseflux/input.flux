0.
72.40
072.40  // == 72.40
2.71828
.26

1s
10d
1h15m  // 1 hour and 15 minutes
5w
1mo5d  // 1 month and 5 days
-1mo5d // negative 1 month and 5 days
5w * 2 // 10 weeks

2018-01-01
2018-01-01T00:00:00.1Z
2018-01-01T00:00:00.00000001Z
2018-01-01T00:00:00.00000002
2018-01-01T00:00:00.00000001+00:00

n = 1
m = 2
x = 5.4
f = () => {
    n = "a"
    m = "b"
    return n + m
}

option severity = ["low", "moderate", "high"]
option alert.severity = ["low", "critical"]

{a: 1, b: 2, c: 3}
{a, b, c}
{o with x: 5, y: 5}
{o with a, b}

() => 1
(a, b) => a + b
(x=1, y=1) => x * y
(a, b, c) => {
    d = a + b
    return d / c
}

add = (a,b) => a + b
mul = (a,b) => a * b

f(a:1, b:9.6)
float(v:1)

add(a: a, b: b) 
add(a, b)

add = (a,b) => a + b
a = 1
b = 2

add(a: a, b)
add(a, b: b)

foo |> bar |> baz

obj.k //asdlkfj
obj["k"]

color = if code == 0 then "green" else if code == 1 then "yellow" else "red"

from( bucket: "the-bucket")
        |> range( start: -2m  , stop: -1m  )
        |> filter( fn: (r) => ( r._measurement == "cpu" ) )
        |> filter( fn: (r) => ( r.cpu == "cpu-total" ) )
