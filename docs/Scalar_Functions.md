# Scalar functions

Create new scalar functions in Go. 

## Required guidelines 

Please help us make the contribution process easier by providing feedback about your experience and any technical hurdles you encountered here. 

### **Standalone Scalar Functions Guidelines**
- Scalar functions must conform to examples in [math](../stdlib/math) or [strings](../stdlib/strings).
- You must submit a unit test in the same folder as the new scalar function implementation. 
- You must submit an end-to-end test in [testdata](../stdlib/testing/testdata). Please look at [End_to_End_Testing.md](/End_to_End_Testing.md)for details. Note that he user cannot directly test functions in Flux end-to-end tests, so he must use it inside of a `map` transformation, for example. Please take a look at [length_test.flux](../stdlib/strings/length_test.flux) to see an example of this pattern.

- You must add a description of the function to [SPEC.md](/SPEC.md).
