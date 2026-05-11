package sponge

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type nbtTag byte

const (
	tagEnd       nbtTag = 0
	tagByte      nbtTag = 1
	tagShort     nbtTag = 2
	tagInt       nbtTag = 3
	tagLong      nbtTag = 4
	tagFloat     nbtTag = 5
	tagDouble    nbtTag = 6
	tagByteArray nbtTag = 7
	tagString    nbtTag = 8
	tagList      nbtTag = 9
	tagCompound  nbtTag = 10
	tagIntArray  nbtTag = 11
	tagLongArray nbtTag = 12
)

type rawSpongeSchematic struct {
	Version       int32
	Width         int32
	Height        int32
	Length        int32
	Offset        []int32
	Palette       map[string]int32
	BlockData     []byte
	BlockEntities []rawBlockEntity
	Blocks        rawSpongeBlocks
}

type rawSpongeBlocks struct {
	Palette       map[string]int32
	Data          []byte
	BlockEntities []rawBlockEntity
}

type rawBlockEntity struct {
	ID       string
	Pos      []int32
	Patterns any
	Base     int32
	HasBase  bool
}

type nbtReader struct {
	r *bufio.Reader
}

func decodeSpongeNBT(r io.Reader) (rawSpongeSchematic, error) {
	dec := nbtReader{r: bufio.NewReader(r)}
	t, name, err := dec.namedTag()
	if err != nil {
		return rawSpongeSchematic{}, err
	}
	if t != tagCompound {
		return rawSpongeSchematic{}, fmt.Errorf("root tag %q has type %d, want compound", name, t)
	}
	var root rawSpongeSchematic
	if err := dec.schematicRoot(&root); err != nil {
		return rawSpongeSchematic{}, err
	}
	return root, nil
}

func (d *nbtReader) schematicRoot(out *rawSpongeSchematic) error {
	for {
		t, name, err := d.namedTag()
		if err != nil {
			return err
		}
		if t == tagEnd {
			return nil
		}
		if name == "Schematic" && t == tagCompound {
			var inner rawSpongeSchematic
			if err := d.schematicCompound(&inner); err != nil {
				return err
			}
			*out = inner
			continue
		}
		if err := d.schematicField(out, t, name); err != nil {
			return err
		}
	}
}

func (d *nbtReader) schematicCompound(out *rawSpongeSchematic) error {
	for {
		t, name, err := d.namedTag()
		if err != nil {
			return err
		}
		if t == tagEnd {
			return nil
		}
		if err := d.schematicField(out, t, name); err != nil {
			return err
		}
	}
}

func (d *nbtReader) schematicField(out *rawSpongeSchematic, t nbtTag, name string) error {
	switch name {
	case "Version":
		v, err := d.numericInt32(t)
		out.Version = v
		return err
	case "Width":
		v, err := d.numericInt32(t)
		out.Width = v
		return err
	case "Height":
		v, err := d.numericInt32(t)
		out.Height = v
		return err
	case "Length":
		v, err := d.numericInt32(t)
		out.Length = v
		return err
	case "Offset":
		var v []int32
		var err error
		switch t {
		case tagIntArray:
			v, err = d.intArray()
		case tagList:
			v, err = d.intList()
		default:
			return d.skip(t)
		}
		out.Offset = v
		return err
	case "Palette":
		if t != tagCompound {
			return d.skip(t)
		}
		v, err := d.palette()
		out.Palette = v
		return err
	case "BlockData":
		var v []byte
		var err error
		switch t {
		case tagByteArray:
			v, err = d.byteArray()
		case tagList:
			v, err = d.byteList()
		default:
			return d.skip(t)
		}
		out.BlockData = v
		return err
	case "BlockEntities":
		if t != tagList {
			return d.skip(t)
		}
		v, err := d.blockEntityList()
		out.BlockEntities = v
		return err
	case "Blocks":
		if t != tagCompound {
			return d.skip(t)
		}
		return d.blocksCompound(&out.Blocks)
	default:
		return d.skip(t)
	}
}

func (d *nbtReader) blocksCompound(out *rawSpongeBlocks) error {
	for {
		t, name, err := d.namedTag()
		if err != nil {
			return err
		}
		if t == tagEnd {
			return nil
		}
		switch name {
		case "Palette":
			if t != tagCompound {
				if err := d.skip(t); err != nil {
					return err
				}
				continue
			}
			v, err := d.palette()
			if err != nil {
				return err
			}
			out.Palette = v
		case "Data":
			var v []byte
			var err error
			switch t {
			case tagByteArray:
				v, err = d.byteArray()
			case tagList:
				v, err = d.byteList()
			default:
				if err := d.skip(t); err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}
			out.Data = v
		case "BlockEntities":
			if t != tagList {
				if err := d.skip(t); err != nil {
					return err
				}
				continue
			}
			v, err := d.blockEntityList()
			if err != nil {
				return err
			}
			out.BlockEntities = v
		default:
			if err := d.skip(t); err != nil {
				return err
			}
		}
	}
}

func (d *nbtReader) palette() (map[string]int32, error) {
	out := map[string]int32{}
	for {
		t, name, err := d.namedTag()
		if err != nil {
			return nil, err
		}
		if t == tagEnd {
			return out, nil
		}
		v, err := d.numericInt32(t)
		if err != nil {
			return nil, fmt.Errorf("palette %q: %w", name, err)
		}
		out[name] = v
	}
}

func (d *nbtReader) blockEntityList() ([]rawBlockEntity, error) {
	elem, n, err := d.listHeader()
	if err != nil {
		return nil, err
	}
	if elem != tagCompound {
		for i := int32(0); i < n; i++ {
			if err := d.skip(elem); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
	out := make([]rawBlockEntity, 0, n)
	for i := int32(0); i < n; i++ {
		var be rawBlockEntity
		if err := d.blockEntityCompound(&be); err != nil {
			return nil, err
		}
		out = append(out, be)
	}
	return out, nil
}

func (d *nbtReader) blockEntityCompound(out *rawBlockEntity) error {
	for {
		t, name, err := d.namedTag()
		if err != nil {
			return err
		}
		if t == tagEnd {
			return nil
		}
		switch name {
		case "Id", "id":
			if t != tagString {
				if err := d.skip(t); err != nil {
					return err
				}
				continue
			}
			out.ID, err = d.string()
			if err != nil {
				return err
			}
		case "Pos":
			switch t {
			case tagIntArray:
				out.Pos, err = d.intArray()
			case tagList:
				out.Pos, err = d.intList()
			default:
				if err := d.skip(t); err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}
		case "Patterns":
			out.Patterns, err = d.any(t)
			if err != nil {
				return err
			}
		case "Base":
			out.Base, err = d.numericInt32(t)
			out.HasBase = err == nil
			if err != nil {
				return err
			}
		case "Data":
			if t != tagCompound {
				if err := d.skip(t); err != nil {
					return err
				}
				continue
			}
			if err := d.blockEntityCompound(out); err != nil {
				return err
			}
		default:
			if err := d.skip(t); err != nil {
				return err
			}
		}
	}
}

func (d *nbtReader) namedTag() (nbtTag, string, error) {
	b, err := d.r.ReadByte()
	if err != nil {
		return 0, "", err
	}
	t := nbtTag(b)
	if t == tagEnd {
		return t, "", nil
	}
	name, err := d.string()
	return t, name, err
}

func (d *nbtReader) listHeader() (nbtTag, int32, error) {
	b, err := d.r.ReadByte()
	if err != nil {
		return 0, 0, err
	}
	n, err := d.int32()
	if err != nil {
		return 0, 0, err
	}
	if n < 0 {
		return 0, 0, fmt.Errorf("negative list length %d", n)
	}
	return nbtTag(b), n, nil
}

func (d *nbtReader) any(t nbtTag) (any, error) {
	switch t {
	case tagByte:
		return d.r.ReadByte()
	case tagShort:
		return d.int16()
	case tagInt:
		return d.int32()
	case tagLong:
		return d.int64()
	case tagFloat:
		return d.float32()
	case tagDouble:
		return d.float64()
	case tagByteArray:
		return d.byteArray()
	case tagString:
		return d.string()
	case tagList:
		elem, n, err := d.listHeader()
		if err != nil {
			return nil, err
		}
		out := make([]any, 0, n)
		for i := int32(0); i < n; i++ {
			v, err := d.any(elem)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case tagCompound:
		out := map[string]any{}
		for {
			nested, name, err := d.namedTag()
			if err != nil {
				return nil, err
			}
			if nested == tagEnd {
				return out, nil
			}
			v, err := d.any(nested)
			if err != nil {
				return nil, err
			}
			out[name] = v
		}
	case tagIntArray:
		return d.intArray()
	case tagLongArray:
		return d.longArray()
	default:
		return nil, fmt.Errorf("unsupported tag %d", t)
	}
}

func (d *nbtReader) skip(t nbtTag) error {
	switch t {
	case tagByte:
		_, err := d.r.ReadByte()
		return err
	case tagShort:
		_, err := d.int16()
		return err
	case tagInt:
		_, err := d.int32()
		return err
	case tagLong:
		_, err := d.int64()
		return err
	case tagFloat:
		_, err := d.float32()
		return err
	case tagDouble:
		_, err := d.float64()
		return err
	case tagByteArray:
		n, err := d.int32()
		if err != nil {
			return err
		}
		if n < 0 {
			return fmt.Errorf("negative byte array length %d", n)
		}
		_, err = io.CopyN(io.Discard, d.r, int64(n))
		return err
	case tagString:
		n, err := d.uint16()
		if err != nil {
			return err
		}
		_, err = io.CopyN(io.Discard, d.r, int64(n))
		return err
	case tagList:
		elem, n, err := d.listHeader()
		if err != nil {
			return err
		}
		for i := int32(0); i < n; i++ {
			if err := d.skip(elem); err != nil {
				return err
			}
		}
		return nil
	case tagCompound:
		for {
			nested, _, err := d.namedTag()
			if err != nil {
				return err
			}
			if nested == tagEnd {
				return nil
			}
			if err := d.skip(nested); err != nil {
				return err
			}
		}
	case tagIntArray:
		n, err := d.int32()
		if err != nil {
			return err
		}
		if n < 0 {
			return fmt.Errorf("negative int array length %d", n)
		}
		_, err = io.CopyN(io.Discard, d.r, int64(n)*4)
		return err
	case tagLongArray:
		n, err := d.int32()
		if err != nil {
			return err
		}
		if n < 0 {
			return fmt.Errorf("negative long array length %d", n)
		}
		_, err = io.CopyN(io.Discard, d.r, int64(n)*8)
		return err
	default:
		return fmt.Errorf("unsupported tag %d", t)
	}
}

func (d *nbtReader) numericInt32(t nbtTag) (int32, error) {
	switch t {
	case tagByte:
		b, err := d.r.ReadByte()
		return int32(int8(b)), err
	case tagShort:
		v, err := d.int16()
		return int32(v), err
	case tagInt:
		return d.int32()
	case tagLong:
		v, err := d.int64()
		return int32(v), err
	default:
		if err := d.skip(t); err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("tag %d is not integer", t)
	}
}

func (d *nbtReader) intList() ([]int32, error) {
	elem, n, err := d.listHeader()
	if err != nil {
		return nil, err
	}
	if elem != tagInt {
		for i := int32(0); i < n; i++ {
			if err := d.skip(elem); err != nil {
				return nil, err
			}
		}
		return nil, fmt.Errorf("list element tag %d is not int", elem)
	}
	out := make([]int32, n)
	for i := range out {
		out[i], err = d.int32()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (d *nbtReader) byteList() ([]byte, error) {
	elem, n, err := d.listHeader()
	if err != nil {
		return nil, err
	}
	if elem != tagByte {
		for i := int32(0); i < n; i++ {
			if err := d.skip(elem); err != nil {
				return nil, err
			}
		}
		return nil, fmt.Errorf("list element tag %d is not byte", elem)
	}
	out := make([]byte, n)
	_, err = io.ReadFull(d.r, out)
	return out, err
}

func (d *nbtReader) byteArray() ([]byte, error) {
	n, err := d.int32()
	if err != nil {
		return nil, err
	}
	if n < 0 {
		return nil, fmt.Errorf("negative byte array length %d", n)
	}
	out := make([]byte, n)
	_, err = io.ReadFull(d.r, out)
	return out, err
}

func (d *nbtReader) intArray() ([]int32, error) {
	n, err := d.int32()
	if err != nil {
		return nil, err
	}
	if n < 0 {
		return nil, fmt.Errorf("negative int array length %d", n)
	}
	out := make([]int32, n)
	for i := range out {
		out[i], err = d.int32()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (d *nbtReader) longArray() ([]int64, error) {
	n, err := d.int32()
	if err != nil {
		return nil, err
	}
	if n < 0 {
		return nil, fmt.Errorf("negative long array length %d", n)
	}
	out := make([]int64, n)
	for i := range out {
		out[i], err = d.int64()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (d *nbtReader) string() (string, error) {
	n, err := d.uint16()
	if err != nil {
		return "", err
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(d.r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func (d *nbtReader) uint16() (uint16, error) {
	var v uint16
	err := binary.Read(d.r, binary.BigEndian, &v)
	return v, err
}

func (d *nbtReader) int16() (int16, error) {
	var v int16
	err := binary.Read(d.r, binary.BigEndian, &v)
	return v, err
}

func (d *nbtReader) int32() (int32, error) {
	var v int32
	err := binary.Read(d.r, binary.BigEndian, &v)
	return v, err
}

func (d *nbtReader) int64() (int64, error) {
	var v int64
	err := binary.Read(d.r, binary.BigEndian, &v)
	return v, err
}

func (d *nbtReader) float32() (float32, error) {
	var v float32
	err := binary.Read(d.r, binary.BigEndian, &v)
	return v, err
}

func (d *nbtReader) float64() (float64, error) {
	var v float64
	err := binary.Read(d.r, binary.BigEndian, &v)
	return v, err
}
