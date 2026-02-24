package zcl

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func TestDecodeEncodeUint8(t *testing.T) {
	data := []byte{0x42}
	val, n, err := DecodeValue(TypeUint8, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("consumed %d, want 1", n)
	}
	if val.(uint8) != 0x42 {
		t.Errorf("got %v, want 0x42", val)
	}

	encoded, err := EncodeValue(TypeUint8, uint8(0x42))
	if err != nil {
		t.Fatal(err)
	}
	if encoded[0] != 0x42 {
		t.Errorf("encoded %X, want 42", encoded)
	}
}

func TestDecodeEncodeUint16(t *testing.T) {
	data := []byte{0x34, 0x12} // little-endian 0x1234
	val, n, err := DecodeValue(TypeUint16, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("consumed %d, want 2", n)
	}
	if val.(uint16) != 0x1234 {
		t.Errorf("got %v, want 0x1234", val)
	}
}

func TestDecodeEncodeBool(t *testing.T) {
	val, _, err := DecodeValue(TypeBool, []byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	if val.(bool) != true {
		t.Error("expected true")
	}

	val, _, err = DecodeValue(TypeBool, []byte{0x00})
	if err != nil {
		t.Fatal(err)
	}
	if val.(bool) != false {
		t.Error("expected false")
	}
}

func TestDecodeCharStr(t *testing.T) {
	data := []byte{5, 'H', 'e', 'l', 'l', 'o'}
	val, n, err := DecodeValue(TypeCharStr, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 6 {
		t.Errorf("consumed %d, want 6", n)
	}
	if val.(string) != "Hello" {
		t.Errorf("got %q, want %q", val, "Hello")
	}
}

func TestDecodeInt16(t *testing.T) {
	// -100 = 0xFF9C in little-endian: 0x9C, 0xFF
	data := []byte{0x9C, 0xFF}
	val, _, err := DecodeValue(TypeInt16, data)
	if err != nil {
		t.Fatal(err)
	}
	if val.(int16) != -100 {
		t.Errorf("got %v, want -100", val)
	}
}

func TestDecodeNotEnoughData(t *testing.T) {
	_, _, err := DecodeValue(TypeUint32, []byte{0x01})
	if err == nil {
		t.Error("expected error for insufficient data")
	}
}

func TestEncodeCharStr(t *testing.T) {
	encoded, err := EncodeValue(TypeCharStr, "Hi")
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) != 3 || encoded[0] != 2 || encoded[1] != 'H' || encoded[2] != 'i' {
		t.Errorf("encoded %X, want [02 48 69]", encoded)
	}
}

// --- New tests ---

func TestDecodeEncodeInt8(t *testing.T) {
	data := []byte{0x80} // -128
	val, n, err := DecodeValue(TypeInt8, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("consumed %d, want 1", n)
	}
	if val.(int8) != -128 {
		t.Errorf("got %v, want -128", val)
	}

	encoded, err := EncodeValue(TypeInt8, int(-128))
	if err != nil {
		t.Fatal(err)
	}
	if encoded[0] != 0x80 {
		t.Errorf("encoded %X, want 80", encoded)
	}
}

func TestDecodeEncodeInt24(t *testing.T) {
	// -1 in int24 = 0xFFFFFF → bytes: FF FF FF
	data := []byte{0xFF, 0xFF, 0xFF}
	val, n, err := DecodeValue(TypeInt24, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("consumed %d, want 3", n)
	}
	if val.(int32) != -1 {
		t.Errorf("got %v, want -1", val)
	}

	// Positive: 100 = 0x000064
	data = []byte{0x64, 0x00, 0x00}
	val, _, err = DecodeValue(TypeInt24, data)
	if err != nil {
		t.Fatal(err)
	}
	if val.(int32) != 100 {
		t.Errorf("got %v, want 100", val)
	}

	// Encode round-trip
	encoded, err := EncodeValue(TypeInt24, int64(-1))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, []byte{0xFF, 0xFF, 0xFF}) {
		t.Errorf("encoded %X, want FFFFFF", encoded)
	}
}

func TestDecodeEncodeUint24(t *testing.T) {
	data := []byte{0x56, 0x34, 0x12} // 0x123456
	val, n, err := DecodeValue(TypeUint24, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("consumed %d, want 3", n)
	}
	if val.(uint32) != 0x123456 {
		t.Errorf("got 0x%X, want 0x123456", val)
	}

	encoded, err := EncodeValue(TypeUint24, uint64(0x123456))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, data) {
		t.Errorf("encoded %X, want %X", encoded, data)
	}
}

func TestDecodeEncodeUint32(t *testing.T) {
	data := []byte{0x78, 0x56, 0x34, 0x12}
	val, n, err := DecodeValue(TypeUint32, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Errorf("consumed %d, want 4", n)
	}
	if val.(uint32) != 0x12345678 {
		t.Errorf("got 0x%X, want 0x12345678", val)
	}

	encoded, err := EncodeValue(TypeUint32, uint64(0x12345678))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, data) {
		t.Errorf("encoded %X, want %X", encoded, data)
	}
}

func TestDecodeEncodeInt32(t *testing.T) {
	v := int32(-100000)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(v))
	val, n, err := DecodeValue(TypeInt32, buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Errorf("consumed %d, want 4", n)
	}
	if val.(int32) != -100000 {
		t.Errorf("got %v, want -100000", val)
	}

	encoded, err := EncodeValue(TypeInt32, int64(-100000))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, buf) {
		t.Errorf("encoded %X, want %X", encoded, buf)
	}
}

func TestDecodeEncodeUint40(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	val, n, err := DecodeValue(TypeUint40, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Errorf("consumed %d, want 5", n)
	}
	expected := uint64(0x01) | uint64(0x02)<<8 | uint64(0x03)<<16 | uint64(0x04)<<24 | uint64(0x05)<<32
	if val.(uint64) != expected {
		t.Errorf("got 0x%X, want 0x%X", val, expected)
	}

	encoded, err := EncodeValue(TypeUint40, expected)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, data) {
		t.Errorf("encoded %X, want %X", encoded, data)
	}
}

func TestDecodeEncodeUint48(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	val, n, err := DecodeValue(TypeUint48, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 6 {
		t.Errorf("consumed %d, want 6", n)
	}
	expected := uint64(0x01) | uint64(0x02)<<8 | uint64(0x03)<<16 | uint64(0x04)<<24 | uint64(0x05)<<32 | uint64(0x06)<<40
	if val.(uint64) != expected {
		t.Errorf("got 0x%X, want 0x%X", val, expected)
	}

	encoded, err := EncodeValue(TypeUint48, expected)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, data) {
		t.Errorf("encoded %X, want %X", encoded, data)
	}
}

func TestDecodeEncodeFloat32(t *testing.T) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(3.14))
	val, n, err := DecodeValue(TypeFloat32, buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Errorf("consumed %d, want 4", n)
	}
	if v := val.(float32); v != 3.14 {
		t.Errorf("got %v, want 3.14", v)
	}

	encoded, err := EncodeValue(TypeFloat32, float64(3.14))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, buf) {
		t.Errorf("encoded %X, want %X", encoded, buf)
	}
}

func TestDecodeEncodeFloat64(t *testing.T) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, math.Float64bits(2.718281828))
	val, n, err := DecodeValue(TypeFloat64, buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 8 {
		t.Errorf("consumed %d, want 8", n)
	}
	if v := val.(float64); v != 2.718281828 {
		t.Errorf("got %v, want 2.718281828", v)
	}
}

func TestDecodeEncodeEUI64(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	val, n, err := DecodeValue(TypeEUI64, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 8 {
		t.Errorf("consumed %d, want 8", n)
	}
	addr := val.([8]byte)
	if !bytes.Equal(addr[:], data) {
		t.Errorf("got %X, want %X", addr, data)
	}

	encoded, err := EncodeValue(TypeEUI64, addr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, data) {
		t.Errorf("encoded %X, want %X", encoded, data)
	}
}

func TestDecodeOctetStr(t *testing.T) {
	data := []byte{3, 0xAA, 0xBB, 0xCC}
	val, n, err := DecodeValue(TypeOctetStr, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Errorf("consumed %d, want 4", n)
	}
	if !bytes.Equal(val.([]byte), []byte{0xAA, 0xBB, 0xCC}) {
		t.Errorf("got %X", val)
	}
}

func TestDecodeCharStr16(t *testing.T) {
	s := "Hello World"
	data := make([]byte, 2+len(s))
	binary.LittleEndian.PutUint16(data[:2], uint16(len(s)))
	copy(data[2:], s)

	val, n, err := DecodeValue(TypeCharStr16, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2+len(s) {
		t.Errorf("consumed %d, want %d", n, 2+len(s))
	}
	if val.(string) != s {
		t.Errorf("got %q, want %q", val, s)
	}
}

func TestDecodeOctetStr16(t *testing.T) {
	payload := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	data := make([]byte, 2+len(payload))
	binary.LittleEndian.PutUint16(data[:2], uint16(len(payload)))
	copy(data[2:], payload)

	val, n, err := DecodeValue(TypeOctetStr16, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2+len(payload) {
		t.Errorf("consumed %d, want %d", n, 2+len(payload))
	}
	if !bytes.Equal(val.([]byte), payload) {
		t.Errorf("got %X, want %X", val, payload)
	}
}

func TestDecodeCharStrInvalid(t *testing.T) {
	// 0xFF length = invalid
	data := []byte{0xFF}
	val, n, err := DecodeValue(TypeCharStr, data)
	if err != nil {
		t.Fatalf("0xFF length should return nil, not error: %v", err)
	}
	if val != nil {
		t.Errorf("got %v, want nil for 0xFF length", val)
	}
	if n != 1 {
		t.Errorf("consumed %d, want 1", n)
	}
}

func TestDecodeNoData(t *testing.T) {
	val, n, err := DecodeValue(TypeNoData, []byte{0xFF})
	if err != nil {
		t.Fatal(err)
	}
	if val != nil || n != 0 {
		t.Errorf("got val=%v, n=%d; want nil, 0", val, n)
	}
}

func TestEncodeOctetStr(t *testing.T) {
	encoded, err := EncodeValue(TypeOctetStr, []byte{0xDE, 0xAD})
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{2, 0xDE, 0xAD}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encoded %X, want %X", encoded, expected)
	}
}

func TestEncodeCharStr16(t *testing.T) {
	encoded, err := EncodeValue(TypeCharStr16, "AB")
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{0x02, 0x00, 'A', 'B'}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encoded %X, want %X", encoded, expected)
	}
}

func TestEncodeOctetStr16(t *testing.T) {
	encoded, err := EncodeValue(TypeOctetStr16, []byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{0x01, 0x00, 0x01}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encoded %X, want %X", encoded, expected)
	}
}

// --- Overflow and edge case tests ---

func TestEncodeUint8Overflow(t *testing.T) {
	_, err := EncodeValue(TypeUint8, uint64(256))
	if err == nil {
		t.Error("expected overflow error for uint8(256)")
	}
}

func TestEncodeInt8Overflow(t *testing.T) {
	_, err := EncodeValue(TypeInt8, int64(128))
	if err == nil {
		t.Error("expected overflow error for int8(128)")
	}
	_, err = EncodeValue(TypeInt8, int64(-129))
	if err == nil {
		t.Error("expected overflow error for int8(-129)")
	}
}

func TestEncodeUint24Overflow(t *testing.T) {
	_, err := EncodeValue(TypeUint24, uint64(0x1000000))
	if err == nil {
		t.Error("expected overflow error for uint24(0x1000000)")
	}
}

func TestEncodeInt24Overflow(t *testing.T) {
	_, err := EncodeValue(TypeInt24, int64(8388608))
	if err == nil {
		t.Error("expected overflow error for int24(8388608)")
	}
	_, err = EncodeValue(TypeInt24, int64(-8388609))
	if err == nil {
		t.Error("expected overflow error for int24(-8388609)")
	}
}

// --- toUint64 negative rejection tests ---

func TestToUint64RejectsNegativeInt(t *testing.T) {
	_, ok := toUint64(int(-1))
	if ok {
		t.Error("toUint64 should reject negative int")
	}
}

func TestToUint64RejectsNegativeInt64(t *testing.T) {
	_, ok := toUint64(int64(-1))
	if ok {
		t.Error("toUint64 should reject negative int64")
	}
}

func TestToUint64RejectsNegativeFloat64(t *testing.T) {
	_, ok := toUint64(float64(-0.5))
	if ok {
		t.Error("toUint64 should reject negative float64")
	}
}

func TestToUint64AcceptsZero(t *testing.T) {
	v, ok := toUint64(int(0))
	if !ok || v != 0 {
		t.Errorf("toUint64(int(0)) = %d, %v; want 0, true", v, ok)
	}
	v, ok = toUint64(float64(0))
	if !ok || v != 0 {
		t.Errorf("toUint64(float64(0)) = %d, %v; want 0, true", v, ok)
	}
}

func TestToUint64AcceptsPositive(t *testing.T) {
	v, ok := toUint64(int(42))
	if !ok || v != 42 {
		t.Errorf("toUint64(int(42)) = %d, %v; want 42, true", v, ok)
	}
	v, ok = toUint64(float64(255))
	if !ok || v != 255 {
		t.Errorf("toUint64(float64(255)) = %d, %v; want 255, true", v, ok)
	}
}

func TestEncodeUint8RejectsNegative(t *testing.T) {
	// This exercises the full path: EncodeValue -> toUint64 -> reject
	_, err := EncodeValue(TypeUint8, int(-1))
	if err == nil {
		t.Error("expected error for encoding negative int as uint8")
	}
}

func TestEncodeUint16RejectsNegativeFloat(t *testing.T) {
	_, err := EncodeValue(TypeUint16, float64(-1.0))
	if err == nil {
		t.Error("expected error for encoding negative float64 as uint16")
	}
}

// --- Type conversion edge cases ---

func TestToInt64FromFloat64(t *testing.T) {
	v, ok := toInt64(float64(-42.9))
	if !ok {
		t.Fatal("toInt64 should accept float64")
	}
	if v != -42 { // truncation expected
		t.Errorf("got %d, want -42", v)
	}
}

func TestToBoolFromFloat64(t *testing.T) {
	v, ok := toBool(float64(0))
	if !ok || v != false {
		t.Error("toBool(0.0) should be false")
	}
	v, ok = toBool(float64(1))
	if !ok || v != true {
		t.Error("toBool(1.0) should be true")
	}
}

func TestEncodeUnsupportedType(t *testing.T) {
	_, err := EncodeValue(0x48, "anything") // array type
	if err == nil {
		t.Error("expected error for unsupported encode type")
	}
}

func TestDecodeEnumTypes(t *testing.T) {
	val, n, err := DecodeValue(TypeEnum8, []byte{0x03})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 || val.(uint8) != 3 {
		t.Errorf("enum8: got %v, consumed %d", val, n)
	}

	val, n, err = DecodeValue(TypeEnum16, []byte{0x01, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 || val.(uint16) != 1 {
		t.Errorf("enum16: got %v, consumed %d", val, n)
	}
}

func TestDecodeBitmapTypes(t *testing.T) {
	val, n, err := DecodeValue(TypeBitmap8, []byte{0xFF})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 || val.(uint8) != 0xFF {
		t.Errorf("map8: got %v, consumed %d", val, n)
	}

	val, n, err = DecodeValue(TypeBitmap16, []byte{0xFF, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 || val.(uint16) != 0x00FF {
		t.Errorf("map16: got %v, consumed %d", val, n)
	}
}

func TestDecodeUTC(t *testing.T) {
	data := []byte{0x78, 0x56, 0x34, 0x12}
	val, n, err := DecodeValue(TypeUTC, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 || val.(uint32) != 0x12345678 {
		t.Errorf("UTC: got %v, consumed %d", val, n)
	}
}

func TestTypeSizeValues(t *testing.T) {
	tests := []struct {
		typeID uint8
		want   int
	}{
		{TypeNoData, 0},
		{TypeBool, 1},
		{TypeUint8, 1},
		{TypeUint16, 2},
		{TypeUint24, 3},
		{TypeUint32, 4},
		{TypeUint40, 5},
		{TypeUint48, 6},
		{TypeInt8, 1},
		{TypeInt16, 2},
		{TypeInt24, 3},
		{TypeInt32, 4},
		{TypeFloat16, 2},
		{TypeFloat32, 4},
		{TypeFloat64, 8},
		{TypeEUI64, 8},
		{TypeCharStr, -1},
		{TypeOctetStr, -1},
	}
	for _, tt := range tests {
		got := TypeSize(tt.typeID)
		if got != tt.want {
			t.Errorf("TypeSize(0x%02X) = %d, want %d", tt.typeID, got, tt.want)
		}
	}
}
