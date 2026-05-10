// Package sponge reads Sponge Schematic v2 (.schem) files. It depends on
// gophertunnel/minecraft/nbt for big-endian Java NBT decoding.
//
// Translation to Dragonfly world.Block is deferred to a hook (resolveBlock)
// so this package can be exercised before the translate package exists.
// Phase 7.5 of the implementation plan replaces resolveBlock with a call
// to translate.Lookup once that package is ready.
package sponge

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"

	"github.com/Clxser/S2D/palette"
	"github.com/Clxser/S2D/schem"
)

// rawSchematic mirrors the Sponge v2 NBT root.
type rawSchematic struct {
	Version     int32            `nbt:"Version"`
	DataVersion int32            `nbt:"DataVersion,omitempty"`
	Width       int16            `nbt:"Width"`
	Height      int16            `nbt:"Height"`
	Length      int16            `nbt:"Length"`
	PaletteMax  int32            `nbt:"PaletteMax,omitempty"`
	Palette     map[string]int32 `nbt:"Palette"`
	BlockData   []byte           `nbt:"BlockData"`
	Offset      []int32          `nbt:"Offset,omitempty"`
}

// lookupResult is the placeholder shape returned by resolveBlock until
// the translate package is wired in (Phase 7.5).
type lookupResult struct {
	Block      world.Block
	Liquid     world.Liquid
	Recognized bool
	RawKey     string
}

// resolveBlock is a stub: it returns Dragonfly air for every canonical key,
// marking the result unrecognized so the reader populates UnknownReport.
// Phase 7.5 of the implementation plan replaces calls to this function with
// translate.Lookup(canonical).
func resolveBlock(canonical string) lookupResult {
	air, _ := world.BlockByName("minecraft:air", nil)
	return lookupResult{
		Block:      air,
		Recognized: false,
		RawKey:     canonical,
	}
}

// Read parses a Sponge v2 schematic from r.
func Read(r io.Reader) (*schem.Schematic, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("sponge v2: gzip: %w", err)
	}
	defer gz.Close()
	body, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("sponge v2: read: %w", err)
	}

	var raw rawSchematic
	if err := nbt.UnmarshalEncoding(body, &raw, nbt.BigEndian); err != nil {
		return nil, fmt.Errorf("sponge v2: nbt: %w", err)
	}
	if raw.Version != 2 {
		return nil, fmt.Errorf("sponge v2: version %d unsupported in v1.0", raw.Version)
	}
	if raw.Width <= 0 || raw.Height <= 0 || raw.Length <= 0 {
		return nil, fmt.Errorf("sponge v2: invalid dimensions %dx%dx%d",
			raw.Width, raw.Height, raw.Length)
	}
	if len(raw.Palette) == 0 {
		return nil, fmt.Errorf("sponge v2: missing palette")
	}
	if len(raw.BlockData) == 0 {
		return nil, fmt.Errorf("sponge v2: missing BlockData")
	}

	w, h, l := int(uint16(raw.Width)), int(uint16(raw.Height)), int(uint16(raw.Length))
	totalCells := w * h * l

	// Build index→key lookup. The palette map is keyed by block name with
	// palette index as value. Build a slice indexed by palette index.
	highest := int32(-1)
	for _, v := range raw.Palette {
		if v > highest {
			highest = v
		}
	}
	_ = totalCells // available for future sanity checks
	indexToKey := make([]string, highest+1)
	for k, v := range raw.Palette {
		if v < 0 || int(v) >= len(indexToKey) {
			return nil, fmt.Errorf("sponge v2: palette index %d out of range", v)
		}
		indexToKey[v] = k
	}

	out := &schem.Schematic{
		Format:   schem.FormatSpongeV2,
		Width:    w,
		Height:   h,
		Length:   l,
		Blocks:   make([]schem.Block, 0, totalCells),
		Unknowns: schem.UnknownReport{Counts: map[string]int{}},
	}
	if len(raw.Offset) == 3 {
		out.Offset = [3]int{int(raw.Offset[0]), int(raw.Offset[1]), int(raw.Offset[2])}
	}

	br := bytes.NewReader(raw.BlockData)
	for y := 0; y < h; y++ {
		for z := 0; z < l; z++ {
			for x := 0; x < w; x++ {
				idx, _, err := readVarint(br)
				if err != nil {
					return nil, fmt.Errorf("sponge v2: varint at (%d,%d,%d): %w", x, y, z, err)
				}
				if int(idx) >= len(indexToKey) || indexToKey[idx] == "" {
					return nil, fmt.Errorf("sponge v2: at (%d,%d,%d): palette index %d out of range",
						x, y, z, idx)
				}
				key := indexToKey[idx]

				js, perr := palette.Decode(key)
				if perr != nil {
					return nil, fmt.Errorf("sponge v2: at (%d,%d,%d): palette key %q: %w",
						x, y, z, key, perr)
				}
				canonical := js.Canonical()
				res := resolveBlock(canonical)
				if !res.Recognized {
					out.Unknowns.Counts[res.RawKey]++
					out.Unknowns.Total++
				}
				out.Blocks = append(out.Blocks, schem.Block{
					Pos: [3]int{x, y, z}, Block: res.Block, Liquid: res.Liquid,
				})
			}
		}
	}
	return out, nil
}
