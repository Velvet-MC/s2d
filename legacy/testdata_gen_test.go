package legacy

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func TestGenerateFixtures(t *testing.T) {
	if os.Getenv("S2D_REGEN_FIXTURES") != "1" {
		t.Skip("set S2D_REGEN_FIXTURES=1 to regenerate")
	}
	dir := "testdata"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// basic.schematic: 2x1x1, two stone cells (id=1, data=0)
	{
		root := map[string]any{
			"Materials": "Alpha",
			"Width":     int16(2),
			"Height":    int16(1),
			"Length":    int16(1),
			"Blocks":    []byte{0x01, 0x01}, // id=1 stone, both cells
			"Data":      []byte{0x00},        // both nibbles 0; one byte covers two cells
		}
		writeFixture(t, filepath.Join(dir, "basic.schematic"), root)
	}

	// addblocks.schematic: 1x1x1, id=256 (out of low-byte range), data=0
	// id=256 → high4=1, low8=0. AddBlocks holds the high nibble:
	// cell index 0 is even → high nibble of AddBlocks[0] is the id high4.
	// For one cell with high4=1: AddBlocks[0] = 0x10 (high nibble 1, low nibble unused).
	{
		root := map[string]any{
			"Materials": "Alpha",
			"Width":     int16(1),
			"Height":    int16(1),
			"Length":    int16(1),
			"Blocks":    []byte{0x00},
			"AddBlocks": []byte{0x10},
			"Data":      []byte{0x00},
		}
		writeFixture(t, filepath.Join(dir, "addblocks.schematic"), root)
	}
}

func writeFixture(t *testing.T, path string, root map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	enc := nbt.NewEncoderWithEncoding(&buf, nbt.BigEndian)
	if err := enc.Encode(root); err != nil {
		t.Fatal(err)
	}
	var gz bytes.Buffer
	gw := gzip.NewWriter(&gz)
	if _, err := gw.Write(buf.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, gz.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}
