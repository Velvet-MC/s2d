package sponge

import (
	"bytes"
	"compress/gzip"
	"os"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func writeFixture(t *testing.T, path string, root map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	enc := nbt.NewEncoderWithEncoding(&buf, nbt.BigEndian)
	if err := enc.Encode(root); err != nil {
		t.Fatal(err)
	}
	var gz bytes.Buffer
	gw := gzip.NewWriter(&gz)
	if _, err := gw.Write(buf.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, gz.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeVarintBuf(buf *bytes.Buffer, v uint32) {
	for {
		if v < 0x80 {
			buf.WriteByte(byte(v))
			return
		}
		buf.WriteByte(byte(v&0x7F | 0x80))
		v >>= 7
	}
}
