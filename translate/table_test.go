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
			key:      "minecraft:gray_glazed_terracotta[facing=west]",
			wantName: "minecraft:gray_glazed_terracotta",
		},
		{
			key:      "minecraft:blue_glazed_terracotta[facing=west]",
			wantName: "minecraft:blue_glazed_terracotta",
		},
		{
			key:      "minecraft:bedrock",
			wantName: "minecraft:bedrock",
		},
		{
			key:      "minecraft:terracotta",
			wantName: "minecraft:hardened_clay",
		},
		{
			key:      "minecraft:smooth_quartz",
			wantName: "minecraft:smooth_quartz",
		},
		{
			key:      "minecraft:polished_diorite_slab[type=bottom,waterlogged=false]",
			wantName: "minecraft:polished_diorite_slab",
		},
		{
			key:      "minecraft:smooth_stone_slab[type=bottom,waterlogged=false]",
			wantName: "minecraft:smooth_stone_slab",
		},
		{
			key:      "minecraft:stone_slab[type=double,waterlogged=false]",
			wantName: "minecraft:smooth_stone_double_slab",
		},
		{
			key:      "minecraft:deepslate_brick_slab[type=top,waterlogged=false]",
			wantName: "minecraft:deepslate_brick_slab",
		},
		{
			key:      "minecraft:redstone_ore[lit=false]",
			wantName: "minecraft:redstone_ore",
		},
		{
			key:      "minecraft:deepslate_redstone_ore[lit=false]",
			wantName: "minecraft:deepslate_redstone_ore",
		},
		{
			key:      "minecraft:cobweb",
			wantName: "minecraft:web",
		},
		{
			key:      "minecraft:dead_bush",
			wantName: "minecraft:deadbush",
		},
		{
			key:      "minecraft:magma_block",
			wantName: "minecraft:magma",
		},
		{
			key:      "minecraft:bricks",
			wantName: "minecraft:brick_block",
		},
		{
			key:      "minecraft:wall_torch[facing=west]",
			wantName: "minecraft:torch",
		},
		{
			key:      "minecraft:chest[facing=south,type=single,waterlogged=false]",
			wantName: "minecraft:chest",
		},
		{
			key:      "minecraft:chain[axis=y,waterlogged=false]",
			wantName: "minecraft:iron_chain",
		},
		{
			key:      "minecraft:lantern[hanging=true,waterlogged=false]",
			wantName: "minecraft:lantern",
		},
		{
			key:      "minecraft:glow_lichen[down=false,east=true,north=false,south=false,up=false,waterlogged=false,west=false]",
			wantName: "minecraft:glow_lichen",
		},
		{
			key:      "minecraft:spawner",
			wantName: "minecraft:mob_spawner",
		},
		{
			key:      "minecraft:bubble_column[drag=true]",
			wantName: "minecraft:water",
		},
		{
			key:      "minecraft:birch_sapling[stage=1]",
			wantName: "minecraft:short_grass",
		},
		{
			key:      "minecraft:anvil[facing=north]",
			wantName: "minecraft:anvil",
		},
		{
			key:      "minecraft:grindstone[face=ceiling,facing=west]",
			wantName: "minecraft:grindstone",
		},
		{
			key:      "minecraft:cartography_table",
			wantName: "minecraft:cartography_table",
		},
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

func TestLookup_PrismBaroqueMissingStates(t *testing.T) {
	tests := []struct {
		key       string
		wantName  string
		wantProps map[string]any
	}{
		{key: "minecraft:end_stone_bricks", wantName: "minecraft:end_bricks"},
		{key: "minecraft:light[level=13,waterlogged=false]", wantName: "minecraft:light_block_13"},
		{key: "minecraft:light[level=15,waterlogged=false]", wantName: "minecraft:light_block_15"},
		{key: "minecraft:sugar_cane[age=0]", wantName: "minecraft:reeds", wantProps: map[string]any{"age": int32(0)}},
		{key: "minecraft:rooted_dirt", wantName: "minecraft:dirt_with_roots"},
		{key: "minecraft:nether_portal[axis=z]", wantName: "minecraft:portal", wantProps: map[string]any{"portal_axis": "z"}},
		{key: "minecraft:nether_portal[axis=x]", wantName: "minecraft:portal", wantProps: map[string]any{"portal_axis": "x"}},
		{key: "minecraft:tall_seagrass[half=upper]", wantName: "minecraft:seagrass", wantProps: map[string]any{"sea_grass_type": "double_top"}},
		{key: "minecraft:tall_seagrass[half=lower]", wantName: "minecraft:seagrass", wantProps: map[string]any{"sea_grass_type": "double_bot"}},
		{key: "minecraft:lily_pad", wantName: "minecraft:waterlily"},
		{key: "minecraft:melon", wantName: "minecraft:melon_block"},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			res := Lookup(tt.key)
			if !res.Recognized {
				t.Fatalf("%s not recognized", tt.key)
			}
			if res.BedrockState.Name != tt.wantName {
				t.Fatalf("%s -> %s, want %s", tt.key, res.BedrockState.Name, tt.wantName)
			}
			for k, want := range tt.wantProps {
				if got := res.BedrockState.Properties[k]; got != want {
					t.Fatalf("%s property %s = %#v (%T), want %#v (%T)", tt.key, k, got, got, want, want)
				}
			}
		})
	}
}

func TestLookup_SkyOlympusMissingStates(t *testing.T) {
	tests := []struct {
		key       string
		wantName  string
		wantProps map[string]any
	}{
		{key: "minecraft:short_grass", wantName: "minecraft:short_grass"},
		{key: "minecraft:dirt_path", wantName: "minecraft:grass_path"},
		{
			key:      "minecraft:light_gray_glazed_terracotta[facing=west]",
			wantName: "minecraft:silver_glazed_terracotta",
			wantProps: map[string]any{
				"facing_direction": int32(4),
			},
		},
		{key: "minecraft:red_nether_bricks", wantName: "minecraft:red_nether_brick"},
		{
			key:      "minecraft:small_dripleaf[facing=east,half=upper,waterlogged=false]",
			wantName: "minecraft:small_dripleaf_block",
			wantProps: map[string]any{
				"minecraft:cardinal_direction": "east",
				"upper_block_bit":              uint8(1),
			},
		},
		{key: "minecraft:snow_block", wantName: "minecraft:snow"},
		{key: "minecraft:polished_tuff", wantName: "minecraft:polished_tuff"},
		{key: "minecraft:waxed_weathered_cut_copper_slab[type=double,waterlogged=false]", wantName: "minecraft:waxed_weathered_double_cut_copper_slab"},
		{
			key:      "minecraft:tuff_stairs[facing=west,half=top,shape=straight,waterlogged=false]",
			wantName: "minecraft:tuff_stairs",
			wantProps: map[string]any{
				"weirdo_direction": int32(1),
				"upside_down_bit":  uint8(1),
			},
		},
		{key: "minecraft:tuff_wall[east=none,north=tall,south=tall,up=true,waterlogged=false,west=tall]", wantName: "minecraft:tuff_wall"},
		{
			key:      "minecraft:red_wall_banner[facing=south]",
			wantName: "minecraft:wall_banner",
			wantProps: map[string]any{
				"facing_direction": int32(3),
			},
		},
		{key: "minecraft:potted_blue_orchid", wantName: "minecraft:flower_pot"},
		{key: "minecraft:cave_vines_plant[berries=true]", wantName: "minecraft:cave_vines_body_with_berries"},
		{key: "minecraft:beetroots[age=0]", wantName: "minecraft:beetroot", wantProps: map[string]any{"growth": int32(0)}},
		{key: "minecraft:water_cauldron[level=3]", wantName: "minecraft:cauldron", wantProps: map[string]any{"cauldron_liquid": "water", "fill_level": int32(6)}},
		{
			key:      "minecraft:waxed_weathered_copper_trapdoor[facing=north,half=bottom,open=true,powered=false,waterlogged=false]",
			wantName: "minecraft:waxed_weathered_copper_trapdoor",
			wantProps: map[string]any{
				"direction":       int32(3),
				"open_bit":        uint8(1),
				"upside_down_bit": uint8(0),
			},
		},
		{
			key:      "minecraft:white_bed[facing=east,occupied=true,part=foot]",
			wantName: "minecraft:bed",
			wantProps: map[string]any{
				"direction":      int32(3),
				"head_piece_bit": uint8(0),
				"occupied_bit":   uint8(1),
			},
		},
		{
			key:      "minecraft:comparator[facing=east,mode=subtract,powered=false]",
			wantName: "minecraft:unpowered_comparator",
			wantProps: map[string]any{
				"minecraft:cardinal_direction": "east",
				"output_lit_bit":               uint8(0),
				"output_subtract_bit":          uint8(1),
			},
		},
		{
			key:      "minecraft:repeater[delay=2,facing=east,locked=false,powered=false]",
			wantName: "minecraft:unpowered_repeater",
			wantProps: map[string]any{
				"minecraft:cardinal_direction": "east",
				"repeater_delay":               int32(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			res := Lookup(tt.key)
			if !res.Recognized {
				t.Fatalf("%s not recognized", tt.key)
			}
			if res.BedrockState.Name != tt.wantName {
				t.Fatalf("%s -> %s, want %s", tt.key, res.BedrockState.Name, tt.wantName)
			}
			for k, want := range tt.wantProps {
				if got := res.BedrockState.Properties[k]; got != want {
					t.Fatalf("%s property %s = %#v (%T), want %#v (%T)", tt.key, k, got, got, want, want)
				}
			}
		})
	}
}

func TestLookupPreservesAttachmentDirections(t *testing.T) {
	tests := []struct {
		key       string
		wantName  string
		wantProps map[string]any
	}{
		{
			key:      "minecraft:oak_trapdoor[facing=east,half=bottom,open=false,powered=false,waterlogged=false]",
			wantName: "minecraft:trapdoor",
			wantProps: map[string]any{
				"direction":       int32(0),
				"open_bit":        uint8(0),
				"upside_down_bit": uint8(0),
			},
		},
		{
			key:      "minecraft:oak_trapdoor[facing=north,half=top,open=true,powered=false,waterlogged=false]",
			wantName: "minecraft:trapdoor",
			wantProps: map[string]any{
				"direction":       int32(3),
				"open_bit":        uint8(1),
				"upside_down_bit": uint8(1),
			},
		},
		{
			key:      "minecraft:oak_wall_sign[facing=south,waterlogged=false]",
			wantName: "minecraft:wall_sign",
			wantProps: map[string]any{
				"facing_direction": int32(3),
			},
		},
		{
			key:      "minecraft:oak_wall_sign[facing=east,waterlogged=false]",
			wantName: "minecraft:wall_sign",
			wantProps: map[string]any{
				"facing_direction": int32(5),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			res := Lookup(tt.key)
			if !res.Recognized {
				t.Fatalf("%s not recognized", tt.key)
			}
			if res.BedrockState.Name != tt.wantName {
				t.Fatalf("%s -> %s, want %s", tt.key, res.BedrockState.Name, tt.wantName)
			}
			for k, want := range tt.wantProps {
				if got := res.BedrockState.Properties[k]; got != want {
					t.Fatalf("%s property %s = %#v (%T), want %#v (%T)", tt.key, k, got, got, want, want)
				}
			}
		})
	}
}

func TestLookupAddsWaterLayerForAquaticPlants(t *testing.T) {
	for _, key := range []string{
		"minecraft:seagrass",
		"minecraft:tall_seagrass[half=lower]",
		"minecraft:tall_seagrass[half=upper]",
	} {
		t.Run(key, func(t *testing.T) {
			res := Lookup(key)
			if !res.Recognized {
				t.Fatalf("%s not recognized", key)
			}
			if res.Liquid == nil {
				t.Fatalf("%s should carry a water liquid layer", key)
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

func TestLookupTreatsDragonflyPlaceholdersAsRecognizedBedrockStates(t *testing.T) {
	direct, ok := world.BlockByName("minecraft:structure_void", nil)
	if !ok {
		t.Skip("Dragonfly does not know minecraft:structure_void")
	}
	if nbtBlock, ok := direct.(world.NBTer); !ok || nbtBlock.EncodeNBT() != nil {
		t.Skip("Dragonfly structure_void is implemented in this version")
	}

	res := Lookup("minecraft:structure_void")
	if !res.Recognized {
		t.Fatal("structure_void resolved to a Bedrock palette state but was reported as unknown")
	}
	if res.BedrockState.Name != "minecraft:structure_void" {
		t.Fatalf("BedrockState.Name = %q, want minecraft:structure_void", res.BedrockState.Name)
	}
}
