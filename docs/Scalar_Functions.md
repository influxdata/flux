# Scalar functions

Create new scalar functions in Go. 

## Required guidelines 

Please help us make the contribution process easier by providing feedback about your experience and any technical hurdles you encountered here. 

### **Pure Standalone Scalar Functions Guidelines**
- Scalar functions must conform to examples in [math](https://github.com/influxdata/flux/tree/master/stdlib/math) or [strings](https://github.com/influxdata/flux/tree/master/stdlib/strings).
- You must submit a unit test in the same folder as the new scalar function implementation. 
- You must submit an end-to-end test in [testdata](https://github.com/influxdata/flux/tree/master/stdlib/testing/testdata). Please look at [End_to_End_Testing.md](https://github.com/influxdata/flux/tree/master/docs/End_to_End_Testing.md)for details.
- You must add a description of the function to [SPEC.md](./docs/SPEC.md).
