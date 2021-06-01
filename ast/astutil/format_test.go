package astutil_test

import (
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/ast/astutil"
	"github.com/influxdata/flux/parser"
)

func TestFormat(t *testing.T) {
	src := `x=1+2`
	pkg := parser.ParseSource(src)
	if ast.Check(pkg) > 0 {
		t.Fatalf("unexpected error: %s", ast.GetError(pkg))
	} else if len(pkg.Files) != 1 {
		t.Fatalf("expected one file in the package, got %d", len(pkg.Files))
	}

	got, err := astutil.Format(pkg.Files[0])
	if err != nil {
		t.Fatal(err)
	}

	if want := `x = 1 + 2`; want != got {
		t.Errorf("unexpected formatted file -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}

func TestFormatWithCommentsBase(t *testing.T) {
	src := `// add two numbers

x=1+2`
	pkg := parser.ParseSource(src)
	if ast.Check(pkg) > 0 {
		t.Fatalf("unexpected error: %s", ast.GetError(pkg))
	} else if len(pkg.Files) != 1 {
		t.Fatalf("expected one file in the package, got %d", len(pkg.Files))
	}

	got, err := astutil.Format(pkg.Files[0])
	if err != nil {
		t.Fatal(err)
	}

	if want := `// add two numbers
x = 1 + 2`; want != got {
		t.Errorf("unexpected formatted file -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}

func TestFormatWithCommentsDict(t *testing.T) {
	src := `[
    "a": 0,
    //comment
    "b": 1,
    ]`
	pkg := parser.ParseSource(src)
	if ast.Check(pkg) > 0 {
		t.Fatalf("unexpected error: %s", ast.GetError(pkg))
	} else if len(pkg.Files) != 1 {
		t.Fatalf("expected one file in the package, got %d", len(pkg.Files))
	}

	got, err := astutil.Format(pkg.Files[0])
	if err != nil {
		t.Fatal(err)
	}

	if want := `[
    "a": 0,
    //comment
    "b": 1,
]`; want != got {
		t.Errorf("unexpected formatted file -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}

func TestFormatWithCommentsParens(t *testing.T) {
	src := `// comment\n(1 * 1)`
	pkg := parser.ParseSource(src)
	if ast.Check(pkg) > 0 {
		t.Fatalf("unexpected error: %s", ast.GetError(pkg))
	} else if len(pkg.Files) != 1 {
		t.Fatalf("expected one file in the package, got %d", len(pkg.Files))
	}

	got, err := astutil.Format(pkg.Files[0])
	if err != nil {
		t.Fatal(err)
	}

	if want := `// comment\n(1 * 1)`; want != got {
		t.Errorf("unexpected formatted file -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}

func TestFormatWithCommentsColon(t *testing.T) {
	src := `// Comment
    builtin foo
    // colon comment
    : int`
	pkg := parser.ParseSource(src)
	if ast.Check(pkg) > 0 {
		t.Fatalf("unexpected error: %s", ast.GetError(pkg))
	} else if len(pkg.Files) != 1 {
		t.Fatalf("expected one file in the package, got %d", len(pkg.Files))
	}

	got, err := astutil.Format(pkg.Files[0])
	if err != nil {
		t.Fatal(err)
	}

	if want := `// Comment
builtin foo
    // colon comment
    : int`; want != got {
		t.Errorf("unexpected formatted file -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}

func TestFormatWithCommentsUnaryExpressions(t *testing.T) {
	src := `// define a
    a = 5.0
    // eval this
    10.0 * -a == -0.5
        // or this
        or a == 6.0`
	pkg := parser.ParseSource(src)
	if ast.Check(pkg) > 0 {
		t.Fatalf("unexpected error: %s", ast.GetError(pkg))
	} else if len(pkg.Files) != 1 {
		t.Fatalf("expected one file in the package, got %d", len(pkg.Files))
	}

	got, err := astutil.Format(pkg.Files[0])
	if err != nil {
		t.Fatal(err)
	}

	if want := `// define a
a = 5.0

// eval this
10.0 * (-a) == -0.5 
    // or this
    or a == 6.0`; want != got {
		t.Errorf("unexpected formatted file -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}

func TestFormatWithCommentsBuiltin(t *testing.T) {
	src := `foo = 1

    foo
    
    builtin bar : int
    builtin rab : int
    
    // comment
    builtin baz : int`
	pkg := parser.ParseSource(src)
	if ast.Check(pkg) > 0 {
		t.Fatalf("unexpected error: %s", ast.GetError(pkg))
	} else if len(pkg.Files) != 1 {
		t.Fatalf("expected one file in the package, got %d", len(pkg.Files))
	}

	got, err := astutil.Format(pkg.Files[0])
	if err != nil {
		t.Fatal(err)
	}

	if want := `foo = 1

foo

builtin bar : int
builtin rab : int

// comment
builtin baz : int`; want != got {
		t.Errorf("unexpected formatted file -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}

func TestFormatWithTestCaseStmt(t *testing.T) {
	src := `testcase my_test { a = 1 }`
	pkg := parser.ParseSource(src)
	if ast.Check(pkg) > 0 {
		t.Fatalf("unexpected error: %s", ast.GetError(pkg))
	} else if len(pkg.Files) != 1 {
		t.Fatalf("expected one file in the package, got %d", len(pkg.Files))
	}

	got, err := astutil.Format(pkg.Files[0])
	if err != nil {
		t.Fatal(err)
	}

	if want := "testcase my_test {\n    a = 1\n}"; want != got {
		t.Errorf("unexpected formatted file -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}

func TestFormatWithComments(t *testing.T) {
	src := `    // hi
    // there
    {_time: r._time, io_time: r._value, // this is the end
    }
    
    // minimal
    foo = (arg=[1, 2]) => (1)
    
    // left
    left = from(bucket: "test")
        |> range(start: 2018-05-22T19:53:00Z
        // i write too many comments
        , stop: 2018-05-22T19:55:00Z)
        // and put them in strange places
        |>  drop
    
            // this hurts my eyes
    (columns: ["_start", "_stop"])
            // just terrible
        |> filter(fn: (r) =>
            (r.user 
    
            // (don't fire me, this is intentional)
            == "user1"))
        |> group(by
        // strange place for a comment
    : ["user"])
    
    right = from(bucket: "test")
        |> range(start: 2018-05-22T19:53:00Z,
                // please stop
                stop: 2018-05-22T19:55:00Z)
        |> drop( // spare me the pain
    // this hurts
    columns: ["_start", "_stop"// what
    ])
        |> filter(
            // just why
            fn: (r) =>
            // user 2 is the best user
            (r.user == "user2"))
        |> group(by: //just stop
    ["_measurement"])
    
    join(tables: {left: left, right: right}, on: ["_time", "_measurement"])
    
    from(bucket, _option // friends
    ,// stick together
    )
    
    i = // definitely
    not true
    // a
    // list
    // of
    // comments
    
    j
    
    // not lost`
	pkg := parser.ParseSource(src)
	if ast.Check(pkg) > 0 {
		t.Fatalf("unexpected error: %s", ast.GetError(pkg))
	} else if len(pkg.Files) != 1 {
		t.Fatalf("expected one file in the package, got %d", len(pkg.Files))
	}

	got, err := astutil.Format(pkg.Files[0])
	if err != nil {
		t.Fatal(err)
	}

	if want := `// hi
// there
{
    _time: r._time,
    io_time: r._value,
// this is the end
}

// minimal
foo = (arg=[1, 2]) => 1

// left
left = from(bucket: "test")
    |> range(
        start: 2018-05-22T19:53:00Z
            // i write too many comments
            ,
            stop: 2018-05-22T19:55:00Z,
        )
    // and put them in strange places
    |> drop
        // this hurts my eyes
        (columns: ["_start", "_stop"])
    // just terrible
    |> filter(
        fn: (r) => r.user 
            // (don't fire me, this is intentional)
            == "user1",
    )
    |> group(
        by
            // strange place for a comment
            : ["user"],
    )

right = from(bucket: "test")
    |> range(
        start: 2018-05-22T19:53:00Z,
        // please stop
        stop: 2018-05-22T19:55:00Z,
    )
    |> drop(
        // spare me the pain
        // this hurts
        columns: [
            "_start",
            "_stop",
        // what
        ],
    )
    |> filter(
        // just why
        fn: (r) => 
            // user 2 is the best user
            (r.user == "user2"),
    )
    |> group(
        by: 
            //just stop
            ["_measurement"],
    )

join(tables: {left: left, right: right}, on: ["_time", "_measurement"])

from(bucket, _option
    // friends
    // stick together
)

i = 
    // definitely
    not true

// a
// list
// of
// comments
j
// not lost`; want != got {
		t.Errorf("unexpected formatted file -want/+got:\n\t- %q\n\t+ %q", want, got)
	}
}
