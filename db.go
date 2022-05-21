package bakelite

import ()

type page struct {
	store [pageSize]byte
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
