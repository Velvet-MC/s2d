package translate

import (
	"testing"

	_ "github.com/df-mc/dragonfly/server/block" // register vanilla blocks for tests
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
