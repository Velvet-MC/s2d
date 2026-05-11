package properties

// Face converts Java attachment face for blocks that use a separate Bedrock
// attachment state.
func Face(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent != "grindstone" {
		return "", nil, false
	}
	switch javaValue {
	case "ceiling":
		return "attachment", "hanging", true
	case "floor":
		return "attachment", "standing", true
	case "wall":
		return "attachment", "side", true
	}
	return "attachment", "standing", true
}
