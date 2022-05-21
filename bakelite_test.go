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

func TestEmptyTable(t *testing.T) {
	// single, empty, table
	db := New()
	ok(t, db.Add("hello", []string{"planet"}, nil)) // no data

	b := &bytes.Buffer{}
	ok(t, db.Write(b))
	file := saveFile(t, b, "emptytable.sqlite")

	sqlite(t, file, ".tables", "hello\n")
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
