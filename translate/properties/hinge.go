package properties

// Hinge converts Java door hinge (left|right) to Bedrock door_hinge_bit.
func Hinge(javaValue, bedrockIdent string) (string, any, bool) {
	return "door_hinge_bit", javaValue == "right", true
}
