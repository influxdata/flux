## **Stream Transformation Functions**

This section is for developers who want to create new stream transformation functions in go. 

Authorship is kept as simple as possible to promote people to develop and submit new functions.

Please help us make the contribution process easier by providing feedback about your experience and any technical hurdles you encountered here. 

### **Stream Transformation Functions Guidelines**

- A stream transformation function must conform to the examples in [universe](https://github.com/influxdata/flux/blob/master/stdlib/universe)
- You must submit a unit test in in the same folder as the new stream transformation function implementation.
- You must submit an end-to-end test in [testdata](https://github.com/influxdata/flux/tree/master/stdlib/testing/testdata).Please look at [End_to_End_Testing.md](https://github.com/influxdata/flux/tree/master/docs/End_to_End_Testing.md)for details.
- You must  add a description to SPEC.md
- They functions must contain the following pieces:

1. **init():**

2. 1. Where you define your function signature and register the methods(**Op-Spec**, **Procedure Spec**, **Transformation Spec**)

3. **FunctionSignature**: 

4. 1. Where you define how you want the user to write out a function in flux and what inputs they need to include. 

   2. For example, with timeShift, this is where you would define that a user must provide a duration and a time columns. 

   3. 1. \```|>timeShift(duration: 10h, columns: ["_start", "_stop", "_time"])```

5. **Op-Spec:**

6. 1. The internal representation or mirror of the function signature. 
   2. The **Op-Spec** defines the function signature and gets converted converted into the **Procedure Spec**. 
   3. The **Op-Spec** identifies and collects function arguments and then encodes them into a JSON-encodable struct. 

7. **Procedure Spec:**

8. 1. Identifies the incoming **Op-Spec**. 
   2. Copies incoming values and converts them if needed.
   3. Creates the plan for the actual transformation of the data. It’s optimized by the query planner which passes the plan into the execution engine after and converted into the Transformation type. 

9. **Transformation Spec:**

10. 1. This type is the meat of the entire script. It is responsible for transforming the data. 

11. **Process Method:**

12. 1. Part of the **Transformation Spec**. Where you actually define the transformation you want. 