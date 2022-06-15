// Package oee provides functions for calculating overall equipment effectiveness (OEE).
//
// ## Metadata
// introduced: 0.112.0
//
package oee


import "contrib/tomhollingworth/events"
import "experimental"

// computeAPQ computes availability, performance, and quality (APQ)
// and overall equipment effectiveness (OEE) using two separate input streams:
// **production events** and **part events**.
//
// ## Output schema
// For each input table, `oee.computeAPQ` outputs a table with a single row and
// the following columns:
//
// - **_time**: Timestamp associated with the APQ calculation.
// - **availability**: Ratio of time production was in a running state.
// - **oee**: Overall equipment effectiveness.
// - **performance**: Ratio of production efficiency.
// - **quality**: Ratio of production quality.
// - **runTime**: Total nanoseconds spent in the running state.
//
// ## Parameters
//
// - productionEvents: Production events stream that contains the production
//   state or start and stop events.
//
//     Each row must contain the following columns:
//
//     - **_stop**: Right time boundary timestamp (typically assigned by `range()` or `window()`).
//     - **_time**: Timestamp of the production event.
//     - **state**: String that represents start or stop events or the production state.
//
//     Use [`runningState`](#runningstate) to specify which value in the `state`
//     column represents a running state.
//
// - partEvents: Part events that contains the running totals of parts produced and
//     parts that do not meet quality standards.
//
//     Each row must contain the following columns:
//
//     - **_stop**: Right time boundary timestamp (typically assigned by
//     `range()` or `window()`).
//     - **_time**: Timestamp of the parts event.
//     - **partCount:** Cumulative total of parts produced.
//     - **badCount** Cumulative total of parts that do not meet quality standards.
//
// - runningState: State value that represents a running state.
// - plannedTime: Total time that equipment is expected to produce parts.
// - idealCycleTime: Ideal minimum time to produce one part.
//
// ## Examples
// ```
// import "array"
// import "experimental/oee"
//
// productionData = array.from(
//     rows: [
//         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T00:00:00Z, state: "running"},
//         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T04:03:53Z, state: "stopped"},
//         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T04:24:53Z, state: "running"},
//         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T08:00:00Z, state: "running"}
//     ],
// )
//     |> group(columns: ["_start", "_stop"])
//
// partsData = array.from(
//     rows: [
//         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T00:00:00Z, partCount: 0, badCount: 0},
//         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T04:03:53Z, partCount: 673, badCount: 5},
//         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T04:24:53Z, partCount: 673, badCount: 5},
//         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T08:00:00Z, partCount: 1298, badCount: 13},
//     ],
// )
//     |> group(columns: ["_start", "_stop"])
//
// oee.computeAPQ(
//     productionEvents: productionData,
//     partEvents: partsData,
//     runningState: "running",
//     plannedTime: 8h,
//     idealCycleTime: 21s,
// )
// >     |> drop(columns: ["_start", "_stop"])
// ```
//
// ## Metadata
// tags: transformations,aggregates
//
computeAPQ = (
    productionEvents,
    partEvents,
    runningState,
    plannedTime,
    idealCycleTime,
) =>
{
    availability =
        productionEvents
            |> events.duration(unit: 1ns, columnName: "runTime")
            |> filter(fn: (r) => r.state == runningState)
            |> sum(column: "runTime")
            |> map(
                fn: (r) => ({r with _time: r._stop, availability: float(v: r.runTime) / float(v: int(v: plannedTime))}),
            )
    totalCount =
        partEvents
            |> difference(columns: ["partCount"], nonNegative: true)
            |> sum(column: "partCount")
            |> duplicate(column: "_stop", as: "_time")
    badCount =
        partEvents
            |> difference(columns: ["badCount"], nonNegative: true)
            |> sum(column: "badCount")
            |> duplicate(column: "_stop", as: "_time")
    performance =
        experimental.join(
            left: availability,
            right: totalCount,
            fn: (left, right) =>
                ({left with performance:
                        float(v: right.partCount) * float(v: int(v: idealCycleTime)) / float(v: left.runTime),
                }),
        )
    quality =
        experimental.join(
            left: badCount,
            right: totalCount,
            fn: (left, right) =>
                ({left with quality: (float(v: right.partCount) - float(v: left.badCount)) / float(v: right.partCount),
                }),
        )

    return
        experimental.join(
            left: performance,
            right: quality,
            fn: (left, right) =>
                ({left with quality: right.quality, oee: left.availability * left.performance * right.quality}),
        )
}

// APQ computes availability, performance, quality (APQ) and overall equipment
// effectiveness (OEE) in producing parts.
//
// Provide the required input schema to ensure this function successfully calculates APQ and OEE.
//
// ### Required input schema
// Input tables must include the following columns:
//
// - **_stop**: Right time boundary timestamp (typically assigned by `range()` or `window()`).
// - **_time**: Timestamp of the production event.
// - **state**: String that represents start or stop events or the production state.
// - **partCount**: Cumulative total of parts produced.
// - **badCount**: Cumulative total of parts that do not meet quality standards.
//
// ### Output schema
// For each input table, `oee.APQ` outputs a table with a single row that includes the following columns:
//
// - **_time**: Timestamp associated with the APQ calculation.
// - **availability**: Ratio of time production was in a running state.
// - **oee**: Overall equipment effectiveness.
// - **performance**: Ratio of production efficiency.
// - **quality**: Ratio of production quality.
// - **runTime**: Total nanoseconds spent in the running state.
//
// ## Parameters
// - runningState: State value that represents a running state.
// - plannedTime: Total time that equipment is expected to produce parts.
// - idealCycleTime: Ideal minimum time to produce one part.
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
// ```
// # import "array"
// import "experimental/oee"
// #
// # productionData = array.from(
// #     rows: [
// #         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T00:00:00Z, state: "running", partCount: 0, badCount: 0},
// #         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T04:03:53Z, state: "stopped", partCount: 673, badCount: 5},
// #         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T04:24:53Z, state: "running", partCount: 673, badCount: 5},
// #         {_start: 2021-01-01T00:00:00Z, _stop: 2021-01-01T08:00:01Z, _time: 2021-01-01T08:00:00Z, state: "running", partCount: 1298, badCount: 13},
// #     ],
// # )
// #     |> group(columns: ["_start", "_stop"])
//
// < productionData
//     |> oee.APQ(runningState: "running", plannedTime: 8h, idealCycleTime: 21s)
// >     |> drop(columns: ["_start", "_stop"])
// ```
//
// ## Metadata
// tags: trasnformations,aggregates
//
APQ = (tables=<-, runningState, plannedTime, idealCycleTime) =>
    computeAPQ(
        productionEvents: tables,
        partEvents: tables,
        runningState: runningState,
        plannedTime: plannedTime,
        idealCycleTime: idealCycleTime,
    )
