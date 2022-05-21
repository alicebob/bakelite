package bakelite

import (
	"encoding/binary"
)

// Write a "leaf table" page. It's on you to make sure the cells fit (for now).
// Cells need to be ordered.
// We ignore cell overflow for now.
// If isPage1 is true we still return a full sized page, but the header starts 100 bytes in (and so there is 100 bytes less available)
func makeTableLeaf(isPage1 bool, cells []tableLeafCell) ([]byte, error) {
	enc := binary.BigEndian

	page := make([]byte, pageSize)
	offset := 0
	if isPage1 {
		offset = 100
	}
	page[offset] = 0x0D // it's a leaf!
	// page[offset + 1,2]: first free block (not relevant)
	// page[offset + 3,4]: number of cells
	// page[offset + 5,6]: start of cell content area (0 for our 64K pages)
	// page[offset + 7]: fragmented free bytes (not relevant)

	cellContentStart := len(page)
	pointer := offset + 8 // where are we writing cell pointers to in page[].
	count := uint16(0)
	for _, cell := range cells {
		payload := cell.Bytes()

		// TODO: check if this doesn't overwrite the cell pointers
		cellContentStart -= len(payload)
		copy(page[cellContentStart:], payload)
		enc.PutUint16(page[pointer:], uint16(cellContentStart))
		pointer += 2
		count += 1
	}

	enc.PutUint16(page[offset+3:], uint16(count))
	if cellContentStart < 1<<16 {
		// "0" means 64K, which happens when there are no rows and not isPage1.
		enc.PutUint16(page[offset+5:], uint16(cellContentStart))
	}
	return page, nil
}
