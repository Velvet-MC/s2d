package properties

import "strconv"

// Redstone maps Java redstone wire "power" 0-15 to Bedrock "redstone_signal" 0-15.
func Redstone(javaValue, bedrockIdent string) (string, any, bool) {
	n, _ := strconv.Atoi(javaValue)
	return "redstone_signal", int32(n), true
}
