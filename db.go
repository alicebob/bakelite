package bakelite

import (
	"github.com/alicebob/bakelite/internal"
)

type DB struct {
	store storer
	// pages  [][]byte    // all these are of the correct length (PageSize)
	master   []masterRow // one entry per table, "sqlite_master" table, which is stored at "page 1" (pages[0])
	lastPage int
	err      error
}

func newDB(store storer) *DB {
	db := &DB{
		store: store,
	}
	db.addPage(db.blankPage()) // setup page1
	return db
}

func (db *DB) blankPage() []byte {
	return make([]byte, PageSize)
}

// add a page and return its ID.
func (db *DB) addPage(p []byte) int {
	id, err := db.store.AddPage(p)
	if err != nil {
		db.err = err
	}
	db.lastPage = id
	return id
	// id := len(db.pages) + 1
	// db.pages = append(db.pages, p)
	// return id
}

// adds all the rows of a table to the database. Returns the root ID.
func (db *DB) storeTable(source *recordSource) int {
	// first fill all the table cell pages, collecting which page(s) we created.
	isPage1 := false
	var leafCells []tableInteriorCell
	for {
		cells := collectTableLeaf(isPage1, source)
		if db.err != nil {
			return 0
		}
		firstKey := 0
		if len(cells) > 0 {
			firstKey = cells[0].rowID
		}

		page := db.blankPage()
		writeTableLeaf(page, isPage1, cells)
		leafCells = append(leafCells, tableInteriorCell{
			left: db.addPage(page),
			key:  firstKey,
		})
		if source.Peek() == nil {
			break
		}
	}

	// ...then the (possibly skipped, possibly nested) interior pages
	return db.buildInterior(leafCells)
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
		placed := writeTableInterior(page, false, cells)
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
	internal.PutUint32(page, uint32(nextID))
	copy(page[4:], car)
	return db.addPage(page)
}

// transform a record to a leaf cell ready to store. Deals with overflow.
func (db *DB) makeLeafCell(rowID int, rec []byte) tableLeafCell {
	fullSize := len(rec)
	maxInPage := cellPayload(rec)
	overflow := 0
	if len(rec) > maxInPage {
		overflow = db.storeOverflow(rec[maxInPage:])
		rec = rec[:maxInPage]
	}
	return leafCell(rowID, fullSize, rec, overflow)
}

// "page 1" is the first page (db.page[0]) of the db. It is a leaf page with
// all the tables in it. The first 100 bytes have the database header.
// updatePage1() should be called when all tables have been added and we're
// about to generate the db file.
func (db *DB) getPage1() []byte {
	recs := db.masterRecords()
	source := newRecordSource(db, "sqlite_master", stream(recs))
	cells := collectTableLeaf(true, source)
	page1 := db.blankPage()

	if source.Peek() == nil {
		// Easy case, all our table definitions fit on page1, no interior pages
		// needed.
		writeTableLeaf(page1, true, cells)
	} else {
		// If we have just a few tables we're lucky and can fit all master
		// tables in page[0]. However, it seems that we have too many tables,
		// so we'll have to go build an interior-cell structure. SQLite can
		// deal with this case nicely; it's used to moving things around, but
		// we're not.
		// So what we do is we build a new page with the leaf cells we just
		// wanted to used as page1, and then we make a normal tree with all the
		// other records. Finally we put in page1 a interiorcell to the page
		// and the tree.
		// SQLite is fine with this.
		page := db.blankPage()
		writeTableLeaf(page, false, cells)
		pageID := db.addPage(page)

		firstKey := source.Peek().rowID
		restRootID := db.storeTable(source)

		writeTableInterior(page1, true, []tableInteriorCell{
			{left: pageID, key: 0},
			{left: restRootID, key: firstKey},
		})
	}

	h := header(db.lastPage)
	copy(page1, h) // overwrite the first 100 bytes
	return page1
}

func (db *DB) masterRecords() [][]any {
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
	return rows
}
