package properties

// Axis converts Java AXIS (x|y|z) to the matching Bedrock state property.
func Axis(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent == "portal" {
		return "portal_axis", javaValue, true
	}
	return "pillar_axis", javaValue, true
}
