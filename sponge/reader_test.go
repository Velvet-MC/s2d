package sponge

import (
	"os"
	"testing"

	_ "github.com/df-mc/dragonfly/server/block" // register vanilla blocks for translation

	"github.com/Clxser/S2D/schem"
)

func TestRead_SingleStone(t *testing.T) {
	f, err := os.Open("testdata/single_stone.schem")
	if err != nil { t.Fatal(err) }
	defer f.Close()

	s, err := Read(f)
	if err != nil { t.Fatalf("Read: %v", err) }

	if s.Format != schem.FormatSpongeV2 {
		t.Errorf("format: %q want %q", s.Format, schem.FormatSpongeV2)
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
	if s.Blocks[0].Block == nil {
		t.Fatal("Block is nil")
	}
	name, _ := s.Blocks[0].Block.EncodeBlock()
	if name != "minecraft:stone" {
		t.Errorf("got %q want minecraft:stone", name)
	}
	if s.Unknowns.Total != 0 {
		t.Errorf("expected 0 unknowns for vanilla stone, got %d: %v",
			s.Unknowns.Total, s.Unknowns.Counts)
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

	// YZX iteration: cells [0..3] should have x=0..3, all y=z=0.
	for x := 0; x < 4; x++ {
		want := [3]int{x, 0, 0}
		if s.Blocks[x].Pos != want {
			t.Errorf("Blocks[%d].Pos = %v, want %v", x, s.Blocks[x].Pos, want)
		}
	}

	// Cell index sequence: 0,1,200,3 = stone, dirt, oak_planks, diamond_block.
	// First row gives us all four palette entries.
	wantNames := []string{
		"minecraft:stone",
		"minecraft:dirt",
		"minecraft:oak_planks",
		"minecraft:diamond_block",
	}
	for i, want := range wantNames {
		if s.Blocks[i].Block == nil {
			t.Fatalf("Blocks[%d].Block is nil", i)
		}
		got, _ := s.Blocks[i].Block.EncodeBlock()
		if got != want {
			t.Errorf("Blocks[%d]: got %q want %q", i, got, want)
		}
	}
	if s.Unknowns.Total != 0 {
		t.Errorf("expected 0 unknowns for vanilla blocks, got %d: %v",
			s.Unknowns.Total, s.Unknowns.Counts)
	}
}
