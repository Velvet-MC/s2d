package legacy

import (
	"compress/gzip"
	"fmt"
	"io"
	"maps"

	"github.com/sandertv/gophertunnel/minecraft/nbt"

	"github.com/Velvet-MC/s2d/palette"
	"github.com/Velvet-MC/s2d/schem"
	"github.com/Velvet-MC/s2d/translate"
)

// rawLegacy mirrors the MCEdit `.schematic` NBT root.
type rawLegacy struct {
	Materials string `nbt:"Materials"`
	Width     int16  `nbt:"Width"`
	Height    int16  `nbt:"Height"`
	Length    int16  `nbt:"Length"`
	Blocks    []byte `nbt:"Blocks"`
	Data      []byte `nbt:"Data"`
	AddBlocks []byte `nbt:"AddBlocks,omitempty"`
}

// Read parses a legacy MCEdit `.schematic` from r.
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

// Scan parses a legacy MCEdit `.schematic` from r and calls yield for each
// translated block without materialising a full schem.Schematic.Blocks slice.
func Scan(r io.Reader, yield schem.BlockHandler) (schem.ScanInfo, error) {
	return ScanWithInfo(r, nil, yield)
}

// ScanWithInfo parses a legacy MCEdit `.schematic` from r, calls onInfo once
// after dimensions are known, then calls yield for each translated block.
func ScanWithInfo(r io.Reader, onInfo schem.InfoHandler, yield schem.BlockHandler) (schem.ScanInfo, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return schem.ScanInfo{}, fmt.Errorf("legacy: gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()
	body, err := io.ReadAll(gz)
	if err != nil {
		return schem.ScanInfo{}, fmt.Errorf("legacy: read: %w", err)
	}

	var raw rawLegacy
	if err := nbt.UnmarshalEncoding(body, &raw, nbt.BigEndian); err != nil {
		return schem.ScanInfo{}, fmt.Errorf("legacy: nbt: %w", err)
	}
	if raw.Materials != "" && raw.Materials != "Alpha" {
		return schem.ScanInfo{}, fmt.Errorf("legacy: unsupported Materials %q (expect Alpha for Java)", raw.Materials)
	}
	if raw.Width <= 0 || raw.Height <= 0 || raw.Length <= 0 {
		return schem.ScanInfo{}, fmt.Errorf("legacy: invalid dimensions %dx%dx%d", raw.Width, raw.Height, raw.Length)
	}

	w, h, l := int(raw.Width), int(raw.Height), int(raw.Length)
	total := w * h * l
	if len(raw.Blocks) != total {
		return schem.ScanInfo{}, fmt.Errorf("legacy: Blocks length %d != %dx%dx%d=%d", len(raw.Blocks), w, h, l, total)
	}
	if len(raw.Data) != total {
		return schem.ScanInfo{}, fmt.Errorf("legacy: Data length %d != %dx%dx%d=%d", len(raw.Data), w, h, l, total)
	}

	info := schem.ScanInfo{
		Format:   schem.FormatLegacy,
		Width:    w,
		Height:   h,
		Length:   l,
		Unknowns: schem.UnknownReport{Counts: map[string]int{}},
	}
	if onInfo != nil {
		if err := onInfo(info); err != nil {
			return info, fmt.Errorf("legacy: info: %w", err)
		}
	}

	for y := 0; y < h; y++ {
		for z := 0; z < l; z++ {
			for x := 0; x < w; x++ {
				i := x + z*w + y*w*l
				low8 := int(raw.Blocks[i])
				high4 := 0
				if len(raw.AddBlocks) > 0 {
					addIdx := i / 2
					if addIdx < len(raw.AddBlocks) {
						b := raw.AddBlocks[addIdx]
						if i%2 == 0 {
							high4 = int(b>>4) & 0xF
						} else {
							high4 = int(b) & 0xF
						}
					}
				}
				id := (high4 << 8) | low8
				data := 0
				if i < len(raw.Data) {
					data = int(raw.Data[i]) & 0xF
				}

				if id == 0 {
					// Air. Translate via canonical key so the Bedrock-side
					// air block (or whatever the missing fallback returns)
					// fills the cell.
					res := translate.Lookup("minecraft:air")
					if !res.Recognized {
						info.Unknowns.Counts["minecraft:air"]++
						info.Unknowns.Total++
					}
					if err := yield(schem.Block{
						Pos:          [3]int{x, y, z},
						Block:        res.Block,
						Liquid:       res.Liquid,
						BedrockState: schem.BedrockState{Name: res.BedrockState.Name, Properties: maps.Clone(res.BedrockState.Properties)},
					}); err != nil {
						return info, fmt.Errorf("legacy: yield at (%d,%d,%d): %w", x, y, z, err)
					}
					continue
				}

				key, ok := Lookup(id, data)
				if !ok {
					rawKey := fmt.Sprintf("legacy:%d:%d", id, data)
					info.Unknowns.Counts[rawKey]++
					info.Unknowns.Total++
					if err := yield(schem.Block{Pos: [3]int{x, y, z}, Block: translate.MissingBlock()}); err != nil {
						return info, fmt.Errorf("legacy: yield at (%d,%d,%d): %w", x, y, z, err)
					}
					continue
				}

				js, perr := palette.Decode(key)
				if perr != nil {
					return info, fmt.Errorf("legacy: at (%d,%d,%d): decoding %q: %w", x, y, z, key, perr)
				}
				canonical := js.Canonical()
				res := translate.Lookup(canonical)
				if !res.Recognized {
					info.Unknowns.Counts[res.RawKey]++
					info.Unknowns.Total++
				}
				if err := yield(schem.Block{
					Pos:          [3]int{x, y, z},
					Block:        res.Block,
					Liquid:       res.Liquid,
					BedrockState: schem.BedrockState{Name: res.BedrockState.Name, Properties: maps.Clone(res.BedrockState.Properties)},
				}); err != nil {
					return info, fmt.Errorf("legacy: yield at (%d,%d,%d): %w", x, y, z, err)
				}
			}
		}
	}
	return info, nil
}
