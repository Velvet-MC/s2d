package properties

// Hanging converts Java hanging (lanterns) to Bedrock hanging.
func Hanging(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent != "lantern" && bedrockIdent != "soul_lantern" {
		return "", nil, false
	}
	return "hanging", javaValue == "true", true
}
