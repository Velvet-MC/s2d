package properties

// Distance is the Java leaves "distance" 1-7 property. Bedrock has no
// equivalent (leaves persistence is handled differently). Drop it.
func Distance(javaValue, bedrockIdent string) (string, any, bool) {
	return "", nil, false
}
