package properties

// Open converts Java open (true|false) to Bedrock open_bit.
func Open(javaValue, bedrockIdent string) (string, any, bool) {
	return "open_bit", javaValue == "true", true
}
