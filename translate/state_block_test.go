package translate

import "testing"

func TestStateBlockHotPathAllocations(t *testing.T) {
	b := NewStateBlock(BedrockState{
		Name: "minecraft:sandstone_stairs",
		Properties: map[string]any{
			"weirdo_direction": int32(3),
			"upside_down_bit":  uint8(0),
		},
	})

	if got := testing.AllocsPerRun(1000, func() {
		_, _ = b.Hash()
	}); got != 0 {
		t.Fatalf("Hash allocations = %v, want 0", got)
	}
	if got := testing.AllocsPerRun(1000, func() {
		_, _ = b.EncodeBlock()
	}); got != 0 {
		t.Fatalf("EncodeBlock allocations = %v, want 0", got)
	}
}
