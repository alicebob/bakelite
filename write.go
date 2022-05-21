package bakelite

import (
	"bytes"
	"encoding/binary"
)

const (
	pageSize = 1 << 12 // 16
)

var (
	headerMagic = "SQLite format 3\x00"
)

func header(pageCount int) []byte {
	ps := pageSize
	if ps == 1<<16 {
		ps = 1
	}
	// the file header, as described in "1.2. The Database Header"
	h := struct {
		Magic                [16]byte
		PageSize             uint16
		WriteVersion         uint8
		ReadVersion          uint8
		ReservedSpace        uint8
		MaxFraction          uint8
		MinFraction          uint8
		LeafFraction         uint8
		ChangeCounter        uint32
		PageCount            uint32
		FirstFreelist        uint32
		FreelistCount        uint32
		SchemaCookie         uint32
		SchemaFormat         uint32
		PageCacheSize        uint32
		_                    uint32
		TextEncoding         uint32
		_                    uint32
		_                    uint32
		_                    uint32
		ReservedForExpansion [20]byte
		VersionValidFor      uint32
		SQLiteVersion        uint32
	}{
		Magic:           asByte(headerMagic),
		PageSize:        uint16(ps),
		WriteVersion:    1, // "journal". "2" is WAL, but sqlittle doesn't read those
		ReadVersion:     1, // "journal"
		ReservedSpace:   0,
		MaxFraction:     64,
		MinFraction:     32,
		LeafFraction:    32,
		ChangeCounter:   42,
		PageCount:       uint32(pageCount),
		FirstFreelist:   0,
		FreelistCount:   0,
		SchemaCookie:    1, // we don't change the schema
		SchemaFormat:    4,
		PageCacheSize:   0,
		TextEncoding:    1,  // "UTF-8"
		VersionValidFor: 42, // must match ChangeCounter
		SQLiteVersion:   0,  // ??
	}
	b := &bytes.Buffer{}
	binary.Write(b, binary.BigEndian, h)
	return b.Bytes()
}

func asByte(s string) [16]byte {
	var r [16]byte
	copy(r[:], s)
	return r
}
