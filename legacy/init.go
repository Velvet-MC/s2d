package legacy

import (
	"bytes"
	"io"

	"github.com/Velvet-MC/s2d/schem"
)

func init() {
	schem.Register(schem.FormatHandler{
		Name:       schem.FormatLegacy,
		Extensions: []string{".schematic"},
		Signature:  signatureMatch,
		Read: func(r io.Reader) (*schem.Schematic, error) {
			return Read(r)
		},
		Scan: func(r io.Reader, onInfo schem.InfoHandler, yield schem.BlockHandler) (schem.ScanInfo, error) {
			return ScanWithInfo(r, onInfo, yield)
		},
	})
}

func signatureMatch(headerPeek []byte) bool {
	return bytes.Contains(headerPeek, []byte("Materials"))
}
