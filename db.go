package bakelite

import (
	"encoding/binary"
)

type DB struct {
	pages  [][]byte    // all these are of the correct length (PageSize)
	master []masterRow // one entry per table, "sqlite_master" table, which is stored at "page 1" (pages[0])
}

func (db *DB) blankPage() []byte {
	return make([]byte, PageSize)
}

// add a page and return its ID.
func (db *DB) addPage(p []byte) int {
	id := len(db.pages) + 1
	db.pages = append(db.pages, p)
	return id
}

// adds all the rows of a table to the database. Returns the root ID.
func (db *DB) storeBtree(rows [][]any) (int, error) {
	cells, err := db.makeLeafCells(rows)
	if err != nil {
		return 0, err
	}

	// first fill all the table cell pages...
	var leafCells []tableInteriorCell
	for {
		page := db.blankPage()
		placed := writeTableLeaf(page, false, cells)
		key := 0
		if len(cells) > 0 {
			// 0 cells is valid. The page will end up as .rightmost
			key = cells[0].left
		}
		leafCells = append(leafCells, tableInteriorCell{
			left: db.addPage(page),
			key:  key,
		})
		cells = cells[placed:]
		if len(cells) == 0 {
			break
		}
	}

	// ...then the (possibly skipped, possibly nested) interior pages
	return db.buildInterior(leafCells), nil
}

// gets a list of page IDs and stores them in a tree of "interior table" pages.
// assumes len(pageIDs) > 0
func (db *DB) buildInterior(cells []tableInteriorCell) int {
	if len(cells) == 1 {
		return cells[0].left
	}

	var leafCells []tableInteriorCell
	for len(cells) > 0 {
		page := db.blankPage()
		placed := writeTableInterior(page, cells)
		leafCells = append(leafCells, tableInteriorCell{
			left: db.addPage(page),
			key:  cells[0].key,
		})
		cells = cells[placed:]
	}
	return db.buildInterior(leafCells)
}

// store arbitrary long overflow in a sequence of linked pages. Returns the root page ID.
func (db *DB) storeOverflow(b []byte) int {
	// First 4 bytes are the page ID of the next page, or 0.
	page := db.blankPage()

	if len(b) < PageSize-4 {
		copy(page[4:], b)
		return db.addPage(page)
	}

	car, cdr := b[:PageSize-4], b[PageSize-4:]
	nextID := db.storeOverflow(cdr)
	binary.BigEndian.PutUint32(page, uint32(nextID))
	copy(page[4:], car)
	return db.addPage(page)
}

func (db *DB) makeLeafCells(rows [][]any) ([]tableLeafCell, error) {
	var cells []tableLeafCell
	for i, row := range rows {
		rec, err := makeRecord(row)
		if err != nil {
			return nil, err
		}
		fullSize := len(rec)
		maxInPage := PageSize - 35 // defined by sqlite for page leaf cells.
		maxInCell := calculateCellInPageBytes(int64(fullSize), PageSize, maxInPage)
		overflow := 0
		if len(rec) > maxInCell {
			overflow = db.storeOverflow(rec[maxInCell:])
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

// "page 1" is the first page (db.page[0]) of the db. It is a leaf page with
// all the tables in it. The first 100 bytes have the database header.
// updatePage1() should be called when all tables have been added and we're
// about to generate the db file.
func (db *DB) updatePage1() {
	cells := db.masterCells()
	page := db.pages[0]
	placed := writeTableLeaf(page, true /* ! */, cells)
	if placed != len(cells) {
		// FIXME
		// for the master table we don't add interior cells yet
		panic("too many tables for now. Fixme.")
	}
	h := header(len(db.pages))
	copy(page, h) // overwrite the first 100 bytes
}

func (db *DB) masterCells() []tableLeafCell {
	var rows [][]any
	for _, master := range db.master {
		rows = append(rows, []any{
			master.typ,
			master.name,
			master.tblName,
			master.rootpage,
			master.sql,
		})
	}
	cells, err := db.makeLeafCells(rows)
	if err != nil {
		panic(err)
	}
	return cells
}
