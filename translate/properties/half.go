package properties

import "strings"

// Half handles Java's "half" property which means different things on
// stairs/trapdoors vs doors/flowers.
func Half(javaValue, bedrockIdent string) (string, any, bool) {
	if strings.HasSuffix(bedrockIdent, "_door") || bedrockIdent == "wooden_door" || bedrockIdent == "iron_door" ||
		bedrockIdent == "tall_grass" || bedrockIdent == "large_fern" ||
		bedrockIdent == "sunflower" || bedrockIdent == "lilac" ||
		bedrockIdent == "rose_bush" || bedrockIdent == "peony" ||
		bedrockIdent == "pitcher_plant" {
		return "upper_block_bit", javaValue == "upper", true
	}
	return "upside_down_bit", javaValue == "top", true
}
