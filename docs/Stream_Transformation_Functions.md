# Stream transformation functions

Create new stream transformation functions in Go. 

## Required guidelines

- Stream transformation functions must conform to the examples in [universe](../stdlib/universe) and include the required function and methods shown in the table below.
- You must submit a unit test in in the same folder as the new stream transformation function implementation.
- You must submit an end-to-end test in [testdata](../stdlib/testing/testdata). Please look at [End_to_End_Testing.md](./docs/End_to_End_Testing.md) for details.
- You must add a description of the function to [SPEC.md](/SPEC.md).

### Attributes of a Stream Transformation Function

1. `init()`: Defines your function signature and registers the methods (`OpSpec`, `ProcedureSpec`, and `Transformation`)

2. `FunctionSignature`: Defines how to write the function in Flux, including necessary inputs. For example, `timeShift`, specifies that duration and time columns must be included:
`|>timeShift(duration: 10h, columns: ["_start", "_stop", "_time"])`
Please look at [shift.go](../stdlib/universe/shift.go)

3. `OpSpec`: An internal representation that defines the function signature and gets converted into Procedure Spec. Identifies and collects function arguments and then encodes them into a JSON-encodable struct.

4. `ProcedureSpec`:
-Identifies and collects function arguments and then encodes them into a JSON-encodable struct.
-Identifies the incoming Op-Spec.
-Copies incoming values and converts them if needed.
-Creates the data transformation plan.
-Is optimized by the query planner, which passes the plan for transforming the data.
-Optimized by the quert planner, which passes the plan into the execution engine.
-This plan is then converted into the Transfromation type.

5. `Transformation`: Bulk of the script responsible for transforming data.
