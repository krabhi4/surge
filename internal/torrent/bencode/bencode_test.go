package bencode

import (
	"bytes"
	"testing"
)

func TestEncodeDecode_RoundTrip(t *testing.T) {
	src := map[string]any{
		"announce": "http://tracker",
		"info": map[string]any{
			"name":         "file.txt",
			"piece length": int64(16384),
			"length":       int64(5),
			"pieces":       []byte("12345678901234567890"),
		},
		"list": []any{
			int64(1),
			[]byte("abc"),
		},
	}

	enc, err := Encode(src)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	dec, err := Decode(enc)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	enc2, err := Encode(dec)
	if err != nil {
		t.Fatalf("re-encode failed: %v", err)
	}

	if !bytes.Equal(enc, enc2) {
		t.Fatalf("round-trip mismatch")
	}
}

func TestDecodeWithSpanKey_Info(t *testing.T) {
	info := map[string]any{
		"name":         "file.txt",
		"piece length": int64(16384),
		"length":       int64(5),
		"pieces":       []byte("12345678901234567890"),
	}
	infoBytes, err := Encode(info)
	if err != nil {
		t.Fatalf("encode info failed: %v", err)
	}

	root := map[string]any{
		"announce": "http://tracker",
		"info":     info,
	}
	rootBytes, err := Encode(root)
	if err != nil {
		t.Fatalf("encode root failed: %v", err)
	}

	_, span, err := DecodeWithSpanKey(rootBytes, "info")
	if err != nil {
		t.Fatalf("decode span failed: %v", err)
	}
	if !bytes.Equal(span, infoBytes) {
		t.Fatalf("info span mismatch")
	}
}
