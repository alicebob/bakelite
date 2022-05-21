package bakelite

// type tableLeaf struct {
// cells []tableLeafCell
// }

// Write a "leaf table" page. It's on you to make sure the cells fit (for now).
// Page 1 is special, since it's 100 bytes shorter.
func makeTableLeaf(isPage1 bool, cells []tableLeafCell) ([]byte, error) {
	size := pageSize
	if isPage1 {
		size -= 100
	}
	page := make([]byte, size)
	page[0] = 0x0D // it's a leaf!
	// page[1,2]: first free block (not relevant)
	// page[3,4]: number of cells
	// page[5,6]: start of cell content area (0 for our 64K pages)
	// page[7]: fragmented free bytes (not relevant)
	return page, nil
}
