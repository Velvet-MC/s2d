package legacy

import "testing"

func TestLookup_KnownIDs(t *testing.T) {
	cases := []struct {
		id, data int
		want     string
		wantOK   bool
	}{
		{1, 0, "minecraft:stone", true},  // basic
		{0, 0, "minecraft:air", true},    // air is in the table
		{9999, 0, "", false},             // nonexistent id
	}
	for _, c := range cases {
		label := itoa(c.id) + ":" + itoa(c.data)
		t.Run(label, func(t *testing.T) {
			got, ok := Lookup(c.id, c.data)
			if ok != c.wantOK {
				t.Fatalf("ok=%v wantOK=%v", ok, c.wantOK)
			}
			if !c.wantOK {
				return
			}
			if got != c.want {
				t.Errorf("got %q want %q", got, c.want)
			}
		})
	}
}

func TestLookup_OakLog(t *testing.T) {
	// id=17 is oak log. Different orientations are distinct (id, data) pairs.
	got17_0, ok := Lookup(17, 0)
	if !ok {
		t.Fatalf("17:0 not found")
	}
	t.Logf("17:0 = %q", got17_0)
	if got17_0 == "" {
		t.Errorf("17:0 empty")
	}

	got17_4, ok := Lookup(17, 4)
	if !ok {
		t.Fatalf("17:4 not found")
	}
	t.Logf("17:4 = %q", got17_4)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
