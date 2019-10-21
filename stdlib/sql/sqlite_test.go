package sql

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	flux "github.com/influxdata/flux"
	"github.com/influxdata/flux/values"
	_ "github.com/mattn/go-sqlite3"
)

func TestBoolTranslation(t *testing.T) {
	// write values to "DB" - as would be done by sql.to() - using sqlite3 as it allows maximum levels of type abuse, and therefore likely the largest
	// potential for issues
	db, err := sql.Open("sqlite3", ":memory:")
	if !cmp.Equal(err, nil) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
	}

	// create table - column type can be anything
	q := fmt.Sprintf("CREATE TABLE IF NOT EXISTS bools (name TEXT, age INT, employed BOOL)")
	_, err = db.Exec(q)
	if !cmp.Equal(err, nil) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
	}

	// insert first row - BOOL as BOOL, which sqlite WILL accept as a write - and silently store true as 1
	q = fmt.Sprintf("INSERT INTO bools (name, age, employed) VALUES (\"albert\",110,true)")
	_, err = db.Exec(q)
	if !cmp.Equal(err, nil) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
	}

	// insert second row - BOOL as INT - again, will accept ok
	q = fmt.Sprintf("INSERT INTO bools (name, age, employed) VALUES (\"mary\",10,1)")
	_, err = db.Exec(q)
	if !cmp.Equal(err, nil) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
	}

	q = fmt.Sprintf("SELECT * FROM bools")
	results, err := db.Query(q)
	if !cmp.Equal(err, nil) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
	}

	// read the data back out and check that Flux fails here - we should not be doing magic type casts
	TestReader := &SqliteRowReader{
		Cursor: results,
	}

	TestReader.columnNames = []string{"name", "age", "employed"}
	TestReader.columnTypes = []flux.ColType{flux.TString, flux.TInt, flux.TInt}

	// all rows should fail to parse into Flux types becuase there is no SQLite boolean type, regardless of what the column "type" is
	for TestReader.Next() {
		row, _ := TestReader.GetNextRow()
		// fail as BOOL
		if cmp.Equal(values.NewBool(true), row[2]) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(values.NewBool(true), row[2]))
		}
		// succeed as INT, which is what it actually is
		if !cmp.Equal(values.NewInt(1), row[2]) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(values.NewInt(1), row[2]))
		}

	}

}

func TestNulTranslation(t *testing.T) {
	// write values to "DB" - as would be done by sql.to() - using sqlite3 as it allows maximum levels of type abuse, and therefore likely the largest
	// potential for issues
	db, err := sql.Open("sqlite3", ":memory:")
	if !cmp.Equal(err, nil) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
	}

	// create table
	q := fmt.Sprintf("CREATE TABLE IF NOT EXISTS magic (name TEXT, age INT, employed BADINTBOOL)")
	_, err = db.Exec(q)
	if !cmp.Equal(err, nil) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
	}

	// insert first row - null string
	q = fmt.Sprintf("INSERT INTO magic (age, employed) VALUES (11,true)")
	_, err = db.Exec(q)
	if !cmp.Equal(err, nil) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
	}

	// insert second row - null int
	q = fmt.Sprintf("INSERT INTO magic (name, employed) VALUES (\"mary\",true)")
	_, err = db.Exec(q)
	if !cmp.Equal(err, nil) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
	}

	// insert third row - null bool
	q = fmt.Sprintf("INSERT INTO magic (name, age) VALUES (\"casper\",10)")
	_, err = db.Exec(q)
	if !cmp.Equal(err, nil) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
	}

	q = fmt.Sprintf("SELECT * FROM magic")
	results, err := db.Query(q)
	if !cmp.Equal(err, nil) {
		t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
	}

	// read the data back out and check
	TestReader := &SqliteRowReader{
		Cursor: results,
	}

	TestReader.columnNames = []string{"name", "age", "employed"}
	TestReader.columnTypes = []flux.ColType{flux.TString, flux.TInt, flux.TBool}

	// number of columns == number of rows - and Nill goes left -> right. otherwise the following loop will fail
	i := 0
	for TestReader.Next() {
		row, err := TestReader.GetNextRow()
		// none should throw errors - expect them all to handle Nulls from DB correctly
		if !cmp.Equal(err, nil) {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(nil, err))
		}
		if !row[i].IsNull() {
			t.Fatalf("unexpected result -want/+got\n\n%s\n\n", cmp.Diff(true, row[i].IsNull()))
		}
		i++
	}

}
