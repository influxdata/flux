// Provides functions for geographic location filtering and grouping base on S2 cells.
package geo

import "strings"

// Calculates grid (set of cell ID tokens) for given box and according to options.
builtin getGrid

// Find parent cell ID token for given cell (specified by token) at specified level.
builtin getParent

// Check whether lat/lon is in a lat/lon box or center/radius circle.
builtin containsLatLon

// Checks for tag presence in a record and its value against a set.
builtin containsTag

// ----------------------------------------
// Filtering functions
// ----------------------------------------

// Filters records by geohash tag value (`_cid1` ... `_cid30`) if exist
// TODO(?): uses hardcoded schema tag keys and Flux does not provide dynamic access, therefore containsTag() is provided.
tokenFilter = (tables=<-, grid) =>
  tables
    |> filter(fn: (r) =>
	  if grid.level == 1 and exists r._cid1 then contains(value: r._cid1, set: grid.set)
	  else if grid.level == 2 and exists r._cid2 then contains(value: r._cid2, set: grid.set)
	  else if grid.level == 3 and exists r._cid3 then contains(value: r._cid3, set: grid.set)
	  else if grid.level == 4 and exists r._cid4 then contains(value: r._cid4, set: grid.set)
	  else if grid.level == 5 and exists r._cid5 then contains(value: r._cid5, set: grid.set)
	  else if grid.level == 6 and exists r._cid6 then contains(value: r._cid6, set: grid.set)
	  else if grid.level == 7 and exists r._cid7 then contains(value: r._cid7, set: grid.set)
	  else if grid.level == 8 and exists r._cid8 then contains(value: r._cid8, set: grid.set)
	  else if grid.level == 9 and exists r._cid9 then contains(value: r._cid9, set: grid.set)
	  else if grid.level == 10 and exists r._cid10 then contains(value: r._cid10, set: grid.set)
	  else if grid.level == 11 and exists r._cid11 then contains(value: r._cid11, set: grid.set)
	  else if grid.level == 12 and exists r._cid12 then contains(value: r._cid12, set: grid.set)
	  else if grid.level == 13 and exists r._cid12 then contains(value: r._cid13, set: grid.set)
	  else if grid.level == 14 and exists r._cid12 then contains(value: r._cid14, set: grid.set)
	  else if grid.level == 15 and exists r._cid12 then contains(value: r._cid15, set: grid.set)
	  else if grid.level == 16 and exists r._cid12 then contains(value: r._cid16, set: grid.set)
	  else if grid.level == 17 and exists r._cid12 then contains(value: r._cid17, set: grid.set)
	  else if grid.level == 18 and exists r._cid12 then contains(value: r._cid18, set: grid.set)
	  else if grid.level == 19 and exists r._cid12 then contains(value: r._cid19, set: grid.set)
	  else if grid.level == 20 and exists r._cid12 then contains(value: r._cid20, set: grid.set)
	  else if grid.level == 21 and exists r._cid12 then contains(value: r._cid21, set: grid.set)
	  else if grid.level == 22 and exists r._cid12 then contains(value: r._cid22, set: grid.set)
	  else if grid.level == 23 and exists r._cid12 then contains(value: r._cid23, set: grid.set)
	  else if grid.level == 24 and exists r._cid12 then contains(value: r._cid24, set: grid.set)
	  else if grid.level == 25 and exists r._cid12 then contains(value: r._cid25, set: grid.set)
	  else if grid.level == 26 and exists r._cid12 then contains(value: r._cid26, set: grid.set)
	  else if grid.level == 27 and exists r._cid12 then contains(value: r._cid27, set: grid.set)
	  else if grid.level == 28 and exists r._cid12 then contains(value: r._cid28, set: grid.set)
	  else if grid.level == 29 and exists r._cid12 then contains(value: r._cid29, set: grid.set)
	  else if grid.level == 30 and exists r._cid12 then contains(value: r._cid30, set: grid.set)
	  else false
    )

// Filters records by cell ID token tag value using custom builtin function.
tokenFilterEx = (tables=<-, grid, prefix="_cid") =>
  tables
    |> filter(fn: (r) =>
      containsTag(row: r, tagKey: prefix + string(v: grid.level), set: grid.set)
    )

// Filters records by lat/lon box or center/radius circle.
// The grid always overlaps specified area and therefore result may contain values outside the box.
// If precise filtering is needed, `boxFilter()` may be used later (after `toRows()`).
gridFilter = (tables=<-, fn=tokenFilterEx, box={}, circle={}, minGridSize=9, maxGridSize=-1, level=-1, maxLevelIndex=30) => {
  grid = getGrid(box: box, circle: circle, minSize: minGridSize, maxSize: maxGridSize, level: level, maxLevel: maxLevelIndex)
  return
    tables
      |> fn(grid: grid)
}

// --- experimental: simpler token tags schema filter

tokenFilter2 = (tables=<-, grid, ciLevel) =>
  tables
    |> filter(fn: (r) =>
      if grid.level == ciLevel then
        contains(value: r._ci, set: grid.set)
      else
        contains(value: getParent(token: r._ci, level: grid.level), set: grid.set)
    )

gridFilter2 = (tables=<-, box={}, circle={}, minGridSize=9, maxGridSize=-1, level=-1, ciLevel) => {
  grid = getGrid(box: box, circle: circle, minSize: minGridSize, maxSize: maxGridSize, level: level, maxLevel: ciLevel)
  return
    tables
      |> tokenFilter2(grid: grid, ciLevel: ciLevel)
}

// --- end ---

// Filters records by lat/lon box or center/radius circle.
// Must be used after `toRows()` because it requires `lat` and `lon` columns in input row set.
strictFilter = (tables=<-, box={}, circle={}) =>
  tables
    |> filter(fn: (r) =>
      containsLatLon(box: box, circle: circle, lat: r.lat, lon: r.lon)
    )

// ----------------------------------------
// Convenience functions
// ----------------------------------------

// Collects values to row-wise sets.
toRows = (tables=<-, correlationKey=["_time"]) =>
  tables
    |> pivot(
      rowKey: correlationKey,
      columnKey: ["_field"],
      valueColumn: "_value"
    )

// Drops geohash indexes columns except those specified.
// It will fail if input tables are grouped by any of them.
stripMeta = (tables=<-, pattern=/_cid\d+/, except=[]) =>
  tables
    |> drop(fn: (column) => column =~ pattern and (length(arr: except) == 0 or not contains(value: column, set: except)))

// ----------------------------------------
// Grouping functions
// ----------------------------------------
// intended to be used row-wise sets (i.e after `toRows()`)

// Grouping levels: https://s2geometry.io/resources/s2cell_statistics.html

// Groups data by area of size specified by level. Result is grouped by `newColumn`.
// Parameter `maxLevelIndex` specifies finest cell ID level tag available in the input tables.
// TODO: can maxLevelIndex be discovered at Flux level?
groupByArea = (tables=<-, newColumn, level, maxLevelIndex, prefix="_cid") => {
  prepared =
    if level <= maxLevelIndex then
      tables
	    |> duplicate(column: prefix + string(v: level), as: newColumn)
    else
      tables
        |> map(fn: (r) => ({ r with _cidx: getParent(token: r.cid, level: level) }))
	    |> rename(columns: { _cidx: newColumn })
  return prepared
    |> group(columns: [newColumn])
}

// Organizes rows into tracks.
// It groups input source and track id and orders by time in ascending order.
asTracks = (tables=<-, groupBy=["id","tid"], orderBy=["_time"]) =>
  tables
    |> group(columns: groupBy)
    |> sort(columns: orderBy)

// --- experimental: simpler token tags schema grouping

groupByArea2 = (tables=<-, newColumn, level, ciLevel) => {
  prepared =
    if level == ciLevel then
      tables
	    |> duplicate(column: "_ci", as: newColumn)
    else
      tables
        |> map(fn: (r) => ({ r with _cix: getParent(token: r._ci, level: level) }))
	    |> rename(columns: { _cix: newColumn })
  return prepared
    |> group(columns: [newColumn])
}
