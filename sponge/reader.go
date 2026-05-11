// Package sponge reads Sponge Schematic (.schem) files. It depends on
// gophertunnel/minecraft/nbt for big-endian Java NBT decoding.
//
// Translation to Dragonfly world.Block is performed by translate.Lookup.
package sponge

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/Velvet-MC/s2d/palette"
	"github.com/Velvet-MC/s2d/schem"
	"github.com/Velvet-MC/s2d/translate"
)

var (
	decodePaletteKey = palette.Decode
	lookupJavaState  = translate.Lookup
)

// Read parses a Sponge schematic from r.
//
// The reader is permissive about both the wrapper shape (real WorldEdit
// output nests fields under a top-level "Schematic" compound; hand-built
// fixtures usually don't) and about optional fields (Metadata,
// BlockEntities, Entities, biome data) which are accepted and silently
// ignored in v1.0.
func Read(r io.Reader) (*schem.Schematic, error) {
	var blocks []schem.Block
	info, err := Scan(r, func(b schem.Block) error {
		blocks = append(blocks, b)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &schem.Schematic{
		Format:   info.Format,
		Width:    info.Width,
		Height:   info.Height,
		Length:   info.Length,
		Offset:   info.Offset,
		Blocks:   blocks,
		Unknowns: info.Unknowns,
	}, nil
}

// Scan parses a Sponge schematic from r and calls yield for each translated
// block without materialising a full schem.Schematic.Blocks slice.
func Scan(r io.Reader, yield schem.BlockHandler) (schem.ScanInfo, error) {
	return ScanWithInfo(r, nil, yield)
}

// ScanWithInfo parses a Sponge schematic from r, calls onInfo once after
// dimensions are known, then calls yield for each translated block.
func ScanWithInfo(r io.Reader, onInfo schem.InfoHandler, yield schem.BlockHandler) (schem.ScanInfo, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return schem.ScanInfo{}, fmt.Errorf("sponge: gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()
	raw, err := decodeSpongeNBT(gz)
	if err != nil {
		return schem.ScanInfo{}, fmt.Errorf("sponge: nbt: %w", err)
	}

	version := raw.Version
	format := schem.FormatSpongeV2
	rawPalette := raw.Palette
	blockData := raw.BlockData
	blockEntities := raw.BlockEntities
	blockDataField := "BlockData"
	switch version {
	case 2:
	case 3:
		format = schem.FormatSpongeV3
		rawPalette = raw.Blocks.Palette
		blockData = raw.Blocks.Data
		blockEntities = raw.Blocks.BlockEntities
		blockDataField = "Blocks.Data"
	default:
		return schem.ScanInfo{}, fmt.Errorf("sponge: version %d unsupported", version)
	}

	if raw.Width <= 0 || raw.Height <= 0 || raw.Length <= 0 {
		return schem.ScanInfo{}, fmt.Errorf("sponge v%d: invalid dimensions %dx%dx%d", version, raw.Width, raw.Height, raw.Length)
	}
	if len(rawPalette) == 0 {
		return schem.ScanInfo{}, fmt.Errorf("sponge v%d: missing palette", version)
	}
	if len(blockData) == 0 {
		return schem.ScanInfo{}, fmt.Errorf("sponge v%d: missing or invalid %s", version, blockDataField)
	}

	w, h, l := int(uint16(raw.Width)), int(uint16(raw.Height)), int(uint16(raw.Length))

	// Build index-to-key lookup. The palette map is keyed by block name with
	// palette index as value. Build a slice indexed by palette index.
	highest := int32(-1)
	for _, v := range rawPalette {
		if v > highest {
			highest = v
		}
	}
	indexToKey := make([]string, highest+1)
	for k, v := range rawPalette {
		if v < 0 || int(v) >= len(indexToKey) {
			return schem.ScanInfo{}, fmt.Errorf("sponge v%d: palette index %d out of range", version, v)
		}
		indexToKey[v] = k
	}
	indexToResult := make([]translate.Result, len(indexToKey))
	indexToBedrockState := make([]schem.BedrockState, len(indexToKey))
	for i, key := range indexToKey {
		if key == "" {
			continue
		}
		js, perr := decodePaletteKey(key)
		if perr != nil {
			return schem.ScanInfo{}, fmt.Errorf("sponge v%d: palette key %q: %w", version, key, perr)
		}
		res := lookupJavaState(js.Canonical())
		indexToResult[i] = res
		indexToBedrockState[i] = schem.BedrockState{Name: res.BedrockState.Name, Properties: res.BedrockState.Properties}
	}
	blockEntityNBT := bannerBlockEntities(blockEntities)

	info := schem.ScanInfo{
		Format:      format,
		Width:       w,
		Height:      h,
		Length:      l,
		PaletteSize: len(indexToKey),
		Unknowns:    schem.UnknownReport{Counts: map[string]int{}},
	}
	if len(raw.Offset) == 3 {
		info.Offset = [3]int{int(raw.Offset[0]), int(raw.Offset[1]), int(raw.Offset[2])}
	}
	if onInfo != nil {
		if err := onInfo(info); err != nil {
			return info, fmt.Errorf("sponge v%d: info: %w", version, err)
		}
	}

	br := bytes.NewReader(blockData)
	for y := 0; y < h; y++ {
		for z := 0; z < l; z++ {
			for x := 0; x < w; x++ {
				idx, _, err := readVarint(br)
				if err != nil {
					return info, fmt.Errorf("sponge v%d: varint at (%d,%d,%d): %w", version, x, y, z, err)
				}
				if int(idx) >= len(indexToKey) || indexToKey[idx] == "" {
					return info, fmt.Errorf("sponge v%d: at (%d,%d,%d): palette index %d out of range", version, x, y, z, idx)
				}
				res := indexToResult[idx]
				pos := [3]int{x, y, z}
				if !res.Recognized {
					info.Unknowns.Counts[res.RawKey]++
					info.Unknowns.Total++
				}
				block := res.Block
				if data, ok := blockEntityNBT[pos]; ok {
					block = translate.MergeBlockNBT(block, data)
				}
				if err := yield(schem.Block{
					Pos:            pos,
					Block:          block,
					Liquid:         res.Liquid,
					BedrockState:   indexToBedrockState[idx],
					PaletteIndex:   idx,
					PaletteIndexOK: true,
				}); err != nil {
					return info, fmt.Errorf("sponge v%d: yield at (%d,%d,%d): %w", version, x, y, z, err)
				}
			}
		}
	}
	return info, nil
}

func bannerBlockEntities(entries []rawBlockEntity) map[[3]int]map[string]any {
	if len(entries) == 0 {
		return nil
	}
	out := map[[3]int]map[string]any{}
	for _, entry := range entries {
		if !strings.Contains(entry.ID, "banner") || len(entry.Pos) != 3 {
			continue
		}
		data := map[string]any{"id": "Banner"}
		if entry.Patterns != nil {
			data["Patterns"] = entry.Patterns
		}
		if entry.HasBase {
			data["Base"] = entry.Base
		}
		out[[3]int{int(entry.Pos[0]), int(entry.Pos[1]), int(entry.Pos[2])}] = data
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
