package schem

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// dummyModel satisfies world.BlockModel.
type dummyModel struct{}

func (dummyModel) BBox(cube.Pos, world.BlockSource) []cube.BBox { return nil }
func (dummyModel) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool { return false }

// dummyBlock satisfies world.Block.
type dummyBlock struct{}

func (dummyBlock) EncodeBlock() (string, map[string]any) { return "minecraft:stone", nil }
func (dummyBlock) Hash() (uint64, uint64)                { return 0, 0 }
func (dummyBlock) Model() world.BlockModel               { return dummyModel{} }

// compile-time assertion
var _ world.Block = dummyBlock{}

func fakeReadFunc(name Format) func(r io.Reader) (*Schematic, error) {
	return func(r io.Reader) (*Schematic, error) {
		_, _ = io.Copy(io.Discard, r)
		return &Schematic{Format: name, Width: 1, Height: 1, Length: 1,
			Blocks: []Block{{Pos: [3]int{0, 0, 0}, Block: dummyBlock{}}}}, nil
	}
}

func TestRead_DispatchesByExtension(t *testing.T) {
	prev := handlers
	handlers = nil
	defer func() { handlers = prev }()

	Register(FormatHandler{
		Name: "fake_a", Extensions: []string{".a"},
		Read: fakeReadFunc("fake_a"),
	})
	Register(FormatHandler{
		Name: "fake_b", Extensions: []string{".b"},
		Read: fakeReadFunc("fake_b"),
	})

	got, err := Read("hello.a", bytes.NewReader(nil))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got.Format != "fake_a" {
		t.Errorf("got %q want fake_a", got.Format)
	}

	got, err = Read("hello.b", bytes.NewReader(nil))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got.Format != "fake_b" {
		t.Errorf("got %q want fake_b", got.Format)
	}
}

func TestRead_UnknownExtension(t *testing.T) {
	prev := handlers
	handlers = nil
	defer func() { handlers = prev }()

	_, err := Read("file.unknown", bytes.NewReader(nil))
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported-format error, got %v", err)
	}
}
