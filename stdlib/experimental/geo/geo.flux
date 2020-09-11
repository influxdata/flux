// Provides functions for geographic location filtering and grouping based on S2 cells.
package geo

import "experimental"
import "influxdata/influxdb/v1"

// Units
option units = {
  distance: "km"
}

//
// Builtin GIS functions
//

// Returns boolean whether the region contains specified geometry.
builtin stContains : (region: A, geometry: B, units: {distance: string}) => bool where A: Record, B: Record

// Returns distance from given region to specified geometry.
builtin stDistance : (region: A, geometry: B, units: {distance: string}) => float where A: Record, B: Record

// Returns length of a curve.
builtin stLength : (geometry: A, units: {distance: string}) => float where A: Record

//
// Flux GIS ST functions
//

ST_Contains = (region, geometry, units=units) =>
  stContains(region: region, geometry: geometry, units: units)

ST_Distance = (region, geometry, units=units) =>
  stDistance(region: region, geometry: geometry, units: units)

ST_DWithin = (region, geometry, distance, units=units) =>
  stDistance(region: region, geometry: geometry, units: units) <= distance

ST_Intersects = (region, geometry, units=units) =>
  stDistance(region: region, geometry: geometry, units: units) <= 0.0

ST_Length = (geometry, units=units) =>
  stLength(geometry: geometry, units: units)

// Non-standard
ST_LineString = (tables=<-) =>
  tables
    |> reduce(fn: (r, accumulator) => ({
        __linestring: accumulator.__linestring + (if accumulator.__count > 0 then ", " else "") + string(v: r.lon) + " " + string(v: r.lat),
        __count: accumulator.__count + 1
      }), identity: {
        __linestring: "",
        __count: 0
      }
    )
    |> drop(columns: ["__count"])
    |> rename(columns: {__linestring: "st_linestring"})

//
// None of the following builtin functions are intended to be used by end users.
//

// Calculates grid (set of cell ID tokens) for given region and according to options.
builtin getGrid : (region: T, ?minSize: int, ?maxSize: int, ?level: int, ?maxLevel: int, units: {distance: string}) => {level: int , set: [string]} where T: Record

// Returns level of specified cell ID token.
builtin getLevel : (token: string) => int

// Returns cell ID token for given cell or lat/lon point at specified level.
builtin s2CellIDToken : (?token: string, ?point: {lat: float , lon: float}, level: int) => string

// Returns lat/lon coordinates of given cell ID token.
builtin s2CellLatLon : (token: string) => {lat: float , lon: float}

//
// Flux functions
//

// Gets level of cell ID tag `s2cellID` from the first record from the first table in the stream.
_detectLevel = (tables=<-) => {
  _r0 =
    tables
      |> tableFind(fn: (key) => exists key.s2_cell_id)
      |> getRecord(idx: 0)
  _level =
    if exists _r0 then
      getLevel(token: _r0.s2_cell_id)
    else
       666
  return _level
}

//
// Convenience functions
//

// Pivots values to row-wise sets.
toRows = (tables=<-) =>
  tables
    |> v1.fieldsAsCols()

// Shapes data to meet the requirements of the geo package.
// Renames fields containing latitude and longitude values to lat and lon.
// Pivots values to row-wise sets.
// Generates an s2_cell_id tag for each reach using lat and lon values.
// Adds the s2_cell_id column to the group key.
shapeData = (tables=<-, latField, lonField, level) =>
  tables
    |> map(fn: (r) => ({ r with
        _field:
          if r._field == latField then "lat"
          else if r._field == lonField then "lon"
          else r._field
      })
    )
    |> toRows()
    |> map(fn: (r) => ({ r with
        s2_cell_id: s2CellIDToken(point: {lat: r.lat, lon: r.lon}, level: level)
      })
    )
    |> experimental.group(
      columns: ["s2_cell_id"],
      mode: "extend"
    )

//
// Filtering functions
//

// Filters records by a box, a circle or a polygon area using S2 cell ID tag.
// It is a coarse filter, as the grid always overlays the region, the result will likely contain records
// with lat/lon outside the specified region.
gridFilter = (tables=<-, region, minSize=24, maxSize=-1, level=-1, s2cellIDLevel=-1, units=units) => {
  _s2cellIDLevel =
    if s2cellIDLevel == -1 then
      tables
        |> _detectLevel()
    else
      s2cellIDLevel
  _grid = getGrid(region: region, minSize: minSize, maxSize: maxSize, level: level, maxLevel: _s2cellIDLevel, units: units)
  return
    tables
      |> filter(fn: (r) =>
        if _grid.level == _s2cellIDLevel then
          contains(value: r.s2_cell_id, set: _grid.set)
        else
          contains(value: s2CellIDToken(token: r.s2_cell_id, level: _grid.level), set: _grid.set)
      )
}

// Filters records by specified region.
// It is an exact filter and must be used after `toRows()` because it requires `lat` and `lon` columns in input row sets.
strictFilter = (tables=<-, region) =>
  tables
    |> filter(fn: (r) =>
      ST_Contains(region: region, geometry: {lat: r.lat, lon: r.lon})
    )

// Two-phase filtering by specified region.
// Checks to see if data is already pivoted and contains a lat column.
// Returns pivoted data.
filterRows = (tables=<-, region, minSize=24, maxSize=-1, level=-1, s2cellIDLevel=-1, strict=true) => {
  _columns =
    tables
      |> columns(column: "_value")
      |> tableFind(fn: (key) => true )
      |> getColumn(column: "_value")
  _rows =
    if contains(value: "lat", set: _columns) then
      tables
        |> gridFilter(region: region, minSize: minSize, maxSize: maxSize, level: level, s2cellIDLevel: s2cellIDLevel)
    else
      tables
        |> gridFilter(region: region, minSize: minSize, maxSize: maxSize, level: level, s2cellIDLevel: s2cellIDLevel)
        |> toRows()
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
	    |> duplicate(column: "s2_cell_id", as: newColumn)
    else
      tables
        |> map(fn: (r) => ({
             r with
               _s2_cell_id_xxx: s2CellIDToken(point: {lat: r.lat, lon: r.lon}, level: level)
           }))
        |> rename(columns: { _s2_cell_id_xxx: newColumn })
  return
    _prepared
      |> group(columns: [newColumn])
}

// Groups rows into tracks.
asTracks = (tables=<-, groupBy=["id","tid"], orderBy=["_time"]) =>
  tables
    |> group(columns: groupBy)
    |> sort(columns: orderBy)
