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
	h, rr, err := handlerFor(filename, r)
	if err != nil {
		return nil, err
	}
	if h.Read == nil {
		var blocks []Block
		info, err := h.Scan(rr, nil, func(b Block) error {
			blocks = append(blocks, b)
			return nil
		})
		if err != nil {
			return nil, err
		}
		return &Schematic{
			Format:   info.Format,
			Width:    info.Width,
			Height:   info.Height,
			Length:   info.Length,
			Offset:   info.Offset,
			Blocks:   blocks,
			Unknowns: info.Unknowns,
		}, nil
	}
	return h.Read(rr)
}

// Scan parses a schematic from r and calls yield for each translated block
// without materialising the full Schematic.Blocks slice.
func Scan(filename string, r io.Reader, yield BlockHandler) (ScanInfo, error) {
	return ScanWithInfo(filename, r, nil, yield)
}

// ScanWithInfo parses a schematic from r, calls onInfo once after dimensions
// are known, then calls yield for each translated block.
func ScanWithInfo(filename string, r io.Reader, onInfo InfoHandler, yield BlockHandler) (ScanInfo, error) {
	if yield == nil {
		return ScanInfo{}, fmt.Errorf("schem: nil block handler")
	}
	h, rr, err := handlerFor(filename, r)
	if err != nil {
		return ScanInfo{}, err
	}
	if h.Scan != nil {
		return h.Scan(rr, onInfo, yield)
	}
	s, err := h.Read(rr)
	if err != nil {
		return ScanInfo{}, err
	}
	info := scanInfoFromSchematic(s)
	if onInfo != nil {
		if err := onInfo(info); err != nil {
			return info, err
		}
	}
	for _, b := range s.Blocks {
		if err := yield(b); err != nil {
			return info, err
		}
	}
	return info, nil
}

func handlerFor(filename string, r io.Reader) (FormatHandler, io.Reader, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		for _, h := range handlers {
			if slices.Contains(h.Extensions, ext) {
				return h, r, nil
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
			return FormatHandler{}, nil, fmt.Errorf("schem: read: %w", err)
		}
		// A peek failure (e.g. not gzip-compressed) means no signature matches.
		if peek, err := peekDecompressed(buf, 64); err == nil {
			for _, h := range handlers {
				if h.Signature != nil && h.Signature(peek) {
					return h, bytes.NewReader(buf), nil
				}
			}
		}
	}
	return FormatHandler{}, nil, fmt.Errorf("schem: unsupported schematic format: %s", filename)
}

func scanInfoFromSchematic(s *Schematic) ScanInfo {
	return ScanInfo{
		Format:   s.Format,
		Width:    s.Width,
		Height:   s.Height,
		Length:   s.Length,
		Offset:   s.Offset,
		Unknowns: s.Unknowns,
	}
}

// peekDecompressed gunzips up to n bytes from buf for signature inspection.
func peekDecompressed(buf []byte, n int) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()
	out := make([]byte, n)
	read, err := io.ReadFull(gz, out)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return out[:read], nil
}
