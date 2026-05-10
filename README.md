# S2D

**Schematic-to-Dragonfly** — read Java Minecraft schematics on Dragonfly Bedrock servers.

S2D parses Sponge v2 (`.schem`) and legacy MCEdit (`.schematic`) Java schematic files
and translates every cell to a Dragonfly `world.Block`, including waterlogged variants
(returned as the base block plus a `world.Liquid`). Blocks with no Bedrock equivalent
fall back to a configurable visible marker (default magenta wool) and are reported
back to the consumer in a structured `UnknownReport`.

The library is intentionally narrow: it does not register commands, store files, manage
clipboards, or write to a `world.Tx`. Consumers wire the parsed result into whatever
they're building (a WorldEdit-style plugin, a world converter, a structure-import tool,
a proxy that needs to display Java content on Bedrock, etc.).

## Install

```sh
go get github.com/Clxser/S2D
```

## Quickstart

```go
package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/df-mc/dragonfly/server/block" // register vanilla Bedrock blocks
	_ "github.com/Clxser/S2D/legacy"            // register .schematic handler
	_ "github.com/Clxser/S2D/sponge"            // register .schem handler

	"github.com/Clxser/S2D/schem"
)

func main() {
	f, err := os.Open("castle.schem")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	s, err := schem.Read(f.Name(), f)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("loaded %s: %dx%dx%d (%d blocks)\n",
		s.Format, s.Width, s.Height, s.Length, len(s.Blocks))

	for _, b := range s.Blocks {
		// place b.Block at b.Pos in your world / clipboard / converter.
		// if b.Liquid != nil, also place the liquid in the same cell
		// (the Java source was waterlogged).
		_ = b
	}

	if s.Unknowns.Total > 0 {
		fmt.Printf("warning: %d cells could not be translated\n", s.Unknowns.Total)
		for key, count := range s.Unknowns.Counts {
			fmt.Printf("  - %s x%d\n", key, count)
		}
	}
}
```

The blank imports of `legacy` and `sponge` register their format handlers with the
dispatcher via package `init()`. The blank import of `dragonfly/server/block`
registers Dragonfly's vanilla block types — required for the translation table to
resolve names through `world.BlockByName`.

## Supported formats (v1.0)

| Extension | Format | Notes |
|---|---|---|
| `.schem` | Sponge Schematic v2 | Java WorldEdit ≥ 1.13 |
| `.schematic` | Legacy MCEdit | Pre-1.13, numeric IDs |

Roadmap (v1.1+): Sponge v3, Litematica `.litematic`, vanilla structure NBT `.nbt`.

## Customizing the missing-block marker

When a Java block has no Bedrock equivalent, S2D fills the cell with a visible
marker (default `minecraft:wool[color=magenta]`) and tallies it in `UnknownReport`.
Servers that want a different marker — for example a custom-registered
`info_update` block on a fork that includes one — can swap it via:

```go
import "github.com/Clxser/S2D/translate"

func init() {
	if b, ok := world.BlockByName("minecraft:info_update", nil); ok {
		translate.SetMissingBlock(b)
	}
}
```

`SetMissingBlock` should be called once at startup, before the first call to
`schem.Read`. After the first read the translation table is built and cached;
later calls to `SetMissingBlock` only affect the per-call fallback for keys
that aren't in the cached table.

## Architecture

| Package | Responsibility |
|---|---|
| `schem` | Public types and format dispatcher. `schem.Read(filename, r) (*Schematic, error)` is the entry point. |
| `palette` | Java state string parser (`minecraft:oak_log[axis=y]` → `JavaState{Namespace, Name, Props}`). No Dragonfly dep. |
| `sponge` | Sponge v2 NBT reader. Registers itself with `schem`. |
| `legacy` | MCEdit `.schematic` reader + vendored numeric-ID table. Registers itself with `schem`. |
| `translate` | Java state string → Dragonfly `world.Block`. Builds a ~25k-entry table at first `Lookup` from vendored Bedrock palette + Java schema. |
| `translate/properties` | Per-property Java↔Bedrock value converters (axis ↔ pillar_axis, facing ↔ direction, half ↔ upside_down_bit, etc.). |

## Vendored data

The translation table is built from data files vendored under
`translate/data/` and `legacy/data/`:

- `translate/data/bedrock_palette.<version>.nbt` — vendored from Geyser (MIT)
- `translate/data/java_blocks.json` — vendored from PrismarineJS minecraft-data (MIT)
- `translate/data/overrides.json` — hand-maintained Java→Bedrock identifier overrides
- `legacy/data/legacy.json` — vendored from EngineHub/WorldEdit

See `translate/data/ATTRIBUTION.md`, `translate/data/REGEN.md`, and
`legacy/data/ATTRIBUTION.md` for upstream sources, pinned commits, and refresh
procedures.

## License

MIT. See `LICENSE`.
