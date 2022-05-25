package bakelite

import (
	"github.com/alicebob/bakelite/internal"
)

// read as many leafs from source which still fit in a single leaf page
func collectTableLeaf(isPage1 bool, source *recordSource) []tableLeafCell {
	var ls []tableLeafCell
	left := PageSize
	if isPage1 {
		left -= 100
	}
	left -= 8 // header
	for {
		next := source.Peek()
		if next == nil {
			break
		}
		needed := 2 + len(next.Bytes()) // FIXME: cache/use this. Now we do it twice
		if left < needed {
			break
		}
		ls = append(ls, *next)
		left -= needed
		source.Pop()
	}
	return ls
}

// Write a "leaf table" page. The list must fit on the page (see collectTableLeaf).
// If isPage1 is true we start 100 bytes in (and so there is 100 bytes less available)
func writeTableLeaf(page []byte, isPage1 bool, cells []tableLeafCell) {
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
	internal.PutUint16(page[offset+3:], uint16(len(cells)))

	contentStart := len(page)
	pointer := offset + 8 // where are we writing cell pointers to in page[].
	for _, cell := range cells {
		payload := cell.Bytes()
		contentStart -= len(payload)
		copy(page[contentStart:], payload)
		internal.PutUint16(page[pointer:], uint16(contentStart))
		pointer += 2
	}

	if contentStart == 1<<16 {
		// "0" means 64K, which happens when page size is 1<<16, there are no rows, and this is not isPage1.
		contentStart = 0
	}
	internal.PutUint16(page[offset+5:], uint16(contentStart))
}

// returns: how many we placed
func writeTableInterior(page []byte, isPage1 bool, cells []tableInteriorCell) int {
	// format:
	// page[0]: type
	// page[offset + 1,2]: first free block (not relevant)
	// page[offset + 3,4]: number of cells
	// page[offset + 5,6]: start of cell content area
	// page[offset + 7]: fragmented free bytes (not relevant)
	// page[offset + 8..12]: rightmost pointer
	offset := 0
	if isPage1 {
		offset += 100
	}
	page[offset] = 0x05 // interior table

	contentStart := len(page)
	pointer := offset + 12 // where are we writing cell pointers to in page[].
	count := 0
	rightmost := cells[0].left
	for _, cell := range cells[1:] {
		payload := interiorCell(rightmost, cell.key)
		rightmost = cell.left

		if contentStart-len(payload) < pointer+2 {
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
	internal.PutUint32(page[offset+8:], uint32(rightmost))

	return count + 1 // we count the .rightmost one as "placed"
}

func cellPayload(payload []byte) int {
	maxInPage := PageSize - 35 // defined by sqlite for page leaf cells.
	return calculateCellInPageBytes(int64(len(payload)), PageSize, maxInPage)
}
