# Refresh procedure

## When Dragonfly upgrades to a new Bedrock palette version

1. Copy the new palette from Geyser:

   ```sh
   cp <geyser-clone>/core/src/main/resources/bedrock/block_palette.<NEW_VERSION>.nbt \
      translate/data/bedrock_palette.<NEW_VERSION>.nbt
   ```

   Or, network fetch:

   ```sh
   curl -L -o translate/data/bedrock_palette.<NEW_VERSION>.nbt \
     https://raw.githubusercontent.com/GeyserMC/Geyser/master/core/src/main/resources/bedrock/block_palette.<NEW_VERSION>.nbt
   ```

2. Update the embedded palette filename in `translate/table.go` (the
   `//go:embed data/bedrock_palette.X.nbt` directive).

3. Update the pinned commit in `ATTRIBUTION.md`.

4. Run integrity tests:

   ```sh
   go test ./translate -run TestPaletteCoverage
   go test ./translate -run TestPaletteMatchesDragonfly
   ```

5. For any newly-flagged divergent blocks, add entries to `overrides.json`.

6. Tag a new S2D release.

## When Java schema (java_blocks.json) needs refreshing

```sh
curl -L -o translate/data/java_blocks.json \
  https://raw.githubusercontent.com/PrismarineJS/minecraft-data/master/data/pc/<MC_VERSION>/blocks.json
```

Update the pinned commit + MC version in `ATTRIBUTION.md`. Re-run integrity tests.
