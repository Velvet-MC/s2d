package properties

// Part converts Java bed part (head|foot) to Bedrock head_piece_bit.
func Part(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent != "bed" {
		return "", nil, false
	}
	if javaValue == "head" {
		return "head_piece_bit", uint8(1), true
	}
	return "head_piece_bit", uint8(0), true
}

// Occupied converts Java bed occupied to Bedrock occupied_bit.
func Occupied(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent != "bed" {
		return "", nil, false
	}
	return "occupied_bit", boolToBit(javaValue), true
}
