// Package expect includes functions to mark
// any expectations for a testcase to be satisfied
// before the testcase finishes running.
//
// These functions are intended to be called at the
// beginning of a testcase, but it doesn't really
// matter when they get invoked within the testcase.
package expect


// planner will cause the present testcase to
// expect the given planner rules will be invoked
// exactly as many times as the number given.
//
// The key is the name of the planner rule.
builtin planner : (rules: [string:int]) => {}
