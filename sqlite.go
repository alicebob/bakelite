package bakelite

func calculateCellInPageBytes(l int64, pageSize int, maxInPagePayload int) int {
	// Overflow calculation described in the file format spec. The
	// variable names and magic constants are from the spec exactly.
	u := int64(pageSize)
	p := l
	x := int64(maxInPagePayload)
	m := ((u - 12) * 32 / 255) - 23
	k := m + ((p - m) % (u - 4))

	if p <= x {
		return int(l)
	} else if k <= x {
		return int(k)
	} else {
		return int(m)
	}
}

func cellPayload(payload []byte) int {
	maxInPage := PageSize - 35 // defined by sqlite for page leaf cells.
	return calculateCellInPageBytes(int64(len(payload)), PageSize, maxInPage)
}
