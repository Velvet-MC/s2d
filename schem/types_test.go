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
