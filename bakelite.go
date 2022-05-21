package bakelite

import (
	"fmt"
	"io"
	"strings"
)

func New() *db {
	return &db{
		pages: make([]page, 1), // pages[0] is master page
	}
}

func (d *db) Add(table string, columns []string, rows [][]string) error {
	sql := sqlCreate(table, columns)
	d.master = append(d.master, masterRow{
		typ:      "table",
		name:     table, // FIXME
		tblName:  table, // FIXME
		rootpage: 0,     // FIXME
		sql:      sql,
	})
	return nil
}

func (d *db) Write(w io.Writer) error {
	if err := writeHeader(w, len(d.pages)); err != nil {
		return err
	}

	panic("wip")
}

func sqlCreate(table string, columns []string) string {
	return fmt.Sprintf(
		`CREATE TABLE %q (%s)`,
		table,
		strings.Join(columns, ", "),
	)
}
