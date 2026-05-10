package properties

// SlabType handles Java's slab "type" property. "double" requires an
// identifier swap (oak_slab → double_oak_slab on Bedrock); the table
// builder honors that via overrides.json. Here we only handle top/bottom.
func SlabType(javaValue, bedrockIdent string) (string, any, bool) {
	if javaValue == "top" {
		return "top_slot_bit", byte(1), true
	}
	return "top_slot_bit", byte(0), true
}
