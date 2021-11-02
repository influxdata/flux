testcase isType {
    assert.equal(isType("a", "string"), true)
    assert.equal(isType("a", "strin"), false)
    assert.equal(isType("a", "int"), false)
    assert.equal(isType({}, "int"), false)
    assert.equal(isType(1, "int"), true)
    assert.equal(isType(2030-01-01T00:00:00Z, "timestamp"), true)
    assert.equal(isType(2030-01-01T00:00:00Z, "int"), false)
}
