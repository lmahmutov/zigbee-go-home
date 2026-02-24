# zigbee-go-home

Zigbee smart home coordinator written in Go. Manages a Zigbee network through an nRF52840 NCP (Network Co-Processor) over USB, providing a web dashboard, REST API, real-time WebSocket events, MQTT bridge with Home Assistant autodiscovery, and Lua scripting with a visual Blockly editor.

```
  +-----------+       USB CDC ACM        +------------------+
  |  nRF52840 | <====================>  |  zigbee-go-home   |
  |   NCP     |    ZBOSS NCP / HDLC     |                   |
  |           |                         |   :8080 HTTP/WS   |
  +-----------+                         +--------+----------+
       |                                         |
       |  802.15.4                     +---------+---------+
       |                               |  Web UI / REST    |
  +----+----+                          |  MQTT (HA)        |
  | Zigbee  |                          |  Lua automations  |
  | devices |                          +-------------------+
  +---------+
```

## Features

- **Web dashboard** — Material Design 3 dark theme, sidebar navigation, i18n (EN/RU)
- **REST API** — device management, attribute read/write, cluster commands
- **WebSocket** — real-time device events, network state changes
- **MQTT bridge** — Home Assistant autodiscovery, state publishing, command handling
- **Lua automation** — scripted rules with `zigbee.*`, `system.*`, `telegram.*` modules
- **Blockly editor** — visual drag-and-drop automation builder
- **61 ZCL clusters** — plus custom/proprietary cluster support via JSON
- **Device definitions** — per-manufacturer config with bind, reporting, property decoding (Xiaomi TLV, Tuya DP)
- **BoltDB storage** — embedded key-value store, no external database

## Hardware

**NCP:** nRF52840 with ZBOSS NCP firmware (Nordic NCS v2.6.1). Tested on ProMicro nRF52840 with Adafruit UF2 bootloader.

Pre-built firmware is in the [`ncp-firmware/`](ncp-firmware/) directory. See [ncp-firmware/README.md](ncp-firmware/README.md) for flashing instructions.

## Quick Start

```bash
# Build
make build

# Or manually
go build -ldflags="-s -w" -o zigbee-home ./cmd/zigbee-home

# Run
./zigbee-home config.yaml
```

Open `http://localhost:8080` in a browser.

## Build

Requires Go 1.23+.

```bash
make build                # local build (stripped)
make build-linux-arm64    # aarch64 (MediaTek OpenWrt, RPi 64-bit)
make build-linux-arm      # ARMv7 (RPi 32-bit)
make build-linux-mipsle   # MIPS LE softfloat (ramips)
make build-linux-mips     # MIPS BE softfloat (ath79)
make build-minimal        # without MQTT and Lua automation (~9 MB)
make test                 # go test ./...
make vet                  # go vet ./...
```

### Build Tags

| Tag | Effect |
|-----|--------|
| `no_mqtt` | Excludes MQTT bridge + `paho.mqtt.golang` dependency |
| `no_automation` | Excludes Lua engine + `gopher-lua` dependency |

```bash
# Minimal build for constrained devices
go build -tags no_mqtt,no_automation -ldflags="-s -w" -o zigbee-home ./cmd/zigbee-home
```

## Configuration

```yaml
ncp:
  type: nrf52840
  port: /dev/ttyACM0
  baud: 460800                             # default: 460800

network:
  channel: 15                              # Zigbee channel (11-26)
  pan_id: 0x1A62
  extended_pan_id: "DD:DD:DD:DD:DD:DD:DD:DD"

web:
  listen: ":8080"
  api_key: "my-secret-key"                 # optional, enables auth
  allowed_origins:                         # optional, WebSocket CORS
    - "http://localhost:*"

store:
  path: "./zigbee-home.db"

devices_dir: "./devices"                   # device definitions (JSON)
scripts_dir: "./scripts"                   # Lua automation scripts

mqtt:
  enabled: false
  broker: "tcp://localhost:1883"
  username: ""
  password: ""
  topic_prefix: "zigbee2mqtt"              # HA-compatible topic prefix

telegram:
  bot_token: ""
  chat_ids: []

exec:
  allowlist: []                            # e.g. ["/usr/bin/curl"]
  timeout: "10s"

log:
  level: info                              # debug, info, warn, error
  format: text                             # text, json
```

## REST API

### Devices

```
GET    /api/devices              List all devices
GET    /api/devices/{ieee}       Get device details
DELETE /api/devices/{ieee}       Remove device
```

### Attributes

```
POST   /api/devices/{ieee}/read     Read attributes
POST   /api/devices/{ieee}/write    Write attribute
POST   /api/devices/{ieee}/command  Send cluster command
```

**Read attributes:**
```json
{ "endpoint": 1, "cluster_id": 6, "attr_ids": [0] }
```

**Write attribute:**
```json
{ "endpoint": 1, "cluster_id": 6, "attr_id": 0, "data_type": 16, "value": true }
```

**Send command:**
```json
{ "endpoint": 1, "cluster_id": 6, "command_id": 1 }
```

### Network

```
GET    /api/network              Network info (channel, PAN ID, state)
POST   /api/network/permit-join  Open network for joining
GET    /api/clusters             List all ZCL cluster definitions
GET    /api/version              Current version
```

### Automation

```
GET    /api/automations              List automations
GET    /api/automations/{id}         Get automation
POST   /api/automations              Create automation
PUT    /api/automations/{id}         Update automation
DELETE /api/automations/{id}         Delete automation
POST   /api/automations/{id}/toggle  Toggle enabled/disabled
POST   /api/automations/{id}/run     Run automation manually
```

## WebSocket

Connect to `ws://host:8080/ws` for real-time events.

| Event | Trigger |
|-------|---------|
| `device_joined` | New device joins the network |
| `device_left` | Device leaves the network |
| `device_announce` | Device announces presence |
| `attribute_report` | Device reports attribute value |
| `cluster_command` | Incoming cluster-specific command (e.g., Tuya DP) |
| `property_update` | Decoded proprietary attribute/command value |
| `network_state` | Network state changes |
| `permit_join` | Permit join status updated |

## MQTT Bridge

When enabled, publishes device states and subscribes to commands using a topic layout compatible with Home Assistant's MQTT autodiscovery.

```
zigbee2mqtt/{device_name}          # state (JSON)
zigbee2mqtt/{device_name}/set      # commands (JSON: {"state":"ON"}, {"brightness":128})
homeassistant/{type}/{id}/config   # HA autodiscovery
zigbee2mqtt/bridge/state           # online/offline
```

Supported HA entity types: `light` (JSON schema, brightness), `switch` (on/off), `sensor` (temperature, humidity, pressure, illuminance, battery, analog, link quality), `binary_sensor` (occupancy, IAS zone).

Supported commands: `state` (ON/OFF/TOGGLE), `brightness` (0-254).

## Authentication

When `api_key` is set in config, all `/api/` routes require the `X-API-Key` header:

```
X-API-Key: my-secret-key
```

Pages and WebSocket are not protected by the API key (browsers cannot send custom headers on page navigation). The API key is injected into HTML pages via a `<meta>` tag for use by client-side JavaScript.

## Device Definitions

JSON files in `devices/` directory configure per-device behavior: cluster binding, attribute reporting, proprietary property decoding. See [`devices/README.md`](devices/README.md) for the full format reference.

## OpenWrt

An OpenWrt package definition is in `openwrt/`. See `openwrt/Makefile` for integration with the OpenWrt build system. The package installs the binary, default config, and a procd init script.

## Dependencies

| Module | Purpose |
|--------|---------|
| [go.bug.st/serial](https://pkg.go.dev/go.bug.st/serial) | Serial port I/O |
| [go.etcd.io/bbolt](https://pkg.go.dev/go.etcd.io/bbolt) | Embedded key-value store |
| [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) | YAML config parsing |
| [nhooyr.io/websocket](https://pkg.go.dev/nhooyr.io/websocket) | WebSocket server |
| [paho.mqtt.golang](https://pkg.go.dev/github.com/eclipse/paho.mqtt.golang) | MQTT client (optional, excluded with `no_mqtt`) |
| [gopher-lua](https://pkg.go.dev/github.com/yuin/gopher-lua) | Lua VM (optional, excluded with `no_automation`) |
