package bakelite

import (
	"bytes"
	"strconv"
	"strings"
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

func TestEmptyTable(t *testing.T) {
	// single, empty, table
	db := New()
	ok(t, db.Add("hello", []string{"planet"}, nil)) // no data

	b := &bytes.Buffer{}
	ok(t, db.Write(b))
	file := saveFile(t, b, "emptytable.sqlite")

	sqlite(t, file, ".tables", "hello\n")
	sqlite(t, file, "SELECT * FROM hello", "")
}

func TestAFewRows(t *testing.T) {
	// single table, few rows
	db := New()
	ok(t, db.Add("planets", []string{"name", "moons"}, [][]any{
		{"Mercury", 0},
		{"Venus", 0},
		{"Earth", 1},
		{"Mars", 2},
		{"Jupiter", 80},
		{"Saturn", 83},
		{"Uranus", 27},
		{"Neptune", 4},
	}))

	b := &bytes.Buffer{}
	ok(t, db.Write(b))
	file := saveFile(t, b, "afewrows.sqlite")

	sqlite(t, file, ".tables", "planets\n")
	sqlite(t, file, "SELECT name FROM planets", "Mercury\nVenus\nEarth\nMars\nJupiter\nSaturn\nUranus\nNeptune\n")
	sqlite(t, file, "SELECT name FROM planets ORDER BY moons", "Mercury\nVenus\nEarth\nMars\nNeptune\nUranus\nJupiter\nSaturn\n")
}

func TestOverflow(t *testing.T) {
	// single table, very long rows.
	db := New()
	ok(t, db.Add("planets", []string{"name", "private_key"}, [][]any{
		{"Mercury", strings.Repeat("a", 10_000)},
		{"Venus", strings.Repeat("b", 59)},
		{"Earth", strings.Repeat("c", 40_000)},
		{"Mars", strings.Repeat("d", 123_456)},
		{"Jupiter", strings.Repeat("e", 1)},
		{"Saturn", strings.Repeat("f", 12)},
		{"Uranus", strings.Repeat("g", 8_000)},
		{"Neptune", strings.Repeat("h", 75_000)},
	}))

	b := &bytes.Buffer{}
	ok(t, db.Write(b))
	file := saveFile(t, b, "overflow.sqlite")

	sqlite(t, file, ".tables", "planets\n")
	sqlite(t, file, "SELECT name FROM planets", "Mercury\nVenus\nEarth\nMars\nJupiter\nSaturn\nUranus\nNeptune\n")
	sqlite(t, file, "SELECT name, length(private_key) FROM planets", "Mercury|10000\nVenus|59\nEarth|40000\nMars|123456\nJupiter|1\nSaturn|12\nUranus|8000\nNeptune|75000\n")
}

func TestUpdates(t *testing.T) {
	// two tables, and then update them, to see what sqlite thinks of it.
	db := New()
	ok(t, db.Add("planets", []string{"name", "moons"}, [][]any{
		{"Mercury", 0},
		{"Venus", 0},
		{"Earth", 1},
		{"Mars", 2},
	}))
	ok(t, db.Add("colors", []string{"name", "r", "g", "b"}, [][]any{
		{"white", 0, 0, 0},
		{"black", 256, 256, 256},
		{"red", 256, 0, 0},
	}))

	b := &bytes.Buffer{}
	ok(t, db.Write(b))
	file := saveFile(t, b, "updates.sqlite")

	sqlite(t, file, ".tables", "colors   planets\n")
	sqlite(t, file, "SELECT name FROM planets", "Mercury\nVenus\nEarth\nMars\n")
	sqlite(t, file, "SELECT name FROM colors", "white\nblack\nred\n")

	sqlite(t, file, "INSERT INTO colors (name, r, g, b) VALUES ('sky blue', 135, 206, 235)", "")
	sqlite(t, file, "SELECT name FROM colors ORDER BY r DESC", "black\nred\nsky blue\nwhite\n")
	sqlite(t, file, "DELETE FROM colors WHERE r < 200", "")
	sqlite(t, file, "SELECT name FROM colors", "black\nred\n")
}

func TestMany(t *testing.T) {
	// enough rows to have multiple levels of interior pages
	db := New()
	var rows [][]any
	// 128_000 uses a single level of interior pages, but that's no fun
	for i := 0; i < 200_000; i++ {
		rows = append(rows, []any{"r" + strconv.Itoa(i)})
	}

	ok(t, db.Add("counts", []string{"count"}, rows))

	b := &bytes.Buffer{}
	ok(t, db.Write(b))
	file := saveFile(t, b, "many.sqlite")

	sqlite(t, file, ".tables", "counts\n")
	sqlite(t, file, "SELECT count(*) FROM counts", "200000\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=1", "r1\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=42", "r42\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=100", "r100\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=1000", "r1000\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=10000", "r10000\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=100000", "r100000\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=199999", "r199999\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=200000", "")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=200001", "")
}
