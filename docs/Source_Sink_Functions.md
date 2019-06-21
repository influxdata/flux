## Source/Sing Functions**

This section is for developers who want to create new standalone source and sink functions in go. 

Authorship is kept as simple as possible to promote people to develop and submit new functions. 

Please help us make the contribution process easier by providing feedback about your experience and any technical hurdles you encountered here. 

### **Pure Standalone Scalar Functions Guidelines**

- A source or sink function must conform to the following examples: 
    [sql](https://github.com/influxdata/flux/tree/master/stdlib/sql)
    [http](https://github.com/Anaisdg/flux/tree/master/stdlib/http)
- You must submit a unit test in [testing](https://github.com/influxdata/flux/tree/master/stdlib/testing)
- You must submit an end-to-end test in [testdata](https://github.com/influxdata/flux/tree/master/stdlib/testing/testdata) 
- You must  add a description to SPEC.md