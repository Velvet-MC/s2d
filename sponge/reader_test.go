package sponge

import (
	"os"
	"reflect"
	"testing"

	"github.com/Velvet-MC/s2d/palette"
	_ "github.com/df-mc/dragonfly/server/block" // register vanilla blocks for translation

	"github.com/Velvet-MC/s2d/schem"
)

func TestRead_SingleStone(t *testing.T) {
	f, err := os.Open("testdata/single_stone.schem")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	s, err := Read(f)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

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
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	s, err := Read(f)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if s.Format != schem.FormatSpongeV2 {
		t.Errorf("format: %q", s.Format)
	}
	if s.Width != 4 || s.Height != 1 || s.Length != 4 {
		t.Fatalf("dims: %dx%dx%d", s.Width, s.Height, s.Length)
	}
	if len(s.Blocks) != 16 {
		t.Fatalf("expected 16 blocks, got %d", len(s.Blocks))
	}

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

func TestScanDecodesEachPaletteEntryOnce(t *testing.T) {
	f, err := os.Open("testdata/multi_palette.schem")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	original := decodePaletteKey
	defer func() { decodePaletteKey = original }()
	calls := 0
	decodePaletteKey = func(key string) (palette.JavaState, error) {
		calls++
		return original(key)
	}

	var blocks int
	if _, err := Scan(f, func(schem.Block) error {
		blocks++
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if blocks != 16 {
		t.Fatalf("blocks = %d, want 16", blocks)
	}
	if calls != 4 {
		t.Fatalf("palette decode calls = %d, want one per palette entry (4)", calls)
	}
}

func TestScanReportsPaletteIndexes(t *testing.T) {
	f, err := os.Open("testdata/multi_palette.schem")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	var infoPaletteSize int
	var got []uint32
	var states []string
	info, err := ScanWithInfo(f, func(info schem.ScanInfo) error {
		infoPaletteSize = info.PaletteSize
		return nil
	}, func(b schem.Block) error {
		if !b.PaletteIndexOK {
			t.Fatalf("block at %v did not report palette index", b.Pos)
		}
		got = append(got, b.PaletteIndex)
		states = append(states, b.BedrockState.Name)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if info.PaletteSize != 201 || infoPaletteSize != 201 {
		t.Fatalf("palette size = %d/%d, want 201", info.PaletteSize, infoPaletteSize)
	}
	want := []uint32{0, 1, 200, 3}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("palette index[%d] = %d, want %d", i, got[i], w)
		}
	}
	for i, state := range states {
		if state == "" {
			t.Fatalf("bedrock state[%d] was empty", i)
		}
	}
}

func TestBannerBlockEntitiesKeepsPatterns(t *testing.T) {
	patterns := []any{map[string]any{"Pattern": "bs", "Color": int32(14)}}
	got := bannerBlockEntities([]any{
		map[string]any{
			"Id":       "minecraft:banner",
			"Pos":      [3]int32{1, 2, 3},
			"Patterns": patterns,
		},
		map[string]any{
			"Id":  "minecraft:chest",
			"Pos": [3]int32{4, 5, 6},
		},
	})
	if len(got) != 1 {
		t.Fatalf("bannerBlockEntities len = %d, want 1", len(got))
	}
	nbt := got[[3]int{1, 2, 3}]
	if nbt["id"] != "Banner" {
		t.Fatalf("id = %#v, want Banner", nbt["id"])
	}
	if !reflect.DeepEqual(nbt["Patterns"], patterns) {
		t.Fatalf("Patterns not preserved: %#v", nbt["Patterns"])
	}
}
