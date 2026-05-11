package properties

// Shape converts Java shape for rail blocks. Other block families, notably
// stairs, either do not need the Java shape on Bedrock or encode it using a
// different block model, so those shapes are deliberately dropped.
func Shape(javaValue, bedrockIdent string) (string, any, bool) {
	if bedrockIdent != "rail" {
		return "", nil, false
	}
	switch javaValue {
	case "north_south":
		return "rail_direction", int32(0), true
	case "east_west":
		return "rail_direction", int32(1), true
	case "ascending_east":
		return "rail_direction", int32(2), true
	case "ascending_west":
		return "rail_direction", int32(3), true
	case "ascending_north":
		return "rail_direction", int32(4), true
	case "ascending_south":
		return "rail_direction", int32(5), true
	case "south_east":
		return "rail_direction", int32(6), true
	case "south_west":
		return "rail_direction", int32(7), true
	case "north_west":
		return "rail_direction", int32(8), true
	case "north_east":
		return "rail_direction", int32(9), true
	}
	return "rail_direction", int32(0), true
}
