package bakelite

import (
	"github.com/alicebob/bakelite/internal"
)

type page struct {
	store []byte // every page will be of size pageSize
}

type masterRow struct {
	typ      string // "table", "index"
	name     string
	tblName  string // ?
	rootpage int
	sql      string
}

type db struct {
	pages  []page
	master []masterRow // one entry per table, "sqlite_master" table, which is stored at "page 1" (pages[0])
}

type tableLeafCell struct {
	left    int64 // rowID
	payload []byte
}

// returns bytes for a "Table B-Tree Leaf Cell (header 0x0d)"
//
// - A varint which is the total number of bytes of payload, including any overflow
// - A varint which is the integer key, a.k.a. "rowid"
// - The initial portion of the payload that does not spill to overflow pages.
// - A 4-byte big-endian integer page number for the first page of the overflow page list - omitted if all payload fits on the b-tree page.
func (c *tableLeafCell) Bytes() []byte {
	b := make([]byte, len(c.payload)+(2*9))
	n := 0
	n += internal.PutUvarint(b[n:], uint64(len(c.payload)))
	n += internal.PutUvarint(b[n:], uint64(c.left))
	n += copy(b[n:], c.payload)
	return b[:n]
}

// convert "rows" to a table, which is currently always a single page (or stuff breaks badly)
func makeTable(rows [][]any) page {
	cells, err := makeCells(rows)
	if err != nil {
		panic(err)
	}
	bs, err := makeTableLeaf(false, cells)
	if err != nil {
		panic(err)
	}
	return page{
		store: bs,
	}
}

func makeCells(rows [][]any) ([]tableLeafCell, error) {
	var cells []tableLeafCell
	for i, row := range rows {
		rec, err := makeRecord(row)
		if err != nil {
			return nil, err
		}
		cells = append(cells, tableLeafCell{
			left:    int64(i),
			payload: rec,
		})
	}
	return cells, nil
}

// "page 1" is the first page(d.page[0]) of the db. It is a leaf page with all the tables in it. The first 100 bytes have the database header.
// Should be called when all tables have been added and we're about to generate the db file.
func (d *db) UpdatePage1() error {
	cells := d.masterCells()
	bs, err := makeTableLeaf(true /* ! */, cells)
	if err != nil {
		return err
	}
	h := header(len(d.pages))
	copy(bs, h)

	d.pages[0] = page{store: bs}
	return nil
}

func (d *db) masterCells() []tableLeafCell {
	var rows [][]any
	for _, master := range d.master {
		rows = append(rows, []any{
			master.typ,
			master.name,
			master.tblName,
			master.rootpage,
			master.sql,
		})
	}
	cells, err := makeCells(rows)
	if err != nil {
		panic(err)
	}
	return cells
}
