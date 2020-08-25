## Package `geo`

The package provides functions for geographic location filtering and grouping.
It uses golang implementation of S2 Geometry Library [https://s2geometry.io/].
It is designed to work on a schema with a tags `s2_cell_id` which contains S2 cell ID (as token)
of level decided by the user, and fields `lat`, `lon` containing WGS-84 coordinates
in decimal degrees.

The `s2_cell_id` tag contains cell ID token (`s2.CellID.ToToken()`) of corresponding level.
The cell levels are shown at [https://s2geometry.io/resources/s2cell_statistics.html].
The level must be decided by the user.
The rule of thumb is that it should be as high as possible for faster filtering
but not too high in order to avoid risk of having high cardinality.
The token can be easily calculated from lat and lon using Google S2 library which is available for many languages.

The schema may further contain a tag which identifies data source (`id` by default),
and a field representing track identification (`tid` by default).
For some use cases a tag denoting point type (eg. with values like `start`/`stop`/`via`) may also be useful.

Examples of line protocol input (`s2_cell_id` with cell ID level 11 token):

```
taxi,pt=start,s2_cell_id=89c2594 tip=3.75,dist=14.3,lat=40.744614,lon=-73.979424,tid=1572566401123234345i 1572566401947779410
```
```
bike,id=biker-007,pt=via,s2_cell_id=89c25dc lat=40.753944,lon=-73.992035,tid=1572588100i 1572567115
```

Some functions in this package works on row-wise sets (as it very likely appears in line protocol),
with fields `lat`, `lon` (and possibly `tid`) as columns.
That can be achieved by calling `v1.fieldsAsCols()` or `toRows()` before these functions.

**Fundamental transformations:**
- `gridFilter`
- `strictFilter`
- `toRows`
- `filterRows`
- `shapeData`

**Aggregate operations:**
- `groupByArea`
- `asTracks`

**S2 geometry functions:**
- `s2CellIDToken`
- `s2CellLatLon`

**GIS functions:**
- `ST_Contains`
- `ST_Distance`
- `ST_DWithin`
- `ST_Intersects`
- `ST_Length`
- `ST_LineString`

**The package uses the following types:**
- `region` - depending on shape, it has the following named float values:
  - box - `minLat`, `maxLat`, `minLon`, `maxLon`
  - circle (cap) - `lat`, `lon`, `radius` (in decimal km)
  - point - `lat`, `lon`
  - polygon - `points` - array of points
- `geometry` - can be any region type (typically point), and also:
  - path  - `linestring` - string with comma-separated pairs of longitude and latitude

**Units:**

Supported units are:
- distance - `m`, `km`, `mile`

Default units:
```js
option units = {
  distance: "km"
}
```

To change units, assign a new value the to `units` option, eg:
```js
import "experimental/geo"

option geo.units = {distance:"mile"}

from(bucket:"rides")
  ...
```

### Function `gridFilter`

The `gridFilter` filters data by specified region.
It calculates grid of tokens that overlays specified region and then uses `s2_cell_id`
to filter against the set.
The grid cells always overlay the region, therefore result may contain data with
latitude and/or longitude outside the region.

This filter function is intended to be fast as it uses `s2_cell_id` tag to filter records.
If precise filtering is needed, `strictFilter()` may be used later (after `toRows()`).

Example:
```js
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "taxi")
  |> geo.gridFilter(region: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
```
```js
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "taxi")
  |> geo.gridFilter(region: {lat: 40.69335938, lon: -73.30078125, radius: 20.0})
```
```js
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "taxi")
  |> geo.gridFilter(region: {points: [{lat: 40.671659, lon: -73.936631}, {lat: 40.706543, lon: -73.749177},{lat: 40.791333, lon: -73.880327}]})
```

**Grid calculation may be customized by following options:**
- `minSize` - minimum number of tiles that cover specified region (default value is `24`).
- `maxSize` - maximum number of tiles (optional)
- `level` - desired cell level of the grid tiles (optional)
- `s2cellIDLevel` - cell level in `s2_cell_id` tag (optional - the function attempts to autodetect it)

The `level` parameter is mutually exclusive with others and must be less or equal to `s2cellIDLevel`.

### Function `strictFilter`

Filters records by lat/lon. Unlike `gridFilter()`, this is a strict filter.
Must be used after `toRows()` because it requires `lat` and `lon` columns in records.

Example:
```js
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "taxi")
  |> geo.toRows()
  |> geo.strictFilter(region: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
```

For best performance, it should be used together with `griFilter()`.

### Function `toRows`

_Note: this function is equivalent to `v1.fieldsAsCols()` and will be removed in the future._

Collects values to row-wise sets.
For geo-temporal data sets the result contains rows with `lat` and `lon`, ie. suitable
for visualization and for functions such as `strictFilter` or `groupByArea`.

Example:
```js
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "taxi")
  |> geo.toRows()
```

#### Function definition

```js
toRows = (tables=<-) =>
  tables
    |> v1.fieldsAsCols()
```

### Function `filterRows`

Combined filter. The sequence is either `gridFilter |> toRows |> strictFilter`
or just `gridFilter |> toRows`, depending on `strict` parameter.
`filterRows` also checks to see if input data has already been pivoted into row-wise
sets and, if so, will skip the call to `toRows`.

Example:
```js
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "taxi")
  |> geo.filterRows(region: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
```

It has the same input parameters as `gridFilter`.
By default it applies strict filtering (`strict=true`).

#### Function definition

```js
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
```

### Function `shapeData`

Shapes data with existing longitude and a latitude fields into the the structure
functions in the Geo package require.
It renames the existing longitude and latitude fields to `lon` and `lat`, pivots
the data into row-wise sets, uses the `lat` and `lon` values to generate and add
the `s2_cell_id` tag based on the specified `level`, and adds the `s2_cell_id`
column to the group key.

Example:
```js
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "migration")
  |> geo.shapeData(lonField: "longitude", latField: "latitude", level: 11)
```

#### Function definition

```js
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
```

### Function `groupByArea`

Groups rows by area blocks of size determined by `level` (see [https://s2geometry.io/resources/s2cell_statistics.html]).
Result is grouped by `newColumn`.

Example:
```js
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "taxi")
  |> geo.gridFilter(region: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
  |> geo.toRows()
  |> geo.groupByArea(newColumn: "l3", level: 3)
```

Optional parameter `s2cellIDLevel` specifies cell level of `s2_cell_id` tag.
By default the function attempts to autodetect it.

### Function `asTracks`

Groups rows into tracks.

Example:
```js
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.gridFilter(region: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
  |> geo.toRows()
  |> geo.asTracks()
```

#### Function definition

```js
asTracks = (tables=<-, groupBy=["id","tid"], orderBy=["_time"]) =>
  tables
    |> group(columns: groupBy)
    |> sort(columns: orderBy)
```

### Function `s2CellIDToken`

Returns S2 cell ID token.

Input parameters are:
- `token` - source token
- `point` - source coordinates
- `level` - requested cell level of the target token

Either `token` or `point` must be specified.

Example:
```js
t = geo.s2CellIDToken(point: {lat: 40.51757813, lon: -73.65234375}, level: 10)
```

### Function `s2CellLatLon`

Returns coordinates of the S2 cell center.

Input parameters are:
- `token` - cell ID token

Example:
```js
ll = geo.s2CellLatLon(token: "89c284")
```

### Function `ST_Contains`

Returns boolean value whether the region contains geometry or not.
Parameter `geometry` can be either a point or a linestring.

Input parameters are:
- `region`
- `geometry`

Example:
```js
box = {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875}

from(bucket:"mta")
    ...
    |> geo.toRows()
    |> map(fn: (r) => ({
      r with st_contains: ST_Contains(region: box, geometry: {lat: r.lat, lon: r.lon})
    }))
```

```js
box = {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875}

from(bucket:"mta")
    ...
    |> geo.toRows()
    |> geo.asTracks()
    |> geo.ST_LineString()
    |> map(fn: (r) => ({
      r with st_contains: ST_Contains(region: box, geometry: {linestring: r.st_linestring})
    }))
```

### Function `ST_Distance`

Returns distance between specified region and geometry.
Parameter `geometry` can be either a point or a linestring.

Input parameters are:
- `region`
- `geometry`

Example:
```js
box = {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875}

from(bucket:"mta")
    ...
    |> geo.toRows()
    |> map(fn: (r) => ({
      r with st_distance: ST_Distance(region: box, geometry: {lat: r.lat, lon: r.lon})
    }))
```

### Function `ST_DWithin`

Returns boolean if geometry is within a distance to specified region.
Parameter `geometry` can be either a point or a linestring.

Input parameters are:
- `region`
- `geometry`

Example:
```js
box = {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875}

from(bucket:"mta")
    ...
    |> geo.toRows()
    |> map(fn: (r) => ({
      r with st_within: ST_DWithin(region: box, geometry: {lat: r.lat, lon: r.lon}, distance: 15.0)
    }))
```

#### Function definition

```js
ST_DWithin = (region, distance, geometry) =>
  ST_Distance(region: region, geometry: geometry) <= distance
```

### Function `ST_Intersects`

Returns boolean whether geometry intersects specified region.
Parameter `geometry` can be either a point or a linestring.

Example:
```js
box = {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875}

from(bucket:"mta")
    ...
    |> geo.toRows()
    |> geo.asTracks()
    |> geo.ST_LineString()
    |> map(fn: (r) => ({
      r with st_intersects: ST_Intersects(region: box, geometry: {linestring: r.st_linestring})
    }))
```

#### Function definition

```js
ST_Intersects = (region, geometry) =>
  ST_Contains(region: region, geometry: geometry)
```

### Function `ST_Length`

Returns spherical length of specified geometry.
Parameter `geometry` can be either a point (result is 0.0) or a linestring.

Example:
```js
from(bucket:"mta")
    ...
    |> geo.toRows()
    |> geo.asTracks()
    |> geo.ST_LineString()
    |> map(fn: (r) => ({
      r with st_length: ST_Length(geometry: {linestring: r.st_linestring})
    }))
```

### Function `ST_LineString`

Constructs a linestring from a series of points.
Input data should be grouped in such way that they represent a meaningful path before calling this function.
Output is a table with `st_linestring` column holding the result.

Example:
```js
from(bucket:"mta")
    ...
    |> geo.toRows()
    |> geo.asTracks()
    |> geo.ST_LineString()
```

### Function definition

```js
ST_LineString = (tables=<-) =>
  tables
    |> reduce(fn: (r, accumulator) => ({
        r with
        __linestring: accumulator.__linestring + (if accumulator.__count > 0 then ", " else "") + string(v: r.lat) + " " + string(v: r.lon),
        __count: accumulator.__count + 1
      }), identity: {
        __linestring: "",
        __count: 0
      }
    )
    |> rename(columns: {__linestring: "st_linestring"})
```

### Geofencing

Geofencing use case can be realized using custom check query.
In the following example, a point that is outside the region is evaluated as `"warn"` level status.
Then, in the notification rule, change from `"ok"` to `"warn"` signals that object left specified region,
and vice versa.

Example:
```js
import "influxdata/influxdb/monitor"
import "experimental/geo"

// Injected
option task = {name: "Geofencing", every: 1m}

// Injected
check = {
    _check_id: "0000000000000001",
    _check_name: "Central Long Island check",
    _type: "custom",
    tags: {},
}

box = {
    minLat: 40.5880775,
    maxLat: 40.8247008,
    minLon: -73.80014,
    maxLon: -73.4630336,
}

from(bucket: "mta")
  |> range(start: -task.every)
  |> geo.toRows()
  |> keep(columns: ["_measurement", "_time", "id", "lat", "lon")
  |> monitor.check(
      data: check,
      messageFn: messageFn: (r) => (if r._level == monitor.levelWarn then "Train ${r.id} is out" else "Train ${r.id} is in"),
      warn: (r) => not geo.ST_Contains(region: box, geometry: {lat: r.lat, lon: r.lon})
  )
```
_Notes: in this example, only a subset of columns is kept, but of course,
it is optional step (columns `_measurement` and `_time` are required
by `monitor.check()`)._
