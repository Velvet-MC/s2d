package properties

import "strings"

// Powered converts Java powered (true|false) to the matching Bedrock state when one exists.
func Powered(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent == "unpowered_comparator" || bedrockIdent == "powered_comparator" {
		return "output_lit_bit", boolToBit(javaValue), true
	}
	if bedrockIdent == "unpowered_repeater" || bedrockIdent == "powered_repeater" {
		return "", nil, false
	}
	if strings.HasSuffix(bedrockIdent, "_door") || bedrockIdent == "wooden_door" || bedrockIdent == "iron_door" ||
		strings.HasSuffix(bedrockIdent, "_trapdoor") || bedrockIdent == "trapdoor" || bedrockIdent == "iron_trapdoor" ||
		strings.HasSuffix(bedrockIdent, "_fence_gate") || bedrockIdent == "fence_gate" || bedrockIdent == "noteblock" {
		return "", nil, false
	}
	if strings.HasSuffix(bedrockIdent, "_button") || bedrockIdent == "wooden_button" || bedrockIdent == "stone_button" {
		return "button_pressed_bit", boolToBit(javaValue), true
	}
	return "powered_bit", boolToBit(javaValue), true
}
