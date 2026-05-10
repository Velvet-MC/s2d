package properties

// Persistent converts Java persistent (true|false) to Bedrock persistent_bit (1|0).
func Persistent(javaValue, bedrockIdent string) (string, any, bool) {
	return "persistent_bit", boolToBit(javaValue), true
}
