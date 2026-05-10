package legacy

import (
	"os"
	"testing"

	_ "github.com/df-mc/dragonfly/server/block" // register vanilla blocks for missing-block resolution
	"github.com/Clxser/S2D/schem"
)

func TestRead_Basic(t *testing.T) {
	f, err := os.Open("testdata/basic.schematic")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	s, err := Read(f)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if s.Format != schem.FormatLegacy {
		t.Errorf("format: %q want %q", s.Format, schem.FormatLegacy)
	}
	if s.Width != 2 || s.Height != 1 || s.Length != 1 {
		t.Fatalf("dims: %dx%dx%d", s.Width, s.Height, s.Length)
	}
	if len(s.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(s.Blocks))
	}
	if s.Blocks[0].Pos != [3]int{0, 0, 0} || s.Blocks[1].Pos != [3]int{1, 0, 0} {
		t.Errorf("positions: %v %v", s.Blocks[0].Pos, s.Blocks[1].Pos)
	}
	// Both cells are stone; canonical = "minecraft:stone".
	// Translate stub returns Recognized=false for every key, so both go into Unknowns.
	if s.Unknowns.Total != 2 {
		t.Errorf("expected 2 unknowns (stub translate), got %d: %v",
			s.Unknowns.Total, s.Unknowns.Counts)
	}
	if s.Unknowns.Counts["minecraft:stone"] != 2 {
		t.Errorf("expected 2x minecraft:stone in unknowns, got %v", s.Unknowns.Counts)
	}
}

func TestRead_AddBlocks(t *testing.T) {
	f, err := os.Open("testdata/addblocks.schematic")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	s, err := Read(f)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	// id=256, data=0. legacy.json may or may not have an entry for 256:0
	// (it's an unusual modded ID). Whether the cell ends up as a known
	// translation or as a "legacy:256:0" unknown, the parser itself
	// must not error.
	if s.Width != 1 || s.Height != 1 || s.Length != 1 {
		t.Fatalf("dims: %dx%dx%d", s.Width, s.Height, s.Length)
	}
	if len(s.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(s.Blocks))
	}
	t.Logf("addblocks Unknowns: %v", s.Unknowns.Counts)
	// Either a real Java state string or "legacy:256:0" must appear.
	if len(s.Unknowns.Counts) == 0 {
		t.Errorf("expected at least one unknown entry, got empty: %+v", s.Unknowns)
	}
}
