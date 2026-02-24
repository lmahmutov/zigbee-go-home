//go:build !no_automation

package automation

import (
	"context"
	"strings"
	"time"

	"zigbee-go-home/internal/store"

	lua "github.com/yuin/gopher-lua"
)

// registerZigbeeModule registers the `zigbee` global table in a Lua state.
func registerZigbeeModule(L *lua.LState, vm *scriptVM, e *Engine) {
	mod := L.NewTable()

	mod.RawSetString("on", L.NewFunction(func(L *lua.LState) int {
		return zigbeeOn(L, vm)
	}))

	mod.RawSetString("turn_on", L.NewFunction(func(L *lua.LState) int {
		return zigbeeSendOnOff(L, e, 1)
	}))

	mod.RawSetString("turn_off", L.NewFunction(func(L *lua.LState) int {
		return zigbeeSendOnOff(L, e, 0)
	}))

	mod.RawSetString("toggle", L.NewFunction(func(L *lua.LState) int {
		return zigbeeSendOnOff(L, e, 2)
	}))

	mod.RawSetString("set_brightness", L.NewFunction(func(L *lua.LState) int {
		return zigbeeSetBrightness(L, e)
	}))

	mod.RawSetString("set_color", L.NewFunction(func(L *lua.LState) int {
		return zigbeeSetColor(L, e)
	}))

	mod.RawSetString("send_command", L.NewFunction(func(L *lua.LState) int {
		return zigbeeSendCommand(L, e)
	}))

	mod.RawSetString("get_property", L.NewFunction(func(L *lua.LState) int {
		return zigbeeGetProperty(L, e)
	}))

	mod.RawSetString("after", L.NewFunction(func(L *lua.LState) int {
		return zigbeeAfter(L, vm, e)
	}))

	mod.RawSetString("log", L.NewFunction(func(L *lua.LState) int {
		return zigbeeLog(L, e)
	}))

	mod.RawSetString("devices", L.NewFunction(func(L *lua.LState) int {
		return zigbeeDevices(L, e)
	}))

	L.SetGlobal("zigbee", mod)
}

const maxHandlersPerScript = 100

// zigbee.on(type, filter, callback)
func zigbeeOn(L *lua.LState, vm *scriptVM) int {
	eventType := L.CheckString(1)
	filterTable := L.CheckTable(2)
	fn := L.CheckFunction(3)

	h := luaEventHandler{
		eventType: eventType,
		fn:        fn,
	}

	if v := filterTable.RawGetString("ieee"); v != lua.LNil {
		h.ieee = v.String()
	}
	if v := filterTable.RawGetString("property"); v != lua.LNil {
		h.property = v.String()
	}

	vm.mu.Lock()
	if len(vm.handlers) >= maxHandlersPerScript {
		vm.mu.Unlock()
		L.RaiseError("too many handlers (max %d)", maxHandlersPerScript)
		return 0
	}
	vm.handlers = append(vm.handlers, h)
	vm.mu.Unlock()

	return 0
}

// zigbee.turn_on/turn_off/toggle(ieee_or_name)
func zigbeeSendOnOff(L *lua.LState, e *Engine, cmdID uint8) int {
	target := L.CheckString(1)
	dev := resolveDevice(e, target)
	if dev == nil {
		e.logger.Warn("device not found", "target", target)
		return 0
	}

	// Find first endpoint with OnOff cluster (0x0006)
	ep := findEndpointWithCluster(dev, 0x0006)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.coord.SendClusterCommand(ctx, dev.ShortAddress, ep, 0x0006, cmdID, nil); err != nil {
		e.logger.Error("send on/off command", "err", err, "target", target, "cmd", cmdID)
	}
	return 0
}

// zigbee.set_brightness(ieee_or_name, level)
func zigbeeSetBrightness(L *lua.LState, e *Engine) int {
	target := L.CheckString(1)
	level := L.CheckInt(2)

	dev := resolveDevice(e, target)
	if dev == nil {
		e.logger.Warn("device not found", "target", target)
		return 0
	}

	// Clamp level to 0-254
	if level < 0 {
		level = 0
	}
	if level > 254 {
		level = 254
	}

	ep := findEndpointWithCluster(dev, 0x0008)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Move to Level with On/Off (cmd 0x04): level (1 byte) + transition time (2 bytes, 1/10s)
	payload := []byte{byte(level), 10, 0} // transition = 1s
	if err := e.coord.SendClusterCommand(ctx, dev.ShortAddress, ep, 0x0008, 0x04, payload); err != nil {
		e.logger.Error("set brightness", "err", err, "target", target, "level", level)
	}
	return 0
}

// zigbee.set_color(ieee_or_name, hue, saturation)
func zigbeeSetColor(L *lua.LState, e *Engine) int {
	target := L.CheckString(1)
	hue := L.CheckInt(2)
	sat := L.CheckInt(3)

	dev := resolveDevice(e, target)
	if dev == nil {
		e.logger.Warn("device not found", "target", target)
		return 0
	}

	// Clamp values to 0-254
	if hue < 0 {
		hue = 0
	}
	if hue > 254 {
		hue = 254
	}
	if sat < 0 {
		sat = 0
	}
	if sat > 254 {
		sat = 254
	}

	ep := findEndpointWithCluster(dev, 0x0300) // Color Control
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// MoveToHueAndSaturation (cmd 0x06): hue (1 byte) + saturation (1 byte) + transition time (2 bytes, 1/10s)
	payload := []byte{byte(hue), byte(sat), 10, 0} // transition = 1s
	if err := e.coord.SendClusterCommand(ctx, dev.ShortAddress, ep, 0x0300, 0x06, payload); err != nil {
		e.logger.Error("set color", "err", err, "target", target, "hue", hue, "sat", sat)
	}
	return 0
}

// zigbee.send_command(ieee, ep, cluster, cmd, payload)
func zigbeeSendCommand(L *lua.LState, e *Engine) int {
	target := L.CheckString(1)
	epVal := L.CheckInt(2)
	clusterVal := L.CheckInt(3)
	cmdVal := L.CheckInt(4)

	if epVal < 0 || epVal > 255 {
		L.ArgError(2, "endpoint must be 0-255")
		return 0
	}
	if clusterVal < 0 || clusterVal > 65535 {
		L.ArgError(3, "cluster must be 0-65535")
		return 0
	}
	if cmdVal < 0 || cmdVal > 255 {
		L.ArgError(4, "command must be 0-255")
		return 0
	}

	ep := uint8(epVal)
	cluster := uint16(clusterVal)
	cmd := uint8(cmdVal)

	var payload []byte
	if L.GetTop() >= 5 {
		if tbl, ok := L.Get(5).(*lua.LTable); ok {
			tbl.ForEach(func(_, v lua.LValue) {
				if n, ok := v.(lua.LNumber); ok {
					payload = append(payload, byte(n))
				}
			})
		}
	}

	dev := resolveDevice(e, target)
	if dev == nil {
		e.logger.Warn("device not found", "target", target)
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.coord.SendClusterCommand(ctx, dev.ShortAddress, ep, cluster, cmd, payload); err != nil {
		e.logger.Error("send command", "err", err, "target", target)
	}
	return 0
}

// zigbee.get_property(ieee_or_name, property)
func zigbeeGetProperty(L *lua.LState, e *Engine) int {
	target := L.CheckString(1)
	prop := L.CheckString(2)

	dev := resolveDevice(e, target)
	if dev == nil {
		L.Push(lua.LNil)
		return 1
	}

	if dev.Properties != nil {
		if v, ok := dev.Properties[prop]; ok {
			L.Push(goToLua(L, v))
			return 1
		}
	}

	L.Push(lua.LNil)
	return 1
}

// zigbee.after(seconds, callback) — delayed execution
func zigbeeAfter(L *lua.LState, vm *scriptVM, e *Engine) int {
	seconds := L.CheckNumber(1)
	fn := L.CheckFunction(2)

	go func() {
		timer := time.NewTimer(time.Duration(float64(seconds) * float64(time.Second)))
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-vm.ctx.Done():
			return
		}

		// Send callback execution to the VM's command channel
		select {
		case vm.commands <- func(L *lua.LState) {
			if err := L.CallByParam(lua.P{
				Fn:      fn,
				NRet:    0,
				Protect: true,
			}); err != nil {
				e.logger.Error("after callback error", "err", err)
			}
		}:
		default:
			e.logger.Warn("after: command channel full")
		}
	}()

	return 0
}

// zigbee.log(msg)
func zigbeeLog(L *lua.LState, e *Engine) int {
	msg := L.CheckString(1)
	e.logger.Info("script log", "msg", msg)
	return 0
}

// zigbee.devices() — returns a table of all devices
func zigbeeDevices(L *lua.LState, e *Engine) int {
	devices, err := e.coord.Devices().ListDevices()
	if err != nil {
		L.Push(L.NewTable())
		return 1
	}

	tbl := L.NewTable()
	for i, dev := range devices {
		d := L.NewTable()
		d.RawSetString("ieee", lua.LString(dev.IEEEAddress))
		name := dev.FriendlyName
		if name == "" {
			if dev.Manufacturer != "" {
				name = dev.Manufacturer
			}
			if dev.Model != "" {
				if name != "" {
					name += " "
				}
				name += dev.Model
			}
		}
		d.RawSetString("name", lua.LString(name))
		d.RawSetString("model", lua.LString(dev.Model))
		d.RawSetString("manufacturer", lua.LString(dev.Manufacturer))
		tbl.RawSetInt(i+1, d)
	}

	L.Push(tbl)
	return 1
}

// resolveDevice finds a device by IEEE address or friendly name.
func resolveDevice(e *Engine, target string) *store.Device {
	// Check if target looks like an IEEE address (16 hex chars)
	if len(target) == 16 && isHexString(target) {
		dev, err := e.coord.Devices().GetDevice(strings.ToUpper(target))
		if err == nil {
			return dev
		}
	}

	// Search by friendly name
	devices, err := e.coord.Devices().ListDevices()
	if err != nil {
		return nil
	}

	target = strings.ToLower(target)
	for _, dev := range devices {
		if strings.ToLower(dev.FriendlyName) == target {
			return dev
		}
	}

	// Also try IEEE match case-insensitively
	for _, dev := range devices {
		if strings.EqualFold(dev.IEEEAddress, target) {
			return dev
		}
	}

	return nil
}

func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// findEndpointWithCluster finds the first endpoint that has the given cluster as input.
func findEndpointWithCluster(dev *store.Device, clusterID uint16) uint8 {
	for _, ep := range dev.Endpoints {
		for _, cid := range ep.InClusters {
			if cid == clusterID {
				return ep.ID
			}
		}
	}
	// Default to endpoint 1
	if len(dev.Endpoints) > 0 {
		return dev.Endpoints[0].ID
	}
	return 1
}
