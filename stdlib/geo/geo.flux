package geo

import "strings"

// Calculates geohash grid for given box and according to options
builtin getGrid

// Checks for tag existence and its value // TEMPORARY(??) WORKAROUND
builtin containsTag

// Defaults and limits
MinGridSize = 9
MaxGridSize = -1
DefaultGeohashPrecision = -1
MaxGeohashPrecision = 12

// -----------
// Filtering
// -----------

// Filters records by geohash tag value (_g1 ... _g12) if exist
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

// Filters records by geohash tag value (_gN) using custom builtin function // TEMPORARY(??) WORKAROUND
geohashFilterInternal = (tables=<-, grid) =>
  tables
    |> filter(fn: (r) =>
        containsTag(row: r, tagKey: "_g" + string(v: grid.precision), set: grid.set)
	)

// Filters records by lat/lon box
boxFilter = (tables=<-, box, minGridSize=MinGridSize, maxGridSize=MaxGridSize, geohashPrecision=DefaultGeohashPrecision, maxGeohashPrecision=MaxGeohashPrecision) =>
  tables
    |> geohashFilter(grid: getGrid(box: box, minSize: minGridSize, maxSize: maxGridSize, precision: geohashPrecision, maxPrecision: maxGeohashPrecision))

// -----------
// Misc
// -----------

// Collects values to row-wise sets. Equivalent to pivot(rowKey: correlationKey, columnKey: ["_field"], valueColumn: "_value")
asRows = (tables=<-, correlationKey=["_time"]) =>
  tables
    |> pivot(
      rowKey: correlationKey,
      columnKey: ["_field"],
      valueColumn: "_value"
    )

// Groups row-wise sets by specified columns and drops _g* columns except those specified
groupByAndStripMeta = (tables=<-, columns, except=[]) =>
  tables
    |> group(columns: columns)
    |> drop(fn: (column) => column =~ /_g\d+/ and (length(arr: except) == 0 or not contains(value: column, set: except)))

// -----------
// Aggregation
// -----------

// Grouping levels (based on geohash length/precision) - cell width x height
//  1 - 5000 x 5000 km
//  2 - 1250 x 625 km
//  3 - 156 x 156 km
//  4 - 39.1 x 19.5 km
//  5 - 4.89 x 4.89 km
//  6 - 1.22 x 0.61 km
//  7 - 153 x 153 m
//  8 - 38.2 x 19.1 m
//  9 - 4.77 x 4.77 m
// 10 - 1.19 x 0.596 m
// 11 - 149 x 149 mm
// 12 - 37.2 x 18.6 mm

// Groups data over area of size specified by geohash precision, result is grouped by gColumnName
groupByArea = (tables=<-, gColumnName, gPrecision, gMaxPrecisionIndex) => {
  grouped =
   if gPrecision <= gMaxPrecisionIndex then
    tables
	  |> duplicate(column: "_g" + string(v: gPrecision), as: gColumnName)
   else
    tables
      |> map(fn: (r) => ({ r with _gx: strings.substring(v: r.geohash, start:0, end: gPrecision) }))
	  |> rename(columns: { _gx: gColumnName })
  return grouped
    |> group(columns: [gColumnName])
}

aggregateArea = (tables=<-, gColumnName, gPrecision, gMaxPrecisionIndex, fn, column) =>
  tables
    |> groupByArea(gColumnName, gPrecision, gMaxPrecisionIndex)
    |> fn(column: column)
