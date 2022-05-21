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

func (d *db) Add(table string, columns []string, rows [][]any) error {
	page := makeTable(rows)
	root := len(d.pages) + 1
	d.pages = append(d.pages, page)
	d.master = append(d.master, masterRow{
		typ:      "table",
		name:     table,
		tblName:  table,
		rootpage: root,
		sql:      sqlCreate(table, columns),
	})
	return nil
}

func (d *db) Write(w io.Writer) error {
	if err := d.UpdatePage1(); err != nil {
		return err
	}
	for _, p := range d.pages {
		if _, err := w.Write(p.store); err != nil {
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
