package translate

import (
	"fmt"
	"maps"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/segmentio/fasthash/fnv1"
)

// StateBlock is an inert world.Block backed only by a Bedrock identifier and
// state properties. It is intended for schematic conversion, where visual
// state fidelity matters even when the server runtime has no safe behaviour
// implementation for the block.
type StateBlock struct {
	state BedrockState
	nbt   map[string]any
}

// NewStateBlock returns an inert world.Block for state.
func NewStateBlock(state BedrockState) StateBlock {
	return StateBlock{state: state.Clone()}
}

// NewStateBlockWithNBT returns an inert world.Block for state carrying
// block-entity NBT for Bedrock clients and chunk storage.
func NewStateBlockWithNBT(state BedrockState, data map[string]any) StateBlock {
	return StateBlock{state: state.Clone(), nbt: maps.Clone(data)}
}

// MergeBlockNBT returns block with additional block-entity data attached when
// it is an inert StateBlock. Existing inferred NBT, such as banner base color,
// is kept unless additional data contains the same key.
func MergeBlockNBT(block world.Block, data map[string]any) world.Block {
	if len(data) == 0 {
		return block
	}
	sb, ok := block.(StateBlock)
	if !ok {
		return block
	}
	merged := maps.Clone(sb.nbt)
	if merged == nil {
		merged = map[string]any{}
	}
	maps.Copy(merged, data)
	return StateBlock{state: sb.state.Clone(), nbt: merged}
}

// EncodeBlock returns the Bedrock identifier and state properties.
func (b StateBlock) EncodeBlock() (string, map[string]any) {
	return b.state.Name, maps.Clone(b.state.Properties)
}

// Hash returns a unique base identity for hot-path caches while keeping
// math.MaxUint64 as the state hash so Dragonfly resolves the runtime ID from
// EncodeBlock instead of expecting a registered concrete block hash.
func (b StateBlock) Hash() (uint64, uint64) {
	return fnv1.HashString64(stateBlockHashKey(b.state) + stateBlockNBTKey(b.nbt)), math.MaxUint64
}

// Model returns a full-cube inert model.
func (StateBlock) Model() world.BlockModel {
	return stateBlockModel{}
}

// EncodeNBT returns empty block-entity data for states whose Bedrock runtime
// ID is NBT-backed. StateBlock deliberately preserves visuals only; it does
// not emulate interactive block behaviour or inventories.
func (b StateBlock) EncodeNBT() map[string]any {
	if len(b.nbt) == 0 {
		return map[string]any{}
	}
	return maps.Clone(b.nbt)
}

// DecodeNBT keeps the inert state block when Dragonfly reloads NBT-backed
// runtime IDs from chunk storage.
func (b StateBlock) DecodeNBT(data map[string]any) any {
	b.nbt = maps.Clone(data)
	return b
}

type stateBlockModel struct{}

func (stateBlockModel) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 1, 1)}
}

func (stateBlockModel) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return true
}

var (
	_ world.Block = StateBlock{}
	_ world.NBTer = StateBlock{}
)

func stateBlockHashKey(state BedrockState) string {
	if len(state.Properties) == 0 {
		return state.Name
	}
	keys := make([]string, 0, len(state.Properties))
	for k := range state.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := state.Name
	for _, k := range keys {
		out += "\x00" + k + "="
		switch v := state.Properties[k].(type) {
		case uint8:
			out += strconv.Itoa(int(v))
		case int32:
			out += strconv.Itoa(int(v))
		case string:
			out += v
		default:
			out += fmt.Sprint(v)
		}
	}
	return out
}

func stateBlockNBTKey(data map[string]any) string {
	if len(data) == 0 {
		return ""
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out strings.Builder
	for _, k := range keys {
		out.WriteByte('\x00')
		out.WriteString(k)
		out.WriteByte('=')
		_, _ = fmt.Fprint(&out, stableNBTValue(data[k]))
	}
	return out.String()
}

func stableNBTValue(v any) any {
	rv := reflect.ValueOf(v)
	if rv.IsValid() && rv.Kind() == reflect.Slice {
		items := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			items[i] = stableNBTValue(rv.Index(i).Interface())
		}
		return items
	}
	if m, ok := v.(map[string]any); ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make([]any, 0, len(keys)*2)
		for _, k := range keys {
			out = append(out, k, stableNBTValue(m[k]))
		}
		return out
	}
	return v
}
