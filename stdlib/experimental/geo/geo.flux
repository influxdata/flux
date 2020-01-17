package geo

import "strings"

// Calculates geohash grid for given box and according to options
builtin getGrid

// Checks for tag presence in a record and its value against a set
builtin containsTag

// ----------------------------------------
// Filtering functions
// ----------------------------------------

// Filters records by geohash tag value (_g1 ... _g12) if exist // TODO ugly hardcoded schema tag keys
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

// Filters records by lat/lon box
boxFilter = (tables=<-, box, minGridSize=9, maxGridSize=-1, geohashPrecision=-1, maxGeohashPrecision=12) =>
  tables
    |> geohashFilter(grid: getGrid(box: box, minSize: minGridSize, maxSize: maxGridSize, precision: geohashPrecision, maxPrecision: maxGeohashPrecision))

// Filters records by geohash tag value using custom builtin function
geohashFilterEx = (tables=<-, grid, prefix="_g") =>
  tables
    |> filter(fn: (r) =>
      containsTag(row: r, tagKey: prefix + string(v: grid.precision), set: grid.set)
	)

// Filters records by lat/lon box using custom builtin function
boxFilterEx = (tables=<-, box, minGridSize=9, maxGridSize=-1, geohashPrecision=-1, maxGeohashPrecision=12) =>
  tables
    |> geohashFilterEx(grid: getGrid(box: box, minSize: minGridSize, maxSize: maxGridSize, precision: geohashPrecision, maxPrecision: maxGeohashPrecision))

// ----------------------------------------
// Convenience functions
// ----------------------------------------

// Collects values to row-wise sets. Equivalent to pivot(rowKey: correlationKey, columnKey: ["_field"], valueColumn: "_value")
toRows = (tables=<-, correlationKey=["_time"]) =>
  tables
    |> pivot(
      rowKey: correlationKey,
      columnKey: ["_field"],
      valueColumn: "_value"
    )

// Drops geohash tags columns except those specified
stripMeta = (tables=<-, pattern=/_g\d+/, except=[]) =>
  tables
    |> drop(fn: (column) => column =~ pattern and (length(arr: except) == 0 or not contains(value: column, set: except)))

// ----------------------------------------
// Grouping functions
// ----------------------------------------
// intended to be used row-wise sets (i.e after toRows())

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

// Groups rows by area of size specified by geohash precision
groupByArea = (tables=<-, newColumn, precision, maxPrecisionIndex, prefix="_g") => {
  prepared =
    if precision <= maxPrecisionIndex then
      tables
	    |> duplicate(column: prefix + string(v: precision), as: newColumn)
    else
      tables
        |> map(fn: (r) => ({ r with _gx: strings.substring(v: r.geohash, start:0, end: precision) }))
	    |> rename(columns: { _gx: newColumn })
  return prepared
    |> group(columns: [newColumn])
}

// Groups rows into tracks (by "id" and "tid" by default) and order by time asc
asTracks = (tables=<-, groupBy=["id","tid"], orderBy=["_time"]) =>
  tables
    |> group(columns: groupBy)
    |> sort(columns: orderBy)
