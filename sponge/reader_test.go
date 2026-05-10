package sponge

import (
	"os"
	"testing"

	"github.com/Clxser/S2D/schem"
)

func TestRead_SingleStone(t *testing.T) {
	f, err := os.Open("testdata/single_stone.schem")
	if err != nil { t.Fatal(err) }
	defer f.Close()

	s, err := Read(f)
	if err != nil { t.Fatalf("Read: %v", err) }

	if s.Format != schem.FormatSpongeV2 {
		t.Errorf("format: %q", s.Format)
	}
	if s.Width != 1 || s.Height != 1 || s.Length != 1 {
		t.Fatalf("dims: %dx%dx%d", s.Width, s.Height, s.Length)
	}
	if len(s.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(s.Blocks))
	}
	if s.Blocks[0].Pos != [3]int{0, 0, 0} {
		t.Errorf("pos: %v", s.Blocks[0].Pos)
	}
	// With the translate stub, every palette entry should appear in Unknowns
	// because no real translation exists yet. Real name assertions land in Phase 7.5.
	if s.Unknowns.Total != 1 {
		t.Errorf("expected 1 unknown (stub translate), got %d", s.Unknowns.Total)
	}
	if _, ok := s.Unknowns.Counts["minecraft:stone"]; !ok {
		t.Errorf("expected minecraft:stone in unknowns, got %v", s.Unknowns.Counts)
	}
}

func TestRead_MultiPalette(t *testing.T) {
	f, err := os.Open("testdata/multi_palette.schem")
	if err != nil { t.Fatal(err) }
	defer f.Close()

	s, err := Read(f)
	if err != nil { t.Fatalf("Read: %v", err) }

	if s.Format != schem.FormatSpongeV2 { t.Errorf("format: %q", s.Format) }
	if s.Width != 4 || s.Height != 1 || s.Length != 4 {
		t.Fatalf("dims: %dx%dx%d", s.Width, s.Height, s.Length)
	}
	if len(s.Blocks) != 16 { t.Fatalf("expected 16 blocks, got %d", len(s.Blocks)) }

	// Verify YZX iteration order: cells [0..3] should have x=0..3, all y=z=0.
	for x := 0; x < 4; x++ {
		want := [3]int{x, 0, 0}
		if s.Blocks[x].Pos != want {
			t.Errorf("Blocks[%d].Pos = %v, want %v", x, s.Blocks[x].Pos, want)
		}
	}

	// Total unknowns should be 16 (all cells, since translate is stub).
	if s.Unknowns.Total != 16 {
		t.Errorf("expected 16 unknowns, got %d", s.Unknowns.Total)
	}
	// Palette had 4 entries: stone, dirt, oak_planks, diamond_block.
	if len(s.Unknowns.Counts) != 4 {
		t.Errorf("expected 4 distinct palette keys, got %d: %v",
			len(s.Unknowns.Counts), s.Unknowns.Counts)
	}
	// oak_planks at index 200 (multi-byte varint) should have been decoded.
	if _, ok := s.Unknowns.Counts["minecraft:oak_planks"]; !ok {
		t.Errorf("oak_planks missing from unknowns (varint multi-byte broken?): %v",
			s.Unknowns.Counts)
	}
}
