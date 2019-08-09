package v1_env_test

import "testing"
import "env"

getStringWithNoOsValueGot = env.getString(name: "ThisEnvironmentVariableWillNeverExist")
getStringWithNoOsValueWant = ""

testing.assertEquals(
  name: "getStringWithNoOsValue",
  got: getStringWithNoOsValueGot,
  want: getStringWithNoOsValueWant
)

//
// How do I set this for the integration test in a safe way?

//getStringWorkingGot = env.getString(name: "HI")
//getStringWorkingWant = "Hello, world!"
//
//testing.assertEquals(
//  name: "getStringWorking",
//  got: getStringWorkingGot,
//  want: getStringWorkingWant
//)
