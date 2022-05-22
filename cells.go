package bakelite

import (
	"github.com/alicebob/bakelite/internal"
)

type tableLeafCell struct {
	left     int    // rowID
	fullSize int    // length with overflow
	payload  []byte // without overflow
	overflow int    // page ID, or zero
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
func (c *tableLeafCell) Bytes() []byte {
	b := make([]byte, len(c.payload)+(2*9)+4)
	n := 0
	n += internal.PutUvarint(b[n:], uint64(c.fullSize))
	n += internal.PutUvarint(b[n:], uint64(c.left))
	n += copy(b[n:], c.payload)
	if c.overflow > 0 {
		internal.PutUint32(b[n:], uint32(c.overflow))
		n += 4
	}
	return b[:n]
}

// returns the bytes for a "Table B-Tree Interior Cell (header 0x05)"
//
// - A 4-byte big-endian page number which is the left child pointer.
// - A varint which is the integer key
func (c *tableInteriorCell) Bytes() []byte {
	b := make([]byte, 9+4)
	n := 0
	internal.PutUint32(b[n:], uint32(c.left))
	n += 4
	n += internal.PutUvarint(b[n:], uint64(c.key))
	return b[:n]
}
