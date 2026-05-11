package properties

// InWall converts Java fence gate in_wall to Bedrock in_wall_bit.
func InWall(javaValue, bedrockIdent string) (string, any, bool) {
	return "in_wall_bit", javaValue == "true", true
}
