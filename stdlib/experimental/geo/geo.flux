// Provides functions for geographic location filtering and grouping based on S2 cells.
package geo

//
// None of the builtin functions are intended to be used by end users.
//

// Check whether lat/lon is in a lat/lon box or center/radius circle.
builtin containsLatLon

// Calculates grid (set of cell ID tokens) for given box and according to options.
builtin getGrid

// Finds parent cell ID token for given cell ID at specified level.
builtin getParent

// Returns level of specified cell ID token.
builtin getLevel

//
// Flux
//

// Gets level of cell ID tag `_ci` from the first record from the first table from the stream.
_detectCiLevel = (tables=<-) => {
  _r0 =
    tables
      |> tableFind(fn: (key) => exists key._ci)
      |> getRecord(idx: 0)
  _level =
    if exists _r0 then
      getLevel(token: _r0._ci)
    else
       666
  return _level
}

//
// Convenience functions
//

// Collects values to row-wise sets.
toRows = (tables=<-, correlationKey=["_time"]) =>
  tables
    |> pivot(
      rowKey: correlationKey,
      columnKey: ["_field"],
      valueColumn: "_value"
    )

//
// Filtering functions
//

// Filters records by a box, a circle or a polygon area using S2 cell ID tag.
// It is a coarse filter, as the grid always overlays the region, the result will likely contain records
// with lat/lon outside the specified region.
gridFilter = (tables=<-, box={}, circle={}, polygon={}, minSize=24, maxSize=-1, level=-1, ciLevel=-1) => {
  _ciLevel =
    if ciLevel == -1 then
      tables
        |> _detectCiLevel()
    else
      ciLevel
  _grid = getGrid(box: box, circle: circle, polygon: polygon, minSize: minSize, maxSize: maxSize, level: level, maxLevel: _ciLevel)
  return
    tables
      |> filter(fn: (r) =>
        if _grid.level == _ciLevel then
          contains(value: r._ci, set: _grid.set)
        else
          contains(value: getParent(token: r._ci, level: _grid.level), set: _grid.set)
      )
}

// Filters records by a box, a circle or a polygon region.
// It is an exact filter and must be used after `toRows()` because it requires `lat` and `lon` columns in input row sets.
strictFilter = (tables=<-, box={}, circle={}, polygon={}) =>
  tables
    |> filter(fn: (r) =>
      containsLatLon(box: box, circle: circle, polygon: polygon, lat: r.lat, lon: r.lon)
    )

// Two-phase filtering by a box, a circle or a polygon region.
// Returns rows of fields correlated by `correlationKey`.
filterRows = (tables=<-, box={}, circle={}, polygon={}, minSize=24, maxSize=-1, level=-1, ciLevel=-1, correlationKey=["_time"], strict=true) => {
  _rows =
    tables
      |> gridFilter(box: box, circle: circle, polygon: polygon, minSize: minSize, maxSize: maxSize, level: level, ciLevel: ciLevel)
      |> toRows(correlationKey)
  _result =
    if strict then
      _rows
        |> strictFilter(box, circle, polygon)
    else
      _rows
  return _result
}

//
// Grouping functions
//
// intended to be used row-wise sets (i.e after `toRows()`)

// Groups data by area of size specified by level. Result is grouped by `newColumn`.
// Grouping levels: https://s2geometry.io/resources/s2cell_statistics.html
groupByArea = (tables=<-, newColumn, level, ciLevel=-1) => {
  _ciLevel =
    if ciLevel == -1 then
      tables
        |> _detectCiLevel()
    else
      ciLevel
  _prepared =
    if level == _ciLevel then
      tables
	    |> duplicate(column: "_ci", as: newColumn)
    else
      tables
        |> map(fn: (r) => ({
             r with
               _cixxx: getParent(point: {lat: r.lat, lon: r.lon}, level: level)
           }))
        |> rename(columns: { _cixxx: newColumn })
  return
    _prepared
      |> group(columns: [newColumn])
}

// Groups rows into tracks.
asTracks = (tables=<-, groupBy=["id","tid"], orderBy=["_time"]) =>
  tables
    |> group(columns: groupBy)
    |> sort(columns: orderBy)
