package bakelite

import (
	"bytes"
	"os"
	"testing"

	sqlittledb "github.com/alicebob/sqlittle/db"
)

func TestEmpty(t *testing.T) {
	b := &bytes.Buffer{}
	ok(t, writeHeader(b, 0))

	dir := t.TempDir()
	file := dir + "/empty.sqlite"
	ok(t, os.WriteFile(file, b.Bytes(), 0666))

	f, err := os.Stat(file)
	ok(t, err)
	eq(t, 100, f.Size())

	db, err := sqlittledb.OpenFile(file)
	ok(t, err)
	// db := must(t, tuple(sqlittledb.OpenFile(file))
	mustEq(t, []string{}, tuple(db.Tables()))
}
