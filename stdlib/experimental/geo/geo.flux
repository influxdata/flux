// Package geo provides tools for working with geotemporal data, such as
// filtering and grouping by geographic location.
//
// ## Geo schema requirements
// The Geo package uses the Go implementation of the [S2 Geometry Library](https://s2geometry.io/).
// Functions in the `geo` package require the following:
//
// - a **`s2_cell_id` tag** containing an **S2 cell ID as a token**
// - a **`lat` field** containing the **latitude in decimal degrees** (WGS 84)
// - a **`lon` field** containing the **longitude in decimal degrees** (WGS 84)
//
// #### Schema recommendations
// - a tag that identifies the data source
// - a tag that identifies the point type (for example: `start`, `stop`, `via`)
// - a field that identifies the track or route (for example: `id`, `tid`)
//
// ##### Examples of geotemporal line protocol
// ```
// taxi,pt=start,s2_cell_id=89c2594 tip=3.75,dist=14.3,lat=40.744614,lon=-73.979424,tid=1572566401123234345i 1572566401947779410
// bike,id=biker-007,pt=via,s2_cell_id=89c25dc lat=40.753944,lon=-73.992035,tid=1572588100i 1572567115
// ```
//
// ## S2 Cell IDs
// Use **latitude** and **longitude** with the `s2.CellID.ToToken` endpoint of the S2
// Geometry Library to generate `s2_cell_id` tags.
// Specify your [S2 Cell ID level](https://s2geometry.io/resources/s2cell_statistics.html).
//
// **Note:** To filter more quickly, use higher S2 Cell ID levels, but know that
// higher levels increase [series cardinality](https://docs.influxdata.com/influxdb/latest/reference/glossary/#series-cardinality).
//
// Language-specific implementations of the S2 Geometry Library provide methods for
// generating S2 Cell ID tokens. For example:
//
// - **Go:** [`s2.CellID.ToToken()`](https://godoc.org/github.com/golang/geo/s2#CellID.ToToken)
// - **Python:** [`s2sphere.CellId.to_token()`](https://s2sphere.readthedocs.io/en/latest/api.html#s2sphere.CellId)
// - **Javascript:** [`s2.cellid.toToken()`](https://github.com/mapbox/node-s2/blob/master/API.md#cellidtotoken---string)
//
// ### Add S2 Cell IDs to existing geotemporal data
// Use `geo.shapeData()` to add `s2_cell_id` tags to data that includes fields
// with latitude and longitude values.
//
// ```no_run
// //...
//   |> shapeData(
//     latField: "latitude",
//     lonField: "longitude",
//     level: 10
//   )
// ```
//
// ## Latitude and longitude values
// Flux supports latitude and longitude values in **decimal degrees** (WGS 84).
//
// | Coordinate | Minimum | Maximum |
// |:---------- | -------:| -------:|
// | Latitude   | -90.0   | 90.0    |
// | Longitude  | -180.0  | 180.0   |
//
// ## Region definitions
// Many functions in the Geo package filter data based on geographic region.
// Define geographic regions using the following shapes:
//
// - [box](#box)
// - [circle](#circle)
// - [point](#point)
// - [polygon](#polygon)
//
// ### box
// Define a box-shaped region by specifying a record containing the following properties:
//
// - **minLat:** minimum latitude in decimal degrees (WGS 84) _(Float)_
// - **maxLat:** maximum latitude in decimal degrees (WGS 84) _(Float)_
// - **minLon:** minimum longitude in decimal degrees (WGS 84) _(Float)_
// - **maxLon:** maximum longitude in decimal degrees (WGS 84) _(Float)_
//
// ##### Example box-shaped region
// ```no_run
// {
//   minLat: 40.51757813,
//   maxLat: 40.86914063,
//   minLon: -73.65234375,
//   maxLon: -72.94921875
// }
// ```
//
// ### circle
// Define a circular region by specifying a record containing the following properties:
//
// - **lat**: latitude of the circle center in decimal degrees (WGS 84) _(Float)_
// - **lon**: longitude of the circle center in decimal degrees (WGS 84) _(Float)_
// - **radius**:  radius of the circle in kilometers (km) _(Float)_
//
// ##### Example circular region
// ```no_run
// {
//   lat: 40.69335938,
//   lon: -73.30078125,
//   radius: 20.0
// }
// ```
//
// ### point
// Define a point region by specifying a record containing the following properties:
//
// - **lat**: latitude in decimal degrees (WGS 84) _(Float)_
// - **lon**: longitude in decimal degrees (WGS 84) _(Float)_
//
// ##### Example point region
// ```no_run
// {
//   lat: 40.671659,
//   lon: -73.936631
// }
// ```
//
// ### polygon
// Define a custom polygon region using a record containing the following properties:
//
// - **points**: points that define the custom polygon _(Array of records)_
//
//     Define each point with a record containing the following properties:
//
//       - **lat**: latitude in decimal degrees (WGS 84) _(Float)_
//       - **lon**: longitude in decimal degrees (WGS 84) _(Float)_
//
// ##### Example polygonal region
// ```no_run
// {
//   points: [
//     {lat: 40.671659, lon: -73.936631},
//     {lat: 40.706543, lon: -73.749177},
//     {lat: 40.791333, lon: -73.880327}
//   ]
// }
// ```
//
// ## GIS geometry definitions
// Many functions in the Geo package manipulate data based on geographic information system (GIS) data.
// Define GIS geometry using the following:
//
// - Any [region type](#region-definitions) _(typically [point](#point))_
// - [linestring](#linestring)
//
// ### linestring
// Define a geographic linestring path using a record containing the following properties:
//
// - **linestring**: string containing comma-separated longitude and latitude
//   coordinate pairs (`lon lat,`):
//
// ```no_run
// {
//   linestring: "39.7515 14.01433, 38.3527 13.9228, 36.9978 15.08433"
// }
// ```
//
// ## Distance units
// The `geo` package supports the following units of measurement for distance:
//
// - `m` - meters
// - `km` - kilometers _(default)_
// - `mile` - miles
//
// ### Define distance units
// Use the `units` option to define custom units of measurement:
//
// ```no_run
// import "experimental/geo"
//
// option geo.units = {distance: "mile"}
// ```
//
// ## Metadata
// introduced: 0.63.0
// tags: geotemporal
//
package geo


import "experimental"
import "influxdata/influxdb/v1"

// units defines the unit of measurment used in geotemporal operations.
//
// ## Metadata
// introduced: 0.78.0
//
option units = {distance: "km"}

// stContains returns boolean indicating whether the defined region contains a specified GIS geometry.
//
// `geo.stContains` is used as a helper function for `geo.ST_Contains()`.
//
// ## Parameters
// - region: Region to test. Specify record properties for the shape.
// - geometry: GIS geometry to test. Can be either point or linestring geometry.
// - units: Record that defines the unit of measurement for distance.
//
// ## Metadata
// introduced: 0.78.0
// tags: geotemporal
//
builtin stContains : (region: A, geometry: B, units: {distance: string}) => bool
    where
    A: Record,
    B: Record

// stDistance returns the distance from a given region to a specified GIS geometry.
//
// `geo.stDistance` is used as a helper function for `geo.ST_Distance()`.
//
// ## Parameters
// - region: Region to test. Specify record properties for the shape.
// - geometry: GIS geometry to test. Can be either point or linestring geometry.
// - units: Record that defines the unit of measurement for distance.
//
// ## Metadata
// introduced: 0.78.0
// tags: geotemporal
//
builtin stDistance : (region: A, geometry: B, units: {distance: string}) => float
    where
    A: Record,
    B: Record

// stLength returns the [spherical length or distance](https://mathworld.wolfram.com/SphericalDistance.html)
// of the specified GIS geometry.
//
// `geo.stLength` is used as a helper function for `geo.ST_Length()`.
//
// ## Parameters
// - geometry: GIS geometry to test. Can be either point or linestring geometry.
//   Point geometry will always return `0.0`.
// - units: Record that defines the unit of measurement for distance.
//
// ## Metadata
// introduced: 0.78.0
// tags: geotemporal
//
builtin stLength : (geometry: A, units: {distance: string}) => float where A: Record

// ST_Contains returns boolean indicating whether the defined region contains a
// specified GIS geometry.
//
// ## Parameters
// - region: Region to test. Specify record properties for the shape.
// - geometry: GIS geometry to test. Can be either point or linestring geometry.
// - units: Record that defines the unit of measurement for distance.
//   Default is the `geo.units` option.
//
// ## Metadata
// introduced: 0.78.0
// tags: geotemporal
//
ST_Contains = (region, geometry, units=units) =>
    stContains(region: region, geometry: geometry, units: units)

// ST_Distance returns the distance from a given region to a specified GIS geometry.
//
// ## Parameters
// - region: Region to test. Specify record properties for the shape.
// - geometry: GIS geometry to test. Can be either point or linestring geometry.
// - units: Record that defines the unit of measurement for distance.
//   Default is the `geo.units` option.
//
// ## Metadata
// introduced: 0.78.0
// tags: geotemporal
//
ST_Distance = (region, geometry, units=units) =>
    stDistance(region: region, geometry: geometry, units: units)

// ST_DWithin tests if the specified region is within a defined distance from
// the specified GIS geometry and returns `true` or `false`.
//
// ## Parameters
// - region: Region to test. Specify record properties for the shape.
// - geometry: GIS geometry to test. Can be either point or linestring geometry.
// - distance: Maximum distance allowed between the region and geometry.
//   Define distance units with the `geo.units` option.
// - units: Record that defines the unit of measurement for distance.
//   Default is the `geo.units` option.
//
// ## Metadata
// introduced: 0.78.0
// tags: geotemporal
//
ST_DWithin = (region, geometry, distance, units=units) =>
    stDistance(region: region, geometry: geometry, units: units) <= distance

// ST_Intersects tests if the specified GIS geometry intersects with the
// specified region and returns `true` or `false`.
//
// ## Parameters
// - region: Region to test. Specify record properties for the shape.
// - geometry: GIS geometry to test. Can be either point or linestring geometry.
// - units: Record that defines the unit of measurement for distance.
//   Default is the `geo.units` option.
//
// ## Metadata
// introduced: 0.78.0
// tags: geotemporal
//
ST_Intersects = (region, geometry, units=units) =>
    stDistance(region: region, geometry: geometry, units: units) <= 0.0

// ST_Length returns the [spherical length or distance](https://mathworld.wolfram.com/SphericalDistance.html)
// of the specified GIS geometry.
//
// ## Parameters
// - geometry: GIS geometry to test. Can be either point or linestring geometry.
//   Point geometry will always return `0.0`.
// - units: Record that defines the unit of measurement for distance.
//
// ## Metadata
// introduced: 0.78.0
// tags: geotemporal
//
ST_Length = (geometry, units=units) => stLength(geometry: geometry, units: units)

// ST_LineString converts a series of geographic points into linestring.
//
// Group data into meaningful, ordered paths to before converting to linestring.
// Rows in each table must have `lat` and `lon` columns.
// Output tables contain a single row with a `st_linestring` column containing
// the resulting linestring.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Convert a series of geographic points into linestring
// ```
// # import "array"
// import "experimental/geo"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", lon: 39.7515, lat: 14.01433},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", lon: 38.3527, lat: 13.9228},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", lon: 36.9978, lat: 15.08433},
// #         {_time: 2021-01-01T00:00:00Z, id: "b546c", lon: 24.0069, lat: -14.5464},
// #         {_time: 2021-01-02T01:00:00Z, id: "b546c", lon: 25.1304, lat: -13.3338},
// #         {_time: 2021-01-03T02:00:00Z, id: "b546c", lon: 26.7899, lat: -12.0433},
// #     ],
// # )
// #     |> group(columns: ["id"])
//
// < data
// >     |> geo.ST_LineString()
// ```
//
// ## Metadata
// introduced: 0.78.0
// tags: geotemporal,transformations,aggregates
//
ST_LineString = (tables=<-) =>
    tables
        |> reduce(
            fn: (r, accumulator) =>
                ({
                    __linestring:
                        accumulator.__linestring + (if accumulator.__count > 0 then ", " else "")
                            +
                            string(v: r.lon) + " " + string(v: r.lat),
                    __count: accumulator.__count + 1,
                }),
            identity: {__linestring: "", __count: 0},
        )
        |> drop(columns: ["__count"])
        |> rename(columns: {__linestring: "st_linestring"})

// getGrid calculates a grid or set of cell ID tokens for a specified region.
//
// **Note**: S2 grid cells may not perfectly align with the defined region,
// so results include S2 grid cells fully and partially covered by the region.
//
// ## Parameters
// - region: Region used to return S2 cell ID tokens.
//   Specify record properties for the region shape.
// - minSize: Minimum number of cells that cover the specified region.
// - maxSize: Minimum number of cells that cover the specified region.
// - level: S2 cell level of grid cells.
// - maxLevel: Maximumn S2 cell level of grid cells.
// - units: Record that defines the unit of measurement for distance.
//
builtin getGrid : (
        region: T,
        ?minSize: int,
        ?maxSize: int,
        ?level: int,
        ?maxLevel: int,
        units: {distance: string},
    ) => {level: int, set: [string]}
    where
    T: Record

// getLevel returns the S2 cell level of specified cell ID token.
//
// ## Parameters
// - token: S2 cell ID token.
//
// ## Examples
// ### Return the S2 cell level of an S2 cell ID token
// ```no_run
// import "experimental/geo"
//
// geo.getLevel(token: "166b59")
//
// // Returns 10
// ```
//
// ## Metadata
// tags: geotemporal
//
builtin getLevel : (token: string) => int

// s2CellIDToken returns and S2 cell ID token for given cell or point at a
// specified S2 cell level.
//
// ## Parameters
// - token: S2 cell ID token to update.
//
//   Useful for changing the S2 cell level of an existing S2 cell ID token.
//
// - point: Record with `lat` and `lon` properties that specify the latitude and
//   longitude in decimal degrees (WGS 84) of a point.
// - level: S2 cell level to use when generating the S2 cell ID token.
//
// ## Examples
//
// ### Use latitude and longitude values to generate S2 cell ID tokens
// ```
// # import "array"
// import "experimental/geo"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", lon: 39.7515, lat: 14.01433},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", lon: 38.3527, lat: 13.9228},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", lon: 36.9978, lat: 15.08433},
// #         {_time: 2021-01-01T00:00:00Z, id: "b546c", lon: 24.0069, lat: -14.5464},
// #         {_time: 2021-01-02T01:00:00Z, id: "b546c", lon: 25.1304, lat: -13.3338},
// #         {_time: 2021-01-03T02:00:00Z, id: "b546c", lon: 26.7899, lat: -12.0433},
// #     ],
// # )
// #     |> group(columns: ["id"])
//
// < data
// >     |> map(fn: (r) => ({r with s2_cell_id: geo.s2CellIDToken(point: {lat: r.lat, lon: r.lon}, level: 10)}))
// ```
//
// ### Update S2 cell ID token level
// ```
// # import "array"
// import "experimental/geo"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", s2_cell_id: "166b59"},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", s2_cell_id: "16696d"},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", s2_cell_id: "166599"},
// #         {_time: 2021-01-01T00:00:00Z, id: "b546c", s2_cell_id: "1960d7"},
// #         {_time: 2021-01-02T01:00:00Z, id: "b546c", s2_cell_id: "1965c7"},
// #         {_time: 2021-01-03T02:00:00Z, id: "b546c", s2_cell_id: "1971dd"},
// #     ],
// # )
// #     |> group(columns: ["id"])
//
// < data
// >    |> map(fn: (r) => ({r with s2_cell_id: geo.s2CellIDToken(token: r.s2_cell_id, level: 5)}))
// ```
//
// ## Metadata
// introduced: 0.64.0
// tags: geotemporal
//
builtin s2CellIDToken : (?token: string, ?point: {lat: float, lon: float}, level: int) => string

// s2CellLatLon returns the latitude and longitude of the center of an S2 cell.
//
// ## Parameters
// - token: S2 cell ID token.
//
// ## Examples
// ### Return the center coordinates of an S2 cell
// ```no_run
// import "experimental/geo"
//
// geo.s2CellLatLon(token: "89c284")
//
// // Returns {lat: 40.812535546624574, lon: -73.55941282728273}
// ```
//
// ## Metadata
// introduced: 0.78.0
// tags: geotemporal
//
builtin s2CellLatLon : (token: string) => {lat: float, lon: float}

// _detectLevel returns the level of the S2 cell ID in the `s2cellID` column of
// the first record from the first table in a stream of tables.
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

// toRows pivots fields into columns based on time.
//
// Latitude and longitude should be stored as fields in InfluxDB.
// Because most `geo` package transformation functions require rows to have
// `lat` and `lon` columns, `lat` and `lot` fields must be pivoted into columns.
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Pivot lat and lon fields into columns
// ```
// # import "array"
// import "experimental/geo"
// #
// # data = array.from(rows: [
// #     {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "lat", _value: 14.01433},
// #     {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "lat", _value: 13.9228},
// #     {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "lat", _value: 15.08433},
// #     {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "lon", _value: 39.7515},
// #     {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "lon", _value: 38.3527},
// #     {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "lon", _value: 36.9978},
// # ])
// #     |> group(columns: ["id", "_field"])
//
// < data
// >     |> geo.toRows()
// ```
//
// ## Metadata
// tags: transformations,geotemporal
//
toRows = (tables=<-) =>
    tables
        |> v1.fieldsAsCols()

// shapeData renames existing latitude and longitude fields to **lat** and **lon**
// and adds an **s2\_cell\_id** tag.
//
// Use `geo.shapeData()` to ensure geotemporal data meets the requirements of the Geo package:
//
// 1. Rename existing latitude and longitude fields to `lat` and `lon`.
// 2. Pivot fields into columns based on `_time`.
// 3. Generate `s2_cell_id` tags using `lat` and `lon` values and a specified [S2 cell level](https://s2geometry.io/resources/s2cell_statistics.html).
//
// ## Parameters
// - latField: Name of the existing field that contains the latitude value in decimal degrees (WGS 84).
//
//   Field is renamed to `lat`.
//
// - lonField: Name of the existing field that contains the longitude value in decimal degrees (WGS 84).
//
//   Field is renamed to `lon`.
//
// - level: [S2 cell level](https://s2geometry.io/resources/s2cell_statistics.html)
//   to use when generating the S2 cell ID token.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ```
// # import "array"
// import "experimental/geo"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "latitude", _value: 14.01433},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "latitude", _value: 13.9228},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "latitude", _value: 15.08433},
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "longitude", _value: 39.7515},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "longitude", _value: 38.3527},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "longitude", _value: 36.9978},
// #     ],
// # )
// #     |> group(columns: ["id", "_field"])
//
// < data
// >     |> geo.shapeData(latField: "latitude", lonField: "longitude", level: 10)
// ```
//
// ## Metadata
// tags: transformations,geotemporal
// introduced: 0.65.0
//
shapeData = (tables=<-, latField, lonField, level) =>
    tables
        |> map(
            fn: (r) =>
                ({r with _field:
                        if r._field == latField then
                            "lat"
                        else if r._field == lonField then
                            "lon"
                        else
                            r._field,
                }),
        )
        |> toRows()
        |> map(
            fn: (r) =>
                ({r with s2_cell_id: s2CellIDToken(point: {lat: r.lat, lon: r.lon}, level: level)}),
        )
        |> experimental.group(columns: ["s2_cell_id"], mode: "extend")

// gridFilter filters data by a specified geographic region.
//
// The function compares input data to a set of S2 cell ID tokens located in the specified region.
// Input data must include an `s2_cell_id` column that is **part of the group key**.
//
// **Note**: S2 Grid cells may not perfectly align with the defined region,
// so results may include data with coordinates outside the region, but inside
// S2 grid cells partially covered by the region.
// Use `geo.toRows()` and `geo.strictFilter()` after `geo.gridFilter()` to precisely filter points.
//
// ## Parameters
// - region: Region containing the desired data points.
//
//   Specify record properties for the shape.
//
// - minSize: Minimum number of cells that cover the specified region.
//   Default is `24`.
// - maxSize: Maximum number of cells that cover the specified region.
//   Default is `-1` (unlimited).
// - level: [S2 cell level](https://s2geometry.io/resources/s2cell_statistics.html)
//   of grid cells. Default is `-1`.
//
//   **Note:** `level` is mutually exclusive with `minSize` and `maxSize` and
//   must be less than or equal to `s2cellIDLevel`.
//
// - s2cellIDLevel: [S2 cell level](https://s2geometry.io/resources/s2cell_statistics.html)
//   used in the `s2_cell_id` tag. Default is `-1` (detects S2 cell level from the S2 cell ID token).
// - units: Record that defines the unit of measurement for distance.
//   Default is the `geo.units` option.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Filter data to a specified region
// ```
// # import "array"
// import "experimental/geo"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "latitude", _value: 41.01433},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "latitude", _value: 40.9228},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "latitude", _value: 39.08433},
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "longitude", _value: -70.7515},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "longitude", _value: -73.3527},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "longitude", _value: -75.9978},
// #     ],
// # )
// #     |> group(columns: ["id", "_field"])
// #     |> geo.shapeData(latField: "latitude", lonField: "longitude", level: 10)
//
// < data
// >     |> geo.gridFilter(region: {lat: 40.69335938, lon: -73.30078125, radius: 20.0})
// ```
//
// ## Metadata
// tags: transformations,filters,geotemporal
//
gridFilter = (
        tables=<-,
        region,
        minSize=24,
        maxSize=-1,
        level=-1,
        s2cellIDLevel=-1,
        units=units,
    ) =>
    {
        _s2cellIDLevel =
            if s2cellIDLevel == -1 then
                tables
                    |> _detectLevel()
            else
                s2cellIDLevel
        _grid =
            getGrid(
                region: region,
                minSize: minSize,
                maxSize: maxSize,
                level: level,
                maxLevel: _s2cellIDLevel,
                units: units,
            )

        return
            tables
                |> filter(
                    fn: (r) =>
                        if _grid.level == _s2cellIDLevel then
                            contains(value: r.s2_cell_id, set: _grid.set)
                        else
                            contains(
                                value: s2CellIDToken(token: r.s2_cell_id, level: _grid.level),
                                set: _grid.set,
                            ),
                )
    }

// strictFilter filters data by latitude and longitude in a specified region.
//
// This filter is more strict than `geo.gridFilter()`, but for the best performance,
// use `geo.strictFilter()` after `geo.gridFilter()`.
// Input rows must have `lat` and `lon` columns.
//
// ## Parameters
// - region: Region containing the desired data points.
//
//   Specify record properties for the shape.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Strictly filter data to a specified region
// ```
// # import "array"
// import "experimental/geo"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "latitude", _value: 41.01433},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "latitude", _value: 40.9228},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "latitude", _value: 39.08433},
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "longitude", _value: -70.7515},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "longitude", _value: -73.3527},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "longitude", _value: -75.9978},
// #     ],
// # )
// #     |> group(columns: ["id", "_field"])
// #     |> geo.shapeData(latField: "latitude", lonField: "longitude", level: 5)
//
// < data
// >     |> geo.strictFilter(region: {lat: 40.69335938, lon: -73.30078125, radius: 50.0})
// ```
//
// ## Metadata
// tags: transformations,filters,geotemporal
//
strictFilter = (tables=<-, region) =>
    tables
        |> filter(fn: (r) => ST_Contains(region: region, geometry: {lat: r.lat, lon: r.lon}))

// filterRows filters data by a specified geographic region with the option of strict filtering.
//
// This function is a combination of `geo.gridFilter()` and `geo.strictFilter()`.
// Input data must include an `s2_cell_id` column that is **part of the group key**.
//
// ## Parameters
// - region: Region containing the desired data points.
//
//   Specify record properties for the shape.
//
// - minSize: Minimum number of cells that cover the specified region.
//   Default is `24`.
// - maxSize: Maximum number of cells that cover the specified region.
//   Default is `-1` (unlimited).
// - level: [S2 cell level](https://s2geometry.io/resources/s2cell_statistics.html)
//   of grid cells. Default is `-1`.
//
//   **Note:** `level` is mutually exclusive with `minSize` and `maxSize` and
//   must be less than or equal to `s2cellIDLevel`.
//
// - s2cellIDLevel: [S2 cell level](https://s2geometry.io/resources/s2cell_statistics.html)
//   used in the `s2_cell_id` tag. Default is `-1` (detects S2 cell level from the `s2_cell_id` tag).
// - strict: Enable strict geographic data filtering. Default is `true`.
//
//   Strict filtering returns only points with coordinates in the defined region.
//   Non-strict filtering returns all points from S2 grid cells that are partially
//   covered by the defined region.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Strictly filter geotemporal data by region
// ```no_run
// # import "array"
// import "experimental/geo"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "latitude", _value: 41.01433},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "latitude", _value: 40.9228},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "latitude", _value: 39.08433},
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "longitude", _value: -70.7515},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "longitude", _value: -73.3527},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "longitude", _value: -75.9978},
// #     ],
// # )
// #     |> group(columns: ["id", "_field"])
// #     |> geo.shapeData(latField: "latitude", lonField: "longitude", level: 5)
//
// < data
// >    |> geo.filterRows(region: {lat: 40.69335938, lon: -73.30078125, radius: 100.0})
// ```
//
// ### Approximately filter geotemporal data by region
// ```no_run
// # import "array"
// import "experimental/geo"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "latitude", _value: 41.01433},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "latitude", _value: 40.9228},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "latitude", _value: 39.08433},
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "longitude", _value: -70.7515},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "longitude", _value: -73.3527},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "longitude", _value: -75.9978},
// #     ],
// # )
// #     |> group(columns: ["id", "_field"])
// #     |> geo.shapeData(latField: "latitude", lonField: "longitude", level: 5)
//
// < data
// >    |> geo.filterRows(region: {lat: 40.69335938, lon: -73.30078125, radius: 100.0}, strict: false)
// ```
//
// ## Metadata
// tags: transformations,filters,geotemporal
//
filterRows = (
        tables=<-,
        region,
        minSize=24,
        maxSize=-1,
        level=-1,
        s2cellIDLevel=-1,
        strict=true,
    ) =>
    {
        _columns =
            tables
                |> columns(column: "columns")
                |> findColumn(column: "columns", fn: (key) => true)
        _rows =
            if contains(value: "lat", set: _columns) then
                tables
                    |> gridFilter(
                        region: region,
                        minSize: minSize,
                        maxSize: maxSize,
                        level: level,
                        s2cellIDLevel: s2cellIDLevel,
                    )
            else
                tables
                    |> gridFilter(
                        region: region,
                        minSize: minSize,
                        maxSize: maxSize,
                        level: level,
                        s2cellIDLevel: s2cellIDLevel,
                    )
                    |> toRows()
        _result =
            if strict then
                _rows
                    |> strictFilter(region)
            else
                _rows

        return _result
    }

// groupByArea groups rows by geographic area.
//
// Area sizes are determined by the specified `level`.
// Each geographic area is assigned a unique identifier (the S2 cell ID token)
// which is stored in the `newColumn`.
// Results are grouped by `newColumn`.
//
// ## Parameters
// - newColumn: Name of the new column for the unique identifier for each geographic area.
// - level: [S2 Cell level](https://s2geometry.io/resources/s2cell_statistics.html)
//   used to determine the size of each geographic area.
// - s2cellIDLevel: [S2 Cell level](https://s2geometry.io/resources/s2cell_statistics.html)
//   used in the `s2_cell_id` tag. Default is `-1` (detects S2 cell level from the `s2_cell_id` tag).
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ### Group geotemporal data by geographic area
// ```
// # import "array"
// import "experimental/geo"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "latitude", _value: 41.01433},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "latitude", _value: 40.9228},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "latitude", _value: 39.08433},
// #         {_time: 2021-01-01T00:00:00Z, id: "a213b", _field: "longitude", _value: -70.7515},
// #         {_time: 2021-01-02T01:00:00Z, id: "a213b", _field: "longitude", _value: -73.3527},
// #         {_time: 2021-01-03T02:00:00Z, id: "a213b", _field: "longitude", _value: -75.9978},
// #     ],
// # )
// #     |> group(columns: ["id", "_field"])
// #     |> geo.shapeData(latField: "latitude", lonField: "longitude", level: 5)
//
// < data
// >     |> geo.groupByArea(newColumn: "foo", level: 4)
// ```
//
// ## Metadata
// tags: transformations,geotemporal
//
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
                |> map(
                    fn: (r) =>
                        ({r with _s2_cell_id_xxx:
                                s2CellIDToken(point: {lat: r.lat, lon: r.lon}, level: level),
                        }),
                )
                |> rename(columns: {_s2_cell_id_xxx: newColumn})

    return
        _prepared
            |> group(columns: [newColumn])
}

// asTracks groups rows into tracks (sequential, related data points).
//
// ## Parameters
// - groupBy: Columns to group by. These columns should uniquely identify each track.
//   Default is `["id","tid"]`.
// - orderBy: Columns to order results by. Default is `["_time"]`.
//
//   Sort precedence is determined by list order (left to right).
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Group geotemporal data into tracks
// ```
// # import "array"
// import "experimental/geo"
// #
// # data =
// #     array.from(
// #         rows: [
// #             {_time: 2021-01-01T00:00:00Z, id: "a213b", lat: 14.01433, lon: -14.5464},
// #             {_time: 2021-01-02T01:00:00Z, id: "a213b", lat: 13.9228, lon: -13.3338},
// #             {_time: 2021-01-03T02:00:00Z, id: "a213b", lat: 15.08433, lon: -12.0433},
// #             {_time: 2021-01-01T00:00:00Z, id: "b546c", lat: 14.01433, lon: 39.7515},
// #             {_time: 2021-01-02T01:00:00Z, id: "b546c", lat: 13.9228, lon: 38.3527},
// #             {_time: 2021-01-03T02:00:00Z, id: "b546c", lat: 15.08433, lon: 36.9978},
// #         ],
// #     )
//
// < data
// >     |> geo.asTracks()
// ```
//
// ### Group geotemporal data into tracks and sort by specified columns
// ```
// # import "array"
// import "experimental/geo"
// #
// # data =
// #     array.from(
// #         rows: [
// #             {_time: 2021-01-01T00:00:00Z, id: "a213b", lat: 14.01433, lon: -14.5464},
// #             {_time: 2021-01-02T01:00:00Z, id: "a213b", lat: 13.9228, lon: -13.3338},
// #             {_time: 2021-01-03T02:00:00Z, id: "a213b", lat: 15.08433, lon: -12.0433},
// #             {_time: 2021-01-01T00:00:00Z, id: "b546c", lat: 14.01433, lon: 39.7515},
// #             {_time: 2021-01-02T01:00:00Z, id: "b546c", lat: 13.9228, lon: 38.3527},
// #             {_time: 2021-01-03T02:00:00Z, id: "b546c", lat: 15.08433, lon: 36.9978},
// #         ],
// #     )
//
// < data
// >     |> geo.asTracks(orderBy: ["lat", "lon"])
// ```
//
// ## Metadata
// tags: transformations,geotemporal
//
asTracks = (tables=<-, groupBy=["id", "tid"], orderBy=["_time"]) =>
    tables
        |> group(columns: groupBy)
        |> sort(columns: orderBy)

// totalDistance calculates the total distance covered by subsequent points
// in each input table.
//
// Each row must contain `lat` (latitude) and `lon` (longitude) columns that
// represent the geographic coordinates of the point.
// Row sort order determines the order in which distance between points is calculated.
// Use the `geo.units` option to specify the unit of distance to return (default is km).
//
// ## Parameters
// - outputColumn: Total distance output column. Default is `_value`.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Return the total distance travelled per input table
// ```
// # import "array"
// import "experimental/geo"
//
// # data =
// #     array.from(
// #         rows: [
// #             {id: "ABC1", _time: 2022-01-01T00:00:00Z, lat: 85.1, lon: 42.2},
// #             {id: "ABC1", _time: 2022-01-01T01:00:00Z, lat: 71.3, lon: 50.8},
// #             {id: "ABC1", _time: 2022-01-01T02:00:00Z, lat: 63.1, lon: 62.3},
// #             {id: "ABC1", _time: 2022-01-01T03:00:00Z, lat: 50.6, lon: 74.9},
// #             {id: "DEF2", _time: 2022-01-01T00:00:00Z, lat: -10.8, lon: -12.2},
// #             {id: "DEF2", _time: 2022-01-01T01:00:00Z, lat: -16.3, lon: -0.8},
// #             {id: "DEF2", _time: 2022-01-01T02:00:00Z, lat: -23.2, lon: 12.3},
// #             {id: "DEF2", _time: 2022-01-01T03:00:00Z, lat: -30.4, lon: 24.9},
// #         ],
// #     )
// #     |> group(columns: ["id"])
// #
// < data
// >     |> geo.totalDistance()
// ```
//
// ### Return the total distance travelled in miles
// ```
// # import "array"
// import "experimental/geo"
//
// option geo.units = {distance: "mile"}
//
// # data =
// #     array.from(
// #         rows: [
// #             {id: "ABC1", _time: 2022-01-01T00:00:00Z, lat: 85.1, lon: 42.2},
// #             {id: "ABC1", _time: 2022-01-01T01:00:00Z, lat: 71.3, lon: 50.8},
// #             {id: "ABC1", _time: 2022-01-01T02:00:00Z, lat: 63.1, lon: 62.3},
// #             {id: "ABC1", _time: 2022-01-01T03:00:00Z, lat: 50.6, lon: 74.9},
// #             {id: "DEF2", _time: 2022-01-01T00:00:00Z, lat: -10.8, lon: -12.2},
// #             {id: "DEF2", _time: 2022-01-01T01:00:00Z, lat: -16.3, lon: -0.8},
// #             {id: "DEF2", _time: 2022-01-01T02:00:00Z, lat: -23.2, lon: 12.3},
// #             {id: "DEF2", _time: 2022-01-01T03:00:00Z, lat: -30.4, lon: 24.9},
// #         ],
// #     )
// #     |> group(columns: ["id"])
// #
// < data
// >     |> geo.totalDistance()
// ```
//
// ## Metadata
// introduced: 0.192.0
// tags: transformations, geotemporal, aggregates
//
totalDistance = (tables=<-, outputColumn="_value") =>
    tables
        |> reduce(
            identity: {index: 0, lat: 0.0, lon: 0.0, totalDistance: 0.0},
            fn: (r, accumulator) => {
                _lastPoint =
                    if accumulator.index == 0 then
                        {lat: r.lat, lon: r.lon}
                    else
                        {lat: accumulator.lat, lon: accumulator.lon}
                _currentPoint = {lat: r.lat, lon: r.lon}

                return {
                    index: accumulator.index + 1,
                    lat: r.lat,
                    lon: r.lon,
                    totalDistance:
                        accumulator.totalDistance + ST_Distance(
                                region: _lastPoint,
                                geometry: _currentPoint,
                            ),
                }
            },
        )
        |> drop(columns: ["index", "lat", "lon"])
        |> rename(columns: {totalDistance: outputColumn})
