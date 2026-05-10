package properties

import "strings"

// Half handles Java's "half" property which means different things on
// stairs vs doors vs flowers.
//   stairs:        "top"|"bottom"  → upside_down_bit 1|0
//   doors/flowers: "lower"|"upper" → upper_block_bit 0|1
func Half(javaValue, bedrockIdent string) (string, any, bool) {
	if strings.HasSuffix(bedrockIdent, "_door") || bedrockIdent == "iron_door" ||
		bedrockIdent == "tall_grass" || bedrockIdent == "large_fern" ||
		bedrockIdent == "sunflower" || bedrockIdent == "lilac" ||
		bedrockIdent == "rose_bush" || bedrockIdent == "peony" ||
		bedrockIdent == "pitcher_plant" {
		if javaValue == "upper" {
			return "upper_block_bit", byte(1), true
		}
		return "upper_block_bit", byte(0), true
	}
	// stairs default
	if javaValue == "top" {
		return "upside_down_bit", byte(1), true
	}
	return "upside_down_bit", byte(0), true
}
