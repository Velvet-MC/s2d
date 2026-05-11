// Package schem is the public entry point for the S2D library.
// It defines the cross-cutting Schematic / Block / UnknownReport types,
// holds the format-handler registry, and exposes Read which dispatches
// to a format reader based on the filename.
package schem

import (
	"io"

	"github.com/df-mc/dragonfly/server/world"
)

// Format names a supported on-disk schematic format.
type Format string

const (
	FormatSpongeV2 Format = "sponge_v2"
	FormatLegacy   Format = "legacy_schematic"
)

// Block is one cell in a parsed schematic.
type Block struct {
	Pos    [3]int       // x, y, z within the schematic local frame
	Block  world.Block  // translated Bedrock block; never nil
	Liquid world.Liquid // non-nil iff the Java source was waterlogged
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
