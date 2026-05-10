package properties

// Axis converts Java AXIS (x|y|z) to Bedrock pillar_axis.
// Logs, basalt, deepslate pillars, etc. all use the same prop name on Bedrock.
func Axis(javaValue, bedrockIdent string) (string, any, bool) {
	return "pillar_axis", javaValue, true
}
