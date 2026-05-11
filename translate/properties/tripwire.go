package properties

import "strings"

// Attached converts Java tripwire attached to Bedrock attached_bit.
func Attached(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent != "trip_wire" && !strings.HasSuffix(bedrockIdent, "_hanging_sign") {
		return "", nil, false
	}
	return "attached_bit", boolToBit(javaValue), true
}

// Disarmed converts Java tripwire disarmed to Bedrock disarmed_bit.
func Disarmed(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent != "trip_wire" {
		return "", nil, false
	}
	return "disarmed_bit", boolToBit(javaValue), true
}
