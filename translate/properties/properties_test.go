package properties

import "testing"

type tc struct {
	prop      string
	javaValue string
	bedrockId string
	wantProp  string
	wantValue any
	wantOK    bool
}

func TestConverters(t *testing.T) {
	cases := []tc{
		// axis
		{"axis", "y", "oak_log", "pillar_axis", "y", true},
		{"axis", "x", "deepslate", "pillar_axis", "x", true},
		// facing — stairs
		{"facing", "north", "oak_stairs", "weirdo_direction", int32(3), true},
		{"facing", "east", "oak_stairs", "weirdo_direction", int32(0), true},
		// facing — generic
		{"facing", "north", "furnace", "direction", int32(2), true},
		{"facing", "south", "furnace", "direction", int32(0), true},
		// half — stairs
		{"half", "top", "oak_stairs", "upside_down_bit", byte(1), true},
		{"half", "bottom", "oak_stairs", "upside_down_bit", byte(0), true},
		// half — door
		{"half", "upper", "oak_door", "upper_block_bit", byte(1), true},
		{"half", "lower", "oak_door", "upper_block_bit", byte(0), true},
		// hinge
		{"hinge", "left", "oak_door", "door_hinge_bit", byte(0), true},
		{"hinge", "right", "oak_door", "door_hinge_bit", byte(1), true},
		// open / powered / persistent
		{"open", "true", "oak_door", "open_bit", byte(1), true},
		{"powered", "true", "oak_button", "powered_bit", byte(1), true},
		{"persistent", "true", "oak_leaves", "persistent_bit", byte(1), true},
		// waterlogged → dropped
		{"waterlogged", "true", "oak_stairs", "", nil, false},
		// distance → dropped
		{"distance", "5", "oak_leaves", "", nil, false},
		// snowy
		{"snowy", "true", "grass", "covered_bit", byte(1), true},
		// age (crops)
		{"age", "7", "wheat", "growth", int32(7), true},
		// age (cactus)
		{"age", "12", "cactus", "age", int32(12), true},
		// level
		{"level", "8", "water", "liquid_depth", int32(8), true},
		// power
		{"power", "10", "redstone_wire", "redstone_signal", int32(10), true},
		// slab type
		{"type", "top", "oak_slab", "top_slot_bit", byte(1), true},
		// walls
		{"up", "true", "cobblestone_wall", "wall_post_bit", byte(1), true},
		{"north", "low", "cobblestone_wall", "wall_connection_type_north", "low", true},
	}
	for _, c := range cases {
		t.Run(c.prop+"/"+c.javaValue+"/"+c.bedrockId, func(t *testing.T) {
			conv, ok := Registry[c.prop]
			if !ok {
				t.Fatalf("no converter for %q", c.prop)
			}
			gotProp, gotValue, gotOK := conv(c.javaValue, c.bedrockId)
			if gotOK != c.wantOK {
				t.Errorf("ok: got %v want %v", gotOK, c.wantOK)
			}
			if !c.wantOK {
				return
			}
			if gotProp != c.wantProp {
				t.Errorf("prop: got %q want %q", gotProp, c.wantProp)
			}
			if gotValue != c.wantValue {
				t.Errorf("value: got %v(%T) want %v(%T)", gotValue, gotValue, c.wantValue, c.wantValue)
			}
		})
	}
}

// Lit converter is exercised separately because it returns a `bool`, not a byte/int32.
func TestLit(t *testing.T) {
	prop, val, ok := Lit("true", "furnace")
	if !ok || prop != "lit" || val != true {
		t.Errorf("got prop=%q val=%v ok=%v", prop, val, ok)
	}
	prop, val, ok = Lit("false", "furnace")
	if !ok || prop != "lit" || val != false {
		t.Errorf("got prop=%q val=%v ok=%v", prop, val, ok)
	}
}
