package bakelite

import (
	"bytes"
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
