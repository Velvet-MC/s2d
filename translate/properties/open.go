package properties

// Open converts Java open (true|false) to Bedrock open_bit (1|0).
func Open(javaValue, bedrockIdent string) (string, any, bool) {
	return "open_bit", boolToBit(javaValue), true
}
