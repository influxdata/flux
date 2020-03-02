// Provides functions for geographic location filtering and grouping based on S2 cells.
package geo

//
// None of the builtin functions are intended to be used by end users.
//

// Check whether lat/lon is in specified region.
builtin containsLatLon

// Calculates grid (set of cell ID tokens) for given region and according to options.
builtin getGrid

// Finds parent cell ID token for given cell ID at specified level.
builtin getParent

// Returns level of specified cell ID token.
builtin getLevel

//
// Flux
//

// Gets level of cell ID tag `s2cellID` from the first record from the first table from the stream.
_detectLevel = (tables=<-) => {
  _r0 =
    tables
      |> tableFind(fn: (key) => exists key.s2cellID)
      |> getRecord(idx: 0)
  _level =
    if exists _r0 then
      getLevel(token: _r0.s2cellID)
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
gridFilter = (tables=<-, region, minSize=24, maxSize=-1, level=-1, s2cellIDLevel=-1) => {
  _s2cellIDLevel =
    if s2cellIDLevel == -1 then
      tables
        |> _detectLevel()
    else
      s2cellIDLevel
  _grid = getGrid(region: region, minSize: minSize, maxSize: maxSize, level: level, maxLevel: _s2cellIDLevel)
  return
    tables
      |> filter(fn: (r) =>
        if _grid.level == _s2cellIDLevel then
          contains(value: r.s2cellID, set: _grid.set)
        else
          contains(value: getParent(token: r.s2cellID, level: _grid.level), set: _grid.set)
      )
}

// Filters records by specified region.
// It is an exact filter and must be used after `toRows()` because it requires `lat` and `lon` columns in input row sets.
strictFilter = (tables=<-, region) =>
  tables
    |> filter(fn: (r) =>
      containsLatLon(region: region, lat: r.lat, lon: r.lon)
    )

// Two-phase filtering by speficied region.
// Returns rows of fields correlated by `correlationKey`.
filterRows = (tables=<-, region, minSize=24, maxSize=-1, level=-1, s2cellIDLevel=-1, correlationKey=["_time"], strict=true) => {
  _rows =
    tables
      |> gridFilter(region: region, minSize: minSize, maxSize: maxSize, level: level, s2cellIDLevel: s2cellIDLevel)
      |> toRows(correlationKey)
  _result =
    if strict then
      _rows
        |> strictFilter(region)
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
groupByArea = (tables=<-, newColumn, level, s2cellIDLevel=-1) => {
  _s2cellIDLevel =
    if s2cellIDLevel == -1 then
      tables
        |> _detectLevel()
    else
      s2cellIDLevel
  _prepared =
    if level == _s2cellIDLevel then
      tables
	    |> duplicate(column: "s2cellID", as: newColumn)
    else
      tables
        |> map(fn: (r) => ({
             r with
               _s2cellIDxxx: getParent(point: {lat: r.lat, lon: r.lon}, level: level)
           }))
        |> rename(columns: { _s2cellIDxxx: newColumn })
  return
    _prepared
      |> group(columns: [newColumn])
}

// Groups rows into tracks.
asTracks = (tables=<-, groupBy=["id","tid"], orderBy=["_time"]) =>
  tables
    |> group(columns: groupBy)
    |> sort(columns: orderBy)
