package gen

builtin tables : (n: int, ?nulls: float, ?tags: [{name: string , cardinality: int}]) => [{A with _time: time , _value: float}]
