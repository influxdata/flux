package oee


import "contrib/tomhollingworth/events"
import "experimental"

// computeAPQ computes availability, performance, quality and overall equipment effectiveness (oee).
// productionEvents - a stream of start/stop events for the production process. Each row contains
//   a _time and state that indicates start and stop events.
// partEvents - a stream of part counts. Each row contains cumulative counts where column partCount
//   represents total number of produced parts and badCount number of parts that did not meet quality standards.
// runningState - production event or state value that corresponds to equipment running state
// plannedTime - total time that equipment is expected to produce
// idealCycleTime - theoretical minimum time to produce one part
computeAPQ = (
        productionEvents,
        partEvents,
        runningState,
        plannedTime,
        idealCycleTime,
) => {
    availability = productionEvents
        |> events.duration(unit: 1ns, columnName: "runTime")
        |> filter(fn: (r) => r.state == runningState)
        |> sum(column: "runTime")
        |> map(fn: (r) => ({r with _time: r._stop, availability: float(v: r.runTime) / float(v: int(v: plannedTime))}))
    totalCount = partEvents
        |> difference(columns: ["partCount"], nonNegative: true)
        |> sum(column: "partCount")
        |> duplicate(column: "_stop", as: "_time")
    badCount = partEvents
        |> difference(columns: ["badCount"], nonNegative: true)
        |> sum(column: "badCount")
        |> duplicate(column: "_stop", as: "_time")
    performance = experimental.join(
        left: availability,
        right: totalCount,
        fn: (left, right) => ({left with
            performance: float(v: right.partCount) * float(v: int(v: idealCycleTime)) / float(v: left.runTime),
        }),
    )
    quality = experimental.join(
        left: badCount,
        right: totalCount,
        fn: (left, right) => ({left with
            quality: (float(v: right.partCount) - float(v: left.badCount)) / float(v: right.partCount),
        }),
    )

    return experimental.join(
        left: performance,
        right: quality,
        fn: (left, right) => ({left with
            quality: right.quality,
            oee: left.availability * left.performance * right.quality,
        }),
    )
}

// APQ computes availability, performance, quality and overall equipment effectiveness (oee).
// Input tables are expected to have rows with _time, state, partCount and badCount columns, where
//   state that indicates start and stop events, partCount represents total number
//   of produced parts and badCount represents number of parts that did not meet quality standards.
// plannedTime - total time that equipment is expected to produce
// idealCycleTime - theoretical minimum time to produce one part
APQ = (tables=<-, runningState, plannedTime, idealCycleTime) => computeAPQ(
    productionEvents: tables,
    partEvents: tables,
    runningState: runningState,
    plannedTime: plannedTime,
    idealCycleTime: idealCycleTime,
)
