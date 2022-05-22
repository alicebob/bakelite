package bakelite

import (
	"fmt"
	"io"
	"strings"
)

func New() *db {
	d := &db{}
	master := d.blankPage() // pages[0] is the master page
	d.addPage(master)
	return d
}

func (d *db) Add(table string, columns []string, rows [][]any) error {
	root, err := d.storeBtree(rows)
	if err != nil {
		return err
	}
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
