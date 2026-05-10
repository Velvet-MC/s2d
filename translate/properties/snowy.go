package properties

// Snowy converts Java snowy (true|false) to Bedrock covered_bit (1|0).
// Used on grass / podzol / mycelium.
func Snowy(javaValue, bedrockIdent string) (string, any, bool) {
	return "covered_bit", boolToBit(javaValue), true
}
