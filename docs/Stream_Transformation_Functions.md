# Stream transformation functions

Create new stream transformation functions in Go. 

## Required guidelines

- Stream transformation functions must conform to the examples in [universe](https://github.com/influxdata/flux/blob/master/stdlib/universe) and include the required function and methods shown in the table below.
- You must submit a unit test in [testing](https://github.com/influxdata/flux/tree/master/stdlib/testing).
- You must submit an end-to-end test in [testdata](https://github.com/influxdata/flux/tree/master/stdlib/testing/testdata). 
- You must add a description of the function to [SPEC.md](./docs/SPEC.md).

### Required in function 

| Name              | Description  |
| :--------         | :-------------------------------------------------------|
| **init()**            | Define your function signature and register the methods (**Op-Spec**, **Procedure Spec**, and **Transformation Spec**). |
| **FunctionSignature** | Define how the user writes the function in Flux and define the inputs that must included. For example, the timeShift FunctionSignature specifies both duration and time columns are required: `timeShift(duration: 10h, columns: ["_start", "_stop", "_time"])`|
|  **Op-Spec**       |  Internal representation of the function signature. Defines the function signature and gets converted into the **Procedure Spec**. Identifies and collects function arguments and then encodes them into a JSON-encodable struct. |
|  **Procedure Spec**    |      Identifies the incoming **Op-Spec**. Copies incoming values and converts them if needed. Creates the plan for transforming the data. Optimized by the query planner, which passes the plan into the execution engine after and converted into the Transformation type. |
| **Transformation Spec**  |  Purpose of the script. Responsible for transforming the data. |
| **Process Method**| Part of the **Transformation Spec**. Define the transformation.  |