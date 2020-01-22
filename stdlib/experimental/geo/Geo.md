## Package `geo`

Provides functions for geographic location filtering and grouping.
It is designed to work on a schema with a set of tags by default named `_gX`,
where X specifies geohash precision (corresponds to its number of characters),
fields `lat`, `lon` and `geohash`.

The `geohash` field holds full geohash precision location value (ie. 12-char string).
The schema may/should further contain a tag which identifies data source (`id` by default),
and a field representing track ID (`tid` by default). For some use cases a tag denoting point
type (with values like `"start"`/`"stop"`/`"via"`, for example) is also helpful.

Examples of line protocol input that conforms to such schema:
```
taxi,pt=end,_g1=d,_g2=dr,_g3=dr5,_g4=dr5r dist=12.7,tip=3.43,lat=40.753944,lon=-73.992035,geohash="dr5ru708u223" 1572567115082153551
bike,id=bike007,pt=via,_g1=d,_g2=dr,_g3=dr5,_g4=dr5r lat=40.753944,lon=-73.992035,geohash="dr5ru708u223",tid=1572588115012345678i 1572567115082153551
```

The grouping functions works on row-wise sets (as it very likely appears in line protocol),
where all the geotemporal data (tags `_gX`, `id` and fields `lat`, `lon`, `geohash` and `tid`) are columns.
That is achieved by correlation by `_time" (and `id` if present) using `pivot()` or provided `geo.toRows()` function.
Therefore it is advised to store time with nanoseconds precision to avoid false matches.

Fundamental transformations:
- `gridFilter`
- `boxFilter`
- `toRows`

Schema changing operations:
- `stripMeta`

Aggregate operations:
- `groupByArea`
- `asTracks`

The package uses the following types:
- `box` - has the following named float values: `minLat`, `maxLat`, `minLon`, `maxLon`.

### Function `gridFilter`

The `gridFilter` filters data by specified lat/lon box.
It calculates geohash grid that overlays specified lat/lon box.
Therefore result may contain data where corresponding latitude and/or longitude is outside the box.

This filter function is intended to be fast as it uses tags values.
If precise filtering is needed, `boxFilter()` may be used later (after `toRows()`).

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.gridFilter(box: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
``` 

Grid calculation may be customized by following options:
- `minSize` - minimum number of tiles that cover specified box (default value is `9`).
- `maxSize` - maximum number of tiles
- `precision` - desired geohash precision of the tiles
- `maxPrecision` - maximum geohash precision of the tiles (default value is `12`).

The `precision` parameter is mutually exclusive with others.
By default the algorithm attempts to 9 tiles grid with finest precision that is less than
or equal to `maxPrecision`.
User is required to specify correct `maxPrecision` that matches existing schema.
For example, when schema has tags `_g1` ... `_g5`, then `5` must be passed as `maxPrecision`. 

### Function `boxFilter`

Filters records by lat/lon box. Unlike `gridFilter()`, this is a strict filter.
Must be used after `toRows()` because it requires `lat` and `lon` columns in input row set.

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.toRows()
  |> geo.boxFilter(box: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
``` 

It makes a lot of sense to use it together with `griFiflter()`.

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

Grouping levels (corresponds to geohash precision) - cell width x height
- `1` - 5000 x 5000 km
- `2` - 1250 x 625 km
- `3` - 156 x 156 km
- `4` - 39.1 x 19.5 km
- `5` - 4.89 x 4.89 km
- `6` - 1.22 x 0.61 km
- `7` - 153 x 153 m
- `8` - 38.2 x 19.1 m
- `9` - 4.77 x 4.77 m
- `10` - 1.19 x 0.596 m
- `11` - 149 x 149 mm
- `12` - 37.2 x 18.6 mm

Groups rows by area blocks of size specified by geohash `precision`.
Result is grouped by `newColumn`.
Parameter `maxPrecisionIndex` specifies finest precision geohash tag available in the input.

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.gridFilter(box: {minLat: 40.51757813, maxLat: 40.86914063, minLon: -73.65234375, maxLon: -72.94921875})
  |> geo.toRows()
  |> geo.groupByArea(newColumn: "gx", precision: 3, maxPrecisionIndex: 5)
```

#### Function definition

```
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

Drops geohash indexes columns (`_gX` by default) except those specified.
It will fail if input tables are grouped by any of them.

Example:
```
from(bucket: "rides")
  |> range(start: 2019-11-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "bike")
  |> geo.toRows()
  |> geo.groupByArea(newColumn: "gx", precision: 5, maxPrecisionIndex: 5)
  |> geo.stripMeta(except: ["_g5"])
```

#### Function definition

```
stripMeta = (tables=<-, pattern=/_g\d+/, except=[]) =>
  tables
    |> drop(fn: (column) => column =~ pattern and (length(arr: except) == 0 or not contains(value: column, set: except)))
```