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

func TestDecode_Cases(t *testing.T) {
	cases := []struct {
		in        string
		ns, name  string
		props     map[string]string
		canonical string
		wantErr   bool
	}{
		{"stone", "minecraft", "stone", nil, "minecraft:stone", false},
		{"minecraft:air", "minecraft", "air", nil, "minecraft:air", false},
		{"minecraft:oak_log[axis=y]", "minecraft", "oak_log",
			map[string]string{"axis": "y"}, "minecraft:oak_log[axis=y]", false},
		{"minecraft:oak_stairs[half=top,facing=north]", "minecraft", "oak_stairs",
			map[string]string{"half": "top", "facing": "north"},
			"minecraft:oak_stairs[facing=north,half=top]", false},
		{"mod:custom[a=1,b=2,c=3]", "mod", "custom",
			map[string]string{"a": "1", "b": "2", "c": "3"},
			"mod:custom[a=1,b=2,c=3]", false},
		{"", "", "", nil, "", true},
		{":stone", "", "", nil, "", true},
		{"minecraft:", "", "", nil, "", true},
		{"minecraft:oak_log[axis=", "", "", nil, "", true},
		{"minecraft:oak_log[axis=y", "", "", nil, "", true},
		{"minecraft:oak_log[=y]", "", "", nil, "", true},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := Decode(c.in)
			if (err != nil) != c.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, c.wantErr)
			}
			if c.wantErr {
				return
			}
			if got.Namespace != c.ns || got.Name != c.name {
				t.Errorf("ns/name: got %q:%q want %q:%q", got.Namespace, got.Name, c.ns, c.name)
			}
			if len(got.Props) != len(c.props) {
				t.Errorf("props len: got %v want %v", got.Props, c.props)
			}
			for k, v := range c.props {
				if got.Props[k] != v {
					t.Errorf("props[%q]=%q want %q", k, got.Props[k], v)
				}
			}
			if g := got.Canonical(); g != c.canonical {
				t.Errorf("canonical: got %q want %q", g, c.canonical)
			}
		})
	}
}
