package legacy

import (
	"bytes"
	"io"

	"github.com/Clxser/S2D/schem"
)

func init() {
	schem.Register(schem.FormatHandler{
		Name:       schem.FormatLegacy,
		Extensions: []string{".schematic"},
		Signature:  signatureMatch,
		Read: func(r io.Reader) (*schem.Schematic, error) {
			return Read(r)
		},
	})
}

func signatureMatch(headerPeek []byte) bool {
	return bytes.Contains(headerPeek, []byte("Materials"))
}
