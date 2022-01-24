// Package multirow provides additional functions for remapping values in rows.
//
package multirow


// map is an alternate implementation of `map()` that is more functional.
//
// `multirow.map()` can modify all columns and groups keys same standard map function, but additional you can:
// *  change the number of rows of tables both up and down (you can filter row by returns empty array or you can make unpivot operation by returns array of rows for each column)
// *  allows you to use aggregate functions to calculate the value of a new row based on the values of adjacent rows
// *  allows you to calculate the new row value based on the accumulator value (previous calculation)
// *  allows you to remove extra(virtual columns of accumulator) columns from the result calculation
// *  allows you to get the current line number and the total number of lines in this group
// *  use scalar value as row (default new column name _value can change by column param)
// ## Parameters
// - tables: Input tables streams for transformation
// - fn: A single argument function to apply to each record.
//   the return value must be: record, array of record, primitive type, tables stream.
//   This function has many parameters, but they are all optional:
//    - index: current row index in current group as int
//    - count: row count in current group as int
//    - row: current process row as record
//    - window: table with rows [index - left: index + right]  as table stream
//    - previous: previous process result in same group as record
// - left: row count or _time duration of records before current for add to window stream
//   default: 0
// - right: row count or _time duration of records after current for add to window stream
//   default: 0
// - column: Name of new column for all primitive results of fn
//   default: "_value"
// - init:  Record then will pass to previous param of first row
//   default: {}
// - virtual: Array of string with virtual column names, than used only for intermediate calculations and should not be included in the final result
//   default: []
// ## Example
// import "csv"
// import "contrib/lazarenkovegor/multirow"
// data =
//     "
//   #datatype,string,long,string,string,long
//   #group,false,false,false,false,false
//   #default,_result,0,,,
//   ,result,table,strcol0,strcol1,intcol3
//   ,,,test1,test10,1
//   ,,,test1,test11,
//   ,,,test2,test12,3
//   ,,,test2,test13,4
//   "
//
// csv.from(csv: data)
//     |> multirow.map(
//         fn: (previous, row) => {
//             x = previous.x_col * 2 - 1
//
//             return {row with concat: (if exists previous.concat then previous.concat + "," else "") + row.strcol1,
//                 x_col: x,
//                 val: x % 100,
//             }
//         },
//         init: {x_col: 100},
//         virtual: ["x_col"],
//     )
builtin map : (
        <-tables: [A],
        ?left: E,
        ?right: F,
        ?init: C,
        ?virtual: [string],
        fn: (
            ?index: int,
            ?count: int,
            ?row: A,
            ?window: [A],
            ?previous: D,
        ) => X,
        ?column: string,
        ?limit: int,
    ) => [B]
    where
    A: Record,
    B: Record,
    C: Record,
    D: Record
