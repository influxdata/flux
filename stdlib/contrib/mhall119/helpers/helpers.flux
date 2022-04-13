// Package helpers provides shortcut functions for common operations.
//
// ## Metadata
// introduced: NEXT
package helpers


import "array"
import "json"

// yieldValue returns a stream of tables containing a specified scalar value.
// This makes it possible to display a scalar value in an InfluxDB visualization.
//
// ## Parameters
// - v: Input value.
// - name: Yield name. Default is `value`.
//
// ## Examples
//
// ### Yield an option value
// ```no_run
// import "contrib/mhall119/helpers"
//
// option testData = {testValue: 125}
// yieldValue(v: testData.testValue)
// ```
//
// ## Metadata
// tags: outputs
//
yieldValue = (v, name="value") =>
    array.from(
      rows: [
          {_time: now(), _value: v},
      ]
    )
    |> yield(name: name)

// yieldRecord returns a stream of tables containing a specified record encoded as a string.
// This makes it possible to display records in an InfluxDB visualization.
//
// ## Parameters
// - o: Object to yield
// - name: Name to pass to the subsequent yield() call.  Default is 'object'.
//
// ## Examples
//
// ### Yield an option value
// ```no_run
// import "contrib/mhall119/helpers"
//
// option testData = {testValue: 125}
// yieldObject(v: testData)
// ```
//
// ## Metadata
// tags: outputs
//
defaultYieldObjName="object"
yieldObject = (o, name=defaultYieldObjName) =>
    array.from(
      rows: [
          {_time: now(), _value: display(v: o)},
      ]
    )
    |> yield(name: name)

// yieldObject lets you display an object in the InfluxDB UI by
// encoding it as a JSON string and wrapping it in an Array
//
// ## Parameters
// - o: Object to yield
// - name: Name to pass to the subsequent yield() call.  Default is 'object'.
//
// ## Examples
//
// ### Yield an option value
// ```no_run
// import "contrib/mhall119/helpers"
//
// option testData = {testValue: 125}
// yieldObject(v: testData)
// ```
//
// ## Metadata
// tags: outputs
//
defaultYieldJSONName="json"
yieldJSON = (o, name=defaultYieldJSONName) =>
    array.from(
      rows: [
          {_time: now(), _value: string(v: json.encode(v: o))},
      ]
    )
    |> yield(name: name)