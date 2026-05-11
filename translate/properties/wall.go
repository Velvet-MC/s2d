package properties

import "strings"

// Walls have per-side connection types (none|low|tall) on Java, plus an
// "up" boolean for the post. Bedrock represents them as
// wall_post_bit + wall_connection_type_<side>.

// WallUp converts Java wall up (true|false) to Bedrock wall_post_bit.
func WallUp(javaValue, bedrockIdent string) (string, any, bool) {
	if !isWall(bedrockIdent) {
		return "", nil, false
	}
	return "wall_post_bit", boolToBit(javaValue), true
}

// WallNorth converts Java wall north connection (none|low|tall) to Bedrock.
func WallNorth(javaValue, bedrockIdent string) (string, any, bool) {
	if !isWall(bedrockIdent) {
		return "", nil, false
	}
	return "wall_connection_type_north", wallConnection(javaValue), true
}

// WallEast converts Java wall east connection.
func WallEast(javaValue, bedrockIdent string) (string, any, bool) {
	if !isWall(bedrockIdent) {
		return "", nil, false
	}
	return "wall_connection_type_east", wallConnection(javaValue), true
}

// WallSouth converts Java wall south connection.
func WallSouth(javaValue, bedrockIdent string) (string, any, bool) {
	if !isWall(bedrockIdent) {
		return "", nil, false
	}
	return "wall_connection_type_south", wallConnection(javaValue), true
}

// WallWest converts Java wall west connection.
func WallWest(javaValue, bedrockIdent string) (string, any, bool) {
	if !isWall(bedrockIdent) {
		return "", nil, false
	}
	return "wall_connection_type_west", wallConnection(javaValue), true
}

func wallConnection(javaValue string) string {
	if javaValue == "low" {
		return "short"
	}
	return javaValue
}

func isWall(bedrockIdent string) bool {
	return bedrockIdent == "cobblestone_wall" || strings.HasSuffix(bedrockIdent, "_wall")
}
