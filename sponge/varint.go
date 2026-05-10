package sponge

import (
	"fmt"
	"io"
)

// readVarint reads a Mojang-style LEB128 varint (7-bit groups, MSB
// continuation flag, little-endian) from r. Returns the decoded uint32
// value, the number of bytes consumed, and an error.
//
// Sponge Schematic v2 BlockData uses this encoding for palette indices.
// It is byte-compatible with Java protocol varints.
func readVarint(r io.ByteReader) (uint32, int, error) {
	var (
		value uint32
		shift uint
		n     int
	)
	for n < 5 {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF && n > 0 {
				return 0, n, io.ErrUnexpectedEOF
			}
			return 0, n, err
		}
		n++
		value |= uint32(b&0x7F) << shift
		if b&0x80 == 0 {
			return value, n, nil
		}
		shift += 7
	}
	return 0, n, fmt.Errorf("sponge: varint exceeds 5 bytes")
}
