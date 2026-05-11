package legacy

import (
	"os"
	"testing"

	_ "github.com/df-mc/dragonfly/server/block" // register vanilla blocks for translation

	"github.com/Velvet-MC/s2d/schem"
)

func TestRead_Basic(t *testing.T) {
	f, err := os.Open("testdata/basic.schematic")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

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
	// Both cells are stone (id=1, data=0). With real translate, both resolve.
	for i, b := range s.Blocks {
		if b.Block == nil {
			t.Fatalf("Blocks[%d].Block is nil", i)
		}
		name, _ := b.Block.EncodeBlock()
		if name != "minecraft:stone" {
			t.Errorf("Blocks[%d]: got %q want minecraft:stone", i, name)
		}
	}
	if s.Unknowns.Total != 0 {
		t.Errorf("expected 0 unknowns for vanilla stone, got %d: %v",
			s.Unknowns.Total, s.Unknowns.Counts)
	}
}

func TestRead_AddBlocks(t *testing.T) {
	f, err := os.Open("testdata/addblocks.schematic")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	s, err := Read(f)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	// id=256 is not in legacy.json (modded territory). The reader should
	// still parse cleanly; the cell ends up in Unknowns under "legacy:256:0"
	// and the block holds the missing fallback.
	if s.Width != 1 || s.Height != 1 || s.Length != 1 {
		t.Fatalf("dims: %dx%dx%d", s.Width, s.Height, s.Length)
	}
	if len(s.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(s.Blocks))
	}
	t.Logf("addblocks Unknowns: %v", s.Unknowns.Counts)
	if s.Unknowns.Counts["legacy:256:0"] != 1 {
		t.Errorf("expected 1x legacy:256:0 in unknowns, got %v", s.Unknowns.Counts)
	}
}
