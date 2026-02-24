# Device Definitions

This directory contains JSON files that describe Zigbee devices: how to bind clusters, configure reporting, and extract named properties from manufacturer-specific attributes.

Each `*.json` file is loaded at startup. One file per brand family (e.g., `xiaomi.json`, `IKEA.json`).

## Quick Start

1. Pair the device and check the logs for `manufacturer` and `model` strings.
2. Find or create a JSON file for that brand.
3. Add a device entry with bind/reporting/properties as needed.
4. Restart the coordinator.

## Two JSON Formats

### Manufacturers (grouped format, preferred)

Groups models under a manufacturer name. The `name` field is automatically set as `manufacturer` for each model — no need to repeat it.

```json
{
  "manufacturers": [
    {
      "name": "LUMI",
      "models": [
        {
          "model": "lumi.sensor_magnet.aq2",
          "friendly_name": "Aqara Door/Window Sensor",
          "bind": [6],
          "properties": [...]
        }
      ]
    }
  ]
}
```

### Devices (flat format, legacy)

Each entry must have explicit `manufacturer` and `model`.

```json
{
  "devices": [
    {
      "manufacturer": "IKEA of Sweden",
      "model": "TRADFRI on/off switch",
      "friendly_name": "IKEA On/Off Switch",
      "bind": [6, 8],
      "reporting": [
        {"cluster": 1, "attribute": 33, "type": 32, "min": 3600, "max": 65534, "change": 10}
      ]
    }
  ]
}
```

Both formats work simultaneously and can coexist in the same file.

## Device Definition Fields

| Field           | Type              | Required | Description |
|-----------------|-------------------|----------|-------------|
| `manufacturer`  | string            | yes*     | Manufacturer string from Basic cluster (0x0004). Not needed in `manufacturers` format — set from group `name`. |
| `model`         | string            | yes      | Model string from Basic cluster (0x0005). |
| `friendly_name` | string            | no       | Human-readable name shown in the UI. Falls back to model if not set. |
| `bind`          | array of uint16   | yes      | Cluster IDs to bind to the coordinator during interview. |
| `reporting`     | array of objects  | no       | Attribute reporting configuration entries. |
| `properties`    | array of objects  | no       | Proprietary attribute decoders for named property extraction. |

## Bind

Array of cluster IDs. During interview, the coordinator binds each listed cluster (if found as an OUT cluster on the device's first endpoint) to the coordinator. This enables the device to send reports.

Common cluster IDs:

| ID     | Hex      | Name |
|--------|----------|------|
| 0      | 0x0000   | Basic |
| 1      | 0x0001   | Power Configuration |
| 6      | 0x0006   | On/Off |
| 8      | 0x0008   | Level Control |
| 768    | 0x0300   | Color Control |
| 1026   | 0x0402   | Temperature Measurement |
| 1029   | 0x0405   | Relative Humidity |
| 1030   | 0x0406   | Occupancy Sensing |

## Reporting

Configures the device to periodically report attribute values. Each entry:

```json
{
  "cluster": 1026,
  "attribute": 0,
  "type": 41,
  "min": 10,
  "max": 300,
  "change": 10
}
```

| Field       | Description |
|-------------|-------------|
| `cluster`   | Cluster ID (uint16). |
| `attribute` | Attribute ID (uint16). |
| `type`      | ZCL data type ID of the attribute (see table below). |
| `min`       | Minimum reporting interval in seconds. |
| `max`       | Maximum reporting interval in seconds. |
| `change`    | Reportable change threshold. For temperature (int16), 10 = 0.1 C. |

### ZCL Data Type IDs

| ID (dec) | ID (hex) | Name     | Size    |
|----------|----------|----------|---------|
| 16       | 0x10     | bool     | 1       |
| 32       | 0x20     | uint8    | 1       |
| 33       | 0x21     | uint16   | 2       |
| 35       | 0x23     | uint32   | 4       |
| 40       | 0x28     | int8     | 1       |
| 41       | 0x29     | int16    | 2       |
| 43       | 0x2B     | int32    | 4       |
| 48       | 0x30     | enum8    | 1       |
| 65       | 0x41     | octstr   | var     |
| 66       | 0x42     | string   | var     |

## Properties

Properties extract named values from proprietary attributes (e.g., Xiaomi's 0xFF01 TLV blob on Basic cluster). This turns opaque binary data into discrete `property_update` events on the WebSocket.

```json
{
  "properties": [
    {
      "cluster": 0,
      "attribute": 65281,
      "decoder": "xiaomi_tlv",
      "values": [
        {"tag": 1, "name": "battery_voltage"},
        {"tag": 1, "name": "battery", "transform": "lumi_battery"},
        {"tag": 3, "name": "device_temperature"},
        {"tag": 100, "name": "contact", "transform": "bool_invert"}
      ]
    }
  ]
}
```

### Property Source Fields

| Field       | Description |
|-------------|-------------|
| `cluster`   | Cluster ID of the attribute report to match. |
| `attribute` | Attribute ID to match (e.g., 65281 = 0xFF01). |
| `decoder`   | Decoder name. Currently supported: `xiaomi_tlv`. |
| `values`    | Array of property definitions to extract from decoded data. |

### Property Value Fields

| Field       | Required | Description |
|-------------|----------|-------------|
| `tag`       | yes      | Tag number in the TLV data. |
| `name`      | yes      | Property name emitted in `property_update` events. |
| `transform` | no       | Optional transform applied to the raw value. |

The same tag can appear multiple times with different names/transforms (e.g., tag 1 used for both raw `battery_voltage` and transformed `battery` percentage).

### Available Transforms

| Name            | Description |
|-----------------|-------------|
| `lumi_battery`  | Millivolts to percentage. 2850 mV = 0%, 3000 mV = 100%, linearly interpolated, clamped to 0-100. |
| `minus_one`     | Subtracts 1 from the value. |
| `lumi_trigger`  | Lower 16 bits of the value, then subtracts 1. |
| `bool_invert`   | Inverts a boolean. Also works on numeric types (0 becomes `true`). |

### Xiaomi TLV Format

The `xiaomi_tlv` decoder parses the binary format used in Xiaomi/LUMI attribute 0xFF01:

```
[tag:1][zcl_type:1][value:N] [tag:1][zcl_type:1][value:N] ...
```

Each entry is a tag byte, followed by a standard ZCL type byte, followed by the value encoded per ZCL rules. The decoder reuses the ZCL type system, so all standard types (uint8, uint16, int8, bool, uint40, etc.) are supported.

Common tags for LUMI devices:

| Tag | Typical Type | Meaning |
|-----|-------------|---------|
| 1   | uint16      | Battery voltage (mV) |
| 3   | int8        | Device temperature (C) |
| 5   | uint16      | Power outage count |
| 6   | uint40      | Trigger count |
| 100 | bool/uint16 | Primary sensor value (contact, occupancy, etc.) |
| 101 | varies      | Secondary sensor value |
| 102 | varies      | Tertiary sensor value |

## How to Add a New Device Step by Step

### 1. Pair and identify

Pair the device and look at the coordinator logs for the interview output:

```
device joined    ieee=2469C908008D1500
attribute report cluster=Basic attr=ManufacturerName value="LUMI"
attribute report cluster=Basic attr=ModelIdentifier value="lumi.sensor_magnet.aq2"
```

Note the exact `manufacturer` and `model` strings.

### 2. Choose a file

- Place the device in an existing file if one exists for the brand (e.g., `xiaomi.json` for LUMI).
- Create a new file for a new brand (e.g., `tuya.json`). Any `*.json` filename works.

### 3. Add the device definition

Minimal example using the manufacturers format:

```json
{
  "manufacturers": [
    {
      "name": "Acme Corp",
      "models": [
        {
          "model": "smart-switch-v1",
          "friendly_name": "Acme Smart Switch",
          "bind": [6]
        }
      ]
    }
  ]
}
```

### 4. Add reporting if needed

For sensors that support configured reporting:

```json
{
  "model": "temp-sensor-v1",
  "friendly_name": "Acme Temperature Sensor",
  "bind": [],
  "reporting": [
    {"cluster": 1026, "attribute": 0, "type": 41, "min": 10, "max": 300, "change": 10},
    {"cluster": 1029, "attribute": 0, "type": 33, "min": 10, "max": 300, "change": 100}
  ]
}
```

### 5. Add proprietary attribute decoding if needed

If the device sends proprietary data blobs (like Xiaomi's 0xFF01), add `properties` with the appropriate decoder and tag mappings. Check device documentation or packet captures for the TLV tag meanings.

### 6. Restart and verify

Restart the coordinator. Check the logs for:

```
loaded device file  path=yourfile.json clusters=0 devices=1
device database loaded  files=2 devices=3
```

Remove and re-pair the device so it goes through a fresh interview with the new definition applied.

## Current Files

- **`xiaomi.json`** — LUMI/Aqara devices with Xiaomi TLV property decoding
- **`IKEA.json`** — IKEA TRADFRI devices
