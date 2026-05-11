package properties

import "strings"

// Lit converts Java lit (true|false) to Bedrock lit (boolean).
// Bedrock 1.20+ uses a "lit" boolean property on most light-emitting blocks.
func Lit(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent == "redstone_ore" || bedrockIdent == "deepslate_redstone_ore" ||
		strings.HasPrefix(bedrockIdent, "lit_") && strings.HasSuffix(bedrockIdent, "redstone_ore") {
		return "", nil, false
	}
	return "lit", javaValue == "true", true
}
