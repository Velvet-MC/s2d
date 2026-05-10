package properties

import "strconv"

// Age maps Java numeric "age" property to Bedrock "growth" or "age" int.
// Most crops (wheat, carrots, potatoes, beetroot) use "growth" 0-7.
// Cactus and sugar cane use "age" 0-15.
func Age(javaValue, bedrockIdent string) (string, any, bool) {
	n, _ := strconv.Atoi(javaValue)
	if bedrockIdent == "cactus" || bedrockIdent == "reeds" /* sugar cane */ {
		return "age", int32(n), true
	}
	return "growth", int32(n), true
}
