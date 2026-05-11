package translate

import "testing"

func TestTransformBedrockStateRotatesRails(t *testing.T) {
	tests := []struct {
		name string
		in   int32
		want int32
	}{
		{name: "north_south_to_east_west", in: 0, want: 1},
		{name: "east_west_to_north_south", in: 1, want: 0},
		{name: "ascending_east_to_south", in: 2, want: 5},
		{name: "ascending_west_to_north", in: 3, want: 4},
		{name: "ascending_north_to_east", in: 4, want: 2},
		{name: "ascending_south_to_west", in: 5, want: 3},
		{name: "south_east_to_south_west", in: 6, want: 7},
		{name: "south_west_to_north_west", in: 7, want: 8},
		{name: "north_west_to_north_east", in: 8, want: 9},
		{name: "north_east_to_south_east", in: 9, want: 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := BedrockState{Name: "minecraft:rail", Properties: map[string]any{"rail_direction": tt.in}}
			got, changed := TransformBedrockState(state, "y", 1, false)
			if !changed {
				t.Fatal("TransformBedrockState reported unchanged")
			}
			if got.Properties["rail_direction"] != tt.want {
				t.Fatalf("rail_direction = %#v, want %#v", got.Properties["rail_direction"], tt.want)
			}
		})
	}
}

func TestTransformBedrockStateRotatesCommonDirectionalProperties(t *testing.T) {
	tests := []struct {
		name  string
		state BedrockState
		prop  string
		want  any
	}{
		{
			name:  "stairs",
			state: BedrockState{Name: "minecraft:sandstone_stairs", Properties: map[string]any{"weirdo_direction": int32(3)}},
			prop:  "weirdo_direction",
			want:  int32(0),
		},
		{
			name:  "trapdoor",
			state: BedrockState{Name: "minecraft:trapdoor", Properties: map[string]any{"direction": int32(3)}},
			prop:  "direction",
			want:  int32(0),
		},
		{
			name:  "bed",
			state: BedrockState{Name: "minecraft:bed", Properties: map[string]any{"direction": int32(3)}},
			prop:  "direction",
			want:  int32(0),
		},
		{
			name:  "grindstone",
			state: BedrockState{Name: "minecraft:grindstone", Properties: map[string]any{"direction": int32(3), "attachment": "side"}},
			prop:  "direction",
			want:  int32(0),
		},
		{
			name:  "lever",
			state: BedrockState{Name: "minecraft:lever", Properties: map[string]any{"lever_direction": "up_north_south"}},
			prop:  "lever_direction",
			want:  "up_east_west",
		},
		{
			name:  "amethyst",
			state: BedrockState{Name: "minecraft:amethyst_cluster", Properties: map[string]any{"minecraft:block_face": "east"}},
			prop:  "minecraft:block_face",
			want:  "south",
		},
		{
			name:  "wall",
			state: BedrockState{Name: "minecraft:cobblestone_wall", Properties: map[string]any{"wall_connection_type_north": "tall", "wall_connection_type_east": "none", "wall_connection_type_south": "none", "wall_connection_type_west": "short"}},
			prop:  "wall_connection_type_east",
			want:  "tall",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed := TransformBedrockState(tt.state, "y", 1, false)
			if !changed {
				t.Fatal("TransformBedrockState reported unchanged")
			}
			if got.Properties[tt.prop] != tt.want {
				t.Fatalf("%s = %#v, want %#v", tt.prop, got.Properties[tt.prop], tt.want)
			}
		})
	}
}
