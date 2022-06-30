package pkga_test


import "testing"
import "array"

option testing.tags = ["foo"]

testcase bar {
    array.from(rows: [{}])
}
