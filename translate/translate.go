// Package translate maps Java edition block-state strings to Dragonfly
// world.Block values. The full Java→Bedrock palette table is built once,
// lazily, on first call to Lookup.
//
// In v1.0 of the implementation plan this file holds only the public API
// and a placeholder Lookup that always returns the missing-block fallback.
// Phase 7.4 of the plan adds the real table builder that replaces the
// stub Lookup.
package translate

import (
	"maps"
	"sync"

	"github.com/df-mc/dragonfly/server/world"
)

// BedrockState is a neutral Bedrock block-state identifier. It describes what
// should be sent/saved as a Minecraft Bedrock state without requiring a
// concrete Dragonfly block implementation to exist for that state.
type BedrockState struct {
	Name       string
	Properties map[string]any
}

// Clone returns a copy of the Bedrock state and its property map.
func (s BedrockState) Clone() BedrockState {
	s.Properties = maps.Clone(s.Properties)
	return s
}

// Result is what Lookup returns. Block is never nil; if Recognized is
// false the missing-block fallback is returned and the canonical Java
// state is echoed in RawKey for caller-side reporting.
type Result struct {
	Block        world.Block
	Liquid       world.Liquid
	BedrockState BedrockState
	Recognized   bool
	RawKey       string
}

var (
	missingMu    sync.RWMutex
	missingBlock world.Block
)

// SetMissingBlock changes the global fallback used for unrecognized
// blocks. Call once at startup, before the first Lookup.
func SetMissingBlock(b world.Block) {
	missingMu.Lock()
	defer missingMu.Unlock()
	missingBlock = b
}

// MissingBlock returns the active fallback block. The first call lazily
// resolves a default (magenta wool, then plain wool, then magenta concrete,
// then stone) via world.BlockByName.
func MissingBlock() world.Block {
	missingMu.RLock()
	if missingBlock != nil {
		defer missingMu.RUnlock()
		return missingBlock
	}
	missingMu.RUnlock()
	b := resolveDefaultMissing()
	missingMu.Lock()
	if missingBlock == nil { // double-check after acquiring write lock
		missingBlock = b
	}
	out := missingBlock
	missingMu.Unlock()
	return out
}

// Lookup translates a canonical Java state string to a Dragonfly Result.
// Safe for concurrent use; the table is built once via sync.Once on first
// call. Unknown keys yield Recognized=false with the missing-block fallback.
func Lookup(canonicalJavaState string) Result {
	tableOnce.Do(buildTable)
	if tableErr != nil {
		return Result{Block: MissingBlock(), RawKey: canonicalJavaState}
	}
	if r, ok := table[canonicalJavaState]; ok {
		return r
	}
	if r, ok := lookupDynamicJavaState(canonicalJavaState); ok {
		return r
	}
	return Result{Block: MissingBlock(), RawKey: canonicalJavaState}
}
