package properties

// Waterlogged returns ok=false because Bedrock has no waterlogged property.
// The translate.table builder detects waterlogged=true and emits a Liquid
// alongside the base block instead.
func Waterlogged(javaValue, bedrockIdent string) (string, any, bool) {
	return "", nil, false
}
