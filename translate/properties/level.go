package properties

import "strconv"

// Level maps Java water/lava "level" 0-15 to Bedrock "liquid_depth" 0-15.
func Level(javaValue, bedrockIdent string) (string, any, bool) {
	n, _ := strconv.Atoi(javaValue)
	if bedrockIdent == "cauldron" {
		if n < 0 {
			n = 0
		}
		if n > 3 {
			n = 3
		}
		return "fill_level", int32(n * 2), true
	}
	return "liquid_depth", int32(n), true
}
