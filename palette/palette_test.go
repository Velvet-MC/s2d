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
