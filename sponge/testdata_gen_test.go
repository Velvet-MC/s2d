package sponge

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

	// single_stone.schem: 1x1x1, palette {minecraft:stone: 0}, BlockData [0x00]
	{
		root := map[string]any{
			"Version":     int32(2),
			"DataVersion": int32(2975), // 1.18.1; arbitrary but valid
			"Width":       int16(1),
			"Height":      int16(1),
			"Length":      int16(1),
			"PaletteMax":  int32(1),
			"Palette":     map[string]int32{"minecraft:stone": 0},
			"BlockData":   []byte{0x00},
		}
		writeFixture(t, filepath.Join(dir, "single_stone.schem"), root)
	}

	// multi_palette.schem: 4x1x4, four palette entries, varint-rich BlockData.
	{
		palette := map[string]int32{
			"minecraft:stone":         0,
			"minecraft:dirt":          1,
			"minecraft:oak_planks":    200, // forces multi-byte varint
			"minecraft:diamond_block": 3,
		}
		// 16 cells; cycle through indices 0,1,200,3,...
		var data bytes.Buffer
		seq := []uint32{0, 1, 200, 3, 0, 1, 200, 3, 0, 1, 200, 3, 0, 1, 200, 3}
		for _, v := range seq {
			writeVarintBuf(&data, v)
		}
		root := map[string]any{
			"Version":     int32(2),
			"DataVersion": int32(2975),
			"Width":       int16(4),
			"Height":      int16(1),
			"Length":      int16(4),
			"PaletteMax":  int32(201),
			"Palette":     palette,
			"BlockData":   data.Bytes(),
		}
		writeFixture(t, filepath.Join(dir, "multi_palette.schem"), root)
	}

	// waterlogged_stairs.schem: 1x1x1, oak_stairs with waterlogged=true.
	{
		root := map[string]any{
			"Version":     int32(2),
			"DataVersion": int32(2975),
			"Width":       int16(1),
			"Height":      int16(1),
			"Length":      int16(1),
			"PaletteMax":  int32(1),
			"Palette": map[string]int32{
				"minecraft:oak_stairs[facing=north,half=bottom,shape=straight,waterlogged=true]": 0,
			},
			"BlockData": []byte{0x00},
		}
		writeFixture(t, filepath.Join(dir, "waterlogged_stairs.schem"), root)
	}

	// unknown_block.schem: 1x1x1, a fictitious mod block.
	{
		root := map[string]any{
			"Version":     int32(2),
			"DataVersion": int32(2975),
			"Width":       int16(1),
			"Height":      int16(1),
			"Length":      int16(1),
			"PaletteMax":  int32(1),
			"Palette":     map[string]int32{"mod:foo[bar=baz]": 0},
			"BlockData":   []byte{0x00},
		}
		writeFixture(t, filepath.Join(dir, "unknown_block.schem"), root)
	}
}

// writeFixture encodes root as big-endian NBT, gzips it, and writes to path.
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

// writeVarintBuf writes a Mojang-style LEB128 varint into buf.
func writeVarintBuf(buf *bytes.Buffer, v uint32) {
	for {
		if v < 0x80 {
			buf.WriteByte(byte(v))
			return
		}
		buf.WriteByte(byte(v&0x7F | 0x80))
		v >>= 7
	}
}
