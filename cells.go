package bakelite

import (
	"github.com/alicebob/bakelite/internal"
)

type tableLeafCell struct {
	rowID   int    // rowID
	payload []byte // what's stored in the leaf, see leafCell()
}

type tableInteriorCell struct {
	left int // page pointer
	key  int // rowID
}

// returns bytes for a "Table B-Tree Leaf Cell (header 0x0d)"
//
// - A varint which is the total number of bytes of payload, including any overflow
// - A varint which is the integer key, a.k.a. "rowid"
// - The initial portion of the payload that does not spill to overflow pages.
// - A 4-byte big-endian integer page number for the first page of the overflow page list - omitted if all payload fits on the b-tree page.
func leafCell(rowID int, fullSize int, payload []byte, overflowPageID int) tableLeafCell {
	b := make([]byte, len(payload)+(2*9)+4)
	n := 0
	n += internal.PutUvarint(b[n:], uint64(fullSize))
	n += internal.PutUvarint(b[n:], uint64(rowID))
	n += copy(b[n:], payload)
	if overflowPageID > 0 {
		internal.PutUint32(b[n:], uint32(overflowPageID))
		n += 4
	}

	return tableLeafCell{
		rowID:   rowID,
		payload: b[:n],
	}
}

// returns the bytes for a "Table B-Tree Interior Cell (header 0x05)"
//
// - A 4-byte big-endian page number which is the left child pointer.
// - A varint which is the integer key
func interiorCell(left int, key int) []byte {
	b := make([]byte, 9+4)
	n := 0
	internal.PutUint32(b[n:], uint32(left))
	n += 4
	n += internal.PutUvarint(b[n:], uint64(key))
	return b[:n]
}
