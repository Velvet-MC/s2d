package sponge

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestReadVarint_Cases(t *testing.T) {
	cases := []struct {
		name  string
		bytes []byte
		want  uint32
	}{
		{"zero", []byte{0x00}, 0},
		{"one", []byte{0x01}, 1},
		{"127", []byte{0x7F}, 127},
		{"128", []byte{0x80, 0x01}, 128},
		{"300", []byte{0xAC, 0x02}, 300},
		{"16383", []byte{0xFF, 0x7F}, 16383},
		{"16384", []byte{0x80, 0x80, 0x01}, 16384},
		{"max32", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x0F}, 0xFFFFFFFF},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, n, err := readVarint(bytes.NewReader(c.bytes))
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if uint32(got) != c.want {
				t.Errorf("got %d want %d", got, c.want)
			}
			if n != len(c.bytes) {
				t.Errorf("consumed %d want %d", n, len(c.bytes))
			}
		})
	}
}

func TestReadVarint_EOF(t *testing.T) {
	// Continuation bit set but no next byte.
	_, _, err := readVarint(bytes.NewReader([]byte{0x80}))
	if !errors.Is(err, io.ErrUnexpectedEOF) && err == nil {
		t.Fatalf("expected EOF/UnexpectedEOF, got %v", err)
	}
}

func TestReadVarint_Overflow(t *testing.T) {
	// 6 continuation bytes — varint can't be longer than 5 for uint32.
	_, _, err := readVarint(bytes.NewReader([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01}))
	if err == nil {
		t.Fatalf("expected overflow error")
	}
}
