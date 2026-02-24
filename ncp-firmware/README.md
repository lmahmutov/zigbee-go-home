# NCP Firmware

Pre-built ZBOSS NCP firmware for nRF52840 (ProMicro form factor) with Adafruit UF2 bootloader.

- **NCS version:** 2.6.1
- **ZBOSS stack:** 3.11.3.0
- **Protocol version:** 277
- **Bootloader:** Adafruit nRF52 UF2 (S140 SoftDevice 7.3.0)

## Files

| File | Description |
|------|-------------|
| `bootloader_s140.hex` | UF2 bootloader + S140 SoftDevice (one-time flash via J-Link) |
| `ncp_firmware.uf2` | ZBOSS NCP application firmware (drag-and-drop update) |

## First-Time Flash (J-Link)

Flash the bootloader via J-Link / nrfjprog:

```bash
nrfjprog --program bootloader_s140.hex --chiperase --verify --reset
```

This erases the chip and writes the UF2 bootloader with SoftDevice. After reset, the board enters bootloader mode and exposes a `NRF52BOOT` USB mass storage drive.

## Firmware Update (UF2)

Once the bootloader is installed, updates are drag-and-drop:

1. Double-tap the reset button — board enters bootloader mode, `NRF52BOOT` drive appears
2. Copy the firmware file:
   ```bash
   cp ncp_firmware.uf2 /media/NRF52BOOT/
   ```
3. Board automatically resets and starts the NCP firmware

## USB Connection

The NCP firmware exposes a USB CDC ACM serial port (typically `/dev/ttyACM0` on Linux). Default baud rate: 460800.

```yaml
# config.yaml
ncp:
  type: nrf52840
  port: /dev/ttyACM0
  baud: 460800
```

## Building from Source

See the [nrf-ncp](https://github.com/lmahmutov/nrf-ncp) repository for build instructions using Nordic NCS SDK.
