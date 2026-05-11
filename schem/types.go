// Package schem is the public entry point for the S2D library.
// It defines the cross-cutting Schematic / Block / UnknownReport types,
// holds the format-handler registry, and exposes Read which dispatches
// to a format reader based on the filename.
package schem

import (
	"io"
	"maps"

	"github.com/df-mc/dragonfly/server/world"
)

// Format names a supported on-disk schematic format.
type Format string

const (
	FormatSpongeV2 Format = "sponge_v2"
	FormatSpongeV3 Format = "sponge_v3"
	FormatLegacy   Format = "legacy_schematic"
)

// BedrockState is a neutral Bedrock block-state identifier. It is the
// minecraft: identifier and state properties a consumer should preserve even
// if its runtime has no concrete behaviour implementation for the block.
// Properties may be shared between blocks that came from the same source
// palette entry; call Clone before mutating them.
type BedrockState struct {
	Name       string
	Properties map[string]any
}

// Clone returns a copy of the Bedrock state and its property map.
func (s BedrockState) Clone() BedrockState {
	s.Properties = maps.Clone(s.Properties)
	return s
}

// Block is one cell in a parsed schematic.
type Block struct {
	Pos            [3]int       // x, y, z within the schematic local frame
	Block          world.Block  // translated Bedrock block; never nil
	Liquid         world.Liquid // non-nil iff the Java source was waterlogged
	BedrockState   BedrockState // neutral Bedrock identifier/properties
	PaletteIndex   uint32       // source palette index when PaletteIndexOK is true
	PaletteIndexOK bool         // true when the source format had palette indexes
}

// Schematic is the parsed result of Read.
type Schematic struct {
	Format                Format
	Width, Height, Length int
	Offset                [3]int
	Blocks                []Block
	Unknowns              UnknownReport
}

// ScanInfo is the metadata returned by Scan after visiting every block.
type ScanInfo struct {
	Format                Format
	Width, Height, Length int
	Offset                [3]int
	PaletteSize           int
	Unknowns              UnknownReport
}

// UnknownReport tallies cells whose Java state could not be translated.
// Counts is keyed by the canonical Java state string.
type UnknownReport struct {
	Counts map[string]int
	Total  int
}

// BlockHandler receives one translated block while scanning a schematic.
type BlockHandler func(Block) error

// InfoHandler receives schematic metadata before the first scanned block.
type InfoHandler func(ScanInfo) error

// FormatHandler describes one supported on-disk format.
// Format-reader packages register their handler at init() time.
type FormatHandler struct {
	Name       Format
	Extensions []string                                                     // e.g. []string{".schem"}
	Signature  func(headerPeek []byte) bool                                 // optional NBT signature check
	Read       func(io.Reader) (*Schematic, error)                          // materialising reader
	Scan       func(io.Reader, InfoHandler, BlockHandler) (ScanInfo, error) // streaming reader
}
