package translate

import (
	"math"
	"testing"

	_ "github.com/df-mc/dragonfly/server/block" // register vanilla blocks for tests
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

func TestSetGetMissingBlock(t *testing.T) {
	prev := MissingBlock()
	defer SetMissingBlock(prev)

	swap, ok := world.BlockByName("minecraft:stone", nil)
	if !ok || swap == nil {
		t.Skip("dragonfly stone block not registered in test env")
	}
	SetMissingBlock(swap)
	got := MissingBlock()
	if got == nil {
		t.Fatal("MissingBlock returned nil")
	}
	name, _ := got.EncodeBlock()
	if name != "minecraft:stone" {
		t.Errorf("got %q want minecraft:stone", name)
	}
}

func TestLookup_FallsBackToMissing(t *testing.T) {
	res := Lookup("minecraft:nonexistent_block_for_test")
	if res.Recognized {
		t.Errorf("expected Recognized=false")
	}
	if res.Block == nil {
		t.Errorf("Block should be non-nil (missing fallback)")
	}
	if res.RawKey != "minecraft:nonexistent_block_for_test" {
		t.Errorf("RawKey: got %q", res.RawKey)
	}
}

func TestLookupRecognizesBedrockPaletteStateWithoutConcreteDragonflyBlock(t *testing.T) {
	got := Lookup("minecraft:pink_wool")
	if !got.Recognized {
		t.Fatalf("pink_wool should be recognized as a Bedrock palette state")
	}
	if got.BedrockState.Name != "minecraft:pink_wool" {
		t.Fatalf("BedrockState.Name = %q, want minecraft:pink_wool", got.BedrockState.Name)
	}
	if len(got.BedrockState.Properties) != 0 {
		t.Fatalf("BedrockState.Properties = %#v, want empty", got.BedrockState.Properties)
	}
	name, _ := got.Block.EncodeBlock()
	if name != "minecraft:pink_wool" {
		t.Fatalf("Block encodes as %q, want minecraft:pink_wool", name)
	}
}

func TestLookupReturnsCanonicalBedrockStateProperties(t *testing.T) {
	got := Lookup("minecraft:oak_leaves[distance=1,persistent=true,waterlogged=false]")
	if !got.Recognized {
		t.Fatalf("oak_leaves should be recognized as a Bedrock palette state")
	}
	if got.BedrockState.Name != "minecraft:oak_leaves" {
		t.Fatalf("BedrockState.Name = %q, want minecraft:oak_leaves", got.BedrockState.Name)
	}
	if got.BedrockState.Properties["persistent_bit"] != uint8(1) {
		t.Fatalf("persistent_bit = %#v (%T), want uint8(1)",
			got.BedrockState.Properties["persistent_bit"], got.BedrockState.Properties["persistent_bit"])
	}
	if got.BedrockState.Properties["update_bit"] != uint8(0) {
		t.Fatalf("update_bit = %#v (%T), want uint8(0)",
			got.BedrockState.Properties["update_bit"], got.BedrockState.Properties["update_bit"])
	}
	name, _ := got.Block.EncodeBlock()
	if name != "minecraft:oak_leaves" {
		t.Fatalf("Block encodes as %q, want minecraft:oak_leaves", name)
	}
}

func TestLookupReturnsInertBlockForTickingBedrockState(t *testing.T) {
	got := Lookup("minecraft:smoker[facing=north,lit=false]")
	if !got.Recognized {
		t.Fatalf("smoker should be recognized as a Bedrock palette state")
	}
	name, _ := got.Block.EncodeBlock()
	if name != "minecraft:smoker" {
		t.Fatalf("Block encodes as %q, want minecraft:smoker", name)
	}
	if _, ticking := got.Block.(interface {
		Tick(int64, cube.Pos, *world.Tx)
	}); ticking {
		t.Fatalf("schematic block %T must be inert, not a random ticker", got.Block)
	}
}

func TestStateBlockHashHasUniqueSlowPathIdentity(t *testing.T) {
	stone := NewStateBlock(BedrockState{Name: "minecraft:stone"})
	dirt := NewStateBlock(BedrockState{Name: "minecraft:dirt"})

	stoneBase, stoneState := stone.Hash()
	dirtBase, dirtState := dirt.Hash()
	if stoneState != math.MaxUint64 || dirtState != math.MaxUint64 {
		t.Fatalf("StateBlock state hashes = %d/%d, want slow-path sentinel", stoneState, dirtState)
	}
	if stoneBase == dirtBase {
		t.Fatalf("StateBlock base hashes collide: %d", stoneBase)
	}
}
