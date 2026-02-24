package zcl

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ZCL data type IDs
const (
	TypeNoData    uint8 = 0x00
	TypeBool      uint8 = 0x10
	TypeUint8     uint8 = 0x20
	TypeUint16    uint8 = 0x21
	TypeUint24    uint8 = 0x22
	TypeUint32    uint8 = 0x23
	TypeUint40    uint8 = 0x24
	TypeUint48    uint8 = 0x25
	TypeInt8      uint8 = 0x28
	TypeInt16     uint8 = 0x29
	TypeInt32     uint8 = 0x2B
	TypeEnum8     uint8 = 0x30
	TypeEnum16    uint8 = 0x31
	TypeFloat16   uint8 = 0x38
	TypeFloat32   uint8 = 0x39
	TypeFloat64   uint8 = 0x3A
	TypeOctetStr  uint8 = 0x41
	TypeCharStr   uint8 = 0x42
	TypeOctetStr16 uint8 = 0x43
	TypeCharStr16  uint8 = 0x44
	TypeArray     uint8 = 0x48
	TypeStruct    uint8 = 0x4C
	TypeBitmap8   uint8 = 0x18
	TypeBitmap16  uint8 = 0x19
	TypeBitmap24  uint8 = 0x1A
	TypeBitmap32  uint8 = 0x1B
	TypeInt24     uint8 = 0x2A
	TypeEUI64     uint8 = 0xF0
	TypeClusterID uint8 = 0xE8
	TypeAttrID    uint8 = 0xE9
	TypeToD       uint8 = 0xE0 // Time of Day
	TypeDate      uint8 = 0xE1
	TypeUTC       uint8 = 0xE2
)

// TypeSize returns the fixed size in bytes of a ZCL type, or -1 for variable-length types.
func TypeSize(typeID uint8) int {
	switch typeID {
	case TypeNoData:
		return 0
	case TypeBool, TypeUint8, TypeInt8, TypeEnum8, TypeBitmap8:
		return 1
	case TypeUint16, TypeInt16, TypeEnum16, TypeBitmap16, TypeClusterID, TypeAttrID:
		return 2
	case TypeUint24, TypeInt24, TypeBitmap24:
		return 3
	case TypeFloat16:
		return 2
	case TypeUint32, TypeInt32, TypeBitmap32, TypeFloat32, TypeToD, TypeDate, TypeUTC:
		return 4
	case TypeUint40:
		return 5
	case TypeUint48:
		return 6
	case TypeFloat64, TypeEUI64:
		return 8
	case TypeOctetStr, TypeCharStr:
		return -1 // 1-byte length prefix
	case TypeOctetStr16, TypeCharStr16:
		return -1 // 2-byte length prefix
	case TypeArray, TypeStruct:
		return -1
	default:
		return -1
	}
}

// TypeName returns a human-readable name for a ZCL type.
func TypeName(typeID uint8) string {
	switch typeID {
	case TypeNoData:
		return "nodata"
	case TypeBool:
		return "bool"
	case TypeUint8:
		return "uint8"
	case TypeUint16:
		return "uint16"
	case TypeUint24:
		return "uint24"
	case TypeUint32:
		return "uint32"
	case TypeUint40:
		return "uint40"
	case TypeUint48:
		return "uint48"
	case TypeInt8:
		return "int8"
	case TypeInt16:
		return "int16"
	case TypeInt24:
		return "int24"
	case TypeInt32:
		return "int32"
	case TypeEnum8:
		return "enum8"
	case TypeEnum16:
		return "enum16"
	case TypeFloat16:
		return "float16"
	case TypeFloat32:
		return "float32"
	case TypeFloat64:
		return "float64"
	case TypeOctetStr:
		return "octstr"
	case TypeCharStr:
		return "string"
	case TypeOctetStr16:
		return "octstr16"
	case TypeCharStr16:
		return "string16"
	case TypeBitmap8:
		return "map8"
	case TypeBitmap16:
		return "map16"
	case TypeBitmap24:
		return "map24"
	case TypeBitmap32:
		return "map32"
	case TypeEUI64:
		return "EUI64"
	case TypeArray:
		return "array"
	case TypeStruct:
		return "struct"
	case TypeUTC:
		return "UTC"
	default:
		return fmt.Sprintf("0x%02X", typeID)
	}
}

// DecodeValue decodes a ZCL typed value from raw bytes, returning the Go value and bytes consumed.
func DecodeValue(typeID uint8, data []byte) (interface{}, int, error) {
	size := TypeSize(typeID)
	if size == 0 {
		return nil, 0, nil
	}

	// Variable-length types
	if size < 0 {
		return decodeVariableValue(typeID, data)
	}

	if len(data) < size {
		return nil, 0, fmt.Errorf("zcl: not enough data for type 0x%02X: need %d, have %d", typeID, size, len(data))
	}

	switch typeID {
	case TypeBool:
		return data[0] != 0, 1, nil
	case TypeUint8, TypeEnum8, TypeBitmap8:
		return data[0], 1, nil
	case TypeUint16, TypeEnum16, TypeBitmap16, TypeClusterID, TypeAttrID, TypeFloat16:
		return binary.LittleEndian.Uint16(data[:2]), 2, nil
	case TypeUint24, TypeBitmap24:
		return uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16, 3, nil
	case TypeInt24:
		v := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16
		if v&0x800000 != 0 {
			v |= 0xFF000000 // sign extend
		}
		return int32(v), 3, nil
	case TypeUint32, TypeBitmap32, TypeUTC:
		return binary.LittleEndian.Uint32(data[:4]), 4, nil
	case TypeUint40:
		v := uint64(data[0]) | uint64(data[1])<<8 | uint64(data[2])<<16 | uint64(data[3])<<24 | uint64(data[4])<<32
		return v, 5, nil
	case TypeUint48:
		v := uint64(data[0]) | uint64(data[1])<<8 | uint64(data[2])<<16 | uint64(data[3])<<24 | uint64(data[4])<<32 | uint64(data[5])<<40
		return v, 6, nil
	case TypeInt8:
		return int8(data[0]), 1, nil
	case TypeInt16:
		return int16(binary.LittleEndian.Uint16(data[:2])), 2, nil
	case TypeInt32:
		return int32(binary.LittleEndian.Uint32(data[:4])), 4, nil
	case TypeFloat32:
		bits := binary.LittleEndian.Uint32(data[:4])
		return math.Float32frombits(bits), 4, nil
	case TypeFloat64:
		bits := binary.LittleEndian.Uint64(data[:8])
		return math.Float64frombits(bits), 8, nil
	case TypeEUI64:
		var addr [8]byte
		copy(addr[:], data[:8])
		return addr, 8, nil
	case TypeToD, TypeDate:
		return binary.LittleEndian.Uint32(data[:4]), 4, nil
	}

	return data[:size], size, nil
}

func decodeVariableValue(typeID uint8, data []byte) (interface{}, int, error) {
	switch typeID {
	case TypeOctetStr, TypeCharStr:
		if len(data) < 1 {
			return nil, 0, fmt.Errorf("zcl: no length byte for string type")
		}
		length := int(data[0])
		if length == 0xFF {
			return nil, 1, nil // invalid
		}
		if len(data) < 1+length {
			return nil, 0, fmt.Errorf("zcl: string truncated: need %d, have %d", length, len(data)-1)
		}
		if typeID == TypeCharStr {
			return string(data[1 : 1+length]), 1 + length, nil
		}
		b := make([]byte, length)
		copy(b, data[1:1+length])
		return b, 1 + length, nil

	case TypeOctetStr16, TypeCharStr16:
		if len(data) < 2 {
			return nil, 0, fmt.Errorf("zcl: no length bytes for string16 type")
		}
		length := int(binary.LittleEndian.Uint16(data[:2]))
		if length == 0xFFFF {
			return nil, 2, nil
		}
		if len(data) < 2+length {
			return nil, 0, fmt.Errorf("zcl: string16 truncated")
		}
		if typeID == TypeCharStr16 {
			return string(data[2 : 2+length]), 2 + length, nil
		}
		b := make([]byte, length)
		copy(b, data[2:2+length])
		return b, 2 + length, nil
	}

	return nil, 0, fmt.Errorf("zcl: unsupported variable type 0x%02X", typeID)
}

// EncodeValue encodes a Go value into ZCL wire format.
func EncodeValue(typeID uint8, val interface{}) ([]byte, error) {
	switch typeID {
	case TypeBool:
		v, ok := toBool(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to bool", val)
		}
		if v {
			return []byte{1}, nil
		}
		return []byte{0}, nil

	case TypeUint8, TypeEnum8, TypeBitmap8:
		v, ok := toUint64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to uint8", val)
		}
		if v > math.MaxUint8 {
			return nil, fmt.Errorf("zcl: value %d overflows uint8 (max %d)", v, math.MaxUint8)
		}
		return []byte{uint8(v)}, nil

	case TypeUint16, TypeEnum16, TypeBitmap16:
		v, ok := toUint64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to uint16", val)
		}
		if v > math.MaxUint16 {
			return nil, fmt.Errorf("zcl: value %d overflows uint16 (max %d)", v, math.MaxUint16)
		}
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(v))
		return buf, nil

	case TypeUint32, TypeBitmap32, TypeUTC:
		v, ok := toUint64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to uint32", val)
		}
		if v > uint64(math.MaxUint32) {
			return nil, fmt.Errorf("zcl: value %d overflows uint32 (max %d)", v, uint64(math.MaxUint32))
		}
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(v))
		return buf, nil

	case TypeInt8:
		v, ok := toInt64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to int8", val)
		}
		if v < math.MinInt8 || v > math.MaxInt8 {
			return nil, fmt.Errorf("zcl: value %d overflows int8 (range %d..%d)", v, math.MinInt8, math.MaxInt8)
		}
		return []byte{byte(int8(v))}, nil

	case TypeInt16:
		v, ok := toInt64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to int16", val)
		}
		if v < math.MinInt16 || v > math.MaxInt16 {
			return nil, fmt.Errorf("zcl: value %d overflows int16 (range %d..%d)", v, math.MinInt16, math.MaxInt16)
		}
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(int16(v)))
		return buf, nil

	case TypeInt24:
		v, ok := toInt64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to int24", val)
		}
		if v < -8388608 || v > 8388607 {
			return nil, fmt.Errorf("zcl: value %d overflows int24 (range %d..%d)", v, -8388608, 8388607)
		}
		u := uint32(int32(v))
		return []byte{byte(u), byte(u >> 8), byte(u >> 16)}, nil

	case TypeUint24, TypeBitmap24:
		v, ok := toUint64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to uint24", val)
		}
		if v > 0xFFFFFF {
			return nil, fmt.Errorf("zcl: value %d overflows uint24 (max %d)", v, 0xFFFFFF)
		}
		return []byte{byte(v), byte(v >> 8), byte(v >> 16)}, nil

	case TypeInt32:
		v, ok := toInt64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to int32", val)
		}
		if v < math.MinInt32 || v > math.MaxInt32 {
			return nil, fmt.Errorf("zcl: value %d overflows int32 (range %d..%d)", v, math.MinInt32, math.MaxInt32)
		}
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(int32(v)))
		return buf, nil

	case TypeUint40:
		v, ok := toUint64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to uint40", val)
		}
		if v > 0xFFFFFFFFFF {
			return nil, fmt.Errorf("zcl: value %d overflows uint40 (max %d)", v, uint64(0xFFFFFFFFFF))
		}
		return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24), byte(v >> 32)}, nil

	case TypeUint48:
		v, ok := toUint64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to uint48", val)
		}
		if v > 0xFFFFFFFFFFFF {
			return nil, fmt.Errorf("zcl: value %d overflows uint48 (max %d)", v, uint64(0xFFFFFFFFFFFF))
		}
		return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24), byte(v >> 32), byte(v >> 40)}, nil

	case TypeFloat16:
		v, ok := toUint64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to float16 (raw uint16)", val)
		}
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(v))
		return buf, nil

	case TypeFloat32:
		v, ok := toFloat64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to float32", val)
		}
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(v)))
		return buf, nil

	case TypeFloat64:
		v, ok := toFloat64(val)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to float64", val)
		}
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, math.Float64bits(v))
		return buf, nil

	case TypeEUI64:
		switch a := val.(type) {
		case [8]byte:
			b := make([]byte, 8)
			copy(b, a[:])
			return b, nil
		case []byte:
			if len(a) != 8 {
				return nil, fmt.Errorf("zcl: EUI64 requires 8 bytes, got %d", len(a))
			}
			b := make([]byte, 8)
			copy(b, a)
			return b, nil
		default:
			return nil, fmt.Errorf("zcl: cannot convert %T to EUI64", val)
		}

	case TypeCharStr:
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to string", val)
		}
		if len(s) > 254 {
			return nil, fmt.Errorf("zcl: string too long for CharStr: %d (max 254)", len(s))
		}
		buf := make([]byte, 1+len(s))
		buf[0] = uint8(len(s))
		copy(buf[1:], s)
		return buf, nil

	case TypeOctetStr:
		b, ok := val.([]byte)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to []byte", val)
		}
		if len(b) > 254 {
			return nil, fmt.Errorf("zcl: data too long for OctetStr: %d (max 254)", len(b))
		}
		buf := make([]byte, 1+len(b))
		buf[0] = uint8(len(b))
		copy(buf[1:], b)
		return buf, nil

	case TypeCharStr16:
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to string", val)
		}
		if len(s) > 65534 {
			return nil, fmt.Errorf("zcl: string too long for CharStr16: %d (max 65534)", len(s))
		}
		buf := make([]byte, 2+len(s))
		binary.LittleEndian.PutUint16(buf[:2], uint16(len(s)))
		copy(buf[2:], s)
		return buf, nil

	case TypeOctetStr16:
		b, ok := val.([]byte)
		if !ok {
			return nil, fmt.Errorf("zcl: cannot convert %T to []byte", val)
		}
		if len(b) > 65534 {
			return nil, fmt.Errorf("zcl: data too long for OctetStr16: %d (max 65534)", len(b))
		}
		buf := make([]byte, 2+len(b))
		binary.LittleEndian.PutUint16(buf[:2], uint16(len(b)))
		copy(buf[2:], b)
		return buf, nil
	}

	return nil, fmt.Errorf("zcl: encode not implemented for type 0x%02X", typeID)
}

func toBool(v interface{}) (bool, bool) {
	switch val := v.(type) {
	case bool:
		return val, true
	case float64:
		return val != 0, true
	case int:
		return val != 0, true
	}
	return false, false
}

func toUint64(v interface{}) (uint64, bool) {
	switch val := v.(type) {
	case uint8:
		return uint64(val), true
	case uint16:
		return uint64(val), true
	case uint32:
		return uint64(val), true
	case uint64:
		return val, true
	case int:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case int64:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case float64:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	}
	return 0, false
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float32:
		return float64(val), true
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint64:
		return float64(val), true
	}
	return 0, false
}

func toInt64(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	case int:
		return int64(val), true
	case uint8:
		return int64(val), true
	case uint16:
		return int64(val), true
	case uint32:
		return int64(val), true
	case uint64:
		if val > math.MaxInt64 {
			return 0, false
		}
		return int64(val), true
	case float64:
		if val > math.MaxInt64 || val < math.MinInt64 {
			return 0, false
		}
		return int64(val), true
	}
	return 0, false
}
