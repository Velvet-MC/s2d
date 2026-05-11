package translate

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Velvet-MC/s2d/translate/properties"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

//go:embed data/java_blocks.json
var javaBlocksJSON []byte

//go:embed data/bedrock_palette.1_26_20.nbt
var bedrockPaletteNBT []byte

//go:embed data/overrides.json
var overridesJSON []byte

// prismaBlock is the subset of the PrismarineJS minecraft-data block schema we
// consume for table construction.
type prismaBlock struct {
	Name         string        `json:"name"`
	ID           int           `json:"id"`
	States       []prismaState `json:"states"`
	DefaultState int           `json:"defaultState"`
}

type prismaState struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"` // "enum" | "bool" | "int"
	NumValues int      `json:"num_values"`
	Values    []string `json:"values"` // present for enum/bool; may be absent for int
}

// override is a manual remap entry from overrides.json. Either field may be
// empty; the comment is purely informational.
type override struct {
	BedrockIdentifier string   `json:"bedrock_identifier,omitempty"`
	SkipProperties    []string `json:"skip_properties,omitempty"`
	Comment           string   `json:"comment,omitempty"`
}

var (
	tableOnce sync.Once
	tableErr  error
	table     map[string]Result
)

// buildTable populates `table` from the embedded data sources. Invoked exactly
// once via tableOnce on the first Lookup call.
func buildTable() {
	defer func() {
		if r := recover(); r != nil {
			tableErr = fmt.Errorf("translate: build panic: %v", r)
		}
	}()

	var overrides map[string]override
	if err := json.Unmarshal(overridesJSON, &overrides); err != nil {
		tableErr = fmt.Errorf("translate: overrides.json: %w", err)
		return
	}

	var javaBlocks []prismaBlock
	if err := json.Unmarshal(javaBlocksJSON, &javaBlocks); err != nil {
		tableErr = fmt.Errorf("translate: java_blocks.json: %w", err)
		return
	}

	// Decode the Bedrock palette purely as a validation step. Dragonfly's
	// world.BlockByName is the authoritative resolver; the palette load is
	// retained as a sanity hook for future palette-coverage diagnostics.
	if _, err := loadBedrockPalette(bedrockPaletteNBT); err != nil {
		tableErr = fmt.Errorf("translate: bedrock palette: %w", err)
		return
	}

	t := make(map[string]Result, 30000)
	for _, jb := range javaBlocks {
		bedrockIdent := jb.Name
		ovr, hasOvr := overrides["minecraft:"+jb.Name]
		if hasOvr && ovr.BedrockIdentifier != "" {
			bedrockIdent = strings.TrimPrefix(ovr.BedrockIdentifier, "minecraft:")
		}

		propNames, propValues := extractProperties(jb)

		if len(propNames) == 0 {
			canonical := "minecraft:" + jb.Name
			res := translateOne(jb.Name, bedrockIdent, nil, false, ovr)
			res.RawKey = canonical
			t[canonical] = res
			continue
		}

		idx := make([]int, len(propNames))
		for {
			javaProps := make(map[string]string, len(propNames))
			for i, pn := range propNames {
				javaProps[pn] = propValues[i][idx[i]]
			}
			canonical := canonicalKey(jb.Name, javaProps)
			waterlogged := strings.EqualFold(javaProps["waterlogged"], "true")
			res := translateOne(jb.Name, bedrockIdent, javaProps, waterlogged, ovr)
			res.RawKey = canonical
			t[canonical] = res

			done := true
			for i := len(idx) - 1; i >= 0; i-- {
				idx[i]++
				if idx[i] < len(propValues[i]) {
					done = false
					break
				}
				idx[i] = 0
			}
			if done {
				break
			}
		}
	}

	if airBlock, ok := world.BlockByName("minecraft:air", nil); ok {
		air := Result{Block: airBlock, Recognized: true, RawKey: "minecraft:air"}
		t["minecraft:air"] = air
		t["minecraft:cave_air"] = air
		t["minecraft:void_air"] = air
	}

	table = t
}

// extractProperties returns the property names and per-name value lists for a
// Java block, sorted by name for deterministic canonical keys.
func extractProperties(jb prismaBlock) (names []string, values [][]string) {
	for _, st := range jb.States {
		var vals []string
		switch {
		case st.Type == "bool":
			// PrismarineJS leaves Values empty for bool; use Java convention.
			vals = []string{"false", "true"}
		case len(st.Values) > 0:
			vals = st.Values
		case st.NumValues > 0:
			for i := 0; i < st.NumValues; i++ {
				vals = append(vals, strconv.Itoa(i))
			}
		}
		if len(vals) == 0 {
			continue
		}
		names = append(names, st.Name)
		values = append(values, vals)
	}
	sortByName(names, values)
	return
}

func sortByName(names []string, values [][]string) {
	type pair struct {
		n string
		v []string
	}
	ps := make([]pair, len(names))
	for i := range names {
		ps[i] = pair{names[i], values[i]}
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].n < ps[j].n })
	for i := range ps {
		names[i] = ps[i].n
		values[i] = ps[i].v
	}
}

func canonicalKey(blockName string, props map[string]string) string {
	if len(props) == 0 {
		return "minecraft:" + blockName
	}
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("minecraft:")
	b.WriteString(blockName)
	b.WriteByte('[')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(props[k])
	}
	b.WriteByte(']')
	return b.String()
}

// translateOne resolves a single Java state combination to a Bedrock Result.
func translateOne(javaName, bedrockIdent string, javaProps map[string]string,
	waterlogged bool, ovr override) Result {

	bedrockProps := map[string]any{}
	skip := map[string]struct{}{}
	for _, sp := range ovr.SkipProperties {
		skip[sp] = struct{}{}
	}

	for k, v := range javaProps {
		if _, drop := skip[k]; drop {
			continue
		}
		conv, ok := properties.Registry[k]
		if !ok {
			continue
		}
		bp, bv, ok := conv(v, bedrockIdent)
		if !ok {
			continue
		}
		bedrockProps[bp] = bv
	}

	full := "minecraft:" + bedrockIdent
	b, ok := world.BlockByName(full, bedrockProps)
	if !ok {
		// Try without props (Bedrock blocks may have implicit defaults).
		if b2, ok2 := world.BlockByName(full, nil); ok2 {
			b = b2
			ok = true
		}
	}
	res := Result{Recognized: ok}
	if ok {
		res.Block = b
	} else {
		res.Block = MissingBlock()
	}
	if waterlogged {
		if l, lok := world.BlockByName("minecraft:water", map[string]any{"liquid_depth": int32(0)}); lok {
			if liq, isLiq := l.(world.Liquid); isLiq {
				res.Liquid = liq
			}
		}
	}
	return res
}

// loadBedrockPalette decodes Geyser's gzipped block_palette.<ver>.nbt. The
// palette is shipped as gzipped Java NBT (BigEndian).
func loadBedrockPalette(data []byte) (map[string]any, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gunzip palette: %w", err)
	}
	defer gz.Close()
	body, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("read palette: %w", err)
	}

	var root map[string]any
	if err := nbt.UnmarshalEncoding(body, &root, nbt.BigEndian); err != nil {
		if err2 := nbt.UnmarshalEncoding(body, &root, nbt.LittleEndian); err2 != nil {
			return nil, fmt.Errorf("decode palette: BigEndian=%v LittleEndian=%v", err, err2)
		}
	}
	return root, nil
}
