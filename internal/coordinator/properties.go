package coordinator

import (
	"fmt"

	"zigbee-go-home/internal/ncp"
	"zigbee-go-home/internal/store"
	"zigbee-go-home/internal/zcl"
)

// processProperties checks if the device has property definitions for this
// cluster/attribute and emits property_update events for each extracted value.
// Accepts the already-loaded device to avoid a redundant DB read.
func (dm *DeviceManager) processProperties(ieee string, dev *store.Device, evt ncp.AttributeReportEvent, decoded interface{}) {
	if ieee == "" || dev == nil || dev.Manufacturer == "" || dev.Model == "" {
		return
	}

	db := dm.coord.DeviceDB()
	if db == nil {
		return
	}
	def := db.Lookup(dev.Manufacturer, dev.Model)
	if def == nil || len(def.Properties) == 0 {
		return
	}

	dirty := false
	for _, ps := range def.Properties {
		if ps.Cluster != evt.ClusterID || ps.Attribute != evt.AttrID {
			continue
		}

		var tlvMap map[int]interface{}
		switch ps.Decoder {
		case "xiaomi_tlv":
			var raw []byte
			switch v := decoded.(type) {
			case []byte:
				raw = v
			case string:
				raw = []byte(v)
			default:
				dm.logger.Warn("property decoder expects []byte or string",
					"ieee", ieee, "decoder", ps.Decoder,
					"got", fmt.Sprintf("%T", decoded))
				continue
			}
			var decErr error
			tlvMap, decErr = decodeXiaomiTLV(raw)
			if decErr != nil {
				dm.logger.Warn("xiaomi TLV decode failed",
					"ieee", ieee, "err", decErr)
				continue
			}
		default:
			dm.logger.Warn("unknown property decoder",
				"ieee", ieee, "decoder", ps.Decoder)
			continue
		}

		for _, v := range ps.Values {
			raw, ok := tlvMap[v.Tag]
			if !ok {
				continue
			}
			value := raw
			if v.Transform != "" {
				value = applyTransform(v.Transform, raw)
			}

			// Collect property value (batched save after loop).
			if dev.Properties == nil {
				dev.Properties = make(map[string]any)
			}
			dev.Properties[v.Name] = value
			dirty = true

			dm.coord.Events().Emit(Event{
				Type: EventPropertyUpdate,
				Data: map[string]interface{}{
					"ieee":     ieee,
					"property": v.Name,
					"value":    value,
					"source": map[string]interface{}{
						"cluster":   ps.Cluster,
						"attribute": ps.Attribute,
						"decoder":   ps.Decoder,
						"tag":       v.Tag,
					},
				},
			})

			dm.logger.Info("property update",
				"ieee", ieee,
				"name", deviceName(dev),
				"property", v.Name,
				"value", value,
			)
		}
	}

	// Persist all collected properties in a single DB write.
	if dirty {
		if saveErr := dm.coord.Store().SaveDevice(dev); saveErr != nil {
			dm.logger.Error("save device properties", "err", saveErr, "ieee", ieee)
		}
	}
}

// decodeXiaomiTLV parses the Xiaomi proprietary TLV format.
// Each entry is: [tag:uint8][zcl_type:uint8][value:variable].
// Reuses zcl.DecodeValue for the value portion.
func decodeXiaomiTLV(data []byte) (map[int]interface{}, error) {
	result := make(map[int]interface{})
	pos := 0

	for pos < len(data) {
		if pos+2 > len(data) {
			break // need at least tag + type
		}
		tag := int(data[pos])
		typeID := data[pos+1]
		pos += 2

		val, consumed, err := zcl.DecodeValue(typeID, data[pos:])
		if err != nil {
			return result, fmt.Errorf("tag %d type 0x%02X at offset %d: %w", tag, typeID, pos, err)
		}
		result[tag] = val
		pos += consumed
	}

	return result, nil
}

// applyTransform converts a raw decoded value using a named transform.
func applyTransform(name string, value interface{}) interface{} {
	switch name {
	case "lumi_battery":
		return lumiBattery(value)
	case "minus_one":
		return minusOne(value)
	case "lumi_trigger":
		return lumiTrigger(value)
	case "bool_invert":
		return boolInvert(value)
	default:
		return value
	}
}

// lumiBattery converts millivolt reading to battery percentage.
// 2850 mV = 0%, 3000 mV = 100%, linearly interpolated and clamped.
func lumiBattery(value interface{}) interface{} {
	mv, ok := toNumeric(value)
	if !ok {
		return value
	}
	const minMV, maxMV = 2850, 3000
	pct := float64(mv-minMV) / float64(maxMV-minMV) * 100
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return int(pct)
}

// minusOne subtracts 1 from a numeric value.
func minusOne(value interface{}) interface{} {
	n, ok := toNumeric(value)
	if !ok {
		return value
	}
	return n - 1
}

// lumiTrigger extracts the lower 16 bits of a uint64, then subtracts 1.
func lumiTrigger(value interface{}) interface{} {
	n, ok := toNumeric(value)
	if !ok {
		return value
	}
	return (n & 0xFFFF) - 1
}

// boolInvert inverts a boolean value.
func boolInvert(value interface{}) interface{} {
	switch v := value.(type) {
	case bool:
		return !v
	case uint8:
		return v == 0
	case uint16:
		return v == 0
	case uint32:
		return v == 0
	case uint64:
		return v == 0
	default:
		return value
	}
}

// toNumeric converts various numeric types to int64 for transform calculations.
func toNumeric(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}
