package properties

import "strconv"

// Rotation maps Java sign/head rotation to Bedrock ground_sign_direction.
func Rotation(javaValue, bedrockIdent string) (string, any, bool) {
	n, _ := strconv.Atoi(javaValue)
	return "ground_sign_direction", int32(n), true
}
