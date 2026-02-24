//go:build !no_automation

package automation

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"zigbee-go-home/internal/coordinator"

	lua "github.com/yuin/gopher-lua"
)

// RunResult is the result of a one-shot script execution.
type RunResult struct {
	OK       bool     `json:"ok"`
	Error    string   `json:"error,omitempty"`
	Logs     []string `json:"logs"`
	Duration string   `json:"duration"`
}

// luaEventHandler is a registered Lua callback for a specific event pattern.
type luaEventHandler struct {
	eventType string
	ieee      string // filter: only match this IEEE (empty = any)
	property  string // filter: only match this property (empty = any)
	fn        *lua.LFunction
}

// scriptVM is a running Lua VM for a single script.
type scriptVM struct {
	state    *lua.LState
	commands chan func(*lua.LState) // serializes Lua access
	handlers []luaEventHandler
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex // protects handlers
}

// Engine manages Lua VMs and dispatches EventBus events to scripts.
type Engine struct {
	coord   *coordinator.Coordinator
	manager *Manager
	logger  *slog.Logger

	systemCfg   SystemConfig
	telegramCfg TelegramConfig

	mu   sync.Mutex
	vms  map[string]*scriptVM // script ID -> running VM
	unsub func()
}

// NewEngine creates a new automation engine.
func NewEngine(coord *coordinator.Coordinator, mgr *Manager, logger *slog.Logger, sysCfg SystemConfig, teleCfg TelegramConfig) *Engine {
	return &Engine{
		coord:       coord,
		manager:     mgr,
		logger:      logger.With("component", "automation"),
		systemCfg:   sysCfg,
		telegramCfg: teleCfg,
		vms:         make(map[string]*scriptVM),
	}
}

// Start subscribes to the EventBus and loads all enabled scripts.
func (e *Engine) Start() {
	e.unsub = e.coord.Events().OnAll(func(event coordinator.Event) {
		e.dispatchEvent(event)
	})

	scripts, err := e.manager.List()
	if err != nil {
		e.logger.Error("load scripts", "err", err)
		return
	}

	for _, s := range scripts {
		if !s.Meta.Enabled {
			continue
		}
		if err := e.startScript(s); err != nil {
			e.logger.Error("start script", "id", s.ID, "err", err)
		}
	}

	e.logger.Info("automation engine started", "scripts", len(e.vms))
}

// Stop cancels all VMs and unsubscribes from EventBus.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for id, vm := range e.vms {
		vm.cancel()
		delete(e.vms, id)
	}

	if e.unsub != nil {
		e.unsub()
	}

	e.logger.Info("automation engine stopped")
}

// ReloadScript stops the old VM (if any) and starts a new one.
func (e *Engine) ReloadScript(id string) error {
	e.stopScript(id)

	s, err := e.manager.Get(id)
	if err != nil {
		return fmt.Errorf("get script: %w", err)
	}

	if !s.Meta.Enabled {
		return nil // disabled, just stop
	}

	return e.startScript(s)
}

// StopScript stops a running script VM.
func (e *Engine) StopScript(id string) {
	e.stopScript(id)
}

// RunScript executes a script in a temporary sandboxed VM for testing.
// It runs the top-level code (which registers handlers via zigbee.on) and
// captures any log output. The VM is destroyed after a short timeout.
func (e *Engine) RunScript(id string) *RunResult {
	start := time.Now()

	s, err := e.manager.Get(id)
	if err != nil {
		return &RunResult{OK: false, Error: "script not found: " + err.Error(), Duration: time.Since(start).String()}
	}

	return e.RunLuaCode(s.LuaCode)
}

// RunLuaCode executes arbitrary Lua code in a temporary sandboxed VM for testing.
func (e *Engine) RunLuaCode(code string) *RunResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L := lua.NewState(lua.Options{SkipOpenLibs: false})
	defer L.Close()

	// Sandbox
	L.SetGlobal("os", lua.LNil)
	L.SetGlobal("io", lua.LNil)
	L.SetGlobal("loadfile", lua.LNil)
	L.SetGlobal("dofile", lua.LNil)
	L.SetGlobal("require", lua.LNil)
	L.SetGlobal("load", lua.LNil)
	L.SetGlobal("debug", lua.LNil)
	L.SetGlobal("package", lua.LNil)

	L.SetContext(ctx)

	vm := &scriptVM{
		state:    L,
		commands: make(chan func(*lua.LState), 64),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Capture logs
	var logs []string
	var logMu sync.Mutex

	// Register zigbee module with log capture
	registerZigbeeModule(L, vm, e)
	registerSystemModule(L, e)
	registerTelegramModule(L, e)

	// Override zigbee.log to capture output
	mod := L.GetGlobal("zigbee")
	if tbl, ok := mod.(*lua.LTable); ok {
		tbl.RawSetString("log", L.NewFunction(func(L *lua.LState) int {
			msg := L.CheckString(1)
			logMu.Lock()
			logs = append(logs, msg)
			logMu.Unlock()
			e.logger.Info("script run log", "msg", msg)
			return 0
		}))
	}

	// Also capture system.log
	sysMod := L.GetGlobal("system")
	if tbl, ok := sysMod.(*lua.LTable); ok {
		tbl.RawSetString("log", L.NewFunction(func(L *lua.LState) int {
			level := L.CheckString(1)
			msg := L.CheckString(2)
			logMu.Lock()
			logs = append(logs, "["+level+"] "+msg)
			logMu.Unlock()
			return 0
		}))
	}

	e.logger.Info("RunLuaCode: executing", "code_len", len(code))

	if err := L.DoString(code); err != nil {
		dur := time.Since(start)
		errStr := err.Error()
		if strings.Contains(errStr, "context deadline exceeded") {
			errStr = "timeout (5s)"
		}
		e.logger.Warn("RunLuaCode: script error", "err", errStr)
		return &RunResult{OK: false, Error: errStr, Logs: logs, Duration: dur.String()}
	}

	// If the script registered event handlers (typical for Blockly-generated code),
	// invoke each one with a synthetic event so the actions actually execute.
	vm.mu.Lock()
	handlers := make([]luaEventHandler, len(vm.handlers))
	copy(handlers, vm.handlers)
	vm.mu.Unlock()

	e.logger.Info("RunLuaCode: invoking handlers", "count", len(handlers))

	for i, h := range handlers {
		eventTable := L.NewTable()
		eventTable.RawSetString("type", lua.LString(h.eventType))
		if h.ieee != "" {
			eventTable.RawSetString("ieee", lua.LString(h.ieee))
		}
		if h.property != "" {
			eventTable.RawSetString("property", lua.LString(h.property))
		}
		// Set a default value=true so "if event.value == true" conditions pass
		eventTable.RawSetString("value", lua.LBool(true))

		e.logger.Info("RunLuaCode: calling handler", "index", i, "event_type", h.eventType, "ieee", h.ieee, "property", h.property)

		if err := L.CallByParam(lua.P{
			Fn:      h.fn,
			NRet:    0,
			Protect: true,
		}, eventTable); err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "context deadline exceeded") {
				errStr = "timeout (5s)"
			}
			e.logger.Warn("RunLuaCode: handler error", "index", i, "err", errStr)
			dur := time.Since(start)
			return &RunResult{OK: false, Error: errStr, Logs: logs, Duration: dur.String()}
		}
	}

	dur := time.Since(start)
	e.logger.Info("RunLuaCode: complete", "ok", true, "logs", len(logs), "duration", dur)
	return &RunResult{OK: true, Logs: logs, Duration: dur.String()}
}

func (e *Engine) stopScript(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if vm, ok := e.vms[id]; ok {
		vm.cancel()
		delete(e.vms, id)
		e.logger.Info("script stopped", "id", id)
	}
}

func (e *Engine) startScript(s *Script) error {
	ctx, cancel := context.WithCancel(context.Background())

	L := lua.NewState(lua.Options{
		SkipOpenLibs: false,
	})

	// Sandbox: remove dangerous libs and functions
	L.SetGlobal("os", lua.LNil)
	L.SetGlobal("io", lua.LNil)
	L.SetGlobal("loadfile", lua.LNil)
	L.SetGlobal("dofile", lua.LNil)
	L.SetGlobal("require", lua.LNil)
	L.SetGlobal("load", lua.LNil)
	L.SetGlobal("debug", lua.LNil)
	L.SetGlobal("package", lua.LNil)

	vm := &scriptVM{
		state:    L,
		commands: make(chan func(*lua.LState), 64),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Register modules
	registerZigbeeModule(L, vm, e)
	registerSystemModule(L, e)
	registerTelegramModule(L, e)

	// Execute the script to register handlers
	if err := L.DoString(s.LuaCode); err != nil {
		cancel()
		L.Close()
		return fmt.Errorf("execute script %s: %w", s.ID, err)
	}

	e.mu.Lock()
	e.vms[s.ID] = vm
	e.mu.Unlock()

	// Start command loop goroutine — exits when context is cancelled.
	go func() {
		defer L.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case fn := <-vm.commands:
				fn(L)
			}
		}
	}()

	e.logger.Info("script started", "id", s.ID, "name", s.Meta.Name)
	return nil
}

// dispatchEvent routes an EventBus event to all matching Lua handlers.
func (e *Engine) dispatchEvent(event coordinator.Event) {
	e.mu.Lock()
	vmsCopy := make(map[string]*scriptVM, len(e.vms))
	for k, v := range e.vms {
		vmsCopy[k] = v
	}
	e.mu.Unlock()

	for _, vm := range vmsCopy {
		vm.mu.Lock()
		handlers := make([]luaEventHandler, len(vm.handlers))
		copy(handlers, vm.handlers)
		vm.mu.Unlock()

		for _, h := range handlers {
			if !matchesHandler(h, event) {
				continue
			}

			fn := h.fn
			// Send to VM's command channel for thread-safe Lua execution.
			// Check context first to avoid sending to a stopped VM.
			select {
			case <-vm.ctx.Done():
				// VM stopped, skip remaining handlers
				break
			case vm.commands <- func(L *lua.LState) {
				e.callHandler(L, fn, event)
			}:
			default:
				e.logger.Warn("script command channel full, dropping event")
			}
		}
	}
}

func matchesHandler(h luaEventHandler, event coordinator.Event) bool {
	if h.eventType != event.Type {
		return false
	}

	data, ok := event.Data.(map[string]interface{})
	if !ok {
		return h.ieee == "" && h.property == ""
	}

	if h.ieee != "" {
		if ieee, _ := data["ieee"].(string); ieee != h.ieee {
			return false
		}
	}

	if h.property != "" {
		if prop, _ := data["property"].(string); prop != h.property {
			return false
		}
	}

	return true
}

func (e *Engine) callHandler(L *lua.LState, fn *lua.LFunction, event coordinator.Event) {
	defer func() {
		if r := recover(); r != nil {
			e.logger.Error("lua handler panic", "err", r)
		}
	}()

	// Build event table
	eventTable := L.NewTable()
	eventTable.RawSetString("type", lua.LString(event.Type))

	if data, ok := event.Data.(map[string]interface{}); ok {
		for k, v := range data {
			eventTable.RawSetString(k, goToLua(L, v))
		}
	}

	if err := L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true,
	}, eventTable); err != nil {
		e.logger.Error("lua handler error", "err", err)
	}
}

// goToLua converts a Go value to a Lua value.
func goToLua(L *lua.LState, v interface{}) lua.LValue {
	switch val := v.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(val)
	case string:
		return lua.LString(val)
	case int:
		return lua.LNumber(val)
	case int64:
		return lua.LNumber(val)
	case float64:
		return lua.LNumber(val)
	case uint8:
		return lua.LNumber(val)
	case uint16:
		return lua.LNumber(val)
	case uint32:
		return lua.LNumber(val)
	case uint64:
		return lua.LNumber(val)
	case int8:
		return lua.LNumber(val)
	case int16:
		return lua.LNumber(val)
	case int32:
		return lua.LNumber(val)
	case float32:
		return lua.LNumber(val)
	case map[string]interface{}:
		t := L.NewTable()
		for k, vv := range val {
			t.RawSetString(k, goToLua(L, vv))
		}
		return t
	case []interface{}:
		t := L.NewTable()
		for i, vv := range val {
			t.RawSetInt(i+1, goToLua(L, vv))
		}
		return t
	default:
		return lua.LString(fmt.Sprintf("%v", val))
	}
}
