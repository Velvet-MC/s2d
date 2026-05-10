# Translation data — Attribution

## Bedrock palette (`bedrock_palette.1_26_20.nbt`)

Vendored from [GeyserMC/Geyser](https://github.com/GeyserMC/Geyser).

- Upstream path: `core/src/main/resources/bedrock/block_palette.1_26_20.nbt`
- Pinned commit: `4f676a745ea02be5ea8bfd849d2a0afce21a91cd`
- Vendored: 2026-05-10
- License: MIT (see `LICENSE` in this directory).

## Java block schema (`java_blocks.json`)

Vendored from [PrismarineJS/minecraft-data](https://github.com/PrismarineJS/minecraft-data).

- Upstream path: `data/pc/1.20.2/blocks.json`
- MC version: 1.20.2 (data version 3578)
- Pinned commit: `28038c4f3168f148e7d6c335b5a9bfcc20ff5258`
- Vendored: 2026-05-10
- License: MIT.

The file is an array of block objects. Each object includes `name`,
`displayName`, `id`, `stackSize`, `states`, and other Minecraft-specific
metadata. Phase 7.4 of the implementation plan parses this into a Java
state schema and walks the cartesian product of property values to build
the translation table.

## Overrides (`overrides.json`)

Hand-maintained by S2D. MIT (S2D).

Currently empty — entries are added when Phase 7.4's table-build integrity
tests surface Java→Bedrock identifier divergences.
