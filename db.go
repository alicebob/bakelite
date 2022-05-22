package bakelite

import (
	"encoding/binary"
	"fmt"

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
	pages  [][]byte    // all these are of the correct length (pageSize)
	master []masterRow // one entry per table, "sqlite_master" table, which is stored at "page 1" (pages[0])
}

// add a page and return its ID.
func (d *db) addPage(p []byte) int {
	id := len(d.pages) + 1
	d.pages = append(d.pages, p)
	return id
}

func (d *db) blankPage() []byte {
	return make([]byte, pageSize)
}

// adds a all the rows of a table to the database. Returns the root ID.
func (d *db) storeBtree(rows [][]any) (int, error) {
	cells, err := d.makeCells(rows)
	if err != nil {
		return 0, err
	}

	page := d.blankPage()
	if err := makeTableLeaf(page, false, cells); err != nil {
		panic(err)
	}

	return d.addPage(page), nil
}

// store arbitrary long overflow in a sequence of linked pages. Returns the root page ID.
func (d *db) storeOverflow(b []byte) int {
	// First 4 bytes are the page ID of the next page, or 0.
	page := d.blankPage()

	if len(b) < pageSize-4 {
		copy(page[4:], b)
		return d.addPage(page)
	}

	car, cdr := b[:pageSize-4], b[pageSize-4:]
	nextID := d.storeOverflow(cdr)
	binary.BigEndian.PutUint32(page, uint32(nextID))
	copy(page[4:], car)
	return d.addPage(page)
}

func (d *db) makeCells(rows [][]any) ([]tableLeafCell, error) {
	var cells []tableLeafCell
	for i, row := range rows {
		rec, err := makeRecord(row)
		if err != nil {
			return nil, err
		}
		fullSize := len(rec)
		maxInPage := pageSize - 35 // defined by sqlite for page leaf cells.
		maxInCell := calculateCellInPageBytes(int64(fullSize), pageSize, maxInPage)
		fmt.Printf("makeCells %d is %d (%d)\n", i, fullSize, maxInCell)
		overflow := 0
		if len(rec) > maxInCell {
			overflow = d.storeOverflow(rec[maxInCell:])
			fmt.Printf("go store overflow: %d\n", overflow)
			rec = rec[:maxInCell]
		}
		fmt.Printf("makeCells %d new: %d (%d)\n", i, len(rec), overflow)
		cells = append(cells, tableLeafCell{
			left:     i,
			fullSize: fullSize,
			payload:  rec,
			overflow: overflow,
		})
	}
	return cells, nil
}

type tableLeafCell struct {
	left     int    // rowID
	fullSize int    // length with overflow
	payload  []byte // without overflow
	overflow int    // page ID, or zero
}

// returns bytes for a "Table B-Tree Leaf Cell (header 0x0d)"
//
// - A varint which is the total number of bytes of payload, including any overflow
// - A varint which is the integer key, a.k.a. "rowid"
// - The initial portion of the payload that does not spill to overflow pages.
// - A 4-byte big-endian integer page number for the first page of the overflow page list - omitted if all payload fits on the b-tree page.
func (c *tableLeafCell) Bytes() []byte {
	b := make([]byte, len(c.payload)+(2*9)+4)
	n := 0
	n += internal.PutUvarint(b[n:], uint64(c.fullSize))
	n += internal.PutUvarint(b[n:], uint64(c.left))
	n += copy(b[n:], c.payload)
	if c.overflow > 0 {
		binary.BigEndian.PutUint32(b[n:], uint32(c.overflow))
		n += 4
	}
	return b[:n]
}

// FIXME: delete this
func makeCells(rows [][]any) ([]tableLeafCell, error) {
	var cells []tableLeafCell
	for i, row := range rows {
		rec, err := makeRecord(row)
		if err != nil {
			return nil, err
		}
		cells = append(cells, tableLeafCell{
			left:     i,
			fullSize: len(rec),
			payload:  rec,
		})
	}
	return cells, nil
}

// "page 1" is the first page(d.page[0]) of the db. It is a leaf page with all the tables in it. The first 100 bytes have the database header.
// Should be called when all tables have been added and we're about to generate the db file.
func (d *db) UpdatePage1() error {
	cells := d.masterCells()
	page := d.pages[0]
	err := makeTableLeaf(page, true /* ! */, cells)
	if err != nil {
		return err
	}
	h := header(len(d.pages))
	copy(page, h) // overwrite the first 100 bytes
	d.pages[0] = page

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
