## Pure Flux Code Functions**

This section is for developers who want to create new flux functions in pure Flux code. 

Authorship is kept as simple as possible to promote people to develop and submit new functions.

Please help us make the contribution process easier by providing feedback about your experience and any technical hurdles you encountered here. 

### **Pure Flux Code Functions Guidelines**

- A pure Flux code function must conform to the examples in [universe.flux](https://github.com/influxdata/flux/blob/master/stdlib/universe/universe.flux)
- You must submit a unit test in the same folder as the new flux function implementation. 
- You must submit an end-to-end test in [testdata](https://github.com/influxdata/flux/tree/master/stdlib/testing/testdata). Please look at [End_to_End_Testing.md](https://github.com/influxdata/flux/tree/master/docs/End_to_End_Testing.md)for details. 
- You must include a comment describing your function in [universe.flux](https://github.com/influxdata/flux/blob/master/stdlib/universe/universe.flux)
- You must  add a description to SPEC.md