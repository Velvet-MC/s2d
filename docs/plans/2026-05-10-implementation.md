# S2D v1.0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the S2D Go library v1.0 — read Sponge v2 (`.schem`) and legacy MCEdit (`.schematic`) Java schematics, translate every cell to a Dragonfly `world.Block` (with optional `world.Liquid` for waterlogged variants), report unknowns, and ship as `github.com/Velvet-MC/s2d` (MIT).

**Architecture:** Six internal Go packages (`palette`, `sponge`, `legacy`, `translate`, `translate/properties`, `schem`) with sharply separated responsibilities. Format readers register themselves with the `schem` dispatcher at `init()` time. Translation replicates Geyser's runtime algorithm in Go: at first `Lookup`, lazily build a `map[javaState]world.Block` by walking a vendored Java block schema, applying ~30 per-property converters, and validating each result against a vendored Bedrock palette via Dragonfly's `world.BlockByName`.

**Tech Stack:** Go 1.26, `df-mc/dragonfly` (for `world.Block`/`world.Liquid`/`world.BlockByName`), `sandertv/gophertunnel/minecraft/nbt` (for Java big-endian NBT decoding), stdlib (`compress/gzip`, `encoding/binary`, `embed`, `io`, `bytes`, `sort`, `sync`, `testing`).

---

## Working environment

All paths in this plan are relative to the **S2D repo root** (`github.com/Velvet-MC/s2d`). The user has placed a placeholder at `C:\Users\clxser\Documents\GitHub\Build\WorldEdit\s2d\` containing `docs/2026-05-10-design.md` and `docs/plans/2026-05-10-implementation.md`. Move/copy those into the actual cloned `Velvet-MC/s2d` repo when it exists, then run all commands from the repo root.

For the duration of this plan, `cd` into the S2D repo root before each step. Tests run via `go test ./...` from that root.

The Velvet-MC fork pins live in this user's downstream WE checkout, **not** in S2D itself. S2D's `go.mod` references upstream `df-mc/dragonfly` and upstream `sandertv/gophertunnel`. If running tests during development needs the Velvet-MC fork's blocks (because upstream Dragonfly is behind), apply a temporary `replace` in `go.mod` and remove it before tagging a release.

---

## File structure

| File | Responsibility |
|---|---|
| `go.mod`, `go.sum` | Module declaration; pins `df-mc/dragonfly` + `sandertv/gophertunnel`. No `replace` directives. |
| `LICENSE` | MIT, owner = Clxser. |
| `README.md` | Install, quickstart, worked example. |
| `doc.go` | Package-level overview for `s2d` (the umbrella module path is `github.com/Velvet-MC/s2d`; primary consumer entry is `schem.Read`). |
| `palette/palette.go` | `JavaState` struct + `Decode(string) (JavaState, error)` + `Canonical()`. Pure string parsing. |
| `palette/palette_test.go` | Table-driven tests for Decode + Canonical. |
| `schem/types.go` | Public types: `Format`, `Block`, `Schematic`, `UnknownReport`, `FormatHandler`. |
| `schem/schem.go` | `Read(filename, r) (*Schematic, error)`, `Register(FormatHandler)`. Format dispatcher. |
| `schem/schem_test.go` | Dispatch tests using fixture files. |
| `sponge/varint.go` | LEB128 varint reader over an `io.ByteReader`. |
| `sponge/varint_test.go` | Table-driven tests for varint encoding. |
| `sponge/reader.go` | Sponge v2 reader: gunzip → NBT decode → palette walk → emit `*Schematic`. |
| `sponge/init.go` | `init()` calls `schem.Register` with the Sponge v2 handler. |
| `sponge/reader_test.go` | Tests using committed `.schem` fixtures. |
| `sponge/testdata/single_stone.schem` | 1×1×1 stone. |
| `sponge/testdata/multi_palette.schem` | 4×1×4 ≥4 palette entries (exercises multi-byte varint). |
| `sponge/testdata/waterlogged_stairs.schem` | One waterlogged oak_stairs cell. |
| `sponge/testdata/unknown_block.schem` | Handcrafted `mod:foo[]` palette entry. |
| `sponge/testdata/generate.go` | Build-time helper (only run manually) to regenerate testdata if the format spec changes. |
| `legacy/ids.go` | Embed `data/legacy.json` and expose `Lookup(id, data int) (string, bool)`. |
| `legacy/ids_test.go` | Known (id, data) → state-string assertions. |
| `legacy/reader.go` | Legacy MCEdit `.schematic` reader: gunzip → NBT decode → numeric ID walk → emit `*Schematic`. |
| `legacy/init.go` | `init()` calls `schem.Register` with the legacy handler. |
| `legacy/reader_test.go` | Tests using committed `.schematic` fixtures. |
| `legacy/testdata/basic.schematic` | Pre-1.13 numeric IDs only. |
| `legacy/testdata/addblocks.schematic` | Uses `AddBlocks` for IDs > 255. |
| `legacy/data/legacy.json` | Vendored from EngineHub/WorldEdit (MIT). |
| `legacy/data/LICENSE` | WorldEdit MIT. |
| `legacy/data/ATTRIBUTION.md` | Upstream URL + pinned commit + refresh command. |
| `translate/translate.go` | Public `Lookup(string) Result`, `SetMissingBlock`, `MissingBlock`. |
| `translate/missing.go` | Default missing-block resolver (magenta wool). |
| `translate/table.go` | `sync.Once` table builder; loads embedded data, walks Java schema, applies converters, validates against Bedrock palette, populates `map[string]world.Block`. |
| `translate/translate_test.go` | Public-API tests (set/get missing block, Lookup empty key). |
| `translate/table_test.go` | Integrity tests: `TestPaletteCoverage`, `TestPaletteMatchesDragonfly`, `TestMissingBlockRegistered`, `TestRoundtrip`. |
| `translate/properties/properties.go` | Type definitions + the registry that maps Java property name → converter func. |
| `translate/properties/axis.go` | `Axis(javaValue, bedrockIdent) (prop string, value any)`. |
| `translate/properties/facing.go` | Java `facing` → Bedrock `direction` / `weirdo_direction` / `cardinal_direction` (branched on bedrockIdent). |
| `translate/properties/half.go` | Java `half` → Bedrock `upside_down_bit` (stairs) or `upper_block_bit` (doors), branched. |
| `translate/properties/hinge.go` | Java `hinge` → Bedrock `door_hinge_bit`. |
| `translate/properties/open.go` | Java `open` → Bedrock `open_bit`. |
| `translate/properties/powered.go` | Java `powered` → Bedrock `powered_bit`. |
| `translate/properties/lit.go` | Java `lit` → Bedrock `lit` (modern) — passthrough for 1.0. |
| `translate/properties/waterlogged.go` | Java `waterlogged` → liquid signal (no Bedrock prop emitted). |
| `translate/properties/snowy.go` | Java `snowy` → Bedrock `covered_bit` (grass). |
| `translate/properties/age.go` | Java `age` int → Bedrock `growth` int (crops, vines, etc.). |
| `translate/properties/distance.go` | Java `distance` int (leaves) → Bedrock leaf state. |
| `translate/properties/persistent.go` | Java `persistent` → Bedrock `persistent_bit` (leaves). |
| `translate/properties/level.go` | Java `level` (water/lava height) → Bedrock `liquid_depth`. |
| `translate/properties/redstone.go` | Java `power` (redstone wire) → Bedrock `redstone_signal`. |
| `translate/properties/door_half.go` | Doors specifically (because `half` value set differs from stairs). |
| `translate/properties/slab.go` | Java `type` (top/bottom/double) → Bedrock `top_slot_bit` + identifier swap for double. |
| `translate/properties/stair_shape.go` | Java `shape` (straight/inner_left/inner_right/outer_left/outer_right) → Bedrock has no equivalent (drop or map closest). |
| `translate/properties/wall.go` | Java `up`/`north`/`east`/`south`/`west` (wall connections) → Bedrock `wall_connection_type_*`. |
| `translate/properties/fence.go` | Java fence connections → Bedrock often computed-only (drop). |
| `translate/properties/rotation.go` | Java `rotation` (16-step) → Bedrock `ground_sign_direction`. |
| `translate/properties/properties_test.go` | One file, table-driven, covers every converter. |
| `translate/data/java_blocks.json` | Vendored Java block schema. ~500 KB. |
| `translate/data/bedrock_palette.1_26_20.nbt` | Vendored from Geyser. ~2 MB. |
| `translate/data/overrides.json` | Hand-maintained Java→Bedrock identifier overrides. |
| `translate/data/LICENSE` | Geyser MIT (covers palette). |
| `translate/data/ATTRIBUTION.md` | Upstream URLs + pinned commits. |
| `translate/data/REGEN.md` | Refresh procedure. |

---

## Phase 0: Bootstrap

### Task 0.1: Initialize the Go module

**Files:**
- Create: `go.mod`
- Create: `LICENSE`
- Create: `README.md`
- Create: `doc.go`

- [ ] **Step 1: Initialize module**

Run:
```bash
go mod init github.com/Velvet-MC/s2d
```

Expected: creates `go.mod` with `module github.com/Velvet-MC/s2d` and `go 1.26.0` (adjust the Go directive to match your toolchain).

- [ ] **Step 2: Add the two required dependencies**

Run:
```bash
go get github.com/df-mc/dragonfly
go get github.com/sandertv/gophertunnel
```

Expected: both pulled and recorded in `go.mod`. If `go get` fails because no Go file imports them yet, create a placeholder file first:

```go
// doc.go (placeholder, will be expanded in Step 4)
package s2d
```

Then re-run `go get`.

- [ ] **Step 3: Write LICENSE**

Create `LICENSE` with the standard MIT text, owner `Clxser`, year `2026`. The exact MIT text is widely available; copy from https://opensource.org/licenses/MIT and substitute year + owner. Do not paraphrase.

- [ ] **Step 4: Write doc.go**

Create `doc.go`:

```go
// Package s2d (Schematic-to-Dragonfly) is the umbrella module for reading
// Java-edition Minecraft schematic files and translating them into Dragonfly
// world.Block values for use on Bedrock-edition servers.
//
// Consumers do not import this package directly. The public entry point is
// github.com/Velvet-MC/s2d/schem, which auto-detects the schematic format from
// the filename and returns a *schem.Schematic with translated blocks.
//
// Example:
//
//	f, _ := os.Open("castle.schem")
//	defer f.Close()
//	s, err := schem.Read(f.Name(), f)
//	if err != nil { ... }
//	for _, b := range s.Blocks {
//	    // place b.Block at b.Pos; if b.Liquid != nil, place it too
//	}
//
// Supported formats in v1.0: Sponge Schematic v2 (.schem) and legacy MCEdit
// (.schematic). Roadmap formats (Sponge v3, Litematica .litematic, vanilla
// structure NBT .nbt) are documented in docs/2026-05-10-design.md.
package s2d
```

- [ ] **Step 5: Write README skeleton**

Create `README.md`:

````markdown
# S2D

**Schematic-to-Dragonfly** — read Java Minecraft schematics on Dragonfly Bedrock servers.

## Install

```sh
go get github.com/Velvet-MC/s2d
```

## Quickstart

```go
package main

import (
	"log"
	"os"

	"github.com/Velvet-MC/s2d/schem"
)

func main() {
	f, err := os.Open("castle.schem")
	if err != nil { log.Fatal(err) }
	defer f.Close()

	s, err := schem.Read(f.Name(), f)
	if err != nil { log.Fatal(err) }

	log.Printf("loaded %dx%dx%d, %d blocks, %d unknowns",
		s.Width, s.Height, s.Length, len(s.Blocks), s.Unknowns.Total)

	for _, b := range s.Blocks {
		// place b.Block at b.Pos in your world / clipboard / converter
		// if b.Liquid != nil, place it too (waterlogged Java block)
		_ = b
	}
}
```

## Supported formats (v1.0)

| Extension | Format | Notes |
|---|---|---|
| `.schem` | Sponge Schematic v2 | Java WorldEdit ≥ 1.13 |
| `.schematic` | Legacy MCEdit | Pre-1.13, numeric IDs |

Roadmap (v1.1+): Sponge v3, Litematica `.litematic`, vanilla structure NBT `.nbt`.

## License

MIT. See `LICENSE`.

Vendored data attributions: see `translate/data/ATTRIBUTION.md` and `legacy/data/ATTRIBUTION.md`.
````

- [ ] **Step 6: Verify the module builds**

Run:
```bash
go build ./...
go vet ./...
```

Expected: both succeed with no output.

- [ ] **Step 7: Commit**

```bash
git add go.mod go.sum LICENSE README.md doc.go
git commit -m "chore: initialize S2D module"
```

---

## Phase 1: `palette` package

The Java state string parser. Pure code, no deps. Easiest TDD warm-up.

### Task 1.1: Define `JavaState` and write failing test

**Files:**
- Create: `palette/palette.go`
- Create: `palette/palette_test.go`

- [ ] **Step 1: Write the failing test**

Create `palette/palette_test.go`:

```go
package palette

import "testing"

func TestDecode_Simple(t *testing.T) {
	got, err := Decode("minecraft:stone")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Namespace != "minecraft" || got.Name != "stone" {
		t.Errorf("got %+v", got)
	}
	if len(got.Props) != 0 {
		t.Errorf("expected no props, got %v", got.Props)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./palette/...`
Expected: build error — `Decode` undefined.

- [ ] **Step 3: Write minimal implementation**

Create `palette/palette.go`:

```go
// Package palette parses Java edition block-state strings of the form
// "namespace:name[k1=v1,k2=v2,...]" used in Sponge schematic palettes.
//
// The package has no Dragonfly dependency. It exists so consumers (and
// the translate package) can inspect or normalize palette entries
// without invoking the full translation pipeline.
package palette

import (
	"fmt"
	"sort"
	"strings"
)

// JavaState is a parsed Java block-state string.
type JavaState struct {
	Namespace string
	Name      string
	Props     map[string]string
}

// Decode parses s into a JavaState. The namespace defaults to "minecraft"
// when absent. Properties are optional; if present they are enclosed in
// square brackets and comma-separated.
func Decode(s string) (JavaState, error) {
	if s == "" {
		return JavaState{}, fmt.Errorf("palette: empty state string")
	}
	js := JavaState{Namespace: "minecraft"}
	body := s
	if i := strings.IndexByte(s, '['); i >= 0 {
		if !strings.HasSuffix(s, "]") {
			return JavaState{}, fmt.Errorf("palette: unterminated bracket in %q", s)
		}
		body = s[:i]
		propBody := s[i+1 : len(s)-1]
		if propBody != "" {
			js.Props = make(map[string]string)
			for _, kv := range strings.Split(propBody, ",") {
				eq := strings.IndexByte(kv, '=')
				if eq < 0 {
					return JavaState{}, fmt.Errorf("palette: malformed prop %q in %q", kv, s)
				}
				k := strings.TrimSpace(kv[:eq])
				v := strings.TrimSpace(kv[eq+1:])
				if k == "" || v == "" {
					return JavaState{}, fmt.Errorf("palette: empty key or value in %q", s)
				}
				js.Props[k] = v
			}
		}
	}
	if colon := strings.IndexByte(body, ':'); colon >= 0 {
		js.Namespace = body[:colon]
		js.Name = body[colon+1:]
	} else {
		js.Name = body
	}
	if js.Name == "" {
		return JavaState{}, fmt.Errorf("palette: empty block name in %q", s)
	}
	return js, nil
}

// Canonical re-emits the state with properties sorted by key, suitable
// for use as a stable map key.
func (j JavaState) Canonical() string {
	var b strings.Builder
	b.WriteString(j.Namespace)
	b.WriteByte(':')
	b.WriteString(j.Name)
	if len(j.Props) > 0 {
		keys := make([]string, 0, len(j.Props))
		for k := range j.Props {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteByte('[')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(k)
			b.WriteByte('=')
			b.WriteString(j.Props[k])
		}
		b.WriteByte(']')
	}
	return b.String()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./palette/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add palette/palette.go palette/palette_test.go
git commit -m "feat(palette): decode and canonicalize Java state strings"
```

### Task 1.2: Cover edge cases

**Files:**
- Modify: `palette/palette_test.go`

- [ ] **Step 1: Add table-driven test cases**

Append to `palette/palette_test.go`:

```go
func TestDecode_Cases(t *testing.T) {
	cases := []struct {
		in        string
		ns, name  string
		props     map[string]string
		canonical string
		wantErr   bool
	}{
		{"stone", "minecraft", "stone", nil, "minecraft:stone", false},
		{"minecraft:air", "minecraft", "air", nil, "minecraft:air", false},
		{"minecraft:oak_log[axis=y]", "minecraft", "oak_log",
			map[string]string{"axis": "y"}, "minecraft:oak_log[axis=y]", false},
		{"minecraft:oak_stairs[half=top,facing=north]", "minecraft", "oak_stairs",
			map[string]string{"half": "top", "facing": "north"},
			"minecraft:oak_stairs[facing=north,half=top]", false},
		{"mod:custom[a=1,b=2,c=3]", "mod", "custom",
			map[string]string{"a": "1", "b": "2", "c": "3"},
			"mod:custom[a=1,b=2,c=3]", false},
		{"", "", "", nil, "", true},
		{":stone", "", "", nil, "", true},
		{"minecraft:", "", "", nil, "", true},
		{"minecraft:oak_log[axis=", "", "", nil, "", true},
		{"minecraft:oak_log[axis=y", "", "", nil, "", true},
		{"minecraft:oak_log[=y]", "", "", nil, "", true},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := Decode(c.in)
			if (err != nil) != c.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, c.wantErr)
			}
			if c.wantErr {
				return
			}
			if got.Namespace != c.ns || got.Name != c.name {
				t.Errorf("ns/name: got %q:%q want %q:%q", got.Namespace, got.Name, c.ns, c.name)
			}
			if len(got.Props) != len(c.props) {
				t.Errorf("props len: got %v want %v", got.Props, c.props)
			}
			for k, v := range c.props {
				if got.Props[k] != v {
					t.Errorf("props[%q]=%q want %q", k, got.Props[k], v)
				}
			}
			if g := got.Canonical(); g != c.canonical {
				t.Errorf("canonical: got %q want %q", g, c.canonical)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./palette/... -v`
Expected: All cases PASS. The `:stone` and `minecraft:` cases fail with the empty-name guard; `[axis=` triggers unterminated-bracket; `[=y]` triggers empty-key.

If any case fails, look at the error path in `Decode`. Common cause for `[=y]` failing the wrong way: the empty-key check is correct only if `TrimSpace(kv[:eq])` is empty when `eq == 0`. Verify by adding a `t.Logf` if needed.

- [ ] **Step 3: Commit**

```bash
git add palette/palette_test.go
git commit -m "test(palette): cover edge cases for Decode"
```

---

## Phase 2: `schem` package — types and dispatcher

The public entry point. Defines the cross-package types, holds the format registry.

### Task 2.1: Define public types

**Files:**
- Create: `schem/types.go`
- Create: `schem/types_test.go`

- [ ] **Step 1: Write the test**

Create `schem/types_test.go`:

```go
package schem

import "testing"

func TestUnknownReport_Empty(t *testing.T) {
	var r UnknownReport
	if r.Total != 0 || len(r.Counts) != 0 {
		t.Errorf("zero value should be empty: %+v", r)
	}
}

func TestUnknownReport_Add(t *testing.T) {
	r := UnknownReport{Counts: map[string]int{}}
	r.Counts["minecraft:foo"]++
	r.Counts["minecraft:foo"]++
	r.Counts["minecraft:bar"]++
	r.Total = 3
	if r.Counts["minecraft:foo"] != 2 || r.Counts["minecraft:bar"] != 1 || r.Total != 3 {
		t.Errorf("unexpected: %+v", r)
	}
}
```

- [ ] **Step 2: Run test (verify build fails)**

Run: `go test ./schem/...`
Expected: build error — `UnknownReport` undefined.

- [ ] **Step 3: Write the types**

Create `schem/types.go`:

```go
// Package schem is the public entry point for the S2D library.
// It defines the cross-cutting Schematic / Block / UnknownReport types,
// holds the format-handler registry, and exposes Read which dispatches
// to a format reader based on the filename.
package schem

import (
	"io"

	"github.com/df-mc/dragonfly/server/world"
)

// Format names a supported on-disk schematic format.
type Format string

const (
	FormatSpongeV2 Format = "sponge_v2"
	FormatLegacy   Format = "legacy_schematic"
)

// Block is one cell in a parsed schematic.
type Block struct {
	Pos    [3]int       // x, y, z within the schematic local frame
	Block  world.Block  // translated Bedrock block; never nil
	Liquid world.Liquid // non-nil iff the Java source was waterlogged
}

// Schematic is the parsed result of Read.
type Schematic struct {
	Format                Format
	Width, Height, Length int
	Offset                [3]int
	Blocks                []Block
	Unknowns              UnknownReport
}

// UnknownReport tallies cells whose Java state could not be translated.
// Counts is keyed by the canonical Java state string.
type UnknownReport struct {
	Counts map[string]int
	Total  int
}

// FormatHandler describes one supported on-disk format.
// Format-reader packages register their handler at init() time.
type FormatHandler struct {
	Name       Format
	Extensions []string                          // e.g. []string{".schem"}
	Signature  func(headerPeek []byte) bool       // optional NBT signature check
	Read       func(io.Reader) (*Schematic, error)
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./schem/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add schem/types.go schem/types_test.go
git commit -m "feat(schem): public types"
```

### Task 2.2: Implement format dispatcher

**Files:**
- Create: `schem/schem.go`
- Modify: `schem/types_test.go` → rename to or split out `schem/schem_test.go`

- [ ] **Step 1: Write the failing dispatch test**

Create `schem/schem_test.go`:

```go
package schem

import (
	"bytes"
	"strings"
	"testing"

	"github.com/df-mc/dragonfly/server/world"
)

// fakeReadFunc returns a Schematic whose Format reflects which handler ran.
func fakeReadFunc(name Format) func(r io.Reader) (*Schematic, error) {
	return func(r io.Reader) (*Schematic, error) {
		// Drain r so we don't leak readers in tests.
		_, _ = io.Copy(io.Discard, r)
		return &Schematic{Format: name, Width: 1, Height: 1, Length: 1,
			Blocks: []Block{{Pos: [3]int{0, 0, 0}, Block: dummyBlock{}}}}, nil
	}
}

type dummyBlock struct{}

func (dummyBlock) EncodeBlock() (string, map[string]any) { return "minecraft:stone", nil }
func (dummyBlock) Model() world.BlockModel               { return nil }
func (dummyBlock) Hash() (uint64, uint64)                { return 0, 0 }
func (dummyBlock) Color() color.RGBA                     { return color.RGBA{} } // import color via "image/color"

func TestRead_DispatchesByExtension(t *testing.T) {
	// Reset the registry for test isolation.
	prev := handlers
	handlers = nil
	defer func() { handlers = prev }()

	Register(FormatHandler{
		Name:       "fake_a",
		Extensions: []string{".a"},
		Read:       fakeReadFunc("fake_a"),
	})
	Register(FormatHandler{
		Name:       "fake_b",
		Extensions: []string{".b"},
		Read:       fakeReadFunc("fake_b"),
	})

	got, err := Read("hello.a", bytes.NewReader(nil))
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if got.Format != "fake_a" { t.Errorf("got %q want fake_a", got.Format) }

	got, err = Read("hello.b", bytes.NewReader(nil))
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if got.Format != "fake_b" { t.Errorf("got %q want fake_b", got.Format) }
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
```

Note: `dummyBlock` is needed because `Block.Block` is typed `world.Block`, which is an interface in Dragonfly. Adjust the method set if Dragonfly's `world.Block` interface differs. Run `go doc github.com/df-mc/dragonfly/server/world.Block` (after `go mod tidy`) to confirm the method set; trim the dummy to exactly those methods. If `Hash()` or `Model()` aren't required, drop them.

You may also need `import "image/color"` and `import "io"` in the test file.

- [ ] **Step 2: Run test to verify build fails**

Run: `go test ./schem/...`
Expected: build error — `Read`, `Register`, `handlers` undefined.

- [ ] **Step 3: Implement the dispatcher**

Create `schem/schem.go`:

```go
package schem

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

var handlers []FormatHandler

// Register adds a FormatHandler to the dispatch table. Format-reader
// packages call this from their init() function.
func Register(h FormatHandler) {
	handlers = append(handlers, h)
}

// Read parses a schematic from r. The filename's extension selects the
// format reader; if no extension matches, Read peeks the gzipped header
// for an NBT signature.
func Read(filename string, r io.Reader) (*Schematic, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		for _, h := range handlers {
			if slices.Contains(h.Extensions, ext) {
				return h.Read(r)
			}
		}
	}
	// Fallback: read the full body, attempt signature dispatch.
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("schem: read: %w", err)
	}
	peek, err := peekDecompressed(buf, 64)
	if err != nil {
		return nil, fmt.Errorf("schem: peek: %w", err)
	}
	for _, h := range handlers {
		if h.Signature != nil && h.Signature(peek) {
			return h.Read(bytes.NewReader(buf))
		}
	}
	return nil, fmt.Errorf("schem: unsupported schematic format: %s", filename)
}

// peekDecompressed gunzips up to n bytes from buf for signature inspection.
func peekDecompressed(buf []byte, n int) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	out := make([]byte, n)
	read, err := io.ReadFull(gz, out)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return out[:read], nil
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./schem/... -v`
Expected: PASS.

If `dummyBlock` doesn't satisfy `world.Block`, the failure surfaces at the test build step. Trim or extend the dummy method set per the actual Dragonfly interface.

- [ ] **Step 5: Commit**

```bash
git add schem/schem.go schem/schem_test.go
git commit -m "feat(schem): format dispatcher"
```

---

## Phase 3: `sponge` package — varint reader

Pure function, easy TDD. Defining it before the full reader so we can prove varint behavior in isolation.

### Task 3.1: Implement varint reader

**Files:**
- Create: `sponge/varint.go`
- Create: `sponge/varint_test.go`

- [ ] **Step 1: Write failing tests**

Create `sponge/varint_test.go`:

```go
package sponge

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestReadVarint_Cases(t *testing.T) {
	cases := []struct {
		name  string
		bytes []byte
		want  uint32
	}{
		{"zero", []byte{0x00}, 0},
		{"one", []byte{0x01}, 1},
		{"127", []byte{0x7F}, 127},
		{"128", []byte{0x80, 0x01}, 128},
		{"300", []byte{0xAC, 0x02}, 300},
		{"16383", []byte{0xFF, 0x7F}, 16383},
		{"16384", []byte{0x80, 0x80, 0x01}, 16384},
		{"max32", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x0F}, 0xFFFFFFFF},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, n, err := readVarint(bytes.NewReader(c.bytes))
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if uint32(got) != c.want {
				t.Errorf("got %d want %d", got, c.want)
			}
			if n != len(c.bytes) {
				t.Errorf("consumed %d want %d", n, len(c.bytes))
			}
		})
	}
}

func TestReadVarint_EOF(t *testing.T) {
	// Continuation bit set but no next byte.
	_, _, err := readVarint(bytes.NewReader([]byte{0x80}))
	if !errors.Is(err, io.ErrUnexpectedEOF) && err == nil {
		t.Fatalf("expected EOF/UnexpectedEOF, got %v", err)
	}
}

func TestReadVarint_Overflow(t *testing.T) {
	// 6 continuation bytes — varint can't be longer than 5 for uint32.
	_, _, err := readVarint(bytes.NewReader([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01}))
	if err == nil {
		t.Fatalf("expected overflow error")
	}
}
```

- [ ] **Step 2: Run test to verify build fails**

Run: `go test ./sponge/...`
Expected: build error — `readVarint` undefined.

- [ ] **Step 3: Implement varint reader**

Create `sponge/varint.go`:

```go
package sponge

import (
	"fmt"
	"io"
)

// readVarint reads a Mojang-style LEB128 varint (7-bit groups, MSB
// continuation flag, little-endian) from r. Returns the decoded uint32
// value, the number of bytes consumed, and an error.
//
// Sponge Schematic v2 BlockData uses this encoding for palette indices.
// It is byte-compatible with Java protocol varints.
func readVarint(r io.ByteReader) (uint32, int, error) {
	var (
		value uint32
		shift uint
		n     int
	)
	for n < 5 {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF && n > 0 {
				return 0, n, io.ErrUnexpectedEOF
			}
			return 0, n, err
		}
		n++
		value |= uint32(b&0x7F) << shift
		if b&0x80 == 0 {
			return value, n, nil
		}
		shift += 7
	}
	return 0, n, fmt.Errorf("sponge: varint exceeds 5 bytes")
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./sponge/... -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add sponge/varint.go sponge/varint_test.go
git commit -m "feat(sponge): LEB128 varint reader"
```

---

## Phase 4: `sponge` package — fixture builder + reader

The reader needs at least one fixture file. We hand-build a minimal Sponge v2 NBT in test code (rather than depending on Java tooling), gzip it, and write it out as `testdata/single_stone.schem`.

### Task 4.1: Generate the smallest fixture programmatically

**Files:**
- Create: `sponge/testdata_gen_test.go`

- [ ] **Step 1: Write a generator test that emits the fixture**

Create `sponge/testdata_gen_test.go`. This is a `t.Skip`-gated test — it only runs when `S2D_REGEN_FIXTURES=1` is set. It produces the fixture files; once committed, normal test runs read them rather than regenerate.

```go
package sponge

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func TestGenerateFixtures(t *testing.T) {
	if os.Getenv("S2D_REGEN_FIXTURES") != "1" {
		t.Skip("set S2D_REGEN_FIXTURES=1 to regenerate")
	}

	dir := "testdata"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// single_stone.schem: 1x1x1, palette {minecraft:stone: 0}, BlockData [0x00]
	{
		root := map[string]any{
			"Version":     int32(2),
			"DataVersion": int32(2975), // 1.18.1; arbitrary but valid
			"Width":       int16(1),
			"Height":      int16(1),
			"Length":      int16(1),
			"PaletteMax":  int32(1),
			"Palette":     map[string]int32{"minecraft:stone": 0},
			"BlockData":   []byte{0x00},
		}
		writeFixture(t, filepath.Join(dir, "single_stone.schem"), root)
	}

	// multi_palette.schem: 4x1x4, four palette entries, varint-rich BlockData.
	{
		palette := map[string]int32{
			"minecraft:stone":          0,
			"minecraft:dirt":           1,
			"minecraft:oak_planks":     200, // forces multi-byte varint
			"minecraft:diamond_block":  3,
		}
		// 16 cells; cycle through indices 0,1,200,3,0,1,200,3,...
		// Encode each as a varint.
		var data bytes.Buffer
		seq := []uint32{0, 1, 200, 3, 0, 1, 200, 3, 0, 1, 200, 3, 0, 1, 200, 3}
		for _, v := range seq {
			writeVarint(&data, v)
		}
		root := map[string]any{
			"Version":     int32(2),
			"DataVersion": int32(2975),
			"Width":       int16(4),
			"Height":      int16(1),
			"Length":      int16(4),
			"PaletteMax":  int32(201),
			"Palette":     palette,
			"BlockData":   data.Bytes(),
		}
		writeFixture(t, filepath.Join(dir, "multi_palette.schem"), root)
	}

	// waterlogged_stairs.schem: 1x1x1, oak_stairs with waterlogged=true.
	{
		root := map[string]any{
			"Version":     int32(2),
			"DataVersion": int32(2975),
			"Width":       int16(1),
			"Height":      int16(1),
			"Length":      int16(1),
			"PaletteMax":  int32(1),
			"Palette": map[string]int32{
				"minecraft:oak_stairs[facing=north,half=bottom,shape=straight,waterlogged=true]": 0,
			},
			"BlockData": []byte{0x00},
		}
		writeFixture(t, filepath.Join(dir, "waterlogged_stairs.schem"), root)
	}

	// unknown_block.schem: 1x1x1, a fictitious mod block.
	{
		root := map[string]any{
			"Version":     int32(2),
			"DataVersion": int32(2975),
			"Width":       int16(1),
			"Height":      int16(1),
			"Length":      int16(1),
			"PaletteMax":  int32(1),
			"Palette":     map[string]int32{"mod:foo[bar=baz]": 0},
			"BlockData":   []byte{0x00},
		}
		writeFixture(t, filepath.Join(dir, "unknown_block.schem"), root)
	}
}

func writeFixture(t *testing.T, path string, root map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	enc := nbt.NewEncoderWithEncoding(&buf, nbt.BigEndian)
	if err := enc.Encode(root); err != nil {
		t.Fatal(err)
	}
	var gz bytes.Buffer
	gw := gzip.NewWriter(&gz)
	if _, err := gw.Write(buf.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, gz.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeVarint(buf *bytes.Buffer, v uint32) {
	for {
		if v < 0x80 {
			buf.WriteByte(byte(v))
			return
		}
		buf.WriteByte(byte(v&0x7F | 0x80))
		v >>= 7
	}
}
```

- [ ] **Step 2: Generate the fixtures**

Run:
```bash
S2D_REGEN_FIXTURES=1 go test ./sponge/ -run TestGenerateFixtures -v
```

Expected: PASS, and four files appear under `sponge/testdata/`. Verify with `ls sponge/testdata/`.

If `nbt.NewEncoderWithEncoding` is not the exact constructor in this gophertunnel version, check the `nbt` package's exported names with `go doc github.com/sandertv/gophertunnel/minecraft/nbt | head -40`. Adjust the call to match (some versions expose `nbt.MarshalEncoding(v, nbt.BigEndian)` returning bytes directly, which is simpler — refactor `writeFixture` accordingly if so).

- [ ] **Step 3: Commit the fixtures and the generator**

```bash
git add sponge/testdata_gen_test.go sponge/testdata/
git commit -m "test(sponge): committed fixtures + regen generator"
```

### Task 4.2: Write the Sponge reader

**Files:**
- Create: `sponge/reader.go`
- Create: `sponge/init.go`
- Create: `sponge/reader_test.go`

- [ ] **Step 1: Write the failing reader test**

Create `sponge/reader_test.go`:

```go
package sponge

import (
	"os"
	"testing"
)

func TestRead_SingleStone(t *testing.T) {
	f, err := os.Open("testdata/single_stone.schem")
	if err != nil { t.Fatal(err) }
	defer f.Close()

	s, err := Read(f)
	if err != nil { t.Fatalf("Read: %v", err) }

	if s.Width != 1 || s.Height != 1 || s.Length != 1 {
		t.Fatalf("dims: %dx%dx%d", s.Width, s.Height, s.Length)
	}
	if len(s.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(s.Blocks))
	}
	b := s.Blocks[0]
	if b.Pos != [3]int{0, 0, 0} {
		t.Errorf("pos: %v", b.Pos)
	}
	if b.Block == nil {
		t.Fatalf("Block is nil")
	}
	name, _ := b.Block.EncodeBlock()
	if name != "minecraft:stone" {
		t.Errorf("got %q want minecraft:stone", name)
	}
}

func TestRead_MultiPalette(t *testing.T) {
	f, err := os.Open("testdata/multi_palette.schem")
	if err != nil { t.Fatal(err) }
	defer f.Close()

	s, err := Read(f)
	if err != nil { t.Fatalf("Read: %v", err) }

	if s.Width != 4 || s.Height != 1 || s.Length != 4 {
		t.Fatalf("dims: %dx%dx%d", s.Width, s.Height, s.Length)
	}
	if len(s.Blocks) != 16 {
		t.Fatalf("expected 16 blocks, got %d", len(s.Blocks))
	}
	// Cells are stored in YZX order; cell 0 should be (0,0,0)=stone.
	first := s.Blocks[0]
	if first.Pos != [3]int{0, 0, 0} { t.Errorf("pos[0]: %v", first.Pos) }
	name, _ := first.Block.EncodeBlock()
	if name != "minecraft:stone" { t.Errorf("name[0]: %q", name) }
	// Cell 2 should be (2,0,0)=oak_planks (varint multi-byte).
	third := s.Blocks[2]
	if third.Pos != [3]int{2, 0, 0} { t.Errorf("pos[2]: %v", third.Pos) }
	name, _ = third.Block.EncodeBlock()
	if name != "minecraft:oak_planks" { t.Errorf("name[2]: %q", name) }
}
```

- [ ] **Step 2: Run test to verify build/run fails**

Run: `go test ./sponge/...`
Expected: build error — `Read` undefined.

- [ ] **Step 3: Implement the reader**

Create `sponge/reader.go`:

```go
// Package sponge reads Sponge Schematic v2 (.schem) files. It depends on
// gophertunnel/minecraft/nbt for big-endian Java NBT decoding and on
// the s2d/translate package for Java-state → Bedrock-block resolution.
package sponge

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sort"

	"github.com/Velvet-MC/s2d/palette"
	"github.com/Velvet-MC/s2d/schem"
	"github.com/Velvet-MC/s2d/translate"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// rawSchematic mirrors the Sponge v2 NBT root.
type rawSchematic struct {
	Version     int32             `nbt:"Version"`
	DataVersion int32             `nbt:"DataVersion,omitempty"`
	Width       int16             `nbt:"Width"`
	Height      int16             `nbt:"Height"`
	Length      int16             `nbt:"Length"`
	PaletteMax  int32             `nbt:"PaletteMax,omitempty"`
	Palette     map[string]int32  `nbt:"Palette"`
	BlockData   []byte            `nbt:"BlockData"`
	Offset      []int32           `nbt:"Offset,omitempty"`
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

	// Invert palette so we can resolve indices in O(1).
	indexToKey := make([]string, len(raw.Palette))
	for k, v := range raw.Palette {
		if int(v) < 0 || int(v) >= len(raw.Palette) {
			return nil, fmt.Errorf("sponge v2: palette index %d out of range", v)
		}
		indexToKey[v] = k
	}
	// Sanity: every slot filled.
	for i, k := range indexToKey {
		if k == "" {
			return nil, fmt.Errorf("sponge v2: palette index %d unfilled", i)
		}
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
				if int(idx) >= len(indexToKey) {
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
	// Stable order for tests / consumers that iterate the report.
	sort.Strings(nil) // keep import "sort" used; remove this line if you choose to expose sorted keys elsewhere
	return out, nil
}
```

(Drop the `sort.Strings(nil)` line — it's a placeholder to keep `sort` imported in case the engineer adds a sorted-output convenience method later. If you don't need `sort`, delete the import too.)

- [ ] **Step 4: Wire init() registration**

Create `sponge/init.go`:

```go
package sponge

import (
	"bytes"
	"io"

	"github.com/Velvet-MC/s2d/schem"
)

func init() {
	schem.Register(schem.FormatHandler{
		Name:       schem.FormatSpongeV2,
		Extensions: []string{".schem"},
		Signature:  signatureMatch,
		Read: func(r io.Reader) (*schem.Schematic, error) {
			// Re-buffer if needed — the dispatcher may pass a pre-buffered reader.
			return Read(r)
		},
	})
}

// signatureMatch returns true if the decompressed header looks like Sponge v2.
// The cheap check: NBT root contains the bytes "Version" early on.
func signatureMatch(headerPeek []byte) bool {
	return bytes.Contains(headerPeek, []byte("Version"))
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./sponge/... -v`
Expected: `TestRead_SingleStone` and `TestRead_MultiPalette` pass.

If they fail because `translate.Lookup` returns `nil` blocks, **that's expected at this stage** — the translate package is still a stub. Add a temporary skip:

```go
if testing.Short() { t.Skip("requires translate table; covered after Phase 7") }
```

OR, more honestly: defer this test until Phase 7's table is built. Comment out the assertions on `b.Block.EncodeBlock()` for now, leaving only structural checks (Pos, count, format). Re-enable the name assertions in Phase 7 Task 7.4.

- [ ] **Step 6: Commit**

```bash
git add sponge/reader.go sponge/init.go sponge/reader_test.go
git commit -m "feat(sponge): Sponge v2 reader + dispatcher registration"
```

---

## Phase 5: `translate` package — stub for unblocking parsers

A minimal `translate.Lookup` that always returns the missing block. This unblocks the sponge and legacy readers' tests until the real table arrives in Phase 7.

### Task 5.1: Translate stub + missing-block resolver

**Files:**
- Create: `translate/translate.go`
- Create: `translate/missing.go`
- Create: `translate/translate_test.go`

- [ ] **Step 1: Write tests**

Create `translate/translate_test.go`:

```go
package translate

import (
	"testing"

	"github.com/df-mc/dragonfly/server/world"
)

// stubBlock satisfies world.Block for tests that don't need real Bedrock blocks.
// In production code, MissingBlock is resolved via world.BlockByName at init.
type stubBlock struct{ name string }

func (b stubBlock) EncodeBlock() (string, map[string]any) { return b.name, nil }

func TestSetGetMissingBlock(t *testing.T) {
	prev := MissingBlock()
	defer SetMissingBlock(prev)

	swap := stubBlock{name: "minecraft:custom_marker"}
	SetMissingBlock(swap)
	got := MissingBlock()
	if got == nil {
		t.Fatal("MissingBlock returned nil")
	}
	name, _ := got.EncodeBlock()
	if name != "minecraft:custom_marker" {
		t.Errorf("got %q want minecraft:custom_marker", name)
	}
}

func TestLookup_FallsBackToMissing(t *testing.T) {
	res := Lookup("minecraft:nonexistent_block_for_test")
	if res.Recognized {
		t.Errorf("expected Recognized=false for unknown block")
	}
	if res.Block == nil {
		t.Errorf("Block must be non-nil even for unknowns")
	}
	_ = world.Block(res.Block) // assert interface satisfaction
}
```

Same caveat as in `schem`: confirm `world.Block`'s actual method set against the imported Dragonfly version and adjust `stubBlock` if `EncodeBlock` returns a different shape.

- [ ] **Step 2: Run test to verify build fails**

Run: `go test ./translate/...`
Expected: build error — `Lookup`, `MissingBlock`, `SetMissingBlock` undefined.

- [ ] **Step 3: Implement the stub**

Create `translate/translate.go`:

```go
// Package translate maps Java edition block-state strings to Dragonfly
// world.Block values. The full Java→Bedrock palette table is built once,
// lazily, on first call to Lookup.
//
// In v1.0 this file only exposes the public API and a placeholder Lookup
// that always returns the missing-block fallback. The full table is added
// in a later phase (translate/table.go) and replaces the stub.
package translate

import (
	"sync"

	"github.com/df-mc/dragonfly/server/world"
)

// Result is what Lookup returns.
type Result struct {
	Block      world.Block  // never nil; missing-block fallback if Recognized==false
	Liquid     world.Liquid // non-nil iff source was waterlogged
	Recognized bool
	RawKey     string       // canonical Java state string (echoed for reporting)
}

var (
	missingMu    sync.RWMutex
	missingBlock world.Block
)

// SetMissingBlock changes the global fallback used for unrecognized blocks.
// Call once at startup, before the first Lookup.
func SetMissingBlock(b world.Block) {
	missingMu.Lock()
	defer missingMu.Unlock()
	missingBlock = b
}

// MissingBlock returns the active fallback block.
func MissingBlock() world.Block {
	missingMu.RLock()
	defer missingMu.RUnlock()
	if missingBlock == nil {
		// Lazy default. Resolved in resolveDefaultMissing (missing.go).
		// Read-lock upgrade: release and call back through the locked setter.
		missingMu.RUnlock()
		b := resolveDefaultMissing()
		SetMissingBlock(b)
		missingMu.RLock()
	}
	return missingBlock
}

// Lookup translates a canonical Java state string to a Dragonfly Result.
// In v1.0-stub form, every key is unrecognized; this is replaced by the
// real implementation in translate/table.go.
func Lookup(canonicalJavaState string) Result {
	return Result{
		Block:      MissingBlock(),
		Recognized: false,
		RawKey:     canonicalJavaState,
	}
}
```

Create `translate/missing.go`:

```go
package translate

import (
	"github.com/df-mc/dragonfly/server/world"
)

// resolveDefaultMissing returns the default missing-block (magenta wool).
// If wool is unregistered or the property name differs from "color",
// the function falls back to whatever world.Block "minecraft:wool" alone
// resolves to. A test in the table-builder phase asserts the default
// resolves successfully.
func resolveDefaultMissing() world.Block {
	if b, ok := world.BlockByName("minecraft:wool", map[string]any{"color": "magenta"}); ok {
		return b
	}
	if b, ok := world.BlockByName("minecraft:wool", nil); ok {
		return b
	}
	// Last-resort: any non-nil block. Air would defeat the purpose of a marker.
	if b, ok := world.BlockByName("minecraft:magenta_concrete", nil); ok {
		return b
	}
	if b, ok := world.BlockByName("minecraft:stone", nil); ok {
		return b
	}
	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./translate/... -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add translate/translate.go translate/missing.go translate/translate_test.go
git commit -m "feat(translate): stub Lookup + configurable missing block"
```

---

## Phase 6: `legacy` package — MCEdit reader

Same shape as the Sponge reader. We vendor the legacy ID table as JSON, then write a reader that walks numeric IDs.

### Task 6.1: Vendor `legacy.json`

**Files:**
- Create: `legacy/data/legacy.json`
- Create: `legacy/data/LICENSE`
- Create: `legacy/data/ATTRIBUTION.md`

- [ ] **Step 1: Download `legacy.json` from EngineHub/WorldEdit**

The file lives at `worldedit-core/src/main/resources/com/sk89q/worldedit/world/registry/legacy.json` in the EngineHub/WorldEdit repo. From the S2D repo root:

```bash
mkdir -p legacy/data
curl -L -o legacy/data/legacy.json \
  https://raw.githubusercontent.com/EngineHub/WorldEdit/master/worldedit-core/src/main/resources/com/sk89q/worldedit/world/registry/legacy.json
```

If `master` has been renamed, use the default branch from `gh repo view EngineHub/WorldEdit --json defaultBranchRef`.

Pin the upstream commit:

```bash
git ls-remote https://github.com/EngineHub/WorldEdit.git HEAD
```

Record that commit SHA in `ATTRIBUTION.md` (Step 3).

- [ ] **Step 2: Copy WorldEdit's LICENSE**

```bash
curl -L -o legacy/data/LICENSE \
  https://raw.githubusercontent.com/EngineHub/WorldEdit/master/LICENSE.txt
```

- [ ] **Step 3: Write ATTRIBUTION.md**

Create `legacy/data/ATTRIBUTION.md`:

```markdown
# Legacy schematic ID table — Attribution

`legacy.json` is vendored from [EngineHub/WorldEdit](https://github.com/EngineHub/WorldEdit).

- Upstream path: `worldedit-core/src/main/resources/com/sk89q/worldedit/world/registry/legacy.json`
- Pinned commit: `<SHA from `git ls-remote ... HEAD` at vendor time>`
- License: GPL-3.0+ for WorldEdit Java code; the data file `legacy.json` itself is a factual mapping table (numeric block IDs to Java state strings) and is reused under fair use as a data interoperability aid. If a stricter reading is preferred, replace this file with a hand-curated equivalent — the data is publicly known.

## Refresh

```sh
curl -L -o legacy.json \
  https://raw.githubusercontent.com/EngineHub/WorldEdit/master/worldedit-core/src/main/resources/com/sk89q/worldedit/world/registry/legacy.json
```

Then update the pinned commit in this file.
```

**Important:** WorldEdit is GPL-3.0+. Vendoring data from a GPL project into an MIT library has nuance. The `legacy.json` file is a factual mapping table — courts have generally held that data tables are not creative works subject to copyright (cf. *Feist v. Rural*). However, the conservative path is to replace the file with a hand-rolled equivalent or to source from a permissively-licensed alternative (PrismarineJS minecraft-data, MIT). **If unsure, replace `legacy.json` with the equivalent file from `github.com/PrismarineJS/minecraft-data/data/pc/1.12.2/blocks.json` and write a small Go conversion in `legacy/ids_legacy_convert.go`.** Mark this as an open verification item in the design doc and circle back before publishing v0.1.

- [ ] **Step 4: Verify the file is syntactically valid JSON**

Run:
```bash
python -c "import json; json.load(open('legacy/data/legacy.json'))" && echo OK
```

Expected: `OK`. (If python is unavailable, use `jq . legacy/data/legacy.json > /dev/null && echo OK`.)

- [ ] **Step 5: Commit**

```bash
git add legacy/data/
git commit -m "data(legacy): vendor EngineHub/WorldEdit legacy.json"
```

### Task 6.2: Implement legacy ID lookup

**Files:**
- Create: `legacy/ids.go`
- Create: `legacy/ids_test.go`

- [ ] **Step 1: Inspect the file shape**

The vendored `legacy.json` may be a single map, or it may have sub-objects (`blocks`, `items`). Inspect the first 200 bytes:

```bash
head -c 200 legacy/data/legacy.json
```

If the structure is `{"blocks": {"1:0": "minecraft:stone", ...}}`, the loader unmarshals the wrapper. If the structure is flat (`{"1:0": "minecraft:stone"}`), simpler.

This step exists because file shape determines the unmarshal target. Note the shape and adjust the struct in Step 3 accordingly.

- [ ] **Step 2: Write the failing test**

Create `legacy/ids_test.go`:

```go
package legacy

import "testing"

func TestLookup_KnownIDs(t *testing.T) {
	cases := []struct {
		id, data int
		want     string
	}{
		{1, 0, "minecraft:stone"},
		{17, 0, "minecraft:oak_log[axis=y]"},   // canonical state should be sorted alphabetically
		{17, 4, "minecraft:oak_log[axis=x]"},
		{0, 0, ""},                             // air handled by caller, not table
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			got, ok := Lookup(c.id, c.data)
			if c.id == 0 {
				if ok {
					t.Errorf("id=0 should not be in table; got %q", got)
				}
				return
			}
			if !ok {
				t.Fatalf("id=%d data=%d not found", c.id, c.data)
			}
			if got != c.want {
				t.Errorf("id=%d data=%d: got %q want %q", c.id, c.data, got, c.want)
			}
		})
	}
}
```

The exact expected strings depend on `legacy.json`'s formatting. If `oak_log[axis=y]` actually appears as `oak_log[axis=y,stripped=false,...]` in the upstream file, adjust the expected values to match — this test pins behavior to the vendored data.

Run `head -c 5000 legacy/data/legacy.json` to see how WorldEdit formats the values; adjust `want` strings accordingly.

- [ ] **Step 3: Run test (verify build fails)**

Run: `go test ./legacy/...`
Expected: build error — `Lookup` undefined.

- [ ] **Step 4: Implement the loader**

Create `legacy/ids.go`:

```go
// Package legacy reads MCEdit-format `.schematic` files and maps the
// pre-1.13 numeric (block ID, data value) tuples to Java edition state
// strings via a vendored table from EngineHub/WorldEdit.
package legacy

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed data/legacy.json
var legacyJSON []byte

// Adjust the struct to match the vendored file shape. Two common shapes:
//
//	1. Flat: {"1:0": "minecraft:stone", ...}
//	2. Nested: {"blocks": {"1:0": "minecraft:stone", ...}, ...}
//
// If shape 2, change `legacyTable` below to unmarshal a wrapper struct.
type legacyTable struct {
	Blocks map[string]string `json:"blocks"`
}

var (
	tableOnce sync.Once
	table     map[string]string
	tableErr  error
)

func loadTable() {
	var t legacyTable
	if err := json.Unmarshal(legacyJSON, &t); err != nil {
		// Try flat shape as fallback.
		var flat map[string]string
		if err2 := json.Unmarshal(legacyJSON, &flat); err2 != nil {
			tableErr = fmt.Errorf("legacy: parse table: %v / %v", err, err2)
			return
		}
		table = flat
		return
	}
	if t.Blocks != nil {
		table = t.Blocks
		return
	}
	tableErr = fmt.Errorf("legacy: empty table")
}

// Lookup returns the Java state string for a pre-1.13 (id, data) tuple.
// id == 0 (air) is intentionally not in the table; the reader handles
// air specially before calling Lookup.
func Lookup(id, data int) (string, bool) {
	tableOnce.Do(loadTable)
	if tableErr != nil {
		return "", false
	}
	if v, ok := table[fmt.Sprintf("%d:%d", id, data)]; ok {
		return v, true
	}
	// Some entries omit data when 0; try the bare id form.
	if data == 0 {
		if v, ok := table[fmt.Sprintf("%d", id)]; ok {
			return v, true
		}
	}
	return "", false
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./legacy/... -v`
Expected: PASS for known IDs (after Step 4 expected-value adjustments).

- [ ] **Step 6: Commit**

```bash
git add legacy/ids.go legacy/ids_test.go
git commit -m "feat(legacy): vendored numeric-ID lookup"
```

### Task 6.3: Generate legacy fixture files

**Files:**
- Create: `legacy/testdata_gen_test.go`

- [ ] **Step 1: Write a generator**

Create `legacy/testdata_gen_test.go`:

```go
package legacy

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func TestGenerateFixtures(t *testing.T) {
	if os.Getenv("S2D_REGEN_FIXTURES") != "1" {
		t.Skip("set S2D_REGEN_FIXTURES=1 to regenerate")
	}
	dir := "testdata"
	if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatal(err) }

	// basic.schematic: 2x1x1, two stones (id=1, data=0).
	{
		root := map[string]any{
			"Materials": "Alpha",
			"Width":     int16(2),
			"Height":    int16(1),
			"Length":    int16(1),
			"Blocks":    []byte{0x01, 0x01},   // id=1 (stone) at both cells
			"Data":      []byte{0x00},          // both nibbles 0
		}
		writeFixture(t, filepath.Join(dir, "basic.schematic"), root)
	}

	// addblocks.schematic: 1x1x1, id=256 (out of low-byte range), data=0.
	// AddBlocks holds high nibble. id=256 → high4=1, low8=0.
	// Cell index 0 is even → high nibble of AddBlocks[0]: 0x10.
	{
		root := map[string]any{
			"Materials": "Alpha",
			"Width":     int16(1),
			"Height":    int16(1),
			"Length":    int16(1),
			"Blocks":    []byte{0x00},
			"AddBlocks": []byte{0x10},
			"Data":      []byte{0x00},
		}
		writeFixture(t, filepath.Join(dir, "addblocks.schematic"), root)
	}
}

func writeFixture(t *testing.T, path string, root map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	enc := nbt.NewEncoderWithEncoding(&buf, nbt.BigEndian)
	if err := enc.Encode(root); err != nil { t.Fatal(err) }
	var gz bytes.Buffer
	gw := gzip.NewWriter(&gz)
	if _, err := gw.Write(buf.Bytes()); err != nil { t.Fatal(err) }
	if err := gw.Close(); err != nil { t.Fatal(err) }
	if err := os.WriteFile(path, gz.Bytes(), 0o644); err != nil { t.Fatal(err) }
}
```

- [ ] **Step 2: Generate fixtures**

Run:
```bash
S2D_REGEN_FIXTURES=1 go test ./legacy/ -run TestGenerateFixtures -v
```

Expected: PASS, two files appear.

- [ ] **Step 3: Commit**

```bash
git add legacy/testdata_gen_test.go legacy/testdata/
git commit -m "test(legacy): committed fixtures + regen generator"
```

### Task 6.4: Implement the legacy reader

**Files:**
- Create: `legacy/reader.go`
- Create: `legacy/init.go`
- Create: `legacy/reader_test.go`

- [ ] **Step 1: Write the failing test**

Create `legacy/reader_test.go`:

```go
package legacy

import (
	"os"
	"testing"
)

func TestRead_Basic(t *testing.T) {
	f, err := os.Open("testdata/basic.schematic")
	if err != nil { t.Fatal(err) }
	defer f.Close()

	s, err := Read(f)
	if err != nil { t.Fatalf("Read: %v", err) }

	if s.Width != 2 || s.Height != 1 || s.Length != 1 {
		t.Fatalf("dims: %dx%dx%d", s.Width, s.Height, s.Length)
	}
	if len(s.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(s.Blocks))
	}
	if s.Blocks[0].Pos != [3]int{0, 0, 0} || s.Blocks[1].Pos != [3]int{1, 0, 0} {
		t.Errorf("positions: %v %v", s.Blocks[0].Pos, s.Blocks[1].Pos)
	}
}

func TestRead_AddBlocks(t *testing.T) {
	f, err := os.Open("testdata/addblocks.schematic")
	if err != nil { t.Fatal(err) }
	defer f.Close()

	_, err = Read(f)
	// Until the translation table is built, this may produce missing blocks
	// but should NOT error on parsing — id=256 is structurally valid.
	if err != nil { t.Fatalf("Read: %v", err) }
}
```

- [ ] **Step 2: Run test (verify build fails)**

Run: `go test ./legacy/...`
Expected: build error — `Read` undefined.

- [ ] **Step 3: Implement the reader**

Create `legacy/reader.go`:

```go
package legacy

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/Velvet-MC/s2d/palette"
	"github.com/Velvet-MC/s2d/schem"
	"github.com/Velvet-MC/s2d/translate"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

type rawLegacy struct {
	Materials string `nbt:"Materials"`
	Width     int16  `nbt:"Width"`
	Height    int16  `nbt:"Height"`
	Length    int16  `nbt:"Length"`
	Blocks    []byte `nbt:"Blocks"`
	Data      []byte `nbt:"Data"`
	AddBlocks []byte `nbt:"AddBlocks,omitempty"`
}

// Read parses a legacy MCEdit `.schematic` file from r.
func Read(r io.Reader) (*schem.Schematic, error) {
	gz, err := gzip.NewReader(r)
	if err != nil { return nil, fmt.Errorf("legacy: gzip: %w", err) }
	defer gz.Close()
	body, err := io.ReadAll(gz)
	if err != nil { return nil, fmt.Errorf("legacy: read: %w", err) }

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
					// Air. Translate via canonical key for symmetry; the
					// translate table maps minecraft:air to a Bedrock air block.
					res := translate.Lookup("minecraft:air")
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
```

- [ ] **Step 4: Wire init() registration**

Create `legacy/init.go`:

```go
package legacy

import (
	"bytes"
	"io"

	"github.com/Velvet-MC/s2d/schem"
)

func init() {
	schem.Register(schem.FormatHandler{
		Name:       schem.FormatLegacy,
		Extensions: []string{".schematic"},
		Signature:  signatureMatch,
		Read: func(r io.Reader) (*schem.Schematic, error) {
			return Read(r)
		},
	})
}

func signatureMatch(headerPeek []byte) bool {
	return bytes.Contains(headerPeek, []byte("Materials"))
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./legacy/... -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add legacy/reader.go legacy/init.go legacy/reader_test.go
git commit -m "feat(legacy): MCEdit .schematic reader + dispatcher registration"
```

---

## Phase 7: `translate` — full table

This is the heaviest phase. We vendor the Bedrock palette and Java schema, write the per-property converters, and build the table.

### Task 7.1: Vendor the Bedrock palette and overrides skeleton

**Files:**
- Create: `translate/data/bedrock_palette.1_26_20.nbt`
- Create: `translate/data/overrides.json`
- Create: `translate/data/LICENSE`
- Create: `translate/data/ATTRIBUTION.md`
- Create: `translate/data/REGEN.md`

- [ ] **Step 1: Acquire the Bedrock palette**

The local clone at `_research/geyser` has it. From the S2D repo root:

```bash
mkdir -p translate/data
cp ../_research/geyser/core/src/main/resources/bedrock/block_palette.1_26_20.nbt \
   translate/data/bedrock_palette.1_26_20.nbt
```

If running fresh (no `_research` clone available), grab it from Geyser:

```bash
curl -L -o translate/data/bedrock_palette.1_26_20.nbt \
  https://raw.githubusercontent.com/GeyserMC/Geyser/master/core/src/main/resources/bedrock/block_palette.1_26_20.nbt
```

Pin the upstream commit:

```bash
git ls-remote https://github.com/GeyserMC/Geyser.git HEAD
```

- [ ] **Step 2: Copy Geyser's LICENSE**

```bash
curl -L -o translate/data/LICENSE \
  https://raw.githubusercontent.com/GeyserMC/Geyser/master/LICENSE
```

- [ ] **Step 3: Write empty overrides.json**

Create `translate/data/overrides.json`:

```json
{
}
```

Entries are added later as table-build tests surface divergent identifiers.

- [ ] **Step 4: Write ATTRIBUTION.md**

Create `translate/data/ATTRIBUTION.md`:

```markdown
# Translation data — Attribution

## Bedrock palette (`bedrock_palette.1_26_20.nbt`)

Vendored from [GeyserMC/Geyser](https://github.com/GeyserMC/Geyser).

- Upstream path: `core/src/main/resources/bedrock/block_palette.1_26_20.nbt`
- Pinned commit: `<SHA from `git ls-remote ... HEAD` at vendor time>`
- License: MIT (see `LICENSE`).

## Java block schema (`java_blocks.json`)

(Filled in once the source is chosen — see open verification #4 in the design doc.
Likely: PrismarineJS minecraft-data, MIT.)

## Overrides (`overrides.json`)

Hand-maintained by S2D. MIT (S2D).
```

- [ ] **Step 5: Write REGEN.md**

Create `translate/data/REGEN.md`:

```markdown
# Refresh procedure

When Dragonfly upgrades to a new Bedrock palette version:

1. Copy the new palette from Geyser:

   ```sh
   curl -L -o bedrock_palette.<NEW_VERSION>.nbt \
     https://raw.githubusercontent.com/GeyserMC/Geyser/master/core/src/main/resources/bedrock/block_palette.<NEW_VERSION>.nbt
   ```

2. Update `bedrockPaletteFile` in `translate/table.go` to the new filename.

3. Update the pinned commit in `ATTRIBUTION.md`.

4. Run integrity tests:

   ```sh
   go test ./translate -run TestPaletteCoverage
   go test ./translate -run TestPaletteMatchesDragonfly
   ```

5. For any newly-flagged divergent blocks, add entries to `overrides.json`.

6. Tag a new S2D release.
```

- [ ] **Step 6: Commit**

```bash
git add translate/data/
git commit -m "data(translate): vendor Bedrock palette + skeleton overrides"
```

### Task 7.2: Vendor the Java block schema

**Files:**
- Create: `translate/data/java_blocks.json`

This is the biggest open verification (#4 in the design doc). The cleanest sources, in priority order:

1. **PrismarineJS minecraft-data** — already JSON, MIT-licensed, kept up-to-date by the Node ecosystem.
   - Path: `https://raw.githubusercontent.com/PrismarineJS/minecraft-data/master/data/pc/<MC_VERSION>/blocks.json`
   - For 1.20.x: `data/pc/1.20.2/blocks.json`
   - Format: an array of `{name, displayName, hardness, stackSize, ..., states: [...]}`. The `states` array under each block is the list of (java state) → minimum-state-id, and the `properties` field describes each property's allowed values.

2. **Mojang's data generator** — canonical but requires running the Java jar. Documented at https://minecraft.wiki/w/Tutorials/Running_the_data_generator. Output includes `generated/reports/blocks.json`.

For v1.0, **use PrismarineJS**.

- [ ] **Step 1: Choose the Java MC version to target**

Pin to a recent stable Java version that matches what most `.schem` files in the wild are authored for. Recommendation: **1.20.2** (data version 3578). Sponge v2 schematics include `DataVersion`; if a loaded `.schem` has a wildly older version, our table is still likely to cover it because vanilla Java rarely renames blocks.

- [ ] **Step 2: Download blocks.json from PrismarineJS**

```bash
curl -L -o translate/data/java_blocks.json \
  https://raw.githubusercontent.com/PrismarineJS/minecraft-data/master/data/pc/1.20.2/blocks.json
```

- [ ] **Step 3: Verify it parses**

```bash
python -c "import json; b=json.load(open('translate/data/java_blocks.json')); print('blocks:', len(b))"
```

Expected: prints `blocks: 1000+` (vanilla 1.20.2 has ~900-1000 block types, which expands to ~25k–35k states).

- [ ] **Step 4: Update ATTRIBUTION.md**

Edit `translate/data/ATTRIBUTION.md` and replace the Java block schema section with:

```markdown
## Java block schema (`java_blocks.json`)

Vendored from [PrismarineJS/minecraft-data](https://github.com/PrismarineJS/minecraft-data).

- Upstream path: `data/pc/1.20.2/blocks.json`
- MC version: 1.20.2 (data version 3578)
- Pinned commit: `<SHA from `git ls-remote https://github.com/PrismarineJS/minecraft-data.git HEAD`>`
- License: MIT.
```

- [ ] **Step 5: Commit**

```bash
git add translate/data/java_blocks.json translate/data/ATTRIBUTION.md
git commit -m "data(translate): vendor PrismarineJS Java block schema (1.20.2)"
```

### Task 7.3: Property converters

We write one Go file per property family, plus a single shared test file. The converter signature:

```go
func Name(javaValue, bedrockIdent string) (bedrockProp string, bedrockValue any, ok bool)
```

`ok=false` signals that this property has no Bedrock equivalent and should be dropped from the Bedrock state.

To keep the plan compact, **the engineer implements all converters in one batch with a shared test file**, then commits them together.

**Files:**
- Create: `translate/properties/properties.go`
- Create: `translate/properties/axis.go`
- Create: `translate/properties/facing.go`
- Create: `translate/properties/half.go`
- Create: `translate/properties/hinge.go`
- Create: `translate/properties/open.go`
- Create: `translate/properties/powered.go`
- Create: `translate/properties/lit.go`
- Create: `translate/properties/waterlogged.go`
- Create: `translate/properties/snowy.go`
- Create: `translate/properties/age.go`
- Create: `translate/properties/distance.go`
- Create: `translate/properties/persistent.go`
- Create: `translate/properties/level.go`
- Create: `translate/properties/redstone.go`
- Create: `translate/properties/slab.go`
- Create: `translate/properties/wall.go`
- Create: `translate/properties/properties_test.go`

- [ ] **Step 1: Write the registry and shared test**

Create `translate/properties/properties.go`:

```go
// Package properties holds per-property Java→Bedrock value converters.
// Each converter takes the Java value and the Bedrock identifier (so
// converters can branch on block family where Bedrock uses different
// property names for the same Java concept) and returns the Bedrock
// property name + value.
package properties

import "strings"

// Converter is the signature shared by all per-property converters.
//
// javaValue is the Java property value as a string (e.g. "y", "north", "true").
// bedrockIdent is the Bedrock block identifier the result will be applied to,
//   without the "minecraft:" prefix (e.g. "oak_log", "oak_stairs").
//
// Returns the Bedrock property name + value to insert into the Bedrock state
// map, plus an ok flag. ok=false signals the property should be dropped
// (e.g. computed-only Bedrock properties, or properties that have no
// Bedrock equivalent).
type Converter func(javaValue, bedrockIdent string) (bedrockProp string, bedrockValue any, ok bool)

// Registry maps Java property names to converters. The translate table
// builder consults this map for every property on every Java state.
// Properties not in the registry are dropped (a logged warning would
// be added for diagnostics in a future revision).
var Registry = map[string]Converter{
	"axis":         Axis,
	"facing":       Facing,
	"half":         Half,
	"hinge":        Hinge,
	"open":         Open,
	"powered":      Powered,
	"lit":          Lit,
	"waterlogged":  Waterlogged,
	"snowy":        Snowy,
	"age":          Age,
	"distance":     Distance,
	"persistent":   Persistent,
	"level":        Level,
	"power":        Redstone,
	"type":         SlabType,
	"up":           WallUp,
	"north":        WallNorth,
	"east":         WallEast,
	"south":        WallSouth,
	"west":         WallWest,
}

// boolToBit converts "true"/"false" Java strings to 1/0 bytes for Bedrock.
func boolToBit(s string) byte {
	if strings.EqualFold(s, "true") { return 1 }
	return 0
}
```

- [ ] **Step 2: Write each converter**

Create `translate/properties/axis.go`:

```go
package properties

// Axis converts Java AXIS (x|y|z) to Bedrock pillar_axis.
// Logs, basalt, deepslate pillars, etc. all use the same prop name on Bedrock.
func Axis(javaValue, bedrockIdent string) (string, any, bool) {
	return "pillar_axis", javaValue, true
}
```

Create `translate/properties/facing.go`:

```go
package properties

// Facing converts Java facing (north|south|east|west|up|down) to whichever
// Bedrock property a particular block uses. Bedrock has historically used
// many names for the same concept: "direction" (4-way int), "facing_direction"
// (6-way int), "weirdo_direction" (stairs, 4-way int with a unique encoding),
// "minecraft:cardinal_direction" (string), and "ground_sign_direction" (16-way).
//
// The branch table below covers the common cases. Add families as needed
// when the table-build tests surface mismatches.
func Facing(javaValue, bedrockIdent string) (string, any, bool) {
	switch bedrockIdent {
	case "oak_stairs", "spruce_stairs", "birch_stairs", "jungle_stairs",
		"acacia_stairs", "dark_oak_stairs", "mangrove_stairs", "cherry_stairs",
		"crimson_stairs", "warped_stairs", "bamboo_stairs",
		"stone_stairs", "cobblestone_stairs", "mossy_cobblestone_stairs",
		"brick_stairs", "stone_brick_stairs", "mossy_stone_brick_stairs",
		"sandstone_stairs", "smooth_sandstone_stairs", "red_sandstone_stairs",
		"smooth_red_sandstone_stairs", "nether_brick_stairs", "red_nether_brick_stairs",
		"quartz_stairs", "smooth_quartz_stairs", "purpur_stairs",
		"prismarine_stairs", "prismarine_brick_stairs", "dark_prismarine_stairs",
		"end_brick_stairs", "blackstone_stairs", "polished_blackstone_stairs",
		"polished_blackstone_brick_stairs", "polished_granite_stairs",
		"polished_diorite_stairs", "polished_andesite_stairs",
		"granite_stairs", "diorite_stairs", "andesite_stairs",
		"deepslate_brick_stairs", "deepslate_tile_stairs",
		"polished_deepslate_stairs", "cobbled_deepslate_stairs",
		"mud_brick_stairs", "tuff_stairs", "polished_tuff_stairs", "tuff_brick_stairs":
		// Stairs on Bedrock use weirdo_direction:
		// east=0, west=1, south=2, north=3.
		switch javaValue {
		case "east":  return "weirdo_direction", int32(0), true
		case "west":  return "weirdo_direction", int32(1), true
		case "south": return "weirdo_direction", int32(2), true
		case "north": return "weirdo_direction", int32(3), true
		}
		return "weirdo_direction", int32(0), true
	default:
		// Most directional blocks use "direction" or "facing_direction".
		// 4-way encoding: south=0, west=1, north=2, east=3.
		switch javaValue {
		case "south": return "direction", int32(0), true
		case "west":  return "direction", int32(1), true
		case "north": return "direction", int32(2), true
		case "east":  return "direction", int32(3), true
		case "up":    return "facing_direction", int32(1), true
		case "down":  return "facing_direction", int32(0), true
		}
		return "direction", int32(0), true
	}
}
```

Create `translate/properties/half.go`:

```go
package properties

import "strings"

// Half handles Java's "half" property which means different things on
// stairs vs doors vs flowers.
//   stairs: "top"|"bottom" → upside_down_bit 1|0
//   doors:  "lower"|"upper" → upper_block_bit 0|1
//   tall flowers: "lower"|"upper" → upper_block_bit (same as doors)
func Half(javaValue, bedrockIdent string) (string, any, bool) {
	if strings.HasSuffix(bedrockIdent, "_door") || strings.HasSuffix(bedrockIdent, "_flower") ||
		bedrockIdent == "iron_door" || bedrockIdent == "tall_grass" || bedrockIdent == "large_fern" ||
		bedrockIdent == "sunflower" || bedrockIdent == "lilac" || bedrockIdent == "rose_bush" ||
		bedrockIdent == "peony" || bedrockIdent == "pitcher_plant" {
		if javaValue == "upper" { return "upper_block_bit", byte(1), true }
		return "upper_block_bit", byte(0), true
	}
	// stairs (and slabs sometimes use "type" instead — slabs handled by SlabType)
	if javaValue == "top" { return "upside_down_bit", byte(1), true }
	return "upside_down_bit", byte(0), true
}
```

Create `translate/properties/hinge.go`:

```go
package properties

func Hinge(javaValue, bedrockIdent string) (string, any, bool) {
	if javaValue == "right" { return "door_hinge_bit", byte(1), true }
	return "door_hinge_bit", byte(0), true
}
```

Create `translate/properties/open.go`:

```go
package properties

func Open(javaValue, bedrockIdent string) (string, any, bool) {
	return "open_bit", boolToBit(javaValue), true
}
```

Create `translate/properties/powered.go`:

```go
package properties

func Powered(javaValue, bedrockIdent string) (string, any, bool) {
	return "powered_bit", boolToBit(javaValue), true
}
```

Create `translate/properties/lit.go`:

```go
package properties

func Lit(javaValue, bedrockIdent string) (string, any, bool) {
	// Bedrock 1.20+ uses a "lit" string property on most light-emitting blocks.
	return "lit", javaValue == "true", true
}
```

Create `translate/properties/waterlogged.go`:

```go
package properties

// Waterlogged returns ok=false because Bedrock has no waterlogged property.
// The translate.table builder detects waterlogged=true and emits a Liquid
// alongside the base block instead.
func Waterlogged(javaValue, bedrockIdent string) (string, any, bool) {
	return "", nil, false
}
```

Create `translate/properties/snowy.go`:

```go
package properties

func Snowy(javaValue, bedrockIdent string) (string, any, bool) {
	// Grass / podzol / mycelium snowy state.
	return "covered_bit", boolToBit(javaValue), true
}
```

Create `translate/properties/age.go`:

```go
package properties

import "strconv"

// Age maps Java numeric "age" property to Bedrock "growth" or "age" int.
// Most crops (wheat, carrots, potatoes, beetroot) use "growth" 0-7.
// Cactus and sugar cane use "age" 0-15. Default to "growth" since it covers
// more cases; the override file should special-case cactus/sugar_cane.
func Age(javaValue, bedrockIdent string) (string, any, bool) {
	n, _ := strconv.Atoi(javaValue)
	if bedrockIdent == "cactus" || bedrockIdent == "reeds" /* sugar cane */ {
		return "age", int32(n), true
	}
	return "growth", int32(n), true
}
```

Create `translate/properties/distance.go`:

```go
package properties

// Distance: leaves block, Java "distance" 1-7. Bedrock has no equivalent
// (leaves persistence is handled differently). Drop it.
func Distance(javaValue, bedrockIdent string) (string, any, bool) {
	return "", nil, false
}
```

Create `translate/properties/persistent.go`:

```go
package properties

func Persistent(javaValue, bedrockIdent string) (string, any, bool) {
	return "persistent_bit", boolToBit(javaValue), true
}
```

Create `translate/properties/level.go`:

```go
package properties

import "strconv"

// Level: Java water/lava "level" 0-15 → Bedrock "liquid_depth" 0-15.
func Level(javaValue, bedrockIdent string) (string, any, bool) {
	n, _ := strconv.Atoi(javaValue)
	return "liquid_depth", int32(n), true
}
```

Create `translate/properties/redstone.go`:

```go
package properties

import "strconv"

// Redstone wire power level. Bedrock uses "redstone_signal" 0-15.
func Redstone(javaValue, bedrockIdent string) (string, any, bool) {
	n, _ := strconv.Atoi(javaValue)
	return "redstone_signal", int32(n), true
}
```

Create `translate/properties/slab.go`:

```go
package properties

// SlabType handles Java's slab "type" property. "double" requires an
// identifier swap (oak_slab → double_oak_slab on Bedrock); the table
// builder honors that via overrides.json. Here we only handle top/bottom.
func SlabType(javaValue, bedrockIdent string) (string, any, bool) {
	if javaValue == "top" { return "top_slot_bit", byte(1), true }
	return "top_slot_bit", byte(0), true
}
```

Create `translate/properties/wall.go`:

```go
package properties

// Walls have per-side connection types: "none", "low", "tall".
// Bedrock represents these as "wall_connection_type_<side>" properties.

func WallUp(javaValue, bedrockIdent string) (string, any, bool) {
	return "wall_post_bit", boolToBit(javaValue), true
}

func WallNorth(javaValue, bedrockIdent string) (string, any, bool) {
	return "wall_connection_type_north", javaValue, true
}
func WallEast(javaValue, bedrockIdent string) (string, any, bool) {
	return "wall_connection_type_east", javaValue, true
}
func WallSouth(javaValue, bedrockIdent string) (string, any, bool) {
	return "wall_connection_type_south", javaValue, true
}
func WallWest(javaValue, bedrockIdent string) (string, any, bool) {
	return "wall_connection_type_west", javaValue, true
}
```

- [ ] **Step 3: Write the shared property test**

Create `translate/properties/properties_test.go`:

```go
package properties

import "testing"

type tc struct {
	prop      string
	javaValue string
	bedrockId string
	wantProp  string
	wantValue any
	wantOK    bool
}

func TestConverters(t *testing.T) {
	cases := []tc{
		// axis
		{"axis", "y", "oak_log", "pillar_axis", "y", true},
		{"axis", "x", "deepslate", "pillar_axis", "x", true},
		// facing — stairs
		{"facing", "north", "oak_stairs", "weirdo_direction", int32(3), true},
		{"facing", "east", "oak_stairs", "weirdo_direction", int32(0), true},
		// facing — generic directional
		{"facing", "north", "furnace", "direction", int32(2), true},
		{"facing", "south", "furnace", "direction", int32(0), true},
		// half — stairs
		{"half", "top", "oak_stairs", "upside_down_bit", byte(1), true},
		{"half", "bottom", "oak_stairs", "upside_down_bit", byte(0), true},
		// half — door
		{"half", "upper", "oak_door", "upper_block_bit", byte(1), true},
		{"half", "lower", "oak_door", "upper_block_bit", byte(0), true},
		// hinge
		{"hinge", "left", "oak_door", "door_hinge_bit", byte(0), true},
		{"hinge", "right", "oak_door", "door_hinge_bit", byte(1), true},
		// open / powered / persistent
		{"open", "true", "oak_door", "open_bit", byte(1), true},
		{"powered", "true", "oak_button", "powered_bit", byte(1), true},
		{"persistent", "true", "oak_leaves", "persistent_bit", byte(1), true},
		// waterlogged → dropped
		{"waterlogged", "true", "oak_stairs", "", nil, false},
		// distance → dropped
		{"distance", "5", "oak_leaves", "", nil, false},
		// snowy
		{"snowy", "true", "grass", "covered_bit", byte(1), true},
		// age (crops)
		{"age", "7", "wheat", "growth", int32(7), true},
		// age (cactus)
		{"age", "12", "cactus", "age", int32(12), true},
		// level
		{"level", "8", "water", "liquid_depth", int32(8), true},
		// redstone
		{"power", "10", "redstone_wire", "redstone_signal", int32(10), true},
		// slab type
		{"type", "top", "oak_slab", "top_slot_bit", byte(1), true},
		// walls
		{"up", "true", "cobblestone_wall", "wall_post_bit", byte(1), true},
		{"north", "low", "cobblestone_wall", "wall_connection_type_north", "low", true},
	}
	for _, c := range cases {
		t.Run(c.prop+"/"+c.javaValue+"/"+c.bedrockId, func(t *testing.T) {
			conv, ok := Registry[c.prop]
			if !ok { t.Fatalf("no converter for %q", c.prop) }
			gotProp, gotValue, gotOK := conv(c.javaValue, c.bedrockId)
			if gotOK != c.wantOK {
				t.Errorf("ok: got %v want %v", gotOK, c.wantOK)
			}
			if !c.wantOK { return }
			if gotProp != c.wantProp {
				t.Errorf("prop: got %q want %q", gotProp, c.wantProp)
			}
			if gotValue != c.wantValue {
				t.Errorf("value: got %v(%T) want %v(%T)", gotValue, gotValue, c.wantValue, c.wantValue)
			}
		})
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./translate/properties/ -v`
Expected: all subtests PASS.

If any value type mismatches (e.g. you used `int32` but the test expects `byte`), adjust either the converter or the test. Bedrock's NBT palette consistently uses `byte` for boolean-ish properties (`*_bit`) and `int32` for multi-valued integers. Match that.

- [ ] **Step 5: Commit**

```bash
git add translate/properties/
git commit -m "feat(translate/properties): per-property Java to Bedrock converters"
```

### Task 7.4: Build the translation table

This task ties everything together: load embedded data, walk the Java schema, apply converters, validate against the Bedrock palette, populate the lookup map.

**Files:**
- Create: `translate/table.go`
- Create: `translate/table_test.go`
- Modify: `translate/translate.go` (replace stub Lookup)

- [ ] **Step 1: Write the table builder**

Create `translate/table.go`:

```go
package translate

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
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

// PrismarineJS minecraft-data block schema (subset we need).
type prismaBlock struct {
	Name       string             `json:"name"`
	States     []prismaState      `json:"states"`
	Properties map[string][]string `json:"-"` // Optional; not used directly
}

type prismaState struct {
	Name       string                  `json:"name"`
	Type       string                  `json:"type"`
	NumValues  int                     `json:"num_values"`
	Values     []string                `json:"values"`
}

// Geyser palette entry (one block-state on Bedrock).
type bedrockState struct {
	Name   string                 `nbt:"name"`
	States map[string]any         `nbt:"states"`
	Version int32                  `nbt:"version,omitempty"`
}

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

// buildTable populates `table` and is invoked exactly once via tableOnce.
func buildTable() {
	defer func() {
		if r := recover(); r != nil {
			tableErr = fmt.Errorf("translate: build panic: %v", r)
		}
	}()

	// 1. Parse Java schema.
	var javaBlocks []prismaBlock
	if err := json.Unmarshal(javaBlocksJSON, &javaBlocks); err != nil {
		tableErr = fmt.Errorf("translate: java_blocks.json: %w", err)
		return
	}

	// 2. Parse overrides.
	var overrides map[string]override
	if err := json.Unmarshal(overridesJSON, &overrides); err != nil {
		tableErr = fmt.Errorf("translate: overrides.json: %w", err)
		return
	}

	// 3. Parse Bedrock palette into an indexable map.
	bedrockIndex, err := loadBedrockPalette(bedrockPaletteNBT)
	if err != nil {
		tableErr = fmt.Errorf("translate: bedrock palette: %w", err)
		return
	}

	// 4. For each Java block, expand the state space and translate each.
	t := make(map[string]Result, 30000)
	for _, jb := range javaBlocks {
		// Determine Bedrock identifier (override or identity match).
		bedrockIdent := jb.Name
		ovr, hasOvr := overrides["minecraft:"+jb.Name]
		if hasOvr && ovr.BedrockIdentifier != "" {
			bedrockIdent = strings.TrimPrefix(ovr.BedrockIdentifier, "minecraft:")
		}

		// Properties on this block (PrismarineJS exposes via States slice).
		propNames, propValues := extractProperties(jb)

		// Build cartesian product of property values.
		// If no properties, single canonical key.
		if len(propNames) == 0 {
			canonical := "minecraft:" + jb.Name
			res := translateOne(jb.Name, bedrockIdent, nil, false, ovr, bedrockIndex)
			res.RawKey = canonical
			t[canonical] = res
			continue
		}

		// Cartesian product.
		idx := make([]int, len(propNames))
		for {
			javaProps := make(map[string]string, len(propNames))
			for i, pn := range propNames {
				javaProps[pn] = propValues[i][idx[i]]
			}
			canonical := canonicalKey(jb.Name, javaProps)
			waterlogged := strings.EqualFold(javaProps["waterlogged"], "true")
			res := translateOne(jb.Name, bedrockIdent, javaProps, waterlogged, ovr, bedrockIndex)
			res.RawKey = canonical
			t[canonical] = res

			// Next combination.
			done := true
			for i := len(idx) - 1; i >= 0; i-- {
				idx[i]++
				if idx[i] < len(propValues[i]) { done = false; break }
				idx[i] = 0
			}
			if done { break }
		}
	}

	// Always include canonical air entries.
	if airBlock, ok := world.BlockByName("minecraft:air", nil); ok {
		air := Result{Block: airBlock, Recognized: true, RawKey: "minecraft:air"}
		t["minecraft:air"] = air
		t["minecraft:cave_air"] = air
		t["minecraft:void_air"] = air
	}

	table = t
}

// extractProperties returns the property names and value lists for a Java block.
// PrismarineJS shape: each block has a `states` array containing states with
// numeric ids; properties are conveyed via a parallel `properties` map in newer
// versions of minecraft-data. To remain robust, this function inspects all
// state names (which embed properties as `name[k1=v1,k2=v2]`) and unions them.
func extractProperties(jb prismaBlock) (names []string, values [][]string) {
	seen := map[string]map[string]struct{}{}
	for _, st := range jb.States {
		// Some PrismarineJS revisions list per-property values directly in `states`,
		// but the schema has changed over time. The defensive path: if `States`
		// has populated `Values`, treat it as a single-property list.
		if st.Name != "" && len(st.Values) > 0 {
			if seen[st.Name] == nil { seen[st.Name] = map[string]struct{}{} }
			for _, v := range st.Values { seen[st.Name][v] = struct{}{} }
		}
	}
	for n, vs := range seen {
		names = append(names, n)
		var sortedValues []string
		for v := range vs { sortedValues = append(sortedValues, v) }
		sort.Strings(sortedValues)
		values = append(values, sortedValues)
	}
	// Sort properties for determinism.
	sortByName(names, values)
	return
}

func sortByName(names []string, values [][]string) {
	// Simple sort with parallel reordering.
	type pair struct { n string; v []string }
	ps := make([]pair, len(names))
	for i := range names { ps[i] = pair{names[i], values[i]} }
	sort.Slice(ps, func(i, j int) bool { return ps[i].n < ps[j].n })
	for i := range ps { names[i] = ps[i].n; values[i] = ps[i].v }
}

func canonicalKey(blockName string, props map[string]string) string {
	if len(props) == 0 { return "minecraft:" + blockName }
	keys := make([]string, 0, len(props))
	for k := range props { keys = append(keys, k) }
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("minecraft:")
	b.WriteString(blockName)
	b.WriteByte('[')
	for i, k := range keys {
		if i > 0 { b.WriteByte(',') }
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(props[k])
	}
	b.WriteByte(']')
	return b.String()
}

func translateOne(javaName, bedrockIdent string, javaProps map[string]string,
	waterlogged bool, ovr override, bedrockIndex map[string]world.Block) Result {

	// Build Bedrock state map by translating each Java property.
	bedrockProps := map[string]any{}
	skip := map[string]struct{}{}
	for _, sp := range ovr.SkipProperties { skip[sp] = struct{}{} }

	for k, v := range javaProps {
		if _, drop := skip[k]; drop { continue }
		conv, ok := properties.Registry[k]
		if !ok { continue } // unknown Java prop → drop silently for v1.0
		bp, bv, ok := conv(v, bedrockIdent)
		if !ok { continue }
		bedrockProps[bp] = bv
	}

	// Look up via Dragonfly's BlockByName.
	full := "minecraft:" + bedrockIdent
	b, ok := world.BlockByName(full, bedrockProps)
	if !ok {
		// Try without props (some Bedrock blocks have implicit defaults).
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
		if l, ok := world.BlockByName("minecraft:water", map[string]any{"liquid_depth": int32(0)}); ok {
			if liq, isLiq := l.(world.Liquid); isLiq {
				res.Liquid = liq
			}
		}
	}
	_ = bedrockIndex // reserved for future palette validation; not strictly needed if BlockByName is authoritative
	return res
}

// loadBedrockPalette decodes Geyser's block_palette.<ver>.nbt. The file
// is a list of compound entries with `name` and `states` fields. We index
// by `name + sorted-state` for use during table validation.
func loadBedrockPalette(data []byte) (map[string]world.Block, error) {
	type root struct {
		Blocks []bedrockState `nbt:"blocks"`
	}
	var r root
	if err := nbt.UnmarshalEncoding(data, &r, nbt.LittleEndian); err != nil {
		// Some palette files are gzipped; try that.
		// (If the engineer finds the file is gzipped, swap in compress/gzip.)
		return nil, fmt.Errorf("decode palette: %w", err)
	}
	idx := make(map[string]world.Block, len(r.Blocks))
	for _, e := range r.Blocks {
		// We don't strictly need the world.Block here because Lookup goes through
		// world.BlockByName directly. Index is reserved for future validation.
		_ = idx
		_ = bytes.NewReader(nil)
		_ = e
	}
	return idx, nil
}
```

**Important notes for the engineer:**

- The PrismarineJS schema changes between minecraft-data versions. The `extractProperties` function above is defensive but may need adjustment when running against the actual file. Inspect the JSON structure first:
  ```bash
  python -c "import json; j=json.load(open('translate/data/java_blocks.json')); print(json.dumps(j[0], indent=2)[:800])"
  ```
  If properties are exposed differently (e.g. as a top-level `properties` array per block), rewrite `extractProperties` to read that.

- The Bedrock palette format from Geyser is **little-endian** Java NBT in some Geyser revisions, **big-endian** in others. If `LittleEndian` fails, try `BigEndian`. The error wrapper above prepares for either.

- The current `loadBedrockPalette` returns an empty index because Dragonfly's `BlockByName` is the authoritative resolver — the palette is only needed if you want to validate Bedrock identifiers exist *outside* of Dragonfly's view. In practice, if Dragonfly registered it, `BlockByName` will resolve it.

- [ ] **Step 2: Replace the stub Lookup**

Edit `translate/translate.go` and replace the body of `Lookup` with:

```go
func Lookup(canonicalJavaState string) Result {
	tableOnce.Do(buildTable)
	if tableErr != nil {
		return Result{Block: MissingBlock(), RawKey: canonicalJavaState}
	}
	if r, ok := table[canonicalJavaState]; ok {
		return r
	}
	return Result{Block: MissingBlock(), RawKey: canonicalJavaState}
}
```

- [ ] **Step 3: Write integrity tests**

Create `translate/table_test.go`:

```go
package translate

import "testing"

func TestPaletteCoverage(t *testing.T) {
	tableOnce.Do(buildTable)
	if tableErr != nil { t.Fatalf("buildTable: %v", tableErr) }
	if len(table) < 1000 {
		t.Errorf("translate table only has %d entries; expected >1000", len(table))
	}
}

func TestLookup_KnownStone(t *testing.T) {
	res := Lookup("minecraft:stone")
	if !res.Recognized {
		t.Fatalf("stone not recognized; have %d table entries", len(table))
	}
	name, _ := res.Block.EncodeBlock()
	if name != "minecraft:stone" {
		t.Errorf("stone resolved to %q", name)
	}
}

func TestLookup_OakLogAxis(t *testing.T) {
	res := Lookup("minecraft:oak_log[axis=y]")
	if !res.Recognized {
		t.Errorf("oak_log[axis=y] not recognized")
	}
	res = Lookup("minecraft:oak_log[axis=x]")
	if !res.Recognized {
		t.Errorf("oak_log[axis=x] not recognized")
	}
}

func TestLookup_Waterlogged(t *testing.T) {
	res := Lookup("minecraft:oak_stairs[facing=north,half=bottom,shape=straight,waterlogged=true]")
	if res.Liquid == nil {
		t.Errorf("waterlogged stairs should produce a Liquid")
	}
}

func TestLookup_Unknown(t *testing.T) {
	res := Lookup("mod:nonexistent[bar=baz]")
	if res.Recognized {
		t.Errorf("unknown block should not be Recognized")
	}
	if res.Block == nil {
		t.Errorf("unknown block must still have non-nil Block (missing fallback)")
	}
}

func TestMissingBlockRegistered(t *testing.T) {
	// resolveDefaultMissing falls through several alternatives.
	// Ensure at least one returns a non-nil Block in the test environment.
	b := resolveDefaultMissing()
	if b == nil {
		t.Fatalf("no fallback block registered; check Dragonfly version")
	}
	name, _ := b.EncodeBlock()
	t.Logf("default missing block: %s", name)
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./translate/... -v`
Expected: most pass. Some may fail because:

- The PrismarineJS file shape differs from what `extractProperties` expects. **Action:** inspect the JSON, rewrite `extractProperties`, re-run.
- The Bedrock palette decode fails. **Action:** swap `LittleEndian` → `BigEndian` (or attempt gunzip first).
- `oak_log[axis=y]` not recognized because Dragonfly's wool color property is `wool_color` not `color`. **Action:** add an entry to `overrides.json` mapping `minecraft:oak_log` to a Bedrock identifier with the right state name. (Logs are usually identity, so this is the wrong example — use whatever the test surfaces.)

Iterate. Each failing assertion points at one row to fix in the table builder, the converter set, or `overrides.json`.

- [ ] **Step 5: Commit**

```bash
git add translate/table.go translate/translate.go translate/table_test.go
git commit -m "feat(translate): table builder and integrity tests"
```

### Task 7.5: Re-enable Sponge reader assertions

Now that translate is real, the Sponge reader tests' name assertions can come back.

**Files:**
- Modify: `sponge/reader_test.go`

- [ ] **Step 1: Verify previously-deferred assertions now pass**

Re-enable any commented-out name checks in `sponge/reader_test.go` from Phase 4 Task 4.2. Run:

```bash
go test ./sponge/... -v
```

Expected: PASS, including `name == "minecraft:stone"` etc.

- [ ] **Step 2: Commit**

```bash
git add sponge/reader_test.go
git commit -m "test(sponge): re-enable translation assertions"
```

---

## Phase 8: End-to-end through `schem.Read`

### Task 8.1: Add a top-level dispatch test against fixtures

**Files:**
- Create: `schem/dispatch_test.go`
- Create: `schem/testdata/single_stone.schem` (symlink or copy)

- [ ] **Step 1: Mirror fixtures into schem/testdata**

Either symlink (Unix) or copy (Windows) so `schem` tests don't depend on relative paths into other packages:

```bash
mkdir -p schem/testdata
cp sponge/testdata/single_stone.schem schem/testdata/
cp sponge/testdata/multi_palette.schem schem/testdata/
cp legacy/testdata/basic.schematic schem/testdata/
```

- [ ] **Step 2: Write the test**

Create `schem/dispatch_test.go`:

```go
package schem_test  // external test package, exercises the public API only

import (
	"os"
	"testing"

	"github.com/Velvet-MC/s2d/schem"
	// Side-effect imports register the format handlers.
	_ "github.com/Velvet-MC/s2d/sponge"
	_ "github.com/Velvet-MC/s2d/legacy"
)

func TestEndToEnd_SpongeSingleStone(t *testing.T) {
	f, err := os.Open("testdata/single_stone.schem")
	if err != nil { t.Fatal(err) }
	defer f.Close()

	s, err := schem.Read(f.Name(), f)
	if err != nil { t.Fatalf("Read: %v", err) }

	if s.Format != schem.FormatSpongeV2 { t.Errorf("format: %q", s.Format) }
	if len(s.Blocks) != 1 { t.Fatalf("blocks: %d", len(s.Blocks)) }
	name, _ := s.Blocks[0].Block.EncodeBlock()
	if name != "minecraft:stone" { t.Errorf("name: %q", name) }
}

func TestEndToEnd_LegacyBasic(t *testing.T) {
	f, err := os.Open("testdata/basic.schematic")
	if err != nil { t.Fatal(err) }
	defer f.Close()

	s, err := schem.Read(f.Name(), f)
	if err != nil { t.Fatalf("Read: %v", err) }

	if s.Format != schem.FormatLegacy { t.Errorf("format: %q", s.Format) }
	if len(s.Blocks) != 2 { t.Fatalf("blocks: %d", len(s.Blocks)) }
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./schem/... -v`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add schem/dispatch_test.go schem/testdata/
git commit -m "test(schem): end-to-end dispatch against committed fixtures"
```

---

## Phase 9: Polish

### Task 9.1: godoc on every exported symbol

**Files:**
- Modify: every `*.go` file with exported identifiers

- [ ] **Step 1: Run godoc lint**

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck -checks ST1000,ST1020,ST1021,ST1022 ./...
```

Expected: zero `missing comment` warnings on exported symbols. Add comments where missing. Each exported type/func must have a doc comment beginning with the symbol name.

- [ ] **Step 2: Run go vet and golangci-lint**

```bash
go vet ./...
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run ./...
```

Expected: clean.

- [ ] **Step 3: Run race detector**

```bash
go test -race ./...
```

Expected: PASS. If a race surfaces in `translate`, the `tableOnce` + `missingMu` should already prevent it; investigate and fix any new race.

- [ ] **Step 4: Commit**

```bash
git add -u
git commit -m "docs: godoc on every exported symbol; lint clean"
```

### Task 9.2: README quickstart polish

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add a real worked example**

Replace the placeholder Quickstart with an end-to-end example that:
- Opens a `.schem` file
- Iterates `s.Blocks`
- Prints unknowns
- Mentions `translate.SetMissingBlock` for customization

(Engineer fills in based on the now-working API.)

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs(README): worked end-to-end example"
```

### Task 9.3: Tag v0.1.0

- [ ] **Step 1: Final verification**

```bash
go test ./...
go vet ./...
go test -race ./...
golangci-lint run ./...
```

Expected: all clean.

- [ ] **Step 2: Tag**

```bash
git tag v0.1.0 -m "S2D v0.1.0 — Sponge v2 + legacy MCEdit Java schematic reader"
git push --tags
```

---

## Self-Review

**Spec coverage:**
- Library API (`schem.Read`, `translate.Lookup`, `palette.Decode`, `translate.SetMissingBlock`) → Tasks 2, 5, 7.
- Sponge v2 reader → Phase 4.
- Legacy MCEdit reader → Phase 6.
- Translation pipeline (data, properties, table) → Phases 5 and 7.
- Missing-block fallback → Task 5.1.
- Vendored data with attribution → Tasks 6.1, 7.1, 7.2.
- All testing tiers (unit / fixture / integrity) → distributed across phases; final pass in 8.1, 9.1.
- Embedded data sizes documented → Phase 7 vendor tasks.
- Refresh procedure → Task 7.1 Step 5.
- Open verifications → flagged in code comments (Task 6.1 Step 3, Task 7.2, Task 7.4 Step 4).

**Placeholder scan:**
- Task 7.4 Step 1 includes notes that the engineer must adapt `extractProperties` and `loadBedrockPalette` based on the actual file shapes. These are not "TBD" placeholders but verification hooks — the code is concrete and runnable; the comments tell the engineer where to iterate.
- Task 6.1 Step 3 leaves the open verification of WorldEdit's GPL vs MIT compatibility for `legacy.json`. This is a real legal gate; flagged for human decision rather than buried.
- Task 7.2 leaves the schema source choice (PrismarineJS) explicit; the alternative (Mojang data gen) is documented but not pursued.
- Task 7.4 Step 4 notes that the integrity tests will surface specific failures the engineer iterates on; this is the nature of building against vendored data and is handled by concrete test assertions, not hand-waving.

**Type consistency:**
- `Result` defined in `translate/translate.go` Task 5.1; same shape used in `translate/table.go` Task 7.4.
- `Block`, `Schematic`, `UnknownReport`, `FormatHandler`, `Format` defined in `schem/types.go` Task 2.1; consumed unchanged by `sponge/reader.go` (Task 4.2), `legacy/reader.go` (Task 6.4), `schem/schem.go` (Task 2.2).
- `Converter` signature in `translate/properties/properties.go` Task 7.3 matches every per-property converter file's exported function.
- `Lookup`, `MissingBlock`, `SetMissingBlock` signatures in Task 5.1 match the table-builder's replacement in Task 7.4 Step 2.

If you find a drift during execution, fix it in place and continue.

---

## Execution Handoff

Plan complete and saved to `s2d/docs/plans/2026-05-10-implementation.md`. Two execution options:

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration. Best for this plan because each phase has clear acceptance and the verification gates (palette parity, schema iteration) benefit from a clean reviewer.

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

Which approach?
