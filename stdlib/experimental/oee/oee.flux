package oee

import "contrib/tomhollingworth/events"
import "experimental"

computeAPQ = (
    productionEvents,
    partEvents,
    runningState,
    plannedTime,
    idealCycleTime
) => {
    availability = productionEvents
        |> events.duration(unit: 1ns, columnName: "runTime")
        |> filter(fn: (r) => r.state == runningState)
        |> sum(column: "runTime")
        |> duplicate(column: "_stop", as: "_time")
        |> map(fn: (r) => ({ r with availability: float(v: r.runTime) / float(v: int(v: plannedTime)) }))
    totalCount = partEvents
        |> difference(columns: ["partCount"], nonNegative: true)
        |> sum(column: "partCount")
        |> duplicate(column: "_stop", as: "_time")
    badCount = partEvents
        |> difference(columns: ["badCount"], nonNegative: true)
        |> sum(column: "badCount")
        |> duplicate(column: "_stop", as: "_time")
    performance = experimental.join(left: availability, right: totalCount, fn: (left, right) => ({left with
        performance: float(v: right.partCount) * float(v: int(v: idealCycleTime)) / float(v: left.runTime)
    }))
    quality = experimental.join(left: badCount, right: totalCount, fn: (left, right) => ({left with
            quality: (float(v: right.partCount) - float(v:left.badCount)) / float(v: right.partCount)
    }))

    return experimental.join(left:performance, right: quality, fn: (left, right) => ({left with
        quality: right.quality,
        oee: left.availability * left.performance * right.quality
    }))
}

APQ = (
    tables=<-,
    runningState,
    plannedTime,
    idealCycleTime
) =>
    computeAPQ(productionEvents: tables, partEvents: tables, runningState: runningState, plannedTime: plannedTime, idealCycleTime: idealCycleTime)
