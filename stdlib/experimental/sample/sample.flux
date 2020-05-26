package sample

import "csv"

// The sample package provides tools for generating and incrementing sample data
// and writing it to InfluxDB at regular intervals, typically in a task.

// SeedData takes annotated CSV and appends a _time column with the current time.
// Input CSV seed data should not include a _start column.
seedData = (seed) =>
  csv.from(csv: seed)
    |> map(fn: (r) => ({ r with _time: now() }))

// CheckForPreviousData takes a stream of tables and checks for a _start column
// in any of the tables. If the _start or _stop columns exists, it indicates the
// stream includes data returned from a query result and not just seed data.
// If it detects previous data, it returns true. If not, it returns false.
checkForPreviousData = (tables) => {
  existing_table = tables |> findColumn(fn: (key) => exists key._start or key._stop, column: "_start")
  isPresent = if length(arr: existing_table) == 0 then false else true
  return isPresent
}

// MultiplyDuration takes a duration value and multiplies it by x.
// d is a duration. x is a float.
// This function is used to ensure the query for previous data more than covers
// the interval at which batches are written.
multiplyDuration = (d,x) => duration(v: int(v: float(v: int(v: d)) * x))

// Generate takes seed data and previously written data, unions the streams together,
// filters the data based on whether or not previously written data is returned,
// increments the data using a map function, and returns the incremented data.
// bucket determines the bucket to query.
// seedCSV is raw annotated CSV used to seed the data set.
// incrementFn is a function that increments values in the data set.
// every should match the interval at which generate runs.
// every * 1.5 determines the query range to query the last batch of data.
// In the context of a task, every should be set to task.every
// predicate defines the predicate function used to query the previous batch.
generate = (bucket, seed, incrementFn, every, predicate=(r) => true) => {
  sData = seedData(seed: seed)
  qData = from(bucket: bucket)
    |> range(start: multiplyDuration(d: -every, x: 1.5))
    |> filter(fn: predicate)
  dataExists = checkForPreviousData(tables: qdata)
  updatedData = union(tables: [sData,qData])
    |> filter(fn: (r) => if dataExists then exists r._start else true)
    |> map(fn: (r) => ({ r with _time: now() }))
    |> map(fn: incrementFn )
  return updatedData
}
