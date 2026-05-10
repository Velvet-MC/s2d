// Package s2d (Schematic-to-Dragonfly) is the umbrella module for reading
// Java-edition Minecraft schematic files and translating them into Dragonfly
// world.Block values for use on Bedrock-edition servers.
//
// Consumers do not import this package directly. The public entry point is
// github.com/Clxser/S2D/schem, which auto-detects the schematic format from
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
