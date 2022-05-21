package bakelite

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
	// sqlittledb "github.com/alicebob/sqlittle/db"
)

func TestNothing(t *testing.T) {
	t.Skip("wip")
	b := &bytes.Buffer{}
	ok(t, writeHeader(b, 0))

	file := saveFile(t, b, "nothing.sqlite")
	f, err := os.Stat(file)
	ok(t, err)
	eq(t, 100, f.Size()) // FIXME: should be 0

	/*
		db, err := sqlittledb.OpenFile(file)
		ok(t, err)
		// db := must(t, tuple(sqlittledb.OpenFile(file))
		mustEq(t, []string{}, tuple(db.Tables()))
	*/

	sqlite(t, file, ".tables", "")
}

// run the sqlite3 command. Can be SQL or command such as ".tables".
func sqlite(t *testing.T, file string, sql string, expected string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "sqlite3", "--batch", file)
	cmd.Stdin = strings.NewReader(sql)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	t.Logf("exec: %q\nout: %q\nerr: %q\n", sql, stdout.String(), stderr.String())
	ok(t, err)
	eq(t, expected, stdout.String())
	eq(t, "", stderr.String())
}
