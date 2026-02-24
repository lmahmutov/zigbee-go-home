package ncp

import "encoding/binary"

const (
	typeSizeVariable   = -1 // variable-length type with 1-byte length prefix
	typeSizeVariable16 = -3 // variable-length type with 2-byte length prefix
	typeSizeUnknown    = -2 // unrecognized type
)

func parseAttributeResponses(data []byte) []AttributeResponse {
	var results []AttributeResponse
	for len(data) >= 3 {
		// ZCL Read Attributes Response: AttrID (2 bytes) + Status (1 byte)
		attrID := binary.LittleEndian.Uint16(data[0:2])
		status := data[2]
		data = data[3:]

		ar := AttributeResponse{AttrID: attrID, Status: status}
		if status != 0 {
			results = append(results, ar)
			continue
		}
		if len(data) < 1 {
			break
		}
		ar.DataType = data[0]
		data = data[1:]

		// Read value based on type size
		size := typeSize(ar.DataType)
		if size == typeSizeUnknown {
			// Unknown type: cannot determine value boundaries.
			// Return what we have so far rather than guessing.
			results = append(results, ar)
			return results
		}
		if size > 0 && len(data) >= size {
			ar.Value = make([]byte, size)
			copy(ar.Value, data[:size])
			data = data[size:]
		} else if size == typeSizeVariable && len(data) >= 1 {
			// Variable length with 1-byte length prefix (octstr, string)
			vlen := int(data[0])
			if len(data) >= 1+vlen {
				ar.Value = make([]byte, 1+vlen)
				copy(ar.Value, data[:1+vlen])
				data = data[1+vlen:]
			}
		} else if size == typeSizeVariable16 && len(data) >= 2 {
			// Variable length with 2-byte length prefix (octstr16, string16)
			vlen := int(binary.LittleEndian.Uint16(data[:2]))
			if len(data) >= 2+vlen {
				ar.Value = make([]byte, 2+vlen)
				copy(ar.Value, data[:2+vlen])
				data = data[2+vlen:]
			}
		}
		results = append(results, ar)
	}
	return results
}

func typeSize(t uint8) int {
	switch {
	case t >= 0x08 && t <= 0x0F: // data8..data64
		return int(t-0x08) + 1
	case t == 0x10: // bool
		return 1
	case t == 0x18: // map8
		return 1
	case t == 0x19: // map16
		return 2
	case t == 0x1A: // map24
		return 3
	case t == 0x1B: // map32
		return 4
	case t >= 0x20 && t <= 0x27: // uint8..uint64
		return int(t-0x20) + 1
	case t >= 0x28 && t <= 0x2F: // int8..int64
		return int(t-0x28) + 1
	case t == 0x30: // enum8
		return 1
	case t == 0x31: // enum16
		return 2
	case t == 0x38: // float16
		return 2
	case t == 0x39: // float32
		return 4
	case t == 0x3A: // float64
		return 8
	case t == 0xE0, t == 0xE1: // ToD, Date
		return 4
	case t == 0xE2: // UTC
		return 4
	case t == 0xE8, t == 0xE9: // ClusterID, AttrID
		return 2
	case t == 0xF0: // EUI64
		return 8
	case t == 0x41, t == 0x42: // octstr, string (1-byte length prefix)
		return typeSizeVariable
	case t == 0x43, t == 0x44: // octstr16, string16 (2-byte length prefix)
		return typeSizeVariable16
	}
	return typeSizeUnknown
}
