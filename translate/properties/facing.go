package properties

import "strings"

// Facing converts Java facing (north|south|east|west|up|down) to whichever
// Bedrock property a particular block uses. Bedrock has historically used
// many names for the same concept: "direction" (4-way int), "facing_direction"
// (6-way int), "weirdo_direction" (stairs, 4-way int with a unique encoding),
// and others.
//
// The branch table below covers the common cases. Add families as needed
// when the table-build tests surface mismatches.
func Facing(javaValue, bedrockIdent string) (string, any, bool) {
	if strings.HasSuffix(bedrockIdent, "_door") || bedrockIdent == "wooden_door" ||
		strings.HasSuffix(bedrockIdent, "_fence_gate") || bedrockIdent == "fence_gate" {
		return "minecraft:cardinal_direction", javaValue, true
	}
	if strings.HasSuffix(bedrockIdent, "_glazed_terracotta") {
		return "facing_direction", blockFaceDirection(javaValue), true
	}
	if strings.HasSuffix(bedrockIdent, "_button") || bedrockIdent == "wooden_button" || bedrockIdent == "stone_button" {
		return "facing_direction", blockFaceDirection(javaValue), true
	}
	if strings.HasSuffix(bedrockIdent, "_trapdoor") || bedrockIdent == "trapdoor" {
		return trapdoorDirection(javaValue)
	}
	if bedrockIdent == "torch" || strings.HasSuffix(bedrockIdent, "_torch") {
		return "torch_facing_direction", javaValue, true
	}
	if bedrockIdent == "chest" || bedrockIdent == "trapped_chest" || bedrockIdent == "ender_chest" {
		return "minecraft:cardinal_direction", javaValue, true
	}
	if bedrockIdent == "anvil" {
		return "minecraft:cardinal_direction", javaValue, true
	}
	if strings.Contains(bedrockIdent, "amethyst") && (strings.HasSuffix(bedrockIdent, "_bud") || strings.HasSuffix(bedrockIdent, "_cluster")) {
		return "minecraft:block_face", javaValue, true
	}
	if strings.HasSuffix(bedrockIdent, "_stairs") {
		return stairsDirection(javaValue)
	}
	if strings.HasSuffix(bedrockIdent, "_wall_sign") || bedrockIdent == "wall_sign" {
		return "facing_direction", wallSignDirection(javaValue), true
	}
	switch bedrockIdent {
	case "oak_stairs", "spruce_stairs", "birch_stairs", "jungle_stairs",
		"acacia_stairs", "dark_oak_stairs", "mangrove_stairs", "cherry_stairs",
		"crimson_stairs", "warped_stairs", "bamboo_stairs",
		"stone_stairs", "cobblestone_stairs", "mossy_cobblestone_stairs",
		"brick_stairs", "stone_brick_stairs", "mossy_stone_brick_stairs",
		"sandstone_stairs", "smooth_sandstone_stairs", "red_sandstone_stairs",
		"smooth_red_sandstone_stairs", "nether_brick_stairs", "red_nether_brick_stairs",
		"quartz_stairs", "smooth_quartz_stairs", "purpur_stairs",
		"prismarine_stairs", "prismarine_brick_stairs", "dark_prismarine_stairs",
		"end_brick_stairs", "blackstone_stairs", "polished_blackstone_stairs",
		"polished_blackstone_brick_stairs", "polished_granite_stairs",
		"polished_diorite_stairs", "polished_andesite_stairs",
		"granite_stairs", "diorite_stairs", "andesite_stairs",
		"deepslate_brick_stairs", "deepslate_tile_stairs",
		"polished_deepslate_stairs", "cobbled_deepslate_stairs",
		"mud_brick_stairs", "tuff_stairs", "polished_tuff_stairs", "tuff_brick_stairs":
		return stairsDirection(javaValue)
	default:
		// Most directional blocks use "direction" or "facing_direction".
		// 4-way encoding: south=0, west=1, north=2, east=3.
		switch javaValue {
		case "south":
			return "direction", int32(0), true
		case "west":
			return "direction", int32(1), true
		case "north":
			return "direction", int32(2), true
		case "east":
			return "direction", int32(3), true
		case "up":
			return "facing_direction", int32(1), true
		case "down":
			return "facing_direction", int32(0), true
		}
		return "direction", int32(0), true
	}
}

func stairsDirection(javaValue string) (string, any, bool) {
	// Stairs on Bedrock use weirdo_direction:
	// east=0, west=1, south=2, north=3.
	switch javaValue {
	case "east":
		return "weirdo_direction", int32(0), true
	case "west":
		return "weirdo_direction", int32(1), true
	case "south":
		return "weirdo_direction", int32(2), true
	case "north":
		return "weirdo_direction", int32(3), true
	}
	return "weirdo_direction", int32(0), true
}

func wallSignDirection(javaValue string) int32 {
	switch javaValue {
	case "north":
		return 2
	case "south":
		return 3
	case "west":
		return 4
	case "east":
		return 5
	}
	return 2
}

func trapdoorDirection(javaValue string) (string, any, bool) {
	switch javaValue {
	case "east":
		return "direction", int32(0), true
	case "west":
		return "direction", int32(1), true
	case "south":
		return "direction", int32(2), true
	case "north":
		return "direction", int32(3), true
	}
	return "direction", int32(3), true
}

func blockFaceDirection(javaValue string) int32 {
	switch javaValue {
	case "down":
		return 0
	case "up":
		return 1
	case "north":
		return 2
	case "south":
		return 3
	case "west":
		return 4
	case "east":
		return 5
	}
	return 2
}
