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
// The length of every row should be allowed to be shorter than the number of
// columns, according to the sqlite docs. What happens when they are longer is
// not defined.
// Don't add the same table twice.
// Don't use this from multiple Go routines at the same time.
//
// Supported Go datatypes:
//   - int
//   - string
//   - nil
// (yup, that's all for now)
func (db *DB) Add(table string, columns []string, rows [][]any) error {
	root, err := db.storeBtree(rows)
	if err != nil {
		return err
	}
	db.master = append(db.master, masterRow{
		typ:      "table",
		name:     table,
		tblName:  table,
		rootpage: root,
		sql:      sqlCreate(table, columns),
	})
	return nil
}

// Write the whole file to the writer. You probably don't want to use the db again.
func (db *DB) Write(w io.Writer) error {
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
