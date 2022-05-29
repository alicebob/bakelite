package bakelite

import (
	"fmt"
	"io"
	"strings"
)

const (
	// database file page size
	PageSize = 1 << 12
)

// Create a new, in memory, db.
func New() *DB {
	db := &DB{}
	db.addPage(db.blankPage()) // pages[0] is the master page
	return db
}

// Add a new table with the given columns and rows.
// Table and column names should probably be lowercase simple strings.
// The number of items in every row is allowed to be shorter than the number of
// columns, according to the sqlite docs. What happens when there are more is
// not defined.
// Don't add the same table twice.
// Don't use this from multiple Go routines at the same time.
//
// Supported Go datatypes:
//   - int
//   - float64
//   - string
//   - []byte
//   - nil
// (yup, that's all for now)
//
// This returns an error if any value is of an unsupported type. From then on
// any other call to AddChan() will return the same error, and so will
// WriteTo(). If there is an error we will drain the rows channel.
func (db *DB) AddChan(table string, columns []string, rows <-chan []any) error {
	if db.err != nil {
		// prevent Go routine leak
		for range rows {
		}
		return db.err
	}

	source := newRecordSource(db, table, rows)
	root := db.storeTable(source)
	db.master = append(db.master, masterRow{
		typ:      "table",
		name:     table,
		tblName:  table,
		rootpage: root,
		sql:      sqlCreate(table, columns),
	})

	// prevent Go routine leak if there was an error
	for range rows {
	}

	return db.err
}

// AddSlice is a helper to call AddChan.
func (db *DB) AddSlice(table string, columns []string, rows [][]any) error {
	return db.AddChan(table, columns, stream(rows))
}

// Write the whole file to the writer. You probably don't want to use the db again.
func (db *DB) WriteTo(w io.Writer) error {
	if db.err != nil {
		return db.err
	}

	db.updatePage1()

	for _, p := range db.pages {
		if _, err := w.Write(p); err != nil {
			return err
		}
	}
	return nil
}

func sqlCreate(table string, columns []string) string {
	return fmt.Sprintf(
		`CREATE TABLE %q (%s)`,
		table,
		strings.Join(columns, ", "),
	)
}

type masterRow struct {
	typ      string // "table", "index"
	name     string // name of the table, index, &c
	tblName  string // which table an index is for
	rootpage int
	sql      string
}
