package bakelite

import (
	"fmt"

	"github.com/alicebob/bakelite/internal"
)

// Write a "leaf table" page. It returns how many cells it managed to fit on this page.
// Cells need to be ordered.
// If isPage1 is true we start 100 bytes in (and so there is 100 bytes less available)
func writeTableLeaf(page []byte, isPage1 bool, cells []tableLeafCell) int {
	// format:
	// page[0]: type
	// page[offset + 1,2]: first free block (not relevant)
	// page[offset + 3,4]: number of cells
	// page[offset + 5,6]: start of cell content area
	// page[offset + 7]: fragmented free bytes (not relevant)
	// page[[2 bytes]]: cell pointers: each points to its content, from start of page
	// page[...]: empty space
	// page[[...]]: cell content

	offset := 0
	if isPage1 {
		// page 1 is the same, but we start 100 bytes in.
		offset = 100
	}
	page[offset] = 0x0D // it's a leaf!

	contentStart := len(page)
	pointer := offset + 8 // where are we writing cell pointers to in page[].
	count := 0
	for _, cell := range cells {
		payload := cell.Bytes()

		if contentStart-len(payload) < pointer+2 {
			fmt.Printf("that leaf won't fit!\n")
			break
		}

		contentStart -= len(payload)
		copy(page[contentStart:], payload)
		internal.PutUint16(page[pointer:], uint16(contentStart))
		pointer += 2
		count += 1
	}

	internal.PutUint16(page[offset+3:], uint16(count))
	if contentStart < 1<<16 {
		// "0" means 64K, which happens when page size is 1<<16, there are no rows, and this is not isPage1.
		internal.PutUint16(page[offset+5:], uint16(contentStart))
	}
	return count
}

func writeTableInterior(page []byte, cells []tableInteriorCell) int {
	// format:
	// page[0]: type
	// page[offset + 1,2]: first free block (not relevant)
	// page[offset + 3,4]: number of cells
	// page[offset + 5,6]: start of cell content area
	// page[offset + 7]: fragmented free bytes (not relevant)
	// page[offset + 8..12]: rightmost pointer
	offset := 0
	page[offset] = 0x05 // interior table
	rightmost := uint32(cells[len(cells)-1].left)
	internal.PutUint32(page[offset+8:], rightmost) // right most cell

	contentStart := len(page)
	pointer := offset + 12 // where are we writing cell pointers to in page[].
	count := 0
	for _, cell := range cells[:len(cells)-1] {
		fmt.Printf("check cell starting at %d (page %d)\n", cell.key, cell.left)
		payload := cell.Bytes()

		if contentStart-len(payload) < pointer+2 {
			fmt.Printf("that interior page won't fit!\n")
			break
		}

		contentStart -= len(payload)
		copy(page[contentStart:], payload)
		internal.PutUint16(page[pointer:], uint16(contentStart))
		pointer += 2
		count += 1
	}
	internal.PutUint16(page[offset+3:], uint16(count))
	internal.PutUint16(page[offset+5:], uint16(contentStart))

	return count + 1 // we count the .rightmost one as "placed"
}
