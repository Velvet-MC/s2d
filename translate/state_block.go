package translate

import (
	"fmt"
	"maps"
	"math"
	"sort"
	"strconv"

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
}

// NewStateBlock returns an inert world.Block for state.
func NewStateBlock(state BedrockState) StateBlock {
	return StateBlock{state: state.Clone()}
}

// EncodeBlock returns the Bedrock identifier and state properties.
func (b StateBlock) EncodeBlock() (string, map[string]any) {
	return b.state.Name, maps.Clone(b.state.Properties)
}

// Hash returns a unique base identity for hot-path caches while keeping
// math.MaxUint64 as the state hash so Dragonfly resolves the runtime ID from
// EncodeBlock instead of expecting a registered concrete block hash.
func (b StateBlock) Hash() (uint64, uint64) {
	return fnv1.HashString64(stateBlockHashKey(b.state)), math.MaxUint64
}

// Model returns a full-cube inert model.
func (StateBlock) Model() world.BlockModel {
	return stateBlockModel{}
}

type stateBlockModel struct{}

func (stateBlockModel) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 1, 1)}
}

func (stateBlockModel) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return true
}

var _ world.Block = StateBlock{}

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
