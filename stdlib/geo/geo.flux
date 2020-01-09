package geo

import "strings"

// Calculates geohash grid for given box and according to options
builtin getGrid

// Checks for tag existence and its value (this is a workaround)
builtin containsTag

// Limits
MaxGeohashPrecision = 12

// Filters records by geohash tag value (_g1 ... _g12) if exist
// DOES NOT WORK YET https://community.influxdata.com/t/conditional-filter-with-exists-operator/12534
// ISSUE: hardcoded tag names
geohashFilter = (tables=<-, grid) =>
  tables
    |> filter(fn: (r) =>
	  if grid.precision == 1 and exists r._g1 then contains(value: r._g1, set: grid.set)
	  else if grid.precision == 2 and exists r._g2 then contains(value: r._g2, set: grid.set)
	  else if grid.precision == 3 and exists r._g3 then contains(value: r._g3, set: grid.set)
	  else if grid.precision == 4 and exists r._g4 then contains(value: r._g4, set: grid.set)
	  else if grid.precision == 5 and exists r._g5 then contains(value: r._g5, set: grid.set)
	  else if grid.precision == 6 and exists r._g6 then contains(value: r._g6, set: grid.set)
	  else if grid.precision == 7 and exists r._g7 then contains(value: r._g7, set: grid.set)
	  else if grid.precision == 8 and exists r._g8 then contains(value: r._g8, set: grid.set)
	  else if grid.precision == 9 and exists r._g9 then contains(value: r._g9, set: grid.set)
	  else if grid.precision == 10 and exists r._g10 then contains(value: r._g10, set: grid.set)
	  else if grid.precision == 11 and exists r._g11 then contains(value: r._g11, set: grid.set)
	  else if grid.precision == 12 and exists r._g12 then contains(value: r._g12, set: grid.set)
	  else false
	)

// Filters records by geohash tag value (_gN) using custom builtin function
geohashFilterInternal = (tables=<-, grid) =>
  tables
    |> filter(fn: (r) =>
        containsTag(row: r, tagKey: "_g" + string(v: grid.precision), set: grid.set)
	)

// Filters records by geohash tag value (_g1 ... _g12) if exist
boxFilter = (tables=<-, box, minGridSize=4, maxGeohashPrecision=MaxGeohashPrecision) =>
  tables
    |> geohashFilter(grid: getGrid(box: box, minSize: minGridSize, maxPrecision: maxGeohashPrecision))

// Collects values to row-wise sets. Equivalent to pivot(rowKey: correlationKey, columnKey: ["_field"], valueColumn: "_value")
asRows = (tables=<-, correlationKey=["_time"]) =>
  tables
    |> pivot(
      rowKey: correlationKey,
      columnKey: ["_field"],
      valueColumn: "_value"
    )

// Aggregates data over area of size specified by geohash precision, result is grouped by aColumnName
aggregateArea = (tables=<-, aColumnName, aPrecision, aMaxPrecisionIndex) => {
  aggregated =
   if aPrecision <= aMaxPrecisionIndex then
    tables
	  |> duplicate(column: "_g" + string(v: aPrecision), as: aColumnName)
   else
    tables
      |> map(fn: (r) => ({ r with _gx: strings.substring(v: r.geohash, start:0, end: aPrecision) }))
	  |> rename(columns: { _gx: aColumnName })
  return aggregated
    |> group(columns: [aColumnName])
}
