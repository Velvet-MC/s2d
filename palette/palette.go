// Package palette parses Java edition block-state strings of the form
// "namespace:name[k1=v1,k2=v2,...]" used in Sponge schematic palettes.
//
// The package has no Dragonfly dependency. It exists so consumers (and
// the translate package) can inspect or normalize palette entries
// without invoking the full translation pipeline.
package palette

import (
	"fmt"
	"sort"
	"strings"
)

// JavaState is a parsed Java block-state string.
type JavaState struct {
	Namespace string
	Name      string
	Props     map[string]string
}

// Decode parses s into a JavaState. The namespace defaults to "minecraft"
// when absent. Properties are optional; if present they are enclosed in
// square brackets and comma-separated.
func Decode(s string) (JavaState, error) {
	if s == "" {
		return JavaState{}, fmt.Errorf("palette: empty state string")
	}
	js := JavaState{Namespace: "minecraft"}
	body := s
	if i := strings.IndexByte(s, '['); i >= 0 {
		if !strings.HasSuffix(s, "]") {
			return JavaState{}, fmt.Errorf("palette: unterminated bracket in %q", s)
		}
		body = s[:i]
		propBody := s[i+1 : len(s)-1]
		if propBody != "" {
			js.Props = make(map[string]string)
			for _, kv := range strings.Split(propBody, ",") {
				eq := strings.IndexByte(kv, '=')
				if eq < 0 {
					return JavaState{}, fmt.Errorf("palette: malformed prop %q in %q", kv, s)
				}
				k := strings.TrimSpace(kv[:eq])
				v := strings.TrimSpace(kv[eq+1:])
				if k == "" || v == "" {
					return JavaState{}, fmt.Errorf("palette: empty key or value in %q", s)
				}
				js.Props[k] = v
			}
		}
	}
	if colon := strings.IndexByte(body, ':'); colon >= 0 {
		js.Namespace = body[:colon]
		js.Name = body[colon+1:]
	} else {
		js.Name = body
	}
	if js.Namespace == "" || js.Name == "" {
		return JavaState{}, fmt.Errorf("palette: empty namespace or block name in %q", s)
	}
	return js, nil
}

// Canonical re-emits the state with properties sorted by key, suitable
// for use as a stable map key.
func (j JavaState) Canonical() string {
	var b strings.Builder
	b.WriteString(j.Namespace)
	b.WriteByte(':')
	b.WriteString(j.Name)
	if len(j.Props) > 0 {
		keys := make([]string, 0, len(j.Props))
		for k := range j.Props {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteByte('[')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(k)
			b.WriteByte('=')
			b.WriteString(j.Props[k])
		}
		b.WriteByte(']')
	}
	return b.String()
}
