package bakelite

import (
	"bytes"
	"testing"
)

func TestEmpty(t *testing.T) {
	t.Skip("wip")
	db := New()
	ok(t, db.Add("hello", []string{"planet"}, nil)) // no data

	b := &bytes.Buffer{}
	ok(t, db.Write(b))
	file := saveFile(t, b, "empty.sqlite")

	/*
		db, err := sqlittledb.OpenFile(file)
		ok(t, err)
		// db := must(t, tuple(sqlittledb.OpenFile(file))
		mustEq(t, []string{}, tuple(db.Tables()))
	*/

	sqlite(t, file, ".tables", "planet")
}
