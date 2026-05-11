// Package properties holds per-property Java→Bedrock value converters.
// Each converter takes the Java value and the Bedrock identifier (so
// converters can branch on block family where Bedrock uses different
// property names for the same Java concept) and returns the Bedrock
// property name + value to put in the Bedrock state map.
package properties

import "strings"

// Converter is the signature shared by all per-property converters.
//
// javaValue is the Java property value as a string (e.g. "y", "north", "true").
// bedrockIdent is the Bedrock block identifier the result will be applied to,
// without the "minecraft:" prefix (e.g. "oak_log", "oak_stairs").
//
// Returns the Bedrock property name + value to insert into the Bedrock state
// map, plus an ok flag. ok=false signals the property should be dropped
// (e.g. computed-only Bedrock properties, or properties that have no
// Bedrock equivalent).
type Converter func(javaValue, bedrockIdent string) (bedrockProp string, bedrockValue any, ok bool)

// Registry maps Java property names to converters. The translate table
// builder consults this map for every property on every Java state.
// Properties not in the registry are dropped silently for v1.0.
var Registry = map[string]Converter{
	"axis":        Axis,
	"facing":      Facing,
	"half":        Half,
	"hinge":       Hinge,
	"open":        Open,
	"powered":     Powered,
	"lit":         Lit,
	"waterlogged": Waterlogged,
	"snowy":       Snowy,
	"age":         Age,
	"distance":    Distance,
	"persistent":  Persistent,
	"level":       Level,
	"power":       Redstone,
	"type":        SlabType,
	"up":          WallUp,
	"north":       WallNorth,
	"east":        WallEast,
	"south":       WallSouth,
	"west":        WallWest,
	"in_wall":     InWall,
	"rotation":    Rotation,
	"shape":       Shape,
	"tilt":        Tilt,
	"hanging":     Hanging,
	"face":        Face,
	"part":        Part,
	"occupied":    Occupied,
	"mode":        Mode,
	"delay":       Delay,
	"attached":    Attached,
	"disarmed":    Disarmed,
}

// boolToBit returns 1 for "true", 0 otherwise.
func boolToBit(s string) byte {
	if strings.EqualFold(s, "true") {
		return 1
	}
	return 0
}
