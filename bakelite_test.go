package bakelite

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNoTables(t *testing.T) {
	db := New()

	b := &bytes.Buffer{}
	ok(t, db.WriteTo(b))
	file := saveFile(t, b, "nothing.sqlite")

	sqlite(t, file, ".tables", "")
}

func TestEmptyTable(t *testing.T) {
	// single, empty, table
	db := New()
	db.AddSlice("hello", []string{"planet"}, nil) // no data

	b := &bytes.Buffer{}
	ok(t, db.WriteTo(b))
	file := saveFile(t, b, "emptytable.sqlite")

	sqlite(t, file, ".tables", "hello\n")
	sqlite(t, file, "SELECT * FROM hello", "")
}

func TestAFewRows(t *testing.T) {
	// single table, few rows
	db := New()
	db.AddSlice("planets", []string{"name", "moons"}, [][]any{
		{"Mercury", 0},
		{"Venus", 0},
		{"Earth", 1},
		{"Mars", 2},
		{"Jupiter", 80},
		{"Saturn", 83},
		{"Uranus", 27},
		{"Neptune", 4},
	})

	b := &bytes.Buffer{}
	ok(t, db.WriteTo(b))
	file := saveFile(t, b, "afewrows.sqlite")

	sqlite(t, file, ".tables", "planets\n")
	sqlite(t, file, "SELECT name FROM planets", "Mercury\nVenus\nEarth\nMars\nJupiter\nSaturn\nUranus\nNeptune\n")
	sqlite(t, file, "SELECT name FROM planets ORDER BY moons", "Mercury\nVenus\nEarth\nMars\nNeptune\nUranus\nJupiter\nSaturn\n")
}

func TestValues(t *testing.T) {
	db := New()
	db.AddSlice("ints", []string{"value"}, [][]any{
		{-2147483649},
		{-2147483648},
		{-32769},
		{-32768},
		{-129},
		{-128},
		{-1},
		{0},
		{1},
		{2},
		{127},
		{128},
		{32767},
		{32768},
		{2147483647},
		{2147483648},
	})
	db.AddSlice("floats", []string{"value"}, [][]any{
		{31415926535.89},
		{3.1415},
		{0.0},
		{-0.0},
		{-3.1415},
	})
	db.AddSlice("bytes", []string{"value"}, [][]any{
		{[]byte("foo")},
		{[]byte("bar")},
		{"hello"},
	})

	b := &bytes.Buffer{}
	ok(t, db.WriteTo(b))
	file := saveFile(t, b, "values.sqlite")

	sqlite(t, file, ".tables", "bytes   floats  ints  \n")
	sqlite(t, file, "SELECT value FROM ints ORDER BY value",
		`-2147483649
-2147483648
-32769
-32768
-129
-128
-1
0
1
2
127
128
32767
32768
2147483647
2147483648
`,
	)

	sqlite(t, file, "SELECT value FROM floats ORDER BY value",
		`-3.1415
0.0
0.0
3.1415
31415926535.89
`,
	)

	sqlite(t, file, "SELECT value FROM bytes ORDER BY value",
		`hello
bar
foo
`,
	)
}

func TestOverflow(t *testing.T) {
	// single table, very long rows.
	db := New()
	db.AddSlice("planets", []string{"name", "private_key"}, [][]any{
		{"Mercury", strings.Repeat("a", 10_000)},
		{"Venus", strings.Repeat("b", 59)},
		{"Earth", strings.Repeat("c", 40_000)},
		{"Mars", strings.Repeat("d", 123_456)},
		{"Jupiter", strings.Repeat("e", 1)},
		{"Saturn", strings.Repeat("f", 12)},
		{"Uranus", strings.Repeat("g", 8_000)},
		{"Neptune", strings.Repeat("h", 75_000)},
	})

	b := &bytes.Buffer{}
	ok(t, db.WriteTo(b))
	file := saveFile(t, b, "overflow.sqlite")

	sqlite(t, file, ".tables", "planets\n")
	sqlite(t, file, "SELECT name FROM planets", "Mercury\nVenus\nEarth\nMars\nJupiter\nSaturn\nUranus\nNeptune\n")
	sqlite(t, file, "SELECT name, length(private_key) FROM planets", "Mercury|10000\nVenus|59\nEarth|40000\nMars|123456\nJupiter|1\nSaturn|12\nUranus|8000\nNeptune|75000\n")
}

func TestUpdates(t *testing.T) {
	// two tables, and then update them, to see what sqlite thinks of it.
	db := New()
	db.AddSlice("planets", []string{"name", "moons"}, [][]any{
		{"Mercury", 0},
		{"Venus", 0},
		{"Earth", 1},
		{"Mars", 2},
	})
	db.AddSlice("colors", []string{"name", "r", "g", "b"}, [][]any{
		{"white", 0, 0, 0},
		{"black", 256, 256, 256},
		{"red", 256, 0, 0},
	})

	b := &bytes.Buffer{}
	ok(t, db.WriteTo(b))
	file := saveFile(t, b, "updates.sqlite")

	sqlite(t, file, ".tables", "colors   planets\n")
	sqlite(t, file, "SELECT name FROM planets", "Mercury\nVenus\nEarth\nMars\n")
	sqlite(t, file, "SELECT name FROM colors", "white\nblack\nred\n")

	sqlite(t, file, "INSERT INTO colors (name, r, g, b) VALUES ('sky blue', 135, 206, 235)", "")
	sqlite(t, file, "SELECT name FROM colors ORDER BY r DESC", "black\nred\nsky blue\nwhite\n")
	sqlite(t, file, "DELETE FROM colors WHERE r < 200", "")
	sqlite(t, file, "SELECT name FROM colors", "black\nred\n")
}

func TestManyRows(t *testing.T) {
	// enough rows to have multiple levels of interior pages
	db := New()
	var rows [][]any
	// 128_000 uses a single level of interior pages, but that's no fun
	for i := 1; i < 200_001; i++ {
		rows = append(rows, []any{"r" + strconv.Itoa(i)})
	}

	db.AddSlice("counts", []string{"count"}, rows)

	b := &bytes.Buffer{}
	ok(t, db.WriteTo(b))
	file := saveFile(t, b, "manyrows.sqlite")

	sqlite(t, file, ".tables", "counts\n")
	sqlite(t, file, "SELECT count(*) FROM counts", "200000\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=1", "r1\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=42", "r42\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=100", "r100\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=1000", "r1000\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=10000", "r10000\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=100000", "r100000\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=199999", "r199999\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=200000", "r200000\n")
	sqlite(t, file, "SELECT count FROM counts WHERE rowid=200001", "")
}

func TestManyTables(t *testing.T) {
	// Enough tables to overflow page 1.
	db := New()
	for i := 42; i < 100; i++ {
		tab := fmt.Sprintf("table_%d", i)
		db.AddSlice(tab, []string{"chairs"}, [][]any{{1}, {2}, {3}, {4}, {5}})
	}

	b := &bytes.Buffer{}
	ok(t, db.WriteTo(b))
	file := saveFile(t, b, "manytables.sqlite")

	sqlite(t, file, "SELECT count(*) FROM table_42", "5\n")
	sqlite(t, file, "SELECT sum(chairs) FROM table_87", "15\n")
}

func TestHuge(t *testing.T) {
	if os.Getenv("HUGE") == "" {
		t.Skip("not huge")
	}

	n := 2_000 // much more and I get OOM

	var (
		db      = New()
		rows    [][]any
		payload = strings.Repeat("x", 512_000) // 1/3 of a floppydisk
	)
	for i := 0; i < n; i++ {
		rows = append(rows, []any{payload, payload})
	}
	db.AddSlice("exes", []string{"xes", "axes"}, rows)

	b := &bytes.Buffer{}
	ok(t, db.WriteTo(b))
	file := saveFile(t, b, "huge.sqlite")

	sqlite(t, file, ".tables", "exes\n")
	sqlite(t, file, "SELECT count(*) FROM exes", fmt.Sprintf("%d\n", n))
}

func TestFailSlice(t *testing.T) {
	msg := `table "plants": unsupported type (struct {})`

	db := New()
	fail(t, msg, db.AddSlice("plants", []string{"name", "size"}, [][]any{{struct{}{}}}))

	fail(t, msg, db.AddSlice("blanb", []string{"name"}, [][]any{{"hi"}}))

	b := &bytes.Buffer{}
	fail(t, msg, db.WriteTo(b))
}

func TestFailChannel(t *testing.T) {
	// failing a write should drain the channel
	msg := `table "plants": unsupported type (struct {})`
	db := New()

	// error halfway the channel
	{
		var (
			ch   = make(chan []any)
			done = make(chan struct{})
		)
		go func() {
			ch <- []any{1}
			ch <- []any{2}
			ch <- []any{struct{}{}}
			ch <- []any{4}
			ch <- []any{5}
			close(done)
			close(ch)
		}()

		fail(t, msg, db.AddChan("plants", []string{"name", "size"}, ch))

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout")
		}
	}

	// we should still drain later channels as well
	{
		var (
			ch   = make(chan []any)
			done = make(chan struct{})
		)
		go func() {
			ch <- []any{1}
			ch <- []any{2}
			close(done)
			close(ch)
		}()
		fail(t, msg, db.AddChan("blanb", []string{"name"}, ch))
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout")
		}
	}

	b := &bytes.Buffer{}
	fail(t, msg, db.WriteTo(b))
}

func BenchmarkCreate(b *testing.B) {
	payload := strings.Repeat("x", 512)
	n := 4000
	var rows [][]any
	for i := 0; i < n; i++ {
		rows = append(rows, []any{payload, 12, 42, payload})
	}

	for i := 0; i < b.N; i++ {
		db := New()

		db.AddSlice("exes", []string{"xes", "axes"}, rows)

		buf := &bytes.Buffer{}
		ok(b, db.WriteTo(buf))
		file := saveFile(b, buf, "bench.sqlite")

		sqlite(b, file, "SELECT count(*) FROM exes", fmt.Sprintf("%d\n", n))
	}
}

func Example() {
	db := New()

	// Table with all data in memory already
	db.AddSlice("planets", []string{"name", "moons"}, [][]any{
		{"Mercury", 0},
		{"Venus", 0},
		{"Earth", 1},
		{"Mars", 2},
		{"Jupiter", 80},
		{"Saturn", 83},
		{"Uranus", 27},
		{"Neptune", 4},
	})

	// Table with all data from a channel
	stars := make(chan []any, 10)
	stars <- []any{"Alpha Centauri", "4"}
	stars <- []any{"Barnard's Star", "6"}
	stars <- []any{"Luhman 16", "6"}
	stars <- []any{"WISE 0855−0714", "7"}
	stars <- []any{"Wolf 359", "7"}
	db.AddChan("stars", []string{"name", "lightyears"}, stars)

	b := &bytes.Buffer{}
	err := db.WriteTo(b)
	_ = err // ..
	err = os.WriteFile("/tmp/universe.sqlite", b.Bytes(), 0600)
	_ = err // ..
}
