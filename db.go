package bakelite

import (
	"github.com/alicebob/bakelite/internal"
)

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
