// Package legacy reads MCEdit-format `.schematic` files and maps the
// pre-1.13 numeric (block ID, data value) tuples to Java-edition state
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

// legacyTable mirrors the nested shape: {"blocks": {...}, "items": {...}}.
// We only use the blocks section for schematic decoding.
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
		tableErr = fmt.Errorf("legacy: parse table: %w", err)
		return
	}
	if len(t.Blocks) == 0 {
		tableErr = fmt.Errorf("legacy: empty blocks table")
		return
	}
	table = t.Blocks
}

// Lookup returns the Java state string for a pre-1.13 (id, data) tuple.
// The table includes id=0 (air); callers don't need a special case.
func Lookup(id, data int) (string, bool) {
	tableOnce.Do(loadTable)
	if tableErr != nil {
		return "", false
	}
	if v, ok := table[fmt.Sprintf("%d:%d", id, data)]; ok {
		return v, true
	}
	// Some entries omit the data suffix when only one variant exists.
	if data == 0 {
		if v, ok := table[fmt.Sprintf("%d", id)]; ok {
			return v, true
		}
	}
	return "", false
}
