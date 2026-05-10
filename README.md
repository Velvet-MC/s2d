# S2D

**Schematic-to-Dragonfly** — read Java Minecraft schematics on Dragonfly Bedrock servers.

## Install

```sh
go get github.com/Clxser/S2D
```

## Quickstart

```go
package main

import (
	"log"
	"os"

	"github.com/Clxser/S2D/schem"
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
