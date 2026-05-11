package translate

import (
	"maps"
	"strings"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// TransformBedrockState applies a simple geometric transform to a Bedrock
// state. It is intended for schematic paste transforms after Java states have
// already been translated to Bedrock states.
func TransformBedrockState(state BedrockState, axis string, turns int, flip bool) (BedrockState, bool) {
	t := bedrockTransform{axis: strings.ToLower(axis), turns: ((turns % 4) + 4) % 4, flip: flip}
	if t.turns == 0 && !t.flip {
		return state.Clone(), false
	}
	props := maps.Clone(state.Properties)
	changed := false
	for _, key := range []string{"minecraft:cardinal_direction", "cardinal_direction"} {
		if v, ok := props[key]; ok {
			if s, ok := v.(string); ok {
				if d, ok := bedrockDirectionFromString(s); ok {
					next := bedrockTransformDirection(d, t)
					if next != d {
						props[key] = next.String()
						changed = true
					}
				}
			}
		}
	}
	if v, ok := props["facing_direction"]; ok {
		if face, ok := bedrockIntFace(v); ok {
			next := bedrockTransformFace(face, t)
			if next != face {
				props["facing_direction"] = int32(next)
				changed = true
			}
		}
	}
	if v, ok := props["weirdo_direction"]; ok {
		if d, ok := bedrockIntStairsDirection(v); ok {
			next := bedrockTransformDirection(d, t)
			if next != d {
				props["weirdo_direction"] = bedrockStairsDirectionInt(next)
				changed = true
			}
		}
	}
	if v, ok := props["direction"]; ok && bedrockIsTrapdoorState(state.Name) {
		if d, ok := bedrockIntStairsDirection(v); ok {
			next := bedrockTransformDirection(d, t)
			if next != d {
				props["direction"] = bedrockStairsDirectionInt(next)
				changed = true
			}
		}
	}
	if v, ok := props["ground_sign_direction"]; ok && t.axis == "y" {
		if o, ok := bedrockIntValue(v); ok {
			next := bedrockTransformOrientation(o, t)
			if next != o {
				props["ground_sign_direction"] = int32(next)
				changed = true
			}
		}
	}
	for _, key := range []string{"pillar_axis", "portal_axis"} {
		if v, ok := props[key]; ok {
			if s, ok := v.(string); ok {
				if a, ok := bedrockAxisFromString(s); ok {
					next := bedrockTransformAxis(a, t)
					if next != a {
						props[key] = next.String()
						changed = true
					}
				}
			}
		}
	}
	if v, ok := props["torch_facing_direction"]; ok {
		if s, ok := v.(string); ok {
			if face, ok := bedrockTorchFaceFromString(s); ok {
				next := bedrockTransformFace(face, t)
				if next != face {
					props["torch_facing_direction"] = bedrockTorchFaceString(next)
					changed = true
				}
			}
		}
	}
	if v, ok := props["lever_direction"]; ok {
		if s, ok := v.(string); ok {
			next := bedrockTransformLeverDirection(s, t)
			if next != s {
				props["lever_direction"] = next
				changed = true
			}
		}
	}
	if v, ok := props["rail_direction"]; ok {
		if rail, ok := bedrockIntValue(v); ok {
			next := bedrockTransformRailDirection(rail, t)
			if next != rail {
				props["rail_direction"] = int32(next)
				changed = true
			}
		}
	}
	if bedrockTransformWallConnections(props, t) {
		changed = true
	}
	if !changed {
		return state.Clone(), false
	}
	return BedrockState{Name: state.Name, Properties: props}, true
}

type bedrockTransform struct {
	axis  string
	turns int
	flip  bool
}

func bedrockTransformDirection(d cube.Direction, t bedrockTransform) cube.Direction {
	if t.flip {
		switch t.axis {
		case "x":
			if d == cube.East || d == cube.West {
				return d.Opposite()
			}
		case "z":
			if d == cube.North || d == cube.South {
				return d.Opposite()
			}
		}
		return d
	}
	if t.axis != "y" {
		return d
	}
	for i := 0; i < t.turns; i++ {
		d = d.RotateRight()
	}
	return d
}

func bedrockTransformFace(f cube.Face, t bedrockTransform) cube.Face {
	vec, ok := bedrockFaceVector(f)
	if !ok {
		return f
	}
	return bedrockVectorFace(bedrockTransformVector(vec, t))
}

func bedrockTransformAxis(a cube.Axis, t bedrockTransform) cube.Axis {
	vec := bedrockAxisVector(a)
	next := bedrockTransformVector(vec, t)
	if next[0] != 0 {
		return cube.X
	}
	if next[1] != 0 {
		return cube.Y
	}
	return cube.Z
}

func bedrockTransformOrientation(o int, t bedrockTransform) int {
	o = ((o % 16) + 16) % 16
	if t.flip {
		switch t.axis {
		case "x":
			return (16 - o) % 16
		case "z":
			return (8 - o + 16) % 16
		default:
			return o
		}
	}
	return (o + t.turns*4) % 16
}

func bedrockTransformVector(v [3]int, t bedrockTransform) [3]int {
	if t.flip {
		switch t.axis {
		case "x":
			v[0] = -v[0]
		case "y":
			v[1] = -v[1]
		case "z":
			v[2] = -v[2]
		}
		return v
	}
	for i := 0; i < t.turns; i++ {
		switch t.axis {
		case "x":
			v = [3]int{v[0], -v[2], v[1]}
		case "z":
			v = [3]int{-v[1], v[0], v[2]}
		default:
			v = [3]int{-v[2], v[1], v[0]}
		}
	}
	return v
}

func bedrockDirectionFromString(s string) (cube.Direction, bool) {
	switch strings.ToLower(s) {
	case "north":
		return cube.North, true
	case "south":
		return cube.South, true
	case "west":
		return cube.West, true
	case "east":
		return cube.East, true
	default:
		return cube.North, false
	}
}

func bedrockAxisFromString(s string) (cube.Axis, bool) {
	switch strings.ToLower(s) {
	case "x":
		return cube.X, true
	case "y":
		return cube.Y, true
	case "z":
		return cube.Z, true
	default:
		return cube.Y, false
	}
}

func bedrockFaceVector(f cube.Face) ([3]int, bool) {
	switch f {
	case cube.FaceDown:
		return [3]int{0, -1, 0}, true
	case cube.FaceUp:
		return [3]int{0, 1, 0}, true
	case cube.FaceNorth:
		return [3]int{0, 0, -1}, true
	case cube.FaceSouth:
		return [3]int{0, 0, 1}, true
	case cube.FaceWest:
		return [3]int{-1, 0, 0}, true
	case cube.FaceEast:
		return [3]int{1, 0, 0}, true
	default:
		return [3]int{}, false
	}
}

func bedrockVectorFace(v [3]int) cube.Face {
	switch v {
	case [3]int{0, -1, 0}:
		return cube.FaceDown
	case [3]int{0, 1, 0}:
		return cube.FaceUp
	case [3]int{0, 0, -1}:
		return cube.FaceNorth
	case [3]int{0, 0, 1}:
		return cube.FaceSouth
	case [3]int{-1, 0, 0}:
		return cube.FaceWest
	case [3]int{1, 0, 0}:
		return cube.FaceEast
	default:
		return cube.FaceUp
	}
}

func bedrockAxisVector(a cube.Axis) [3]int {
	switch a {
	case cube.X:
		return [3]int{1, 0, 0}
	case cube.Y:
		return [3]int{0, 1, 0}
	default:
		return [3]int{0, 0, 1}
	}
}

func bedrockIntFace(v any) (cube.Face, bool) {
	n, ok := bedrockIntValue(v)
	if !ok || n < int(cube.FaceDown) || n > int(cube.FaceEast) {
		return cube.FaceDown, false
	}
	return cube.Face(n), true
}

func bedrockIntValue(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case uint8:
		return int(n), true
	case float64:
		return int(n), n == float64(int(n))
	default:
		return 0, false
	}
}

func bedrockIsTrapdoorState(name string) bool {
	name = strings.TrimPrefix(name, "minecraft:")
	return name == "trapdoor" || strings.HasSuffix(name, "_trapdoor")
}

func bedrockIntStairsDirection(v any) (cube.Direction, bool) {
	n, ok := bedrockIntValue(v)
	if !ok {
		return cube.North, false
	}
	switch n {
	case 0:
		return cube.East, true
	case 1:
		return cube.West, true
	case 2:
		return cube.South, true
	case 3:
		return cube.North, true
	default:
		return cube.North, false
	}
}

func bedrockStairsDirectionInt(d cube.Direction) int32 {
	switch d {
	case cube.East:
		return 0
	case cube.West:
		return 1
	case cube.South:
		return 2
	case cube.North:
		return 3
	default:
		return 3
	}
}

func bedrockTorchFaceFromString(s string) (cube.Face, bool) {
	switch strings.ToLower(s) {
	case "top":
		return cube.FaceDown, true
	case "north":
		return cube.FaceNorth, true
	case "south":
		return cube.FaceSouth, true
	case "west":
		return cube.FaceWest, true
	case "east":
		return cube.FaceEast, true
	default:
		return cube.FaceDown, false
	}
}

func bedrockTorchFaceString(f cube.Face) string {
	if f == cube.FaceDown {
		return "top"
	}
	return f.String()
}

func bedrockTransformLeverDirection(s string, t bedrockTransform) string {
	if d, ok := bedrockDirectionFromString(s); ok {
		return bedrockTransformDirection(d, t).String()
	}
	switch s {
	case "up_north_south", "up_east_west", "down_north_south", "down_east_west":
		prefix := "up_"
		axisPart := strings.TrimPrefix(s, "up_")
		if strings.HasPrefix(s, "down_") {
			prefix = "down_"
			axisPart = strings.TrimPrefix(s, "down_")
		}
		axis := cube.Z
		if axisPart == "east_west" {
			axis = cube.X
		}
		next := bedrockTransformAxis(axis, t)
		if next == cube.X {
			return prefix + "east_west"
		}
		return prefix + "north_south"
	default:
		return s
	}
}

func bedrockTransformRailDirection(rail int, t bedrockTransform) int {
	for i := 0; i < t.turns; i++ {
		if t.axis == "y" && !t.flip {
			rail = map[int]int{0: 1, 1: 0, 2: 5, 3: 4, 4: 2, 5: 3, 6: 7, 7: 8, 8: 9, 9: 6}[rail]
		}
	}
	return rail
}

func bedrockTransformWallConnections(props map[string]any, t bedrockTransform) bool {
	keys := map[string]cube.Direction{
		"wall_connection_type_north": cube.North,
		"wall_connection_type_east":  cube.East,
		"wall_connection_type_south": cube.South,
		"wall_connection_type_west":  cube.West,
	}
	values := make(map[cube.Direction]any, len(keys))
	seen := false
	for key, dir := range keys {
		if v, ok := props[key]; ok {
			values[dir] = v
			seen = true
		}
	}
	if !seen {
		return false
	}
	for key := range keys {
		props[key] = "none"
	}
	changed := false
	for from, value := range values {
		to := bedrockTransformDirection(from, t)
		key := bedrockWallConnectionKey(to)
		if key == "" {
			continue
		}
		if props[key] != value {
			changed = true
		}
		props[key] = value
	}
	return changed
}

func bedrockWallConnectionKey(d cube.Direction) string {
	switch d {
	case cube.North:
		return "wall_connection_type_north"
	case cube.East:
		return "wall_connection_type_east"
	case cube.South:
		return "wall_connection_type_south"
	case cube.West:
		return "wall_connection_type_west"
	default:
		return ""
	}
}
