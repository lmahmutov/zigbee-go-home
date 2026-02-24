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

const commandChanBuffer = 64

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

	mu    sync.Mutex
	vms   map[string]*scriptVM // script ID -> running VM
	vmWg  sync.WaitGroup       // tracks command loop goroutines
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
	if e.unsub != nil {
		e.unsub()
	}
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

// Stop cancels all VMs, waits for goroutines to exit, and unsubscribes from EventBus.
func (e *Engine) Stop() {
	if e.unsub != nil {
		e.unsub()
		e.unsub = nil
	}

	e.mu.Lock()
	for id, vm := range e.vms {
		vm.cancel()
		delete(e.vms, id)
	}
	e.mu.Unlock()

	e.vmWg.Wait()

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

	L := newSandboxedLuaState()
	// Cancel context BEFORE closing LState so zigbee.after() goroutines exit
	// before the Lua state is freed. Defers execute LIFO.
	defer L.Close()
	defer cancel()

	L.SetContext(ctx)

	vm := &scriptVM{
		state:    L,
		commands: make(chan func(*lua.LState), commandChanBuffer),
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

	L := newSandboxedLuaState()

	vm := &scriptVM{
		state:    L,
		commands: make(chan func(*lua.LState), commandChanBuffer),
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
	e.vmWg.Add(1)
	go func() {
		defer e.vmWg.Done()
		defer L.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case fn := <-vm.commands:
				func() {
					defer func() {
						if r := recover(); r != nil {
							e.logger.Error("command loop panic", "script", s.ID, "err", r)
						}
					}()
					fn(L)
				}()
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

	nextVM:
		for _, h := range handlers {
			if !matchesHandler(h, event) {
				continue
			}

			fn := h.fn
			// Send to VM's command channel for thread-safe Lua execution.
			// Check context first to avoid sending to a stopped VM.
			select {
			case <-vm.ctx.Done():
				// VM stopped, skip remaining handlers for this VM.
				break nextVM
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

// newSandboxedLuaState creates a Lua VM with only safe libraries loaded.
// Uses whitelist approach: only base, table, string, math are available.
// Dangerous functions (loadfile, dofile, load) are removed from base.
func newSandboxedLuaState() *lua.LState {
	L := lua.NewState(lua.Options{
		SkipOpenLibs:    true,
		CallStackSize:   120,
		RegistrySize:    1024 * 20,
		RegistryMaxSize: 1024 * 80,
	})
	// Load only safe standard libraries.
	for _, lib := range []struct {
		name string
		fn   lua.LGFunction
	}{
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
	} {
		L.Push(L.NewFunction(lib.fn))
		L.Push(lua.LString(lib.name))
		L.Call(1, 0)
	}
	// Remove remaining dangerous base functions.
	L.SetGlobal("loadfile", lua.LNil)
	L.SetGlobal("dofile", lua.LNil)
	L.SetGlobal("load", lua.LNil)

	sandboxStringRep(L)
	return L
}

const maxStringRepLen = 1 << 20 // 1 MB limit for string.rep result

// sandboxStringRep overrides string.rep with a length-limited wrapper.
func sandboxStringRep(L *lua.LState) {
	strMod := L.GetField(L.GetField(L.Get(lua.EnvironIndex), "string"), "rep")
	if strMod == lua.LNil {
		return
	}
	origRep, ok := strMod.(*lua.LFunction)
	if !ok {
		return
	}
	L.SetField(L.GetField(L.Get(lua.EnvironIndex), "string"), "rep", L.NewFunction(func(L *lua.LState) int {
		s := L.CheckString(1)
		n := L.CheckInt(2)
		if int64(len(s))*int64(n) > maxStringRepLen {
			L.ArgError(2, "resulting string exceeds 1MB limit")
			return 0
		}
		// Call the original string.rep.
		if err := L.CallByParam(lua.P{Fn: origRep, NRet: 1, Protect: true}, lua.LString(s), lua.LNumber(n)); err != nil {
			L.RaiseError("%s", err.Error())
			return 0
		}
		ret := L.Get(-1)
		L.Pop(1)
		L.Push(ret)
		return 1
	}))
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
