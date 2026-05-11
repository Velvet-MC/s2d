package properties

// Tilt converts Java big dripleaf tilt to Bedrock's big_dripleaf_tilt state.
func Tilt(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent != "big_dripleaf" {
		return "", nil, false
	}
	switch javaValue {
	case "none", "unstable":
		return "big_dripleaf_tilt", javaValue, true
	case "partial":
		return "big_dripleaf_tilt", "partial_tilt", true
	case "full":
		return "big_dripleaf_tilt", "full_tilt", true
	default:
		return "big_dripleaf_tilt", "none", true
	}
}
