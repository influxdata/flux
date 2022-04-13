package helpers_test


import "testing"
import "contrib/mhall119/helpers"

option testData = {testValue: 125}

t_yieldValue = (v) =>
    helpers.yieldValue(v:v)

test _yieldValue = () =>
    ({input: testData.testValue, want: "", fn: t_yieldValue})