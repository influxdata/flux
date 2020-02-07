## Package `geo`

The package provides functions for geographic location filtering and grouping.
It is designed to work on a schema with a set of tags by default named `_cidX`,
where X specifies S2 Cell level, fields `lat`, `lon` and `cid`.

The maximum level X is up to user to decide with respect to cardinality.
The number of cells for each level is shown at [https://s2geometry.io/resources/s2cell_statistics.html].
These tags hold value of cell ID as hex token (`s2.CellID.ToToken()`) of corresponding level.

The schema may/should further contain a tag which identifies data source (`id` by default),
and a field representing track ID (`tid` by default). For some use cases a tag denoting point
type (with values like `"start"`/`"stop"`/`"via"`, for example) is also helpful.

Examples of line protocol input that conforms to such schema:
```
taxi,_pt=start,_cid1=8c,_cid2=89,_cid3=89c,_cid4=89d,_cid5=89c4,_cid6=89c3,_cid7=89c24,_cid8=89c25,_cid9=89c25c,_cid10=89c259,_cid11=89c2594 tip=3.75,dist=14.3,lat=40.744614,lon=-73.979424,cid="89c2590882ea0441",tid=1572566401947779410i 1572566401947779410
bike,id=bike007,_pt=via,_cid1=8c,_cid2=89,_cid3=89c,_cid4=89d,_cid5=89c4,_cid6=89c3,_cid7=89c24,_cid8=89c25,_cid9=89c25c,_cid10=89c259,_cid11=89c2594 lat=40.753944,lon=-73.992035,cid="89c2590882ea0441",tid=1572588115012345678i 1572567115082153551
```

The grouping functions works on row-wise sets (as it very likely appears in line protocol),
with geo-temporal values (tags `_cidX`, `id` and fields `lat`, `lon`, `cid` and `tid`) as columns.
That is achieved by correlation by `_time` (and `id` if present) using `pivot()` or provided `geo.toRows()` function.
Therefore it is advised to store time with nanoseconds precision to avoid false matches.

Fundamental transformations:
- `gridFilter`
- `strictFilter`
- `toRows`

Schema changing operations:
- `stripMeta`

Aggregate operations:
- `groupByArea`
- `asTracks`

The package uses the following types:
- `box` - has the following named float values: `minLat`, `maxLat`, `minLon`, `maxLon`.
- `circle` - has the following named float values: `lat`, `lon`, `radius`.

**Experimental alternative simple schema**

Single tag `_ci` containing cell ID as token with level decided by user.
Also, `cid` field is not needed in schema with `lat` and `lon` fields.

Examples of line protocol input (cell level 11 token in `_ci`):
```
taxi,_pt=start,_ci=89c2594 tip=3.75,dist=14.3,lat=40.744614,lon=-73.979424,tid=1572566401947779410i 1572566401947779410
bike,id=bike007,_pt=via,_ci=89c2594 lat=40.753944,lon=-73.992035,tid=1572588115012345678i 1572567115082153551
```

Corresponding functions: `gridFilter2`, `groupByArea2`.

### Function `gridFilter`

The `gridFilter` filters data by specified box or circle.
It calculates tokens grid that overlays specified box or circle.
Therefore result may contain data where corresponding latitude and/or longitude is outside the box.

This filter function is intended to be fast as it uses tags values.
If precise filtering is needed, `strictFilter()` may be used later (after `toRows()`).

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.gridFilter(box: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
``` 
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.gridFilter(circle: {lat: 40.69335938, lon: -73.30078125, radius: 20.0})
``` 

Grid calculation may be customized by following options:
- `minSize` - minimum number of tiles that cover specified box (default value is `9`).
- `maxSize` - maximum number of tiles (optional)
- `level` - desired cell level of the tiles (optional)
- `maxLevelIndex` - highest cell level X available in `_cidX` tag.

The `level` parameter is mutually exclusive with others and must be less or equal to `maxLevelIndex`.
By default the algorithm attempts to 9 tiles grid with finest level that is less than
or equal to `maxLevelIndex`.
User is required to specify correct `maxLevelIndex` that matches existing schema.
For example, when schema has tags `_cid1` ... `_cid5`, then `5` must be passed as `maxLevelIndex`. 

#### Function definition

```
gridFilter = (tables=<-, fn=tokenFilterEx, box={}, circle={}, minGridSize=9, maxGridSize=-1, level=-1, maxLevelIndex=30) => {
  grid = getGrid(box: box, circle: circle, minSize: minGridSize, maxSize: maxGridSize, level: level, maxLevel: maxLevelIndex)
  return
    tables
      |> fn(grid: grid)
}
```

### Function `strictFilter`

Filters records by lat/lon box or center/radius circle. Unlike `gridFilter()`, this is a strict filter.
Must be used after `toRows()` because it requires `lat` and `lon` columns in input row set.

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.toRows()
  |> geo.strictFilter(box: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
``` 

It makes a lot of sense to use it together with `griFilter()`.

#### Function definition

```
strictFilter = (tables=<-, box={}, circle={}) =>
  tables
    |> filter(fn: (r) =>
      containsLatLon(box: box, circle: circle, lat: r.lat, lon: r.lon)
    )
```

### Function `toRows`

Collects values to row-wise sets.

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.toRows()
```

#### Function definition

```
toRows = (tables=<-, correlationKey=["_time"]) =>
  tables
    |> pivot(
      rowKey: correlationKey,
      columnKey: ["_field"],
      valueColumn: "_value"
    )
```

### Function `groupByArea`

Groups rows by area blocks of size determined by `level` (see [https://s2geometry.io/resources/s2cell_statistics.html]). 
Result is grouped by `newColumn`.
Parameter `maxLevelIndex` specifies highest cell level X in the input rows.

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.gridFilter(box: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
  |> geo.toRows()
  |> geo.groupByArea(newColumn: "cidx", level: 3, maxLevelIndex: 5)
```

#### Function definition

```
groupByArea = (tables=<-, newColumn, level, maxLevel, prefix="_cid") => {
  prepared =
    if level <= maxlevel then
      tables
	    |> duplicate(column: prefix + string(v: level), as: newColumn)
    else
      tables
        |> map(fn: (r) => ({ r with _cidx: getParent(v: r.cid, level: level) }))
	    |> rename(columns: { _cidx: newColumn })
  return prepared
    |> group(columns: [newColumn])
}
```

### Function `asTracks`

Organizes rows into tracks.

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.gridFilter(box: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
  |> geo.toRows(correlationKey: ["_time", "id"])
  |> geo.asTracks()
```

#### Function definition

```
asTracks = (tables=<-, groupBy=["id","tid"], orderBy=["_time"]) =>
  tables
    |> group(columns: groupBy)
    |> sort(columns: orderBy)
```

### Function `stripMeta`

Drops cell level indexes columns (`_cidX` by default) except those specified.
It will fail if input tables are grouped by any of them.

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.toRows()
  |> geo.groupByArea(newColumn: "cidx", level: 5, maxLevelIndex: 5)
  |> geo.stripMeta(except: ["_cid5"])
```

#### Function definition

```
stripMeta = (tables=<-, pattern=/_cid\d+/, except=[]) =>
  tables
    |> drop(fn: (column) => column =~ pattern and (length(arr: except) == 0 or not contains(value: column, set: except)))
```

## *Experimental alternative simple schema functions*

### Function `gridFilter2`

Grid filtering function that works on simplified schema.
- `ciLevel` - cell level of token stored in `_ci` tag 

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.gridFilter2(box: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875}, ciLevel=11)
``` 

User is required to specify correct `ciLevel` that matches existing schema.

### Function `groupByArea2`

Groups rows by area blocks of size determined by `level` (see [https://s2geometry.io/resources/s2cell_statistics.html]). 
Result is grouped by `newColumn`.
Parameter `ciLevel` specifies cell level of token stored in `_ci` tag.

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.gridFilter(box: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
  |> geo.toRows()
  |> geo.groupByArea(newColumn: "cix", level: 3, ciLevel: 11)
```

#### Function definition

```
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
```
