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

	"github.com/Velvet-MC/s2d/palette"
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

	bedrockPaletteForLookup *bedrockPaletteIndex
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

	bedrockPalette, err := loadBedrockPalette(bedrockPaletteNBT)
	if err != nil {
		tableErr = fmt.Errorf("translate: bedrock palette: %w", err)
		return
	}
	bedrockPaletteForLookup = bedrockPalette

	t := make(map[string]Result, 30000)
	for _, jb := range javaBlocks {
		ovr := overrides["minecraft:"+jb.Name]
		bedrockIdent := baseBedrockIdentifier(jb.Name, ovr)

		propNames, propValues := extractProperties(jb)

		if len(propNames) == 0 {
			canonical := "minecraft:" + jb.Name
			res := translateOne(bedrockPalette, jb.Name, bedrockIdent, nil, false, ovr)
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
			resolvedIdent := adjustBedrockIdentifier(jb.Name, bedrockIdent, javaProps)
			canonical := canonicalKey(jb.Name, javaProps)
			waterlogged := strings.EqualFold(javaProps["waterlogged"], "true")
			res := translateOne(bedrockPalette, jb.Name, resolvedIdent, javaProps, waterlogged, ovr)
			res.RawKey = canonical
			t[canonical] = res
			addLookupAliases(t, jb.Name, javaProps, res)

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
		air := Result{
			Block:        airBlock,
			BedrockState: BedrockState{Name: "minecraft:air"},
			Recognized:   true,
			RawKey:       "minecraft:air",
		}
		t["minecraft:air"] = air
		t["minecraft:cave_air"] = air
		t["minecraft:void_air"] = air
	}

	table = t
}

func lookupDynamicJavaState(raw string) (Result, bool) {
	js, err := palette.Decode(raw)
	if err != nil || js.Namespace != "minecraft" {
		return Result{}, false
	}
	bedrockIdent := baseBedrockIdentifier(js.Name, override{})
	resolvedIdent := adjustBedrockIdentifier(js.Name, bedrockIdent, js.Props)
	waterlogged := strings.EqualFold(js.Props["waterlogged"], "true")
	res := translateOne(bedrockPaletteForLookup, js.Name, resolvedIdent, js.Props, waterlogged, override{})
	if !res.Recognized {
		return Result{}, false
	}
	res.RawKey = canonicalKey(js.Name, js.Props)
	return res, true
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

func addLookupAliases(t map[string]Result, javaName string, javaProps map[string]string, res Result) {
	if !strings.HasSuffix(javaName, "_leaves") && javaName != "flowering_azalea_leaves" {
		return
	}
	withoutDistance := cloneStringMap(javaProps)
	delete(withoutDistance, "distance")
	putAlias(t, javaName, withoutDistance, res)

	if javaProps["waterlogged"] == "false" {
		withoutWaterlogged := cloneStringMap(javaProps)
		delete(withoutWaterlogged, "waterlogged")
		putAlias(t, javaName, withoutWaterlogged, res)

		withoutBoth := cloneStringMap(withoutWaterlogged)
		delete(withoutBoth, "distance")
		putAlias(t, javaName, withoutBoth, res)
	}
}

func putAlias(t map[string]Result, javaName string, javaProps map[string]string, res Result) {
	key := canonicalKey(javaName, javaProps)
	if _, exists := t[key]; !exists {
		alias := res
		alias.RawKey = key
		t[key] = alias
	}
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
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
func translateOne(palette *bedrockPaletteIndex, javaName, bedrockIdent string, javaProps map[string]string,
	waterlogged bool, ovr override) Result {

	bedrockProps := map[string]any{}
	if bedrockIdent == "glow_lichen" {
		bedrockProps["multi_face_direction_bits"] = glowLichenFaces(javaProps)
	}
	skip := map[string]struct{}{}
	for _, sp := range ovr.SkipProperties {
		skip[sp] = struct{}{}
	}

	for k, v := range javaProps {
		if bedrockIdent == "glow_lichen" && isGlowLichenFace(k) {
			continue
		}
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
	applyJavaDerivedBedrockProperties(javaName, bedrockIdent, bedrockProps)
	applyImplicitBedrockProperties(bedrockIdent, bedrockProps)

	full := "minecraft:" + bedrockIdent
	state, ok := palette.lookup(full, bedrockProps)
	if !ok {
		state, ok = palette.defaultState(full)
	}
	res := Result{BedrockState: state.Clone(), Recognized: ok}
	if ok {
		res.Block = NewStateBlock(state)
	} else {
		res.Block = MissingBlock()
	}
	if waterlogged || needsWaterLayer(bedrockIdent) {
		if l, lok := world.BlockByName("minecraft:water", map[string]any{"liquid_depth": int32(0)}); lok {
			if liq, isLiq := l.(world.Liquid); isLiq {
				res.Liquid = liq
			}
		}
	}
	return res
}

func baseBedrockIdentifier(javaName string, ovr override) string {
	if ovr.BedrockIdentifier != "" {
		return strings.TrimPrefix(ovr.BedrockIdentifier, "minecraft:")
	}
	if alias, ok := bedrockIdentifierAliases[javaName]; ok {
		return alias
	}
	if strings.HasPrefix(javaName, "potted_") {
		return "flower_pot"
	}
	if strings.HasSuffix(javaName, "_wall_hanging_sign") {
		return strings.TrimSuffix(javaName, "_wall_hanging_sign") + "_hanging_sign"
	}
	if strings.HasSuffix(javaName, "_wall_banner") {
		return "wall_banner"
	}
	if strings.HasSuffix(javaName, "_banner") {
		return "standing_banner"
	}
	if strings.HasSuffix(javaName, "_bed") {
		return "bed"
	}
	if strings.HasSuffix(javaName, "_wall_skull") {
		return strings.TrimSuffix(javaName, "_wall_skull") + "_skull"
	}
	if strings.HasSuffix(javaName, "_wall_head") {
		return strings.TrimSuffix(javaName, "_wall_head") + "_head"
	}
	return javaName
}

func needsWaterLayer(bedrockIdent string) bool {
	switch bedrockIdent {
	case "seagrass", "kelp", "kelp_plant":
		return true
	default:
		return false
	}
}

var bedrockIdentifierAliases = map[string]string{
	"chain":                        "iron_chain",
	"cobblestone_stairs":           "stone_stairs",
	"cobweb":                       "web",
	"dead_bush":                    "deadbush",
	"bubble_column":                "water",
	"end_stone_bricks":             "end_bricks",
	"end_stone_brick_stairs":       "end_brick_stairs",
	"flowering_azalea_leaves":      "azalea_leaves_flowered",
	"grass":                        "short_grass",
	"lily_pad":                     "waterlily",
	"magma_block":                  "magma",
	"melon":                        "melon_block",
	"nether_portal":                "portal",
	"note_block":                   "noteblock",
	"oak_button":                   "wooden_button",
	"oak_door":                     "wooden_door",
	"oak_fence_gate":               "fence_gate",
	"oak_pressure_plate":           "wooden_pressure_plate",
	"oak_sign":                     "standing_sign",
	"oak_trapdoor":                 "trapdoor",
	"oak_wall_sign":                "wall_sign",
	"dark_oak_sign":                "darkoak_standing_sign",
	"dark_oak_wall_sign":           "darkoak_wall_sign",
	"spruce_sign":                  "spruce_standing_sign",
	"spruce_wall_sign":             "spruce_wall_sign",
	"birch_sign":                   "birch_standing_sign",
	"birch_wall_sign":              "birch_wall_sign",
	"jungle_sign":                  "jungle_standing_sign",
	"jungle_wall_sign":             "jungle_wall_sign",
	"acacia_sign":                  "acacia_standing_sign",
	"acacia_wall_sign":             "acacia_wall_sign",
	"mangrove_sign":                "mangrove_standing_sign",
	"mangrove_wall_sign":           "mangrove_wall_sign",
	"cherry_sign":                  "cherry_standing_sign",
	"cherry_wall_sign":             "cherry_wall_sign",
	"crimson_sign":                 "crimson_standing_sign",
	"crimson_wall_sign":            "crimson_wall_sign",
	"warped_sign":                  "warped_standing_sign",
	"warped_wall_sign":             "warped_wall_sign",
	"prismarine_brick_stairs":      "prismarine_bricks_stairs",
	"stone_slab":                   "smooth_stone_slab",
	"spawner":                      "mob_spawner",
	"terracotta":                   "hardened_clay",
	"wall_torch":                   "torch",
	"bricks":                       "brick_block",
	"budding_amethyst":             "amethyst_block",
	"pointed_dripstone":            "dripstone_block",
	"rooted_dirt":                  "dirt_with_roots",
	"sugar_cane":                   "reeds",
	"tall_seagrass":                "seagrass",
	"oak_sapling":                  "short_grass",
	"spruce_sapling":               "short_grass",
	"birch_sapling":                "short_grass",
	"jungle_sapling":               "short_grass",
	"acacia_sapling":               "short_grass",
	"dark_oak_sapling":             "short_grass",
	"mangrove_propagule":           "short_grass",
	"cherry_sapling":               "short_grass",
	"bamboo_stairs":                "oak_stairs",
	"bamboo_mosaic_stairs":         "oak_stairs",
	"iron_trapdoor":                "trapdoor",
	"iron_door":                    "wooden_door",
	"dirt_path":                    "grass_path",
	"light_gray_glazed_terracotta": "silver_glazed_terracotta",
	"red_nether_bricks":            "red_nether_brick",
	"small_dripleaf":               "small_dripleaf_block",
	"snow_block":                   "snow",
	"beetroots":                    "beetroot",
	"water_cauldron":               "cauldron",
	"comparator":                   "unpowered_comparator",
	"repeater":                     "unpowered_repeater",
	"cave_vines_plant":             "cave_vines",
	"attached_melon_stem":          "melon_stem",
	"attached_pumpkin_stem":        "pumpkin_stem",
	"bamboo_sign":                  "bamboo_standing_sign",
	"bamboo_wall_sign":             "bamboo_wall_sign",
	"big_dripleaf_stem":            "big_dripleaf",
	"frogspawn":                    "frog_spawn",
	"jack_o_lantern":               "lit_pumpkin",
	"kelp_plant":                   "kelp",
	"lava_cauldron":                "cauldron",
	"moving_piston":                "moving_block",
	"nether_bricks":                "nether_brick",
	"nether_quartz_ore":            "quartz_ore",
	"piston_head":                  "piston_arm_collision",
	"powder_snow_cauldron":         "cauldron",
	"powered_rail":                 "golden_rail",
	"redstone_wall_torch":          "redstone_torch",
	"shulker_box":                  "undyed_shulker_box",
	"slime_block":                  "slime",
	"soul_wall_torch":              "soul_torch",
	"tripwire":                     "trip_wire",
	"twisting_vines_plant":         "twisting_vines",
	"waxed_copper_block":           "waxed_copper",
	"weeping_vines_plant":          "weeping_vines",
}

func adjustBedrockIdentifier(javaName, bedrockIdent string, javaProps map[string]string) string {
	switch javaName {
	case "comparator":
		if javaProps["powered"] == "true" {
			return "powered_comparator"
		}
		return "unpowered_comparator"
	case "repeater":
		if javaProps["powered"] == "true" {
			return "powered_repeater"
		}
		return "unpowered_repeater"
	case "cave_vines_plant":
		if javaProps["berries"] == "true" {
			return "cave_vines_body_with_berries"
		}
		return "cave_vines"
	case "redstone_wall_torch":
		if javaProps["lit"] == "false" {
			return "unlit_redstone_torch"
		}
		return "redstone_torch"
	}
	if javaName == "light" {
		return fmt.Sprintf("light_block_%s", javaProps["level"])
	}
	if javaProps["lit"] == "true" {
		switch javaName {
		case "redstone_ore":
			return "lit_redstone_ore"
		case "deepslate_redstone_ore":
			return "lit_deepslate_redstone_ore"
		}
	}
	if javaProps["type"] == "double" && strings.HasSuffix(bedrockIdent, "_slab") {
		if strings.HasSuffix(bedrockIdent, "cut_copper_slab") {
			return strings.TrimSuffix(bedrockIdent, "cut_copper_slab") + "double_cut_copper_slab"
		}
		return strings.TrimSuffix(bedrockIdent, "_slab") + "_double_slab"
	}
	return bedrockIdent
}

func applyJavaDerivedBedrockProperties(javaName, bedrockIdent string, props map[string]any) {
	switch javaName {
	case "lava_cauldron":
		props["cauldron_liquid"] = "lava"
		if _, ok := props["fill_level"]; !ok {
			props["fill_level"] = int32(6)
		}
	case "powder_snow_cauldron":
		props["cauldron_liquid"] = "powder_snow"
	case "tripwire":
		if _, ok := props["suspended_bit"]; !ok {
			props["suspended_bit"] = uint8(0)
		}
	}
	if strings.HasSuffix(bedrockIdent, "_hanging_sign") {
		if strings.HasSuffix(javaName, "_wall_hanging_sign") {
			props["attached_bit"] = uint8(1)
			props["hanging"] = uint8(1)
		} else if _, ok := props["attached_bit"]; !ok {
			props["attached_bit"] = uint8(0)
		}
		if _, ok := props["hanging"]; !ok {
			props["hanging"] = uint8(0)
		}
		if _, ok := props["facing_direction"]; !ok {
			props["facing_direction"] = int32(0)
		}
		if _, ok := props["ground_sign_direction"]; !ok {
			props["ground_sign_direction"] = int32(0)
		}
	}
}

func applyImplicitBedrockProperties(bedrockIdent string, props map[string]any) {
	if strings.HasSuffix(bedrockIdent, "_leaves") || bedrockIdent == "azalea_leaves_flowered" {
		if _, ok := props["update_bit"]; !ok {
			props["update_bit"] = uint8(0)
		}
	}
	switch bedrockIdent {
	case "bedrock":
		if _, ok := props["infiniburn_bit"]; !ok {
			props["infiniburn_bit"] = uint8(0)
		}
	case "smooth_quartz", "quartz_block", "chiseled_quartz_block":
		if _, ok := props["pillar_axis"]; !ok {
			props["pillar_axis"] = "y"
		}
	case "bone_block":
		if _, ok := props["deprecated"]; !ok {
			props["deprecated"] = int32(0)
		}
	case "torch":
		if _, ok := props["torch_facing_direction"]; !ok {
			props["torch_facing_direction"] = "top"
		}
	case "water":
		if _, ok := props["liquid_depth"]; !ok {
			props["liquid_depth"] = int32(0)
		}
	case "cauldron":
		if _, ok := props["cauldron_liquid"]; !ok {
			props["cauldron_liquid"] = "water"
		}
	}
}

func glowLichenFaces(javaProps map[string]string) int32 {
	var bits int32
	if javaProps["down"] == "true" {
		bits |= 1
	}
	if javaProps["up"] == "true" {
		bits |= 2
	}
	if javaProps["north"] == "true" {
		bits |= 4
	}
	if javaProps["south"] == "true" {
		bits |= 8
	}
	if javaProps["west"] == "true" {
		bits |= 16
	}
	if javaProps["east"] == "true" {
		bits |= 32
	}
	return bits
}

func isGlowLichenFace(prop string) bool {
	switch prop {
	case "down", "up", "north", "south", "west", "east":
		return true
	}
	return false
}

type bedrockPaletteIndex struct {
	states   map[string]BedrockState
	defaults map[string]BedrockState
}

// loadBedrockPalette decodes the embedded Bedrock palette and indexes states
// by minecraft: identifier plus state properties.
func loadBedrockPalette(data []byte) (*bedrockPaletteIndex, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gunzip palette: %w", err)
	}
	defer func() { _ = gz.Close() }()
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
	blocks, ok := root["blocks"].([]any)
	if !ok {
		return nil, fmt.Errorf("decode palette: blocks list missing")
	}
	idx := &bedrockPaletteIndex{
		states:   make(map[string]BedrockState, len(blocks)),
		defaults: make(map[string]BedrockState),
	}
	for _, raw := range blocks {
		m, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("decode palette: block entry has type %T", raw)
		}
		name, ok := m["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("decode palette: block entry has invalid name %T", m["name"])
		}
		props, ok := m["states"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("decode palette: %s states have type %T", name, m["states"])
		}
		state := BedrockState{Name: name, Properties: normaliseBedrockProperties(props)}
		idx.states[bedrockStateKey(state.Name, state.Properties)] = state
		if _, exists := idx.defaults[name]; !exists {
			idx.defaults[name] = state
		}
	}
	return idx, nil
}

func (p *bedrockPaletteIndex) lookup(name string, properties map[string]any) (BedrockState, bool) {
	if p == nil {
		return BedrockState{}, false
	}
	state, ok := p.states[bedrockStateKey(name, normaliseBedrockProperties(properties))]
	return state, ok
}

func (p *bedrockPaletteIndex) defaultState(name string) (BedrockState, bool) {
	if p == nil {
		return BedrockState{}, false
	}
	state, ok := p.defaults[name]
	return state, ok
}

func normaliseBedrockProperties(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		switch x := v.(type) {
		case uint8:
			out[k] = x
		case bool:
			if x {
				out[k] = uint8(1)
			} else {
				out[k] = uint8(0)
			}
		case int:
			out[k] = int32(x)
		case int8:
			out[k] = int32(x)
		case int16:
			out[k] = int32(x)
		case int32:
			out[k] = x
		case int64:
			out[k] = int32(x)
		case string:
			out[k] = x
		default:
			out[k] = x
		}
	}
	return out
}

func bedrockStateKey(name string, properties map[string]any) string {
	return name + "\x00" + hashBedrockProperties(properties)
}

func hashBedrockProperties(properties map[string]any) string {
	if len(properties) == 0 {
		return ""
	}
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		switch v := properties[k].(type) {
		case uint8:
			b.WriteString(strconv.Itoa(int(v)))
		case int32:
			b.WriteString(strconv.Itoa(int(v)))
		case string:
			b.WriteString(v)
		default:
			_, _ = fmt.Fprintf(&b, "%v", v)
		}
		b.WriteByte(';')
	}
	return b.String()
}
