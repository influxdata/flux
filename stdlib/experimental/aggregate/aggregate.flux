package aggregate

import "experimental"

rate = (tables=<-, every, groupColumns=[], unit=1s) =>
    tables
        |> derivative(nonNegative:true, unit:unit)
        |> aggregateWindow(every: every, fn : (tables=<-, column) =>
            tables
                |> mean(column: column)
                |> group(columns: groupColumns)
                |> experimental.group(columns: ["_start", "_stop"], mode:"extend")
                |> sum()
        )
