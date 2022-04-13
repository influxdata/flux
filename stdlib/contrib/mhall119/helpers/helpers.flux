// Package helpers provides shortcut functions for small common tasks
//
// ## Metadata
// introduced: NEXT
package helpers

import "array"
import "json"
// yieldValue lets you display a scalar in the InfluxDB UI by
// wrapping it in an Array
//
// ## Parameters
// - v: Input data.
// - name: Name to pass to the subsequent yield() call.  Default is 'value'.
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
// tags: 
//
defaultYieldValName="value"
yieldValue = (v, name=defaultYieldValName) =>
    array.from(
      rows: [
          {_time: now(), _value: v},
      ]
    )
    |> yield(name: name)

// yieldObject lets you display an object in the InfluxDB UI by
// encoding it as String
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
// tags: 
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
// tags: 
//
defaultYieldJSONName="json"
yieldJSON = (o, name=defaultYieldJSONName) =>
    array.from(
      rows: [
          {_time: now(), _value: string(v: json.encode(v: o))},
      ]
    )
    |> yield(name: name)