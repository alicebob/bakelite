package bakelite

import (
	"fmt"
)

type recordSource struct {
	db     *DB
	table  string // only for error messages
	source <-chan []any
	rowID  int
	peek   *tableLeafCell
}

func newRecordSource(db *DB, table string, source <-chan []any) *recordSource {
	return &recordSource{
		db:     db,
		table:  table,
		source: source,
		rowID:  1,
	}
}

// Get next tableLeafCell from source. It'll keep returning the same cell until
// Pop() is called. This is because creating a leaf cell might store overflow
// in the DB as a side effect.
// If source is empty we'll return nil
func (r *recordSource) Peek() *tableLeafCell {
	if r.peek == nil {
		row, ok := <-r.source
		if !ok {
			return nil
		}

		rec, err := makeRecord(row)
		if err != nil {
			r.db.err = fmt.Errorf("table %q: %w", r.table, err)
			return nil
		}
		cell := r.db.makeLeafCell(r.rowID, rec)
		r.peek = &cell
		r.rowID++
	}
	return r.peek
}

func (r *recordSource) Pop() {
	r.peek = nil
}

// helper to go from slices to a channel
func stream(rows [][]any) <-chan []any {
	source := make(chan []any)
	go func() {
		defer close(source)
		for _, row := range rows {
			source <- row
		}
	}()
	return source
}
