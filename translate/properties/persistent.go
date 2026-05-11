package properties

// Persistent converts Java leaves persistent (true|false) to Bedrock persistent_bit.
func Persistent(javaValue, bedrockIdent string) (string, any, bool) {
	return "persistent_bit", javaValue == "true", true
}
