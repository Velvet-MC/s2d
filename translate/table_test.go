package translate

import (
	"testing"

	_ "github.com/df-mc/dragonfly/server/block" // register vanilla blocks
	"github.com/df-mc/dragonfly/server/world"
)

func TestPaletteCoverage(t *testing.T) {
	tableOnce.Do(buildTable)
	if tableErr != nil {
		t.Fatalf("buildTable: %v", tableErr)
	}
	if len(table) < 1000 {
		t.Errorf("translate table only has %d entries; expected >1000", len(table))
	}
}

func TestLookup_KnownStone(t *testing.T) {
	res := Lookup("minecraft:stone")
	if !res.Recognized {
		t.Fatalf("stone not recognized; have %d table entries", len(table))
	}
	name, _ := res.Block.EncodeBlock()
	if name != "minecraft:stone" {
		t.Errorf("stone resolved to %q", name)
	}
}

func TestLookup_OakLogAxis(t *testing.T) {
	res := Lookup("minecraft:oak_log[axis=y]")
	if !res.Recognized {
		t.Errorf("oak_log[axis=y] not recognized")
	}
	res = Lookup("minecraft:oak_log[axis=x]")
	if !res.Recognized {
		t.Errorf("oak_log[axis=x] not recognized")
	}
}

func TestLookup_Leaves(t *testing.T) {
	for _, key := range []string{
		"minecraft:oak_leaves[distance=1,persistent=false]",
		"minecraft:oak_leaves[distance=7,persistent=true]",
		"minecraft:spruce_leaves[distance=3,persistent=false]",
		"minecraft:jungle_leaves[distance=4,persistent=true]",
		"minecraft:azalea_leaves[distance=4,persistent=false]",
		"minecraft:flowering_azalea_leaves[distance=4,persistent=false]",
	} {
		t.Run(key, func(t *testing.T) {
			res := Lookup(key)
			if !res.Recognized {
				t.Fatalf("%s not recognized", key)
			}
			name, _ := res.Block.EncodeBlock()
			if name == "minecraft:magenta_wool" {
				t.Fatalf("%s fell back to missing-block marker", key)
			}
		})
	}
}

func TestLookup_ShortGrass(t *testing.T) {
	res := Lookup("minecraft:grass")
	if !res.Recognized {
		t.Fatal("minecraft:grass not recognized")
	}
	name, _ := res.Block.EncodeBlock()
	if name != "minecraft:short_grass" {
		t.Fatalf("minecraft:grass -> %s, want minecraft:short_grass", name)
	}
}

func TestLookup_CommonArenaBlocks(t *testing.T) {
	tests := []struct {
		key      string
		wantName string
	}{
		{
			key:      "minecraft:stone_brick_wall[east=low,north=none,south=tall,up=true,waterlogged=false,west=none]",
			wantName: "minecraft:stone_brick_wall",
		},
		{
			key:      "minecraft:cobblestone_stairs[facing=north,half=bottom,shape=straight,waterlogged=false]",
			wantName: "minecraft:stone_stairs",
		},
		{
			key:      "minecraft:oak_door[facing=north,half=lower,hinge=right,open=false,powered=false]",
			wantName: "minecraft:wooden_door",
		},
		{
			key:      "minecraft:oak_trapdoor[facing=west,half=top,open=true,powered=false,waterlogged=false]",
			wantName: "minecraft:trapdoor",
		},
		{
			key:      "minecraft:note_block[instrument=harp,note=0,powered=false]",
			wantName: "minecraft:noteblock",
		},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			res := Lookup(tt.key)
			if !res.Recognized {
				t.Fatalf("%s not recognized", tt.key)
			}
			name, _ := res.Block.EncodeBlock()
			if name != tt.wantName {
				t.Fatalf("%s -> %s, want %s", tt.key, name, tt.wantName)
			}
		})
	}
}

func TestLookup_Waterlogged(t *testing.T) {
	res := Lookup("minecraft:oak_stairs[facing=north,half=bottom,shape=straight,waterlogged=true]")
	if res.Liquid == nil {
		t.Errorf("waterlogged stairs should produce a Liquid")
	}
}

func TestLookup_Unknown(t *testing.T) {
	res := Lookup("mod:nonexistent[bar=baz]")
	if res.Recognized {
		t.Errorf("unknown block should not be Recognized")
	}
	if res.Block == nil {
		t.Errorf("unknown block must still have non-nil Block (missing fallback)")
	}
}

func TestMissingBlockRegistered(t *testing.T) {
	b := resolveDefaultMissing()
	if b == nil {
		t.Fatalf("no fallback block registered; check Dragonfly version")
	}
	name, _ := b.EncodeBlock()
	t.Logf("default missing block: %s", name)
}

func TestLookupTreatsDragonflyPlaceholdersAsUnknown(t *testing.T) {
	direct, ok := world.BlockByName("minecraft:structure_void", nil)
	if !ok {
		t.Skip("Dragonfly does not know minecraft:structure_void")
	}
	if nbtBlock, ok := direct.(world.NBTer); !ok || nbtBlock.EncodeNBT() != nil {
		t.Skip("Dragonfly structure_void is implemented in this version")
	}

	res := Lookup("minecraft:structure_void")
	if res.Recognized {
		t.Fatal("structure_void resolved to a Dragonfly placeholder but was reported as recognized")
	}
	if nbtBlock, ok := res.Block.(world.NBTer); ok && nbtBlock.EncodeNBT() == nil {
		t.Fatal("placeholder block with nil NBT leaked through instead of missing-block fallback")
	}
}
