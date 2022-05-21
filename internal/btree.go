// This file is copied from sqlittle (also MIT)
package internal

import (
	"encoding/binary"
	"errors"
)

const (
	// "Define the depth of a leaf b-tree to be 1 and the depth of any interior
	// b-tree to be one more than the maximum depth of any of its children. In
	// a well-formed database, all children of an interior b-tree have the same
	// depth."
	// and:
	// "A "pointer" in an interior b-tree page is just the 31-bit integer page
	// number of the child page."
	// Ergo, in a well-formed database where every interal page only links to a
	// single left branch (highly unlikely), we can't ever go deeper than 31
	// levels.
	maxRecursion = 31
)

type CellPayload struct {
	Length   int64
	Payload  []byte
	Overflow int
}

type TableLeafCell struct {
	Left    int64 // rowID
	Payload CellPayload
}
type TableLeaf struct {
	Cells []TableLeafCell
}

type tableInteriorCell struct {
	left int
	key  int64
}
type tableInterior struct {
	cells     []tableInteriorCell
	rightmost int
}

type indexLeaf struct {
	cells []CellPayload
}

type indexInteriorCell struct {
	left    int // pageID
	payload CellPayload
}
type indexInterior struct {
	cells     []indexInteriorCell
	rightmost int
}

var headerSize = 100

func NewBtree(b []byte, isFileHeader bool, pageSize int) (interface{}, error) {
	hb := b
	if isFileHeader {
		hb = b[headerSize:]
	}
	cells := int(binary.BigEndian.Uint16(hb[3:5]))
	switch typ := int(hb[0]); typ {
	case 0x0d:
		return newLeafTableBtree(cells, hb[8:], b, pageSize)
	case 0x05:
		rightmostPointer := int(binary.BigEndian.Uint32(hb[8:12]))
		return newInteriorTableBtree(cells, hb[12:], b, rightmostPointer)
	case 0x0a:
		return newLeafIndex(cells, b[8:], b, pageSize)
	case 0x02:
		rightmostPointer := int(binary.BigEndian.Uint32(b[8:12]))
		return newInteriorIndex(cells, b[12:], b, rightmostPointer, pageSize)
	default:
		return nil, errors.New("unsupported page type")
	}
}

func newLeafTableBtree(
	count int,
	pointers []byte,
	content []byte,
	pageSize int,
) (*TableLeaf, error) {
	cells, err := parseCellpointers(count, pointers, len(content))
	if err != nil {
		return nil, err
	}
	leafs := make([]TableLeafCell, len(cells))
	for i, start := range cells {
		leafs[i], err = parseTableLeaf(content[start:], pageSize)
		if err != nil {
			return nil, err
		}
	}
	return &TableLeaf{
		Cells: leafs,
	}, nil
}

func newInteriorTableBtree(
	count int,
	pointers []byte,
	content []byte,
	rightmost int,
) (*tableInterior, error) {
	cells, err := parseCellpointers(count, pointers, len(content))
	if err != nil {
		return nil, err
	}
	cs := make([]tableInteriorCell, len(cells))
	for i, start := range cells {
		cs[i], err = parseTableInterior(content[start:])
		if err != nil {
			return nil, err
		}
	}
	return &tableInterior{
		cells:     cs,
		rightmost: rightmost,
	}, nil
}

func newLeafIndex(
	count int,
	pointers []byte,
	content []byte,
	pageSize int,
) (*indexLeaf, error) {
	cells, err := parseCellpointers(count, pointers, len(content))
	if err != nil {
		return nil, err
	}
	cs := make([]CellPayload, len(cells))
	for i, start := range cells {
		cs[i], err = parseIndexLeaf(content[start:], pageSize)
		if err != nil {
			return nil, err
		}
	}
	return &indexLeaf{
		cells: cs,
	}, nil
}

func newInteriorIndex(
	count int,
	pointers []byte,
	content []byte,
	rightmost int,
	pageSize int,
) (*indexInterior, error) {
	cells, err := parseCellpointers(count, pointers, len(content))
	if err != nil {
		return nil, err
	}
	cs := make([]indexInteriorCell, len(cells))
	for i, start := range cells {
		cs[i], err = parseIndexInterior(content[start:], pageSize)
		if err != nil {
			return nil, err
		}
	}
	return &indexInterior{
		cells:     cs,
		rightmost: rightmost,
	}, nil
}

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

var ErrCorrupted = errors.New("corrupted")

// shared code for parsing payload from cells
func parsePayload(l int64, c []byte, pageSize int, maxInPagePayload int) (CellPayload, error) {
	overflow := 0
	inPageBytes := calculateCellInPageBytes(l, pageSize, maxInPagePayload)
	if l < 0 {
		return CellPayload{}, ErrCorrupted
	}

	if int64(inPageBytes) == l {
		return CellPayload{l, c, 0}, nil
	}

	if len(c) < inPageBytes+4 {
		return CellPayload{}, ErrCorrupted
	}

	c, overflow = c[:inPageBytes], int(binary.BigEndian.Uint32(c[inPageBytes:inPageBytes+4]))
	if overflow == 0 {
		return CellPayload{}, ErrCorrupted
	}
	return CellPayload{l, c, overflow}, nil
}

func parseTableLeaf(c []byte, pageSize int) (TableLeafCell, error) {
	l, n := ReadVarint(c)
	if n < 0 {
		return TableLeafCell{}, ErrCorrupted
	}
	c = c[n:]
	rowid, n := ReadVarint(c)
	if n < 0 {
		return TableLeafCell{}, ErrCorrupted
	}

	pl, err := parsePayload(l, c[n:], pageSize, pageSize-35)
	return TableLeafCell{
		Left:    rowid,
		Payload: pl,
	}, err
}

func parseTableInterior(c []byte) (tableInteriorCell, error) {
	if len(c) < 4 {
		return tableInteriorCell{}, ErrCorrupted
	}
	left := int(binary.BigEndian.Uint32(c[:4]))
	key, n := ReadVarint(c[4:])
	if n < 0 {
		return tableInteriorCell{}, ErrCorrupted
	}
	return tableInteriorCell{
		left: left,
		key:  key,
	}, nil
}

func parseIndexLeaf(c []byte, pageSize int) (CellPayload, error) {
	l, n := ReadVarint(c)
	if n < 0 {
		return CellPayload{}, ErrCorrupted
	}
	return parsePayload(l, c[n:], pageSize, ((pageSize-12)*64/255)-23)
}

func parseIndexInterior(c []byte, pageSize int) (indexInteriorCell, error) {
	if len(c) < 4 {
		return indexInteriorCell{}, ErrCorrupted
	}
	left := int(binary.BigEndian.Uint32(c[:4]))
	c = c[4:]
	l, n := ReadVarint(c)
	if n < 0 {
		return indexInteriorCell{}, ErrCorrupted
	}
	pl, err := parsePayload(l, c[n:], pageSize, ((pageSize-12)*64/255)-23)
	return indexInteriorCell{
		left:    int(left),
		payload: pl,
	}, err
}

// Parse the list of pointers to cells into byte offsets for each cell
// This format is used in all four page types.
// N is the nr of cells, pointers point to the start of the cells, until end of
// the page, maxLen is the length of the page (because cell pointers use page
// offsets).
func parseCellpointers(
	n int,
	pointers []byte,
	maxLen int,
) ([]int, error) {
	if len(pointers) < n*2 {
		return nil, errors.New("invalid cell pointer array")
	}
	cs := make([]int, n)
	// cell pointers go [p1, p2, p3], actual cell content can be in any order.
	for i := range cs {
		start := int(binary.BigEndian.Uint16(pointers[2*i : 2*i+2]))
		if start > maxLen {
			return nil, errors.New("invalid cell pointer")
		}
		cs[i] = start
	}
	return cs, nil
}
