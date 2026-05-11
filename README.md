# S2D

**Schematic to Dragonfly**, with Bedrock-friendly block translation.

S2D reads Java Minecraft schematic files and turns them into Dragonfly-ready blocks for Bedrock servers. It is designed for projects that need reliable schematic imports without tying the conversion layer to a command system or a specific server plugin.

## Highlights

- Reads Sponge `.schem` files, including modern v2 and v3 layouts.
- Reads legacy MCEdit `.schematic` files.
- Translates Java block states into Bedrock/Dragonfly block states.
- Preserves neutral `minecraft:` identifiers and state data for blocks Dragonfly does not model directly.
- Streams schematics when you do not want to materialize every block at once.
- Reports unknown states so conversion gaps are easy to audit.

## Usage

```go
package main

import (
	"os"

	"github.com/Velvet-MC/s2d/schem"
	_ "github.com/Velvet-MC/s2d/sponge"
)

func load(path string) (*schem.Schematic, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return schem.Read(path, f)
}
```

## Packages

- `schem` provides the public read and scan APIs.
- `sponge` reads Sponge schematic files.
- `legacy` reads classic MCEdit schematics.
- `translate` owns Java-to-Bedrock block conversion.
- `palette` parses Java block-state palette keys.

## License

MIT