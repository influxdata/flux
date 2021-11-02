package universe_test

import "testing"
import "testing/assert"

testcase isType {
    assert.equal(want: true, got: isType(v: "a", type: "string"))
    assert.equal(want: false, got: isType(v: "a", type: "strin"))
    assert.equal(want: false, got: isType(v: "a", type: "int"))
    assert.equal(want: false, got: isType(v: {}, type: "int"))
    assert.equal(want: true, got: isType(v: 1, type: "int"))
    assert.equal(want: true, got: isType(v: 2030-01-01T00:00:00Z, type: "timestamp"))
    assert.equal(want: false, got: isType(v: 2030-01-01T00:00:00Z, type: "int"))
}
