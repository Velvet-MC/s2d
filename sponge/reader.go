// Package sponge reads Sponge Schematic v2 (.schem) files. It depends on
// gophertunnel/minecraft/nbt for big-endian Java NBT decoding.
//
// Translation to Dragonfly world.Block is performed by translate.Lookup.
package sponge

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"reflect"

	"github.com/sandertv/gophertunnel/minecraft/nbt"

	"github.com/Velvet-MC/s2d/palette"
	"github.com/Velvet-MC/s2d/schem"
	"github.com/Velvet-MC/s2d/translate"
)

var (
	decodePaletteKey = palette.Decode
	lookupJavaState  = translate.Lookup
)

// Read parses a Sponge v2 schematic from r.
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

// Scan parses a Sponge v2 schematic from r and calls yield for each translated
// block without materialising a full schem.Schematic.Blocks slice.
func Scan(r io.Reader, yield schem.BlockHandler) (schem.ScanInfo, error) {
	return ScanWithInfo(r, nil, yield)
}

// ScanWithInfo parses a Sponge v2 schematic from r, calls onInfo once after
// dimensions are known, then calls yield for each translated block.
func ScanWithInfo(r io.Reader, onInfo schem.InfoHandler, yield schem.BlockHandler) (schem.ScanInfo, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return schem.ScanInfo{}, fmt.Errorf("sponge v2: gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()
	body, err := io.ReadAll(gz)
	if err != nil {
		return schem.ScanInfo{}, fmt.Errorf("sponge v2: read: %w", err)
	}

	root, err := decodePermissive(body)
	if err != nil {
		return schem.ScanInfo{}, fmt.Errorf("sponge v2: nbt: %w", err)
	}

	// Real WorldEdit files wrap the schematic in a top-level "Schematic"
	// compound; some other writers put fields at the root.
	if inner, ok := root["Schematic"].(map[string]any); ok {
		root = inner
	}

	version, _ := asInt32(root["Version"])
	if version != 2 {
		return schem.ScanInfo{}, fmt.Errorf("sponge v2: version %d unsupported in v1.0", version)
	}

	width, _ := asInt32(root["Width"])
	height, _ := asInt32(root["Height"])
	length, _ := asInt32(root["Length"])
	if width <= 0 || height <= 0 || length <= 0 {
		return schem.ScanInfo{}, fmt.Errorf("sponge v2: invalid dimensions %dx%dx%d", width, height, length)
	}

	rawPalette, ok := root["Palette"].(map[string]any)
	if !ok || len(rawPalette) == 0 {
		return schem.ScanInfo{}, fmt.Errorf("sponge v2: missing palette")
	}
	blockData, err := asByteSlice(root["BlockData"])
	if err != nil || len(blockData) == 0 {
		return schem.ScanInfo{}, fmt.Errorf("sponge v2: missing or invalid BlockData")
	}

	w, h, l := int(uint16(width)), int(uint16(height)), int(uint16(length))

	// Build index-to-key lookup. The palette map is keyed by block name with
	// palette index as value. Build a slice indexed by palette index.
	highest := int32(-1)
	for _, anyV := range rawPalette {
		v, _ := asInt32(anyV)
		if v > highest {
			highest = v
		}
	}
	indexToKey := make([]string, highest+1)
	for k, anyV := range rawPalette {
		v, _ := asInt32(anyV)
		if v < 0 || int(v) >= len(indexToKey) {
			return schem.ScanInfo{}, fmt.Errorf("sponge v2: palette index %d out of range", v)
		}
		indexToKey[v] = k
	}
	indexToResult := make([]translate.Result, len(indexToKey))
	for i, key := range indexToKey {
		if key == "" {
			continue
		}
		js, perr := decodePaletteKey(key)
		if perr != nil {
			return schem.ScanInfo{}, fmt.Errorf("sponge v2: palette key %q: %w", key, perr)
		}
		indexToResult[i] = lookupJavaState(js.Canonical())
	}

	info := schem.ScanInfo{
		Format:      schem.FormatSpongeV2,
		Width:       w,
		Height:      h,
		Length:      l,
		PaletteSize: len(indexToKey),
		Unknowns:    schem.UnknownReport{Counts: map[string]int{}},
	}
	if off, err := asInt32Slice(root["Offset"]); err == nil && len(off) == 3 {
		info.Offset = [3]int{int(off[0]), int(off[1]), int(off[2])}
	}
	if onInfo != nil {
		if err := onInfo(info); err != nil {
			return info, fmt.Errorf("sponge v2: info: %w", err)
		}
	}

	br := bytes.NewReader(blockData)
	for y := 0; y < h; y++ {
		for z := 0; z < l; z++ {
			for x := 0; x < w; x++ {
				idx, _, err := readVarint(br)
				if err != nil {
					return info, fmt.Errorf("sponge v2: varint at (%d,%d,%d): %w", x, y, z, err)
				}
				if int(idx) >= len(indexToKey) || indexToKey[idx] == "" {
					return info, fmt.Errorf("sponge v2: at (%d,%d,%d): palette index %d out of range", x, y, z, idx)
				}
				res := indexToResult[idx]
				if !res.Recognized {
					info.Unknowns.Counts[res.RawKey]++
					info.Unknowns.Total++
				}
				if err := yield(schem.Block{
					Pos:            [3]int{x, y, z},
					Block:          res.Block,
					Liquid:         res.Liquid,
					PaletteIndex:   idx,
					PaletteIndexOK: true,
				}); err != nil {
					return info, fmt.Errorf("sponge v2: yield at (%d,%d,%d): %w", x, y, z, err)
				}
			}
		}
	}
	return info, nil
}

// decodePermissive decodes Java big-endian NBT into a generic map so the
// reader can ignore unknown/optional fields and handle byte-array tags
// regardless of whether they were emitted as TAG_ByteArray or TAG_List.
func decodePermissive(body []byte) (map[string]any, error) {
	var root map[string]any
	if err := nbt.UnmarshalEncoding(body, &root, nbt.BigEndian); err != nil {
		return nil, err
	}
	return root, nil
}

// asInt32 normalises any integer-shaped NBT value to int32. Sponge v2 uses
// short for dimensions but writes them as signed int16; some fields are
// int32 directly. Both are accepted.
func asInt32(v any) (int32, bool) {
	switch x := v.(type) {
	case int32:
		return x, true
	case int16:
		return int32(x), true
	case int8:
		return int32(x), true
	case int64:
		return int32(x), true
	case int:
		return int32(x), true
	}
	return 0, false
}

// asInt32Slice extracts a slice of int32 from an NBT int-array tag.
func asInt32Slice(v any) ([]int32, error) {
	if v == nil {
		return nil, fmt.Errorf("nil")
	}
	if s, ok := v.([]int32); ok {
		return s, nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Array && rv.Type().Elem().Kind() == reflect.Int32 {
		out := make([]int32, rv.Len())
		reflect.Copy(reflect.ValueOf(out), rv)
		return out, nil
	}
	return nil, fmt.Errorf("expected int32 array, got %T", v)
}

// asByteSlice extracts a []byte from an NBT byte-array tag. gophertunnel's
// strict decoder produces TAG_ByteArray as a fixed-size [N]byte when the
// destination is `any`; this helper copies it into a slice.
func asByteSlice(v any) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("nil")
	}
	if b, ok := v.([]byte); ok {
		return b, nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Array && rv.Type().Elem().Kind() == reflect.Uint8 {
		out := make([]byte, rv.Len())
		reflect.Copy(reflect.ValueOf(out), rv)
		return out, nil
	}
	// Some encoders emit byte arrays as TAG_List of TAG_Byte. Accept []any
	// of byte/int8 too.
	if s, ok := v.([]any); ok {
		out := make([]byte, len(s))
		for i, e := range s {
			switch x := e.(type) {
			case byte:
				out[i] = x
			case int8:
				out[i] = byte(x)
			default:
				return nil, fmt.Errorf("byte slice element %d: %T", i, e)
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("expected byte array, got %T", v)
}
