## Scalar Functions**

This section is for developers who want to create new standalone scalar flux functions in go. 

Authorship is kept as simple as possible to promote people to develop and submit new functions. 

Please help us make the contribution process easier by providing feedback about your experience and any technical hurdles you encountered here. 

### **Pure Standalone Scalar Functions Guidelines**

- A pure standalone scalar function must conform to the examples in [math](https://github.com/influxdata/flux/tree/master/stdlib/math) or [strings](https://github.com/influxdata/flux/tree/master/stdlib/strings)
- You must submit a unit test in the same folder as the new scalar function implementation. 
- You must submit an end-to-end test in [testdata](https://github.com/influxdata/flux/tree/master/stdlib/testing/testdata). Please look at [End_to_End_Testing.md](https://github.com/influxdata/flux/tree/master/docs/End_to_End_Testing.md)for details. 
- You must  add a description to SPEC.md