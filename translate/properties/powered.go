package properties

// Powered converts Java powered (true|false) to Bedrock powered_bit (1|0).
func Powered(javaValue, bedrockIdent string) (string, any, bool) {
	return "powered_bit", boolToBit(javaValue), true
}
