package properties

// Hinge converts Java door hinge (left|right) to Bedrock door_hinge_bit (0|1).
func Hinge(javaValue, bedrockIdent string) (string, any, bool) {
	if javaValue == "right" {
		return "door_hinge_bit", byte(1), true
	}
	return "door_hinge_bit", byte(0), true
}
