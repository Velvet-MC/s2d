package sponge

import (
	"bytes"
	"io"

	"github.com/Velvet-MC/s2d/schem"
)

func init() {
	schem.Register(schem.FormatHandler{
		Name:       schem.FormatSpongeV2,
		Extensions: []string{".schem"},
		Signature:  signatureMatch,
		Read: func(r io.Reader) (*schem.Schematic, error) {
			return Read(r)
		},
		Scan: func(r io.Reader, onInfo schem.InfoHandler, yield schem.BlockHandler) (schem.ScanInfo, error) {
			return ScanWithInfo(r, onInfo, yield)
		},
	})
}

// signatureMatch returns true if the decompressed header looks like Sponge v2.
// Cheap heuristic: NBT root contains the bytes "Version" near the start.
func signatureMatch(headerPeek []byte) bool {
	return bytes.Contains(headerPeek, []byte("Version"))
}
