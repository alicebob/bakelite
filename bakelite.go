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
// Does nothing on db.Close()
func New() *DB {
	store := &memStore{}
	return newDB(store)
}

// Create a new db, with a tmp file stored in `dir`. See os.CreateTemp: if dir
// is empty this uses the default directory for temporary files.
// Unlinks the file on db.Close()
func NewTmp(dir string) (*DB, error) {
	store, err := newTmpStore(dir)
	if err != nil {
		return nil, err
	}
	return newDB(store), nil
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
// If any previous AddChan() or AddSlice() returned an error, then this will
// return the same error.
func (db *DB) WriteTo(w io.Writer) error {
	if db.err != nil {
		return db.err
	}

	page1 := db.getPage1()
	return db.store.WriteTo(page1, w)
}

// Cleanup resources. If the database was created with NewTmp() Close() will remove the temp file.
func (db *DB) Close() error {
	return db.store.Close()
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
