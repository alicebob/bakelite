package bakelite

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type page struct {
	store []byte // every page will be of size pageSize
}

type masterRow struct {
	typ      string // "table", "index"
	name     string // name of the table, index, &c
	tblName  string // which table an index is for
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
	cells, err := d.makeLeafCells(rows)
	if err != nil {
		return 0, err
	}

	// first fill all the table cell pages
	var leafCells []tableInteriorCell
	for {
		page := d.blankPage()
		placed := writeTableLeaf(page, false, cells)
		key := 0
		if len(cells) > 0 {
			// 0 cells is valid. The page will end up as .rightmost
			key = cells[0].left
		}
		fmt.Printf("we placed %d rows (leftmost rowid: %d)\n", placed, key)
		leafCells = append(leafCells, tableInteriorCell{
			left: d.addPage(page),
			key:  key,
		})
		cells = cells[placed:]
		if len(cells) == 0 {
			break
		}
	}
	return d.buildInterior(leafCells), nil
}

// gets a list of page IDs and stores them in a tree of "interior table" pages.
// assumes len(pageIDs) > 0
func (d *db) buildInterior(pageIDs []tableInteriorCell) int {
	fmt.Printf("buildInterior with %d pages\n", len(pageIDs))
	if len(pageIDs) == 1 {
		return pageIDs[0].left
	}

	page := d.blankPage()
	placed := writeTableInterior(page, pageIDs)
	if placed != len(pageIDs) {
		panic("nest me")
	}
	return d.addPage(page)
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

func (d *db) makeLeafCells(rows [][]any) ([]tableLeafCell, error) {
	var cells []tableLeafCell
	for i, row := range rows {
		rec, err := makeRecord(row)
		if err != nil {
			return nil, err
		}
		fullSize := len(rec)
		maxInPage := pageSize - 35 // defined by sqlite for page leaf cells.
		maxInCell := calculateCellInPageBytes(int64(fullSize), pageSize, maxInPage)
		overflow := 0
		if len(rec) > maxInCell {
			overflow = d.storeOverflow(rec[maxInCell:])
			rec = rec[:maxInCell]
		}
		cells = append(cells, tableLeafCell{
			left:     i,
			fullSize: fullSize,
			payload:  rec,
			overflow: overflow,
		})
	}
	return cells, nil
}

// "page 1" is the first page(d.page[0]) of the db. It is a leaf page with all the tables in it. The first 100 bytes have the database header.
// Should be called when all tables have been added and we're about to generate the db file.
func (d *db) UpdatePage1() error {
	cells := d.masterCells()
	page := d.pages[0]
	placed := writeTableLeaf(page, true /* ! */, cells)
	if placed != len(cells) {
		// FIXME
		// for the master table we don't add interior cells yet
		return errors.New("too many tables for now. Fixme.")
	}
	h := header(len(d.pages))
	copy(page, h) // overwrite the first 100 bytes
	// d.pages[0] = page

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
	cells, err := d.makeLeafCells(rows)
	if err != nil {
		panic(err)
	}
	return cells
}
