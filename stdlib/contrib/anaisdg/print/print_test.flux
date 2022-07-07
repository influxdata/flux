package print_test

import "testing"
import "array"
import "contrib/anaisdg/print"

testcase printIntTest {
    got =
        print.print(val:2, result_name:"Int")
    want =
        array.from(
            rows: [
                {"_value": "2"}
            ]
        )
    
    testing.diff(got, want)
}

testcase printFoatTest {
    got =
        print.print(val:2.0, result_name:"Float")
    want =
        array.from(
            rows: [
                {"_value": "2"}
            ]
        )
    
    testing.diff(got, want)
}

testcase printStringTest {
    got =
        print.print(val:"2", result_name:"String")
    want =
        array.from(
            rows: [
                {"_value": "2"}
            ]
        )
    
    testing.diff(got, want)
}

testcase printBoolTest {
    got =
        print.print(val:true, result_name:"Bool")
    want =
        array.from(
            rows: [
                {"_value": "true"}
            ]
        )
    
    testing.diff(got, want)
}

testcase printJSONTest {
    got =
        print.print(val:[{val:2.0}], result_name:"JSON")
    want =
        array.from(
            rows: [
                {"_value": "[{val: 2}]"}
            ]
        )
    
    testing.diff(got, want)
}
