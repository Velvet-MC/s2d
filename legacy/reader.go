package legacy

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/sandertv/gophertunnel/minecraft/nbt"

	"github.com/Clxser/S2D/palette"
	"github.com/Clxser/S2D/schem"
	"github.com/Clxser/S2D/translate"
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
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("legacy: gzip: %w", err)
	}
	defer gz.Close()
	body, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("legacy: read: %w", err)
	}

	var raw rawLegacy
	if err := nbt.UnmarshalEncoding(body, &raw, nbt.BigEndian); err != nil {
		return nil, fmt.Errorf("legacy: nbt: %w", err)
	}
	if raw.Materials != "" && raw.Materials != "Alpha" {
		return nil, fmt.Errorf("legacy: unsupported Materials %q (expect Alpha for Java)", raw.Materials)
	}
	if raw.Width <= 0 || raw.Height <= 0 || raw.Length <= 0 {
		return nil, fmt.Errorf("legacy: invalid dimensions %dx%dx%d",
			raw.Width, raw.Height, raw.Length)
	}

	w, h, l := int(raw.Width), int(raw.Height), int(raw.Length)
	total := w * h * l
	if len(raw.Blocks) != total {
		return nil, fmt.Errorf("legacy: Blocks length %d != %dx%dx%d=%d",
			len(raw.Blocks), w, h, l, total)
	}

	out := &schem.Schematic{
		Format:   schem.FormatLegacy,
		Width:    w,
		Height:   h,
		Length:   l,
		Blocks:   make([]schem.Block, 0, total),
		Unknowns: schem.UnknownReport{Counts: map[string]int{}},
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
				if len(raw.Data) > 0 {
					dIdx := i / 2
					if dIdx < len(raw.Data) {
						b := raw.Data[dIdx]
						if i%2 == 0 {
							data = int(b>>4) & 0xF
						} else {
							data = int(b) & 0xF
						}
					}
				}

				if id == 0 {
					// Air. Translate via canonical key so the Bedrock-side
					// air block (or whatever the missing fallback returns)
					// fills the cell.
					res := translate.Lookup("minecraft:air")
					if !res.Recognized {
						out.Unknowns.Counts["minecraft:air"]++
						out.Unknowns.Total++
					}
					out.Blocks = append(out.Blocks, schem.Block{
						Pos: [3]int{x, y, z}, Block: res.Block, Liquid: res.Liquid,
					})
					continue
				}

				key, ok := Lookup(id, data)
				if !ok {
					rawKey := fmt.Sprintf("legacy:%d:%d", id, data)
					out.Unknowns.Counts[rawKey]++
					out.Unknowns.Total++
					out.Blocks = append(out.Blocks, schem.Block{
						Pos: [3]int{x, y, z}, Block: translate.MissingBlock(),
					})
					continue
				}

				js, perr := palette.Decode(key)
				if perr != nil {
					return nil, fmt.Errorf("legacy: at (%d,%d,%d): decoding %q: %w",
						x, y, z, key, perr)
				}
				canonical := js.Canonical()
				res := translate.Lookup(canonical)
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
