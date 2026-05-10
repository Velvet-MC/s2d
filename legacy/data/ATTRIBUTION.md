# Legacy schematic ID table — Attribution

`legacy.json` is vendored from [EngineHub/WorldEdit](https://github.com/EngineHub/WorldEdit).

- Upstream path: `worldedit-core/src/main/resources/com/sk89q/worldedit/world/registry/legacy.json`
- Pinned commit: `0ff4a5ffd946c793542b70d377db7c260fae9b56`
- Vendored: 2026-05-10
- Upstream license: GPL-3.0+ (see `LICENSE` in this directory)

## License nuance

WorldEdit is GPL-3.0+. S2D is MIT. Vendoring a single data file from a GPL project
into an MIT library has nuance:

- `legacy.json` is a factual mapping table (numeric block IDs to Java state strings)
  representing publicly known Minecraft data. Courts have generally held that pure
  data tables are not creative works subject to copyright (cf. *Feist v. Rural*).
- The conservative path is to replace `legacy.json` with an equivalent table sourced
  from a permissively-licensed project, e.g. `github.com/PrismarineJS/minecraft-data`
  (`data/pc/1.12.2/blocks.json` under MIT).
- **Open verification:** before tagging v0.1.0 publicly, decide which interpretation
  to ship under. If conservative, replace `legacy.json` with a PrismarineJS-derived
  build and update this file.

## Refresh

```sh
curl -L -o legacy.json \
  https://raw.githubusercontent.com/EngineHub/WorldEdit/<branch>/worldedit-core/src/main/resources/com/sk89q/worldedit/world/registry/legacy.json
```

Then update the pinned commit above.
