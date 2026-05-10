package properties

import "strconv"

// Level maps Java water/lava "level" 0-15 to Bedrock "liquid_depth" 0-15.
func Level(javaValue, bedrockIdent string) (string, any, bool) {
	n, _ := strconv.Atoi(javaValue)
	return "liquid_depth", int32(n), true
}
