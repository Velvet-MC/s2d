package properties

// SlabType handles Java's slab "type" property. "double" requires an
// identifier swap (oak_slab → oak_double_slab on Bedrock); the table builder
// handles that before resolving the block. Here we only set the vertical half.
func SlabType(javaValue, bedrockIdent string) (string, any, bool) {
	if len(bedrockIdent) < len("_slab") || bedrockIdent[len(bedrockIdent)-len("_slab"):] != "_slab" {
		return "", nil, false
	}
	if javaValue == "top" {
		return "minecraft:vertical_half", "top", true
	}
	return "minecraft:vertical_half", "bottom", true
}
