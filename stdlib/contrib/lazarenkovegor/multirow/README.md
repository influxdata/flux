# MultiRow Package

Use this Flux Package to perform more complex transformation operations than the standard map. 
This map allow:
- change row count, useful for filter on unpivot operations.
- use row index and total row count in current group for same operations as row_number, ntile, dense_rank.
- use window stream with current row and optional N rows left and N rows right from current for calc new row value with any stream trafsfrormations
- use preveuis value and init value for make cummulative trasfformations, same cummulative sum, moving average, smoothing
- use virtual columns to calculate new row values, but don't include the value of those columns in the result
## Contact
- Author: Egor Lazarenkov
- Email: lazarenkov.egor@gmail.com
