# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

zigbee-go-home is a Go-based Zigbee coordinator that manages a Zigbee smart home network through a Network Co-Processor (NCP) connected via serial port. It uses an nRF52840 dongle with ZBOSS NCP firmware over USB CDC ACM. Features: web dashboard, REST API, WebSocket events, MQTT bridge with Home Assistant autodiscovery, and Lua-based automation scripting with a visual Blockly editor.

## Build & Test Commands

```bash
make build                                      # build (stripped, with version tag)
go build -o zigbee-home ./cmd/zigbee-home       # build without version/strip
go test ./...                                   # all tests
go test ./internal/automation/...               # single package
go test -v -run TestManagerSaveAndGet ./internal/automation  # single test
go vet ./...                                    # static analysis
```

**Build tags:** `no_mqtt` excludes MQTT bridge + paho dependency; `no_automation` excludes Lua engine + gopher-lua. Use `make build-minimal` or `go build -tags no_mqtt,no_automation`.

**Cross-compile targets:** `make build-linux-arm64` (RPi/OpenWrt aarch64), `make build-linux-arm` (ARMv7), `make build-linux-mipsle`/`make build-linux-mips` (OpenWrt ramips/ath79).

Running the binary requires `config.yaml` and serial hardware: `./zigbee-home config.yaml`

Note: `cmd/zigbee-home` is in `.gitignore` (matches binary name `zigbee-home`), so use `git add -f cmd/zigbee-home/main.go` when staging changes.

## Architecture

**Startup flow:** `main.go` loads YAML config → registers 61 ZCL clusters → loads device definitions from `devices/` → opens BoltDB store → creates NCP backend → starts Coordinator (form/resume network) → starts automation engine (Lua VMs) → starts HTTP server → starts MQTT bridge if enabled.

**Shutdown order:** HTTP server (stop accepting) → WebSocket hub → automation engine → MQTT bridge → coordinator (closes NCP). HTTP stops first to prevent in-flight requests hitting stopped backends.

**Network lifecycle:** Coordinator tries to resume existing network from NVRAM first (checks `NetworkState.Formed` + matching channel/PAN ID/ext PAN ID in BoltDB). Only forms a new network if no saved state or parameters changed. This prevents orphaning paired devices on restart.

**NCP interface** (`internal/ncp/ncp.go`) is the core abstraction — all hardware communication goes through it. The `NCP` interface defines 14 methods (network, ZDO, ZCL) + 6 indication callbacks + Close. Coordinator calls NCP methods; NCP calls back via `OnDeviceJoined`, `OnDeviceLeft`, `OnDeviceAnnounce`, `OnAttributeReport`, `OnClusterCommand`, `OnNwkAddrUpdate`. Implementation: `NRF52840NCP` (ZBOSS NCP over USB CDC ACM with HDLC-Lite framing).

**Key layers:**

- **`cmd/zigbee-home/main.go`** — Entry point, config parsing, wiring all components. Registers all 61 standard ZCL clusters via `registerStandardClusters()`. Config struct has NCP, Network, Web, Store, MQTT, Log, DevicesDir, ScriptsDir fields.
- **`internal/coordinator/`** — Central orchestrator. `DeviceManager` handles join/leave/interview lifecycle with concurrent interview goroutines tracked via `interviewWg`/`interviewCancels` with generation counter. Interview uses `UpdateDevice` (atomic merge) instead of `SaveDevice` to avoid overwriting concurrent attribute report updates; checks `ctx.Err()` before saving to prevent resurrecting deleted devices. `EventBus` provides pub/sub with panic recovery (8 event types). `attribute_reader.go` wraps ZCL read/write/command. `properties.go` decodes proprietary formats (Xiaomi TLV, Tuya DP). `devicedb.go` loads device definitions from JSON files (immutable after startup, no mutex needed).
- **`internal/ncp/`** — nRF52840 NCP backend: ZBOSS protocol codec (`nrf52840_zboss.go`), HDLC-Lite framing (`nrf52840_hdlc.go`), async request/response tracking via TSN, ZCL response parsing (`zcl_parse.go` handles variable-length types). Serial I/O with separate mutexes for pending requests vs write serialization. **Important:** NCP indication callbacks (from `readLoop`) must NOT call `request()` inline — it deadlocks because `request()` waits for ACK/response that only `readLoop` can deliver. Use `go` for any callback that needs to send commands (see OTA handler). `GetNCPInfo()` returns a copy to prevent data races. `NetworkInfo()` returns error if all queries fail. `request()` handles nil `*zbossFrame` from closed channels (NCP reset/close).
- **`internal/zcl/`** — ZCL type system (`types.go`) handles 44 data types with encode/decode; `toInt64`/`toUint64` converters validate bounds (uint64 > MaxInt64, float64 range). `Registry` stores cluster definitions with deep-copy-on-read semantics to prevent races. `clusters/` contains declarative Go struct definitions for each standard cluster.
- **`internal/store/`** — BoltDB with `devices` and `network` buckets. `NetworkState.NetworkKey` uses `json:"-"` to hide from API but `networkStateStorage` internal struct preserves it on disk. `UpdateDevice(ieee, fn)` provides atomic read-modify-write transactions.
- **`internal/web/`** — HTTP server with REST API, WebSocket hub (`WSHub`, max 100 clients) for real-time event broadcast, Go HTML templates (embedded via `//go:embed`), and automation page with Blockly visual editor. Auth via `X-API-Key` header with constant-time comparison — only `/api/` routes require auth (pages and WebSocket are unprotected so browsers can navigate without custom headers; designed for trusted LAN). API key injected into HTML via `<meta name="api-key">` tag, read by JS for AJAX calls. CORS protection for REST API with configurable `allowed_origins`. `DeviceView` enrichment adds derived fields (HasOnOff, HasLevel, DeviceType, OnOffState, battery %, LQI quality, etc.) with path traversal protection on device model names. Client-side i18n (EN/RU) via `data-i18n` attributes and `t()` JS function with localStorage persistence. Templates: layout.html, index.html, devices.html, device_detail.html, network.html, automations.html.
- **`internal/automation/`** — Lua scripting engine (gopher-lua). `Manager` handles script file I/O (`.lua` files in `scripts/` dir with JSON metadata header + optional Blockly XML); `Save` validates script IDs and caps unique-name generation at 1000 iterations. `Engine` spawns sandboxed Lua VMs per script with dedicated goroutine + command channel for thread-safe access. **Lua sandbox** uses whitelist approach (`SkipOpenLibs: true`): only `base`, `table`, `string`, `math` are loaded; `loadfile`/`dofile`/`load` removed from base. `newSandboxedLuaState()` factory used by both `startScript` and `RunLuaCode`. Three Lua modules: `zigbee.*` (on, turn_on/off, toggle, set_brightness, set_color, send_command, get_property, after, log, devices), `system.*` (exec, datetime, time_between, log), `telegram.*` (send — uses `json.Marshal` for proper JSON encoding). `RunLuaCode` provides one-shot execution for syntax checking and log capture — runs top-level code only (does not invoke registered handlers). Sandbox limits: `CallStackSize=120`, `RegistryMaxSize=80K`, `string.rep` capped at 1MB. `Engine.Stop()` waits for all VM goroutines and `zigbee.after()` goroutines via `vmWg`. EventBus subscription routes events to matching Lua handlers.
- **`internal/mqtt/`** — MQTT bridge with Home Assistant autodiscovery. Async event processing via buffered channel (256) with context-based cancellation. Publishes state updates, subscribes to per-device command topics (`{prefix}/{name}/set`). HA discovery for: `light` (JSON schema, brightness), `switch` (on/off), `sensor` (temperature, humidity, pressure, illuminance, battery, analog, LQI), `binary_sensor` (occupancy, IAS zone). `RemoveDevice` cleans up both state and discovery topics. `Store.UpdateDevice` used for atomic device property updates. `findEndpointWithCluster` returns endpoint 1 as fallback for devices with no endpoints. `publish()` tracks token-wait goroutines via `pubWg`; `Stop()` waits for all goroutines (`eventWg` + `discWg` + `pubWg`) before disconnect. Config validation requires `mqtt.broker` when `mqtt.enabled` is true.

**Concurrency patterns:**
- `DeviceManager` uses `sync.WaitGroup` + generation counter to track/cancel concurrent device interviews
- `DeviceManager.addrIndex` (short addr → IEEE mapping) uses `sync.RWMutex` with lock-free reads; `lookupOrRebuild()` does double-check locking
- nRF52840 backend uses separate mutexes: `hlMu` for pending HL requests, `writeMu` for serial writes, `zclMu` for ZCL response channels, `llSeqMu` for packet sequence
- `EventBus.Emit()` snapshots handler list under RLock, calls handlers outside lock
- `WSHub` uses channel-based register/unregister/broadcast pattern with slow-client eviction and max 100 client limit
- Automation VMs: per-script command channel serializes all Lua access; `zigbee.after()` goroutines tracked via `vmWg` for clean shutdown; `Engine.Stop()` waits via `vmWg`
- MQTT Bridge: async event processing via buffered channel (256) with context-based cancellation; `eventWg` tracks event loop, `discWg` tracks discovery, `pubWg` tracks publish goroutines; shutdown waits all three before disconnect

## Conventions

- Structured logging via `log/slog` throughout (pass `*slog.Logger` to constructors, add `"component"` field)
- Errors wrapped with context: `fmt.Errorf("verb noun: %w", err)`
- ZCL cluster definitions are declarative Go structs (see any file in `zcl/clusters/` for the pattern)
- All binary protocol encoding is little-endian
- Module path is `zigbee-go-home` (not a domain-based import path)
- IEEE addresses formatted as 16-char uppercase hex: `fmt.Sprintf("%016X", ieeeBytes)`
- Custom/proprietary clusters loaded from JSON files in `devices/` dir with merge semantics (new attributes added to existing clusters)
- Web server uses `ServerOption` functional options pattern (`WithAPIKey`, `WithDevicesDir`, `WithAutomation`, etc.)
- Test stubs: `stubNCP` in web tests, `memStore` in coordinator tests — both implement full interfaces with minimal stubs
- Script files: `.lua` with `-- {"name":...}` JSON on line 1, optional `--[[BLOCKLY_XML...BLOCKLY_XML]]--` block, then Lua code

## NCP Hardware

- **nRF52840:** Custom firmware from `../nrf-ncp/`, NCS v2.6.1, Adafruit UF2 bootloader. Flash bootloader via J-Link, then drag-and-drop .uf2 firmware updates.
