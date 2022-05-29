package bakelite

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/alicebob/bakelite/internal"
)

// makeRecord encodes a record (row) as bytes.
func makeRecord(row []any) ([]byte, error) {
	var (
		header = make([]byte, 1000)
		p      = 0 // where we are in header[p]
		body   = &bytes.Buffer{}
	)
	for _, col := range row {
		switch v := col.(type) {
		case int:
			switch {
			case v == 0:
				p += internal.PutUvarint(header[p:], 8)
			case v == 1:
				p += internal.PutUvarint(header[p:], 9)
			case v >= math.MinInt8 && v <= math.MaxInt8:
				p += internal.PutUvarint(header[p:], 1)
				if err := binary.Write(body, binary.BigEndian, int8(v)); err != nil {
					return nil, err
				}
			case v >= math.MinInt16 && v <= math.MaxInt16:
				p += internal.PutUvarint(header[p:], 2)
				if err := binary.Write(body, binary.BigEndian, int16(v)); err != nil {
					return nil, err
				}
				// skipped Int24 (type 3)
			case v >= math.MinInt32 && v <= math.MaxInt32:
				p += internal.PutUvarint(header[p:], 4)
				if err := binary.Write(body, binary.BigEndian, int32(v)); err != nil {
					return nil, err
				}
				// skipped Int48 (type 5)
			default:
				p += internal.PutUvarint(header[p:], 6)
				if err := binary.Write(body, binary.BigEndian, int64(v)); err != nil {
					return nil, err
				}
			}
		case string:
			l := 13 + 2*len(v)
			p += internal.PutUvarint(header[p:], uint64(l))
			body.WriteString(v)
		case nil:
			p += internal.PutUvarint(header[p:], 0)
		default:
			return nil, fmt.Errorf("unsupported type (%T)", col)
		}
	}
	ret := make([]byte, 1+p+body.Len())
	if n := internal.PutUvarint(ret, uint64(p+1)); n != 1 {
		// Header length is the length _including_ our varint. So the length
		// varint size depends on the value including itself. No idea how
		// you're supposed to calculate that.
		panic("record header was too big")
	}
	copy(ret[1:], header[:p])
	copy(ret[1+p:], body.Bytes())
	return ret, nil
}
