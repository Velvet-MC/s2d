package properties

import "strconv"

// Mode converts Java comparator mode to Bedrock output_subtract_bit.
func Mode(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent != "unpowered_comparator" && bedrockIdent != "powered_comparator" {
		return "", nil, false
	}
	if javaValue == "subtract" {
		return "output_subtract_bit", uint8(1), true
	}
	return "output_subtract_bit", uint8(0), true
}

// Delay converts Java repeater delay 1..4 to Bedrock repeater_delay 0..3.
func Delay(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent != "unpowered_repeater" && bedrockIdent != "powered_repeater" {
		return "", nil, false
	}
	n, _ := strconv.Atoi(javaValue)
	if n < 1 {
		n = 1
	}
	if n > 4 {
		n = 4
	}
	return "repeater_delay", int32(n - 1), true
}
