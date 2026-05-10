package properties

// Lit converts Java lit (true|false) to Bedrock lit (boolean).
// Bedrock 1.20+ uses a "lit" boolean property on most light-emitting blocks.
func Lit(javaValue, bedrockIdent string) (string, any, bool) {
	return "lit", javaValue == "true", true
}
