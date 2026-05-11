package schem_test

import (
	"os"
	"testing"

	_ "github.com/Velvet-MC/s2d/legacy"         // register legacy handler via init()
	_ "github.com/Velvet-MC/s2d/sponge"         // register sponge handler via init()
	_ "github.com/df-mc/dragonfly/server/block" // register vanilla Bedrock blocks

	"github.com/Velvet-MC/s2d/schem"
)

func TestEndToEnd_SpongeSingleStone(t *testing.T) {
	f, err := os.Open("testdata/single_stone.schem")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	s, err := schem.Read(f.Name(), f)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if s.Format != schem.FormatSpongeV2 {
		t.Errorf("format: %q", s.Format)
	}
	if len(s.Blocks) != 1 {
		t.Fatalf("blocks: %d", len(s.Blocks))
	}
	if s.Blocks[0].Block == nil {
		t.Fatalf("Block is nil")
	}
	name, _ := s.Blocks[0].Block.EncodeBlock()
	if name != "minecraft:stone" {
		t.Errorf("name: %q want minecraft:stone", name)
	}
}

func TestEndToEnd_SpongeMultiPalette(t *testing.T) {
	f, err := os.Open("testdata/multi_palette.schem")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	s, err := schem.Read(f.Name(), f)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if s.Format != schem.FormatSpongeV2 {
		t.Errorf("format: %q", s.Format)
	}
	if len(s.Blocks) != 16 {
		t.Fatalf("blocks: %d", len(s.Blocks))
	}
	// First four cells exercise all four palette entries via the multi-byte varint.
	wantNames := []string{
		"minecraft:stone",
		"minecraft:dirt",
		"minecraft:oak_planks",
		"minecraft:diamond_block",
	}
	for i, want := range wantNames {
		if s.Blocks[i].Block == nil {
			t.Fatalf("Blocks[%d] nil", i)
		}
		got, _ := s.Blocks[i].Block.EncodeBlock()
		if got != want {
			t.Errorf("Blocks[%d]: got %q want %q", i, got, want)
		}
	}
}

func TestScan_SpongeMultiPalette(t *testing.T) {
	f, err := os.Open("testdata/multi_palette.schem")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	var blocks []schem.Block
	info, err := schem.Scan(f.Name(), f, func(b schem.Block) error {
		blocks = append(blocks, b)
		return nil
	})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if info.Format != schem.FormatSpongeV2 {
		t.Errorf("format: %q", info.Format)
	}
	if info.Width != 4 || info.Height != 1 || info.Length != 4 {
		t.Errorf("dimensions: %dx%dx%d, want 4x1x4", info.Width, info.Height, info.Length)
	}
	if len(blocks) != 16 {
		t.Fatalf("blocks: %d", len(blocks))
	}
	wantNames := []string{
		"minecraft:stone",
		"minecraft:dirt",
		"minecraft:oak_planks",
		"minecraft:diamond_block",
	}
	for i, want := range wantNames {
		got, _ := blocks[i].Block.EncodeBlock()
		if got != want {
			t.Errorf("blocks[%d]: got %q want %q", i, got, want)
		}
	}
}

func TestEndToEnd_LegacyBasic(t *testing.T) {
	f, err := os.Open("testdata/basic.schematic")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	s, err := schem.Read(f.Name(), f)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if s.Format != schem.FormatLegacy {
		t.Errorf("format: %q", s.Format)
	}
	if len(s.Blocks) != 2 {
		t.Fatalf("blocks: %d", len(s.Blocks))
	}
	for i, b := range s.Blocks {
		if b.Block == nil {
			t.Fatalf("Blocks[%d] nil", i)
		}
		name, _ := b.Block.EncodeBlock()
		if name != "minecraft:stone" {
			t.Errorf("Blocks[%d]: got %q want minecraft:stone", i, name)
		}
	}
}

func TestScan_LegacyBasic(t *testing.T) {
	f, err := os.Open("testdata/basic.schematic")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	var got int
	info, err := schem.Scan(f.Name(), f, func(b schem.Block) error {
		got++
		if b.Block == nil {
			t.Fatalf("block %d nil", got)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if info.Format != schem.FormatLegacy {
		t.Errorf("format: %q", info.Format)
	}
	if got != 2 {
		t.Fatalf("blocks: %d, want 2", got)
	}
}
