# Scalar functions

Create new scalar functions in Go. 

## Required guidelines 

Please help us make the contribution process easier by providing feedback about your experience and any technical hurdles you encountered here. 

### **Standalone Scalar Functions Guidelines**
- Scalar functions must conform to examples in [math](../stdlib/math) or [strings](../stdlib/strings).
- You must submit a unit test in the same folder as the new scalar function implementation. 
- You must submit an end-to-end test in [testdata](../stdlib/testing/testdata). Please look at [End_to_End_Testing.md](/End_to_End_Testing.md)for details. Please take a look at [length_test.flux](../stdlib/strings/length_test.flux) as it is a more sophisticated Standale Scalar Function Test example (i.e. the user cannot directly test the function and must use it inside of a ```map``` transformation)
- You must add a description of the function to [SPEC.md](/SPEC.md).
