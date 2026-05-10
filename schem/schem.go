package schem

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

var handlers []FormatHandler

// Register adds a FormatHandler to the dispatch table. Format-reader
// packages call this from their init() function.
func Register(h FormatHandler) {
	handlers = append(handlers, h)
}

// Read parses a schematic from r. The filename's extension selects the
// format reader; if no extension matches, Read peeks the gzipped header
// for an NBT signature.
func Read(filename string, r io.Reader) (*Schematic, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		for _, h := range handlers {
			if slices.Contains(h.Extensions, ext) {
				return h.Read(r)
			}
		}
	}
	// Fallback: read the full body, attempt signature dispatch.
	// Only bother if at least one handler advertises a Signature check.
	hasSig := false
	for _, h := range handlers {
		if h.Signature != nil {
			hasSig = true
			break
		}
	}
	if hasSig {
		buf, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("schem: read: %w", err)
		}
		// A peek failure (e.g. not gzip-compressed) means no signature matches.
		if peek, err := peekDecompressed(buf, 64); err == nil {
			for _, h := range handlers {
				if h.Signature != nil && h.Signature(peek) {
					return h.Read(bytes.NewReader(buf))
				}
			}
		}
	}
	return nil, fmt.Errorf("schem: unsupported schematic format: %s", filename)
}

// peekDecompressed gunzips up to n bytes from buf for signature inspection.
func peekDecompressed(buf []byte, n int) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	out := make([]byte, n)
	read, err := io.ReadFull(gz, out)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return out[:read], nil
}
