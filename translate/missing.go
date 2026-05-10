package translate

import (
	"github.com/df-mc/dragonfly/server/world"
)

// resolveDefaultMissing returns the default missing-block (magenta wool).
// It tries a chain of fallbacks so the package degrades gracefully if a
// particular Dragonfly version doesn't register the preferred block.
//
// Returns nil only if Dragonfly registered no blocks at all (which would
// only happen if the consumer never imported server/block).
func resolveDefaultMissing() world.Block {
	if b, ok := world.BlockByName("minecraft:wool", map[string]any{"color": "magenta"}); ok && b != nil {
		return b
	}
	if b, ok := world.BlockByName("minecraft:magenta_wool", nil); ok && b != nil {
		return b
	}
	if b, ok := world.BlockByName("minecraft:wool", nil); ok && b != nil {
		return b
	}
	if b, ok := world.BlockByName("minecraft:magenta_concrete", nil); ok && b != nil {
		return b
	}
	if b, ok := world.BlockByName("minecraft:stone", nil); ok && b != nil {
		return b
	}
	return nil
}
