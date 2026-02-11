package bencode

import (
	"fmt"
)

type Decoder struct {
	data []byte
	pos  int
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{data: data}
}

func (d *Decoder) Pos() int {
	return d.pos
}

func (d *Decoder) Decode() (any, error) {
	v, err := d.decodeValue()
	if err != nil {
		return nil, err
	}
	if d.pos != len(d.data) {
		return nil, fmt.Errorf("trailing data at %d", d.pos)
	}
	return v, nil
}

func Decode(data []byte) (any, error) {
	return NewDecoder(data).Decode()
}

// DecodeWithSpanKey decodes a top-level dict and returns raw bencoded bytes
// for the value associated with key (if present).
func DecodeWithSpanKey(data []byte, key string) (any, []byte, error) {
	d := NewDecoder(data)
	if d.peek() != 'd' {
		return nil, nil, fmt.Errorf("expected dict at root")
	}
	_ = d.readByte() // consume 'd'

	out := make(map[string]any)
	var span []byte

	for {
		if d.pos >= len(d.data) {
			return nil, nil, fmt.Errorf("unexpected end of data")
		}
		if d.peek() == 'e' {
			d.pos++
			break
		}
		kb, err := d.decodeString()
		if err != nil {
			return nil, nil, err
		}
		k := string(kb)
		start := d.pos
		v, err := d.decodeValue()
		if err != nil {
			return nil, nil, err
		}
		end := d.pos
		out[k] = v
		if k == key {
			span = d.data[start:end]
		}
	}

	if d.pos != len(d.data) {
		return nil, nil, fmt.Errorf("trailing data at %d", d.pos)
	}
	return out, span, nil
}

func (d *Decoder) decodeValue() (any, error) {
	switch d.peek() {
	case 'i':
		return d.decodeInt()
	case 'l':
		return d.decodeList()
	case 'd':
		return d.decodeDict()
	default:
		if d.peek() >= '0' && d.peek() <= '9' {
			return d.decodeString()
		}
		return nil, fmt.Errorf("invalid token at %d", d.pos)
	}
}

func (d *Decoder) decodeInt() (int64, error) {
	if d.readByte() != 'i' {
		return 0, fmt.Errorf("expected int at %d", d.pos)
	}
	neg := false
	if d.peek() == '-' {
		neg = true
		d.pos++
	}
	if d.pos >= len(d.data) {
		return 0, fmt.Errorf("unexpected end of int at %d", d.pos)
	}
	if d.peek() == 'e' {
		return 0, fmt.Errorf("empty int at %d", d.pos)
	}
	var n int64
	for {
		if d.pos >= len(d.data) {
			return 0, fmt.Errorf("unexpected end of int at %d", d.pos)
		}
		c := d.peek()
		if c == 'e' {
			d.pos++
			break
		}
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid int char at %d", d.pos)
		}
		n = n*10 + int64(c-'0')
		d.pos++
	}
	if neg {
		n = -n
	}
	return n, nil
}

func (d *Decoder) decodeString() ([]byte, error) {
	if d.pos >= len(d.data) {
		return nil, fmt.Errorf("unexpected end of string at %d", d.pos)
	}
	var n int
	for {
		if d.pos >= len(d.data) {
			return nil, fmt.Errorf("unexpected end of string length at %d", d.pos)
		}
		c := d.peek()
		if c == ':' {
			d.pos++
			break
		}
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("invalid string length at %d", d.pos)
		}
		n = n*10 + int(c-'0')
		d.pos++
	}
	if d.pos+n > len(d.data) {
		return nil, fmt.Errorf("string length out of bounds at %d", d.pos)
	}
	out := d.data[d.pos : d.pos+n]
	d.pos += n
	return out, nil
}

func (d *Decoder) decodeList() ([]any, error) {
	if d.readByte() != 'l' {
		return nil, fmt.Errorf("expected list at %d", d.pos)
	}
	var out []any
	for {
		if d.pos >= len(d.data) {
			return nil, fmt.Errorf("unexpected end of list at %d", d.pos)
		}
		if d.peek() == 'e' {
			d.pos++
			break
		}
		v, err := d.decodeValue()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func (d *Decoder) decodeDict() (map[string]any, error) {
	if d.readByte() != 'd' {
		return nil, fmt.Errorf("expected dict at %d", d.pos)
	}
	out := make(map[string]any)
	for {
		if d.pos >= len(d.data) {
			return nil, fmt.Errorf("unexpected end of dict at %d", d.pos)
		}
		if d.peek() == 'e' {
			d.pos++
			break
		}
		kb, err := d.decodeString()
		if err != nil {
			return nil, err
		}
		k := string(kb)
		v, err := d.decodeValue()
		if err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, nil
}

func (d *Decoder) peek() byte {
	if d.pos >= len(d.data) {
		return 0
	}
	return d.data[d.pos]
}

func (d *Decoder) readByte() byte {
	if d.pos >= len(d.data) {
		return 0
	}
	b := d.data[d.pos]
	d.pos++
	return b
}
