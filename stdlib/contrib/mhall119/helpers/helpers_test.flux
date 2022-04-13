package helpers_test


import "testing"
import "contrib/mhall119/helpers"

option testData = {testValue: 125}

// Test yielding a single scalar value
t_yieldValue = (v) =>
    helpers.yieldValue(v:v)

test _yieldValue = () =>
    ({input: testData.testValue, want: "", fn: t_yieldValue})

// Test yielding a Flux object
t_yieldObject = (o) =>
    helpers.yieldObject(o:o)

test _yieldObject = () =>
    ({input: testData, want: "", fn: t_yieldObject})

// Test yielding a Flux object as JSON
t_yieldJSON = (o) =>
    helpers.yieldJSON(o:o)

test _yieldJSON = () =>
    ({input: testData, want: "", fn: t_yieldJSON})

    